import { Injectable } from '@nestjs/common';
import type { Prisma, StepRun } from '@prisma/client';
import { WorkflowRunRepository, StepRunRepository } from '@lunie/db-access';
import { ErrorDefinitions } from 'src/common/http/errors/error-codes';
import { AppError } from 'src/common/http/errors/app-error';
import { buildPaginationMeta } from 'src/common/pagination/pagination';

@Injectable()
export class StepRunService {
  constructor(
    private readonly repo: StepRunRepository,
    private readonly runRepo: WorkflowRunRepository,
  ) {}

  private async assertWorkflowRunExists(
    workflowRunId: string,
    workflowId?: string,
  ) {
    const run = await this.runRepo.findById(workflowRunId);
    if (!run) throw AppError.notFound(ErrorDefinitions.WORKFLOW_RUN.NOT_FOUND);
    if (workflowId && run.workflowId !== workflowId) {
      throw AppError.notFound(ErrorDefinitions.WORKFLOW_RUN.NOT_FOUND);
    }
    return run;
  }

  async create(params: {
    workflowRunId: string;
    stepKey: string;
    status?: StepRun['status'];
    attempt?: number;
    input?: Prisma.InputJsonValue;
    output?: Prisma.InputJsonValue;
    error?: Prisma.InputJsonValue;
    logs?: Prisma.InputJsonValue;
    lastErrorAt?: Date;
    durationMs?: number;
    startedAt?: Date;
    finishedAt?: Date;
  }): Promise<StepRun> {
    await this.assertWorkflowRunExists(params.workflowRunId);

    return this.repo.create({
      workflowRun: { connect: { id: params.workflowRunId } },
      stepKey: params.stepKey,
      status: params.status ?? 'QUEUED',
      attempt: params.attempt ?? 0,
      input: params.input ?? {},
      output: params.output,
      error: params.error,
      logs: params.logs,
      lastErrorAt: params.lastErrorAt,
      durationMs: params.durationMs,
      startedAt: params.startedAt,
      finishedAt: params.finishedAt,
    });
  }

  async list(params: {
    workflowId: string;
    workflowRunId: string;
    page: number;
    pageSize: number;
    sortBy: 'createdAt' | 'updatedAt';
    sortOrder: 'asc' | 'desc';
  }): Promise<{
    items: StepRun[];
    pagination: ReturnType<typeof buildPaginationMeta>;
  }> {
    await this.assertWorkflowRunExists(params.workflowRunId, params.workflowId);
    const { items, total } = await this.repo.findPageByWorkflowRun(params);
    return {
      items,
      pagination: buildPaginationMeta({
        page: params.page,
        pageSize: params.pageSize,
        total,
        sortBy: params.sortBy,
        sortOrder: params.sortOrder,
      }),
    };
  }

  async get(
    workflowId: string,
    workflowRunId: string,
    id: string,
  ): Promise<StepRun> {
    await this.assertWorkflowRunExists(workflowRunId, workflowId);
    const step = await this.repo.findById(id);

    if (!step || step.workflowRunId !== workflowRunId)
      throw AppError.notFound(ErrorDefinitions.STEP_RUN.NOT_FOUND);

    return step;
  }

  async getByStepKey(
    workflowId: string,
    workflowRunId: string,
    stepKey: string,
  ): Promise<StepRun> {
    await this.assertWorkflowRunExists(workflowRunId, workflowId);
    const step = await this.repo.findByWorkflowRunAndStepKey(
      workflowRunId,
      stepKey,
    );

    if (!step) {
      throw AppError.notFound(ErrorDefinitions.STEP_RUN.NOT_FOUND);
    }

    return step;
  }

  async update(
    workflowRunId: string,
    id: string,
    patch: {
      status?: StepRun['status'];
      attempt?: number;
      input?: Prisma.InputJsonValue;
      output?: Prisma.InputJsonValue;
      error?: Prisma.InputJsonValue;
      logs?: Prisma.InputJsonValue;
      lastErrorAt?: Date;
      durationMs?: number;
      startedAt?: Date;
      finishedAt?: Date;
    },
  ): Promise<StepRun> {
    await this.assertWorkflowRunExists(workflowRunId);
    const step = await this.repo.findById(id);
    if (!step || step.workflowRunId !== workflowRunId) {
      throw AppError.notFound(ErrorDefinitions.STEP_RUN.NOT_FOUND);
    }
    return this.repo.update(id, patch);
  }
}
