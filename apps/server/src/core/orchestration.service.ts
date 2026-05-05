import { Injectable, Logger } from '@nestjs/common';
import { Prisma } from '@prisma/client';
import { PrismaService } from '@lunie/db-access';
import { StepRunJobPayload, WorkflowDefinition } from '@lunie/contracts';
import type { JobsOptions } from 'bullmq';

import { StepRunQueueService } from '../queue/step-run-queue.service';
import {
  buildUniqueKey,
  deriveReservedKey,
} from 'src/common/identifiers/key.util';

interface StepDef {
  key: string;
  type: string;
  dependsOn?: string[];
  request?: Record<string, unknown>;
}

type EnqueueItem = {
  stepType: string;
  stepKey: string;
  stepRunId: string;
  dependsOn: string[];
  payload: StepRunJobPayload;
};

@Injectable()
export class OrchestrationService {
  private readonly logger = new Logger(OrchestrationService.name);

  constructor(
    private readonly prisma: PrismaService,
    private readonly stepRunQueueService: StepRunQueueService,
  ) {}

  async startWorkflow(params: {
    workflowId: string;
    workflowVersionId: string;
    triggerId?: string;
    eventType?: string;
    eventExternalId?: string;
    eventPayload?: Record<string, unknown>;
    input?: Record<string, unknown>;
    overrides?: Record<
      string,
      {
        query?: Record<string, string | number | boolean>;
        body?: unknown;
      }
    >;
  }): Promise<{ workflowRunId: string; stepRunIds: string[] }> {
    const {
      workflowId,
      workflowVersionId,
      triggerId,
      eventType,
      eventExternalId,
      eventPayload,
      input,
      overrides,
    } = params;

    this.logger.log(
      `Starting workflow workflowId=${workflowId} workflowVersionId=${workflowVersionId} triggerId=${triggerId ?? '-'} eventType=${eventType ?? '-'}`,
    );

    const { workflowRunId, stepRunIds, enqueueItems } =
      await this.prisma.$transaction(async (tx) => {
        const triggerIdToUse = await this.resolveTriggerIdForEvent(tx, {
          workflowId,
          triggerId,
        });

        const event = await tx.event.create({
          data: {
            triggerId: triggerIdToUse,
            type: eventType ?? 'MANUAL',
            externalId: eventExternalId,
            payload: (eventPayload ??
              input ??
              {}) as unknown as Prisma.InputJsonValue,
            receivedAt: new Date(),
          },
        });

        const workflowVersion = await tx.workflowVersion.findUniqueOrThrow({
          where: { id: workflowVersionId },
        });

        const definition =
          workflowVersion.definition as unknown as WorkflowDefinition;
        const steps = (definition.steps ?? []) as StepDef[];

        this.applyInferredTemplateDependencies(steps);

        const workflowInput = definition.input ?? {};
        const triggerPayload = input ?? {};

        const mergedInput = {
          ...workflowInput,
          ...triggerPayload,
        };

        const overridesToPersist = this.pickAndNormalizeOverridesForSteps(
          steps,
          overrides,
        );

        const workflowSequence = await tx.workflow.update({
          where: { id: workflowId },
          data: { runSequence: { increment: 1 } },
          select: { runSequence: true },
        });

        const workflowRunData: Prisma.WorkflowRunUncheckedCreateInput = {
          workflowId,
          workflowVersionId,
          number: workflowSequence.runSequence,
          triggerId: triggerIdToUse,
          eventId: event.id,
          status: 'QUEUED',
          input: mergedInput as Prisma.InputJsonValue,
          overrides: overridesToPersist
            ? (overridesToPersist as unknown as Prisma.InputJsonValue)
            : Prisma.DbNull,
        };

        const workflowRun = await tx.workflowRun.create({
          data: workflowRunData,
        });

        if (steps.length === 0) {
          await tx.workflowRun.update({
            where: { id: workflowRun.id },
            data: { status: 'SUCCEEDED', finishedAt: new Date() },
          });

          return {
            workflowRunId: workflowRun.id,
            stepRunIds: [] as string[],
            enqueueItems: [] as EnqueueItem[],
          };
        }

        const stepRunMap = new Map<string, string>();
        for (const step of steps) {
          const stepRun = await tx.stepRun.create({
            data: {
              workflowRunId: workflowRun.id,
              stepKey: step.key,
              status: 'QUEUED',
              attempt: 0,
              input: mergedInput as Prisma.InputJsonValue,
              requestOverride: overridesToPersist?.[step.key]
                ? (overridesToPersist[
                    step.key
                  ] as unknown as Prisma.InputJsonValue)
                : Prisma.DbNull,
            },
            select: { id: true, stepKey: true },
          });
          stepRunMap.set(step.key, stepRun.id);
        }

        const stepRunIds = Array.from(stepRunMap.values());

        const enqueueItems: EnqueueItem[] = steps.map((step) => {
          const stepRunId = stepRunMap.get(step.key)!;
          const stepDeps = step.dependsOn || [];

          return {
            stepType: step.type,
            stepKey: step.key,
            stepRunId,
            dependsOn: stepDeps,
            payload: {
              workflowRunId: workflowRun.id,
              stepRunId,
              stepKey: step.key,
              workflowVersionId,
              input: mergedInput,
              dependsOn: stepDeps,
              requestOverride: overridesToPersist?.[step.key],
            },
          };
        });

        return { workflowRunId: workflowRun.id, stepRunIds, enqueueItems };
      });

    const jobIdMap = new Map<string, string>();
    const enqueuedStepRunIds: string[] = [];

    try {
      const batches = this.getExecutionBatches(
        enqueueItems.map((item) => ({
          key: item.stepKey,
          type: item.stepType,
          dependsOn: item.dependsOn,
        })),
      );

      for (const batch of batches) {
        const batchJobs: Promise<string>[] = [];

        for (const stepKey of batch) {
          const item = enqueueItems.find((i) => i.stepKey === stepKey)!;
          const dependsOnJobIds = item.dependsOn
            .map((depKey) => jobIdMap.get(depKey))
            .filter((id): id is string => id !== undefined);

          const options: JobsOptions | undefined =
            dependsOnJobIds.length > 0
              ? ({ dependsOn: dependsOnJobIds } as JobsOptions)
              : undefined;

          const jobPromise = this.stepRunQueueService
            .enqueueStepRun(item.stepType, item.payload, options)
            .then((job) => {
              jobIdMap.set(stepKey, job.id!);
              enqueuedStepRunIds.push(item.stepRunId);
              return job.id!;
            });

          batchJobs.push(jobPromise);
        }

        await Promise.all(batchJobs);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);

      this.logger.error(
        `Failed to enqueue step runs for workflowRunId=${workflowRunId}: ${message} (enqueued ${enqueuedStepRunIds.length}/${stepRunIds.length})`,
      );

      const notEnqueued = stepRunIds.filter(
        (id) => !enqueuedStepRunIds.includes(id),
      );

      await this.prisma.$transaction(async (tx) => {
        if (notEnqueued.length > 0) {
          await tx.stepRun.updateMany({
            where: { id: { in: notEnqueued }, status: 'QUEUED' },
            data: {
              status: 'FAILED',
              finishedAt: new Date(),
              error: {
                message: 'Failed to enqueue step run',
                details: message,
              } as Prisma.InputJsonValue,
            },
          });
        }

        await tx.workflowRun.update({
          where: { id: workflowRunId },
          data: {
            status: 'FAILED',
            finishedAt: new Date(),
          },
        });
      });

      throw err;
    }

    this.logger.log(
      `Workflow started workflowRunId=${workflowRunId} steps=${stepRunIds.length}`,
    );

    return { workflowRunId, stepRunIds };
  }

  private getExecutionBatches(steps: StepDef[]): string[][] {
    const dependencies = new Map<string, string[]>();
    const inDegree = new Map<string, number>();
    const graph = new Map<string, string[]>();
    const allStepKeys = new Set(steps.map((s) => s.key));

    for (const step of steps) {
      dependencies.set(step.key, [...(step.dependsOn || [])]);
      inDegree.set(step.key, 0);
      graph.set(step.key, []);
    }

    for (const [stepKey, deps] of dependencies) {
      for (const dep of deps) {
        if (allStepKeys.has(dep)) {
          const children = graph.get(dep) || [];
          children.push(stepKey);
          graph.set(dep, children);
          inDegree.set(stepKey, (inDegree.get(stepKey) || 0) + 1);
        }
      }
    }

    const queue: string[] = [];
    for (const [key, degree] of inDegree) {
      if (degree === 0) queue.push(key);
    }

    const batches: string[][] = [];

    while (queue.length > 0) {
      const batch = [...queue];
      batches.push(batch);
      queue.length = 0;

      for (const stepKey of batch) {
        const children = graph.get(stepKey) || [];
        for (const child of children) {
          const newDegree = (inDegree.get(child) || 0) - 1;
          inDegree.set(child, newDegree);
          if (newDegree === 0) queue.push(child);
        }
      }
    }

    return batches;
  }

  private applyInferredTemplateDependencies(steps: StepDef[]): void {
    const allStepKeys = new Set(steps.map((s) => s.key));
    const templatePattern =
      /\{\{\s*steps\.([a-zA-Z0-9_-]+)(?:\.[^}]*)?\s*\}\}/g;

    for (const step of steps) {
      if (!step.request) continue;

      const requestStr = JSON.stringify(step.request);
      const inferred: string[] = [];
      let match: RegExpExecArray | null;

      while ((match = templatePattern.exec(requestStr)) !== null) {
        const referencedStep = match[1];
        if (!referencedStep) continue;
        if (referencedStep === step.key) continue;
        if (!allStepKeys.has(referencedStep)) continue;
        if (!inferred.includes(referencedStep)) inferred.push(referencedStep);
      }

      if (inferred.length === 0) continue;

      const explicit = step.dependsOn ?? [];
      const merged = Array.from(new Set([...explicit, ...inferred]));
      step.dependsOn = merged;
    }
  }

  private async resolveTriggerIdForEvent(
    tx: Prisma.TransactionClient,
    params: {
      workflowId: string;
      triggerId?: string;
    },
  ): Promise<string> {
    if (params.triggerId) return params.triggerId;

    const manual = await tx.trigger.findFirst({
      where: { workflowId: params.workflowId, type: 'MANUAL' },
      select: { id: true },
    });

    if (manual) return manual.id;

    const existingTriggers = await tx.trigger.findMany({
      where: { workflowId: params.workflowId },
      select: { key: true, name: true, type: true },
    });
    const manualTriggerKey = buildUniqueKey(
      'Manual',
      'manual',
      existingTriggers.map((trigger) =>
        deriveReservedKey({
          key: trigger.key,
          source: trigger.name ?? trigger.type,
          fallback: trigger.type.toLowerCase(),
        }),
      ),
    );

    const created = await tx.trigger.create({
      data: {
        workflowId: params.workflowId,
        key: manualTriggerKey,
        type: 'MANUAL',
        name: 'Manual',
        isActive: true,
        config: {},
      },
      select: { id: true },
    });

    return created.id;
  }

  private pickAndNormalizeOverridesForSteps(
    steps: Array<{ key: string }>,
    overrides:
      | Record<
          string,
          { query?: Record<string, string | number | boolean>; body?: unknown }
        >
      | undefined,
  ): Record<
    string,
    { query?: Record<string, string | number | boolean>; body?: unknown }
  > | null {
    if (!overrides) return null;

    const picked: Record<
      string,
      { query?: Record<string, string | number | boolean>; body?: unknown }
    > = {};

    for (const step of steps) {
      const raw = overrides[step.key];
      if (!raw) continue;

      const normalized: {
        query?: Record<string, string | number | boolean>;
        body?: unknown;
      } = {};

      if (raw.query && Object.keys(raw.query).length > 0) {
        normalized.query = raw.query;
      }

      if (raw.body !== undefined) {
        normalized.body = raw.body;
      }

      if (normalized.query !== undefined || normalized.body !== undefined) {
        picked[step.key] = normalized;
      }
    }

    return Object.keys(picked).length > 0 ? picked : null;
  }
}
