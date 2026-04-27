import { Processor, WorkerHost } from '@nestjs/bullmq';
import { Logger } from '@nestjs/common';
import { Job } from 'bullmq';
import {
  StepRunRepository,
  WorkflowRunRepository,
  WorkflowVersionRepository,
  SecretRepository,
} from '@lunie/db-access';
import {
  StepRunJobPayload,
  type NotificationEvent,
  type WorkflowNotification,
} from '@lunie/contracts';
import { ExecutorRegistry } from '../executors/executor-registry';
import { TemplateResolver } from '../utils/template-resolver';
import { wrapForDb } from '../utils/persisted-json';
import { redactSecrets } from '../utils/redact';
import { CryptoService } from '../crypto/crypto.service';
import { Inject } from '@nestjs/common';
import type Redis from 'ioredis';
import { REDIS_CLIENT } from '../redis/redis.constants';
import { checkFixedWindowRateLimit } from '../utils/rate-limit';
import { WorkerCacheService } from '../cache/worker-cache.service';
import { RunNotifierService } from '../notifications/run-notifier.service';

interface StepDefinition {
  key: string;
  type: string;
  dependsOn?: string[];
  input?: Record<string, unknown>;
  request?: Record<string, unknown>;
  outputPolicy?: {
    truncate?: boolean;
    maxBytes?: number;
  };
  rateLimit?: {
    key: string;
    max: number;
    perSeconds: number;
  };
}

@Processor('step-runs')
export class StepRunProcessor extends WorkerHost {
  private readonly logger = new Logger(StepRunProcessor.name);
  private readonly templateResolver = new TemplateResolver();

  constructor(
    private readonly stepRunRepository: StepRunRepository,
    private readonly workflowRunRepository: WorkflowRunRepository,
    private readonly workflowVersionRepository: WorkflowVersionRepository,
    private readonly secretRepository: SecretRepository,
    private readonly crypto: CryptoService,
    @Inject(REDIS_CLIENT) private readonly redis: Redis,
    private readonly executorRegistry: ExecutorRegistry,
    private readonly cache: WorkerCacheService,
    private readonly runNotifier: RunNotifierService,
  ) {
    super();
  }

  async process(job: Job<StepRunJobPayload>): Promise<void> {
    const {
      workflowRunId,
      stepRunId,
      stepKey,
      workflowVersionId,
      dependsOn,
      input: triggerInput,
    } = job.data;

    try {
      await this.workflowRunRepository.markRunningIfQueued(workflowRunId);

      await this.stepRunRepository.update(stepRunId, {
        status: 'RUNNING',
        startedAt: new Date(),
      });

      const workflowVersion = await this.cache.getWorkflowVersion(
        workflowVersionId,
        () => this.workflowVersionRepository.findById(workflowVersionId),
      );

      if (!workflowVersion) {
        throw new Error(`WorkflowVersion not found: ${workflowVersionId}`);
      }

      const definition = workflowVersion.definition as unknown as {
        input?: Record<string, unknown>;
        steps: StepDefinition[];
      };
      const stepDef = definition.steps.find((s) => s.key === stepKey);

      if (!stepDef) {
        throw new Error(`Step definition not found for key: ${stepKey}`);
      }

      const mergedInput = (triggerInput as Record<string, unknown>) ?? {};
      const stepLevelInput = stepDef.input ?? {};

      const stepOutputs = await this.fetchStepOutputs(workflowRunId, dependsOn || []);

      const secretNames = this.findSecretRefs(stepDef.request ?? {});
      const secretMap: Record<string, string> = {};
      const missingSecrets: string[] = [];

      for (const name of secretNames) {
        const cached = this.cache.getSecret(name);
        if (cached !== undefined) {
          secretMap[name] = cached;
        } else {
          missingSecrets.push(name);
        }
      }

      if (missingSecrets.length > 0) {
        const secrets = await this.secretRepository.findManyByNames(missingSecrets);
        for (const s of secrets) {
          const value = this.crypto.decryptSecret(s.value);
          secretMap[s.name] = value;
          this.cache.setSecret(s.name, value);
        }
      }
      const secretValues = Object.values(secretMap);

      // Build context: merged input + step-level input + previous step outputs
      const context = {
        input: {
          ...mergedInput,
          ...stepLevelInput,
        },
        steps: stepOutputs,
        secret: secretMap,
      };

      const requestDef = stepDef.request ?? {};
      const { resolved: resolvedRequest, referencedSteps } = this.templateResolver.resolve(
        requestDef,
        context,
      );

      const request = this.applyHttpOverrides(
        stepDef.type,
        resolvedRequest,
        job.data.requestOverride,
      );

      const executor = this.executorRegistry.get(stepDef.type);

      if (stepDef.type === 'http') {
        await this.maybeRateLimitHttp(workflowRunId, stepRunId, stepDef);
      }

      const rawOutput = await executor.execute({
        request,
        input: context,
      });

      const outputUnwrapped = unwrapExecutorOutput(rawOutput);

      const stepRun = await this.stepRunRepository.findById(stepRunId);
      if (!stepRun) {
        throw new Error(`StepRun not found: ${stepRunId}`);
      }

      const durationMs = stepRun.startedAt ? Date.now() - stepRun.startedAt.getTime() : 0;

      const outputPolicy = stepDef.outputPolicy ?? {};
      const redactedOutput = redactSecrets(outputUnwrapped, { secretValues });
      const outputEnvelope = wrapForDb(redactedOutput, {
        maxBytes: outputPolicy.maxBytes ?? 256 * 1024,
        truncate: outputPolicy.truncate ?? true,
        reason: 'step_output',
      });

      const redactedInput = redactSecrets(context.input, { secretValues });
      const inputEnvelope = wrapForDb(redactedInput, {
        maxBytes: 256 * 1024,
        truncate: true,
        reason: 'step_input',
      });

      await this.stepRunRepository.update(stepRunId, {
        status: 'SUCCEEDED',
        finishedAt: new Date(),
        output: outputEnvelope as unknown as object,
        input: inputEnvelope as unknown as object,
        durationMs,
      });

      this.logger.log(`Step ${stepKey} completed successfully (${durationMs}ms)`);

      await this.checkWorkflowCompletion(
        workflowRunId,
        workflowVersionId,
      );
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'unknown error';
      this.logger.error(`Step ${stepKey} failed: ${errorMessage}`);

      // Best-effort: if we already loaded secrets earlier in the run, they are not in scope here.
      // We still redact common sensitive keys.

      const redactedError = redactSecrets(
        {
          message: errorMessage,
          stack: error instanceof Error ? error.stack : undefined,
        },
        { secretValues: [] },
      );
      const errorEnvelope = wrapForDb(redactedError, {
        maxBytes: 64 * 1024,
        truncate: true,
        reason: 'step_error',
      });

      await this.stepRunRepository.update(stepRunId, {
        status: 'FAILED',
        finishedAt: new Date(),
        lastErrorAt: new Date(),
        error: errorEnvelope as unknown as object,
      });

      await this.checkWorkflowCompletion(
        workflowRunId,
        workflowVersionId,
      );

      throw error;
    }
  }

  private async maybeRateLimitHttp(
    workflowRunId: string,
    stepRunId: string,
    stepDef: StepDefinition,
  ): Promise<void> {
    const workflowRun = await this.cache.getWorkflowRun(workflowRunId, () =>
      this.workflowRunRepository.findById(workflowRunId),
    );
    if (!workflowRun) return;

    const enabled = process.env.WORKER_DEFAULT_HTTP_RATE_LIMIT_ENABLED !== 'false';
    const parsePositiveInt = (value: string | number | undefined, fallback: number) => {
      const parsed = Number(value);
      return Number.isFinite(parsed) && parsed > 0 ? Math.floor(parsed) : fallback;
    };
    const defaultMax = parsePositiveInt(process.env.WORKER_DEFAULT_HTTP_RATE_LIMIT_MAX, 300);
    const defaultPerSeconds = parsePositiveInt(
      process.env.WORKER_DEFAULT_HTTP_RATE_LIMIT_PER_SECONDS,
      60,
    );

    const limiter = stepDef.rateLimit
      ? {
          key: stepDef.rateLimit.key,
          max: parsePositiveInt(stepDef.rateLimit.max, defaultMax),
          perSeconds: parsePositiveInt(stepDef.rateLimit.perSeconds, defaultPerSeconds),
        }
      : enabled
        ? { key: '__default_http__', max: defaultMax, perSeconds: defaultPerSeconds }
        : null;

    if (!limiter) return;

    const redisKey = `ratelimit:wf:${workflowRun.workflowId}:${limiter.key}`;
    let allowed = true;
    let ttlSeconds = 0;
    let current = 0;
    try {
      const res = await checkFixedWindowRateLimit({
        redis: this.redis,
        key: redisKey,
        max: limiter.max,
        perSeconds: limiter.perSeconds,
      });
      allowed = res.allowed;
      ttlSeconds = res.ttlSeconds;
      current = res.current;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'unknown error';
      this.logger.warn(
        `Rate limit check failed (fail-open) for ${redisKey}: ${errorMessage}`,
      );
      return;
    }

    if (allowed) return;

    const jitter = Math.floor(Math.random() * 500);
    const delayMs = Math.max(1000, ttlSeconds * 1000 + jitter);
    const when = Date.now() + delayMs;

    const logEnvelope = wrapForDb(
      {
        code: 'RATE_LIMITED',
        limiterKey: limiter.key,
        max: limiter.max,
        perSeconds: limiter.perSeconds,
        current,
        ttlSeconds,
        sleepMs: delayMs,
        resumeAt: new Date(when).toISOString(),
      },
      { maxBytes: 128 * 1024, truncate: true, reason: 'step_logs' },
    );

    await this.stepRunRepository.update(stepRunId, {
      logs: logEnvelope as unknown as object,
    });

    await new Promise<void>((resolve) => setTimeout(resolve, delayMs));
  }

  private findSecretRefs(value: unknown): string[] {
    const found = new Set<string>();
    const pattern = /\{\{\s*secret\.([a-zA-Z0-9_-]+)\s*\}\}/g;

    const walk = (node: unknown) => {
      if (typeof node === 'string') {
        pattern.lastIndex = 0;
        let m: RegExpExecArray | null;
        while ((m = pattern.exec(node)) !== null) {
          if (m[1]) found.add(m[1]);
        }
        return;
      }
      if (Array.isArray(node)) {
        for (const x of node) walk(x);
        return;
      }
      if (node && typeof node === 'object') {
        for (const v of Object.values(node as Record<string, unknown>)) walk(v);
      }
    };

    walk(value);
    return Array.from(found);
  }

  private async fetchStepOutputs(
    workflowRunId: string,
    stepKeys: string[],
  ): Promise<Record<string, unknown>> {
    const outputs: Record<string, unknown> = {};

    for (const key of stepKeys) {
      const stepRun = await this.stepRunRepository.findFirst({
        where: { workflowRunId, stepKey: key, status: 'SUCCEEDED' },
        orderBy: { createdAt: 'desc' },
      });

      if (stepRun?.output) {
        const o = stepRun.output as unknown as { data?: unknown };
        outputs[key] = o && typeof o === 'object' && 'data' in o ? (o as any).data : stepRun.output;
      }
    }

    return outputs;
  }

  private async checkWorkflowCompletion(
    workflowRunId: string,
    workflowVersionId: string,
  ): Promise<void> {
    const siblingSteps = await this.stepRunRepository.findManyByWorkflowRun(workflowRunId);

    const allDone = siblingSteps.every((s) => s.status === 'SUCCEEDED' || s.status === 'FAILED');
    if (!allDone) return;

    const hasFailure = siblingSteps.some((s) => s.status === 'FAILED');
    const finalStatus = hasFailure ? 'FAILED' : 'SUCCEEDED';

    const finalized = await this.workflowRunRepository.finalizeIfOpen(
      workflowRunId,
      finalStatus,
    );

    if (!finalized) {
      return;
    }

    const finalizedRun = await this.workflowRunRepository.findById(workflowRunId);
    if (!finalizedRun) {
      this.logger.warn(
        `WorkflowRun ${workflowRunId} finalized but could not be loaded for notifications`,
      );
      return;
    }

    const workflowVersion = await this.cache.getWorkflowVersion(
      workflowVersionId,
      () => this.workflowVersionRepository.findById(workflowVersionId),
    );

    const notifications = this.extractNotifications(workflowVersion?.definition);
    if (notifications.length > 0) {
      await this.runNotifier.notifyRunCompletion({
        workflowRun: finalizedRun,
        notifications,
        finalStatus,
      });
    }

    this.logger.log(`WorkflowRun ${workflowRunId} completed with status: ${finalStatus}`);
  }

  private extractNotifications(definition: unknown): WorkflowNotification[] {
    if (!definition || typeof definition !== 'object') return [];
    const raw = (definition as { notifications?: unknown }).notifications;
    if (!Array.isArray(raw)) return [];

    const notifications: WorkflowNotification[] = [];
    for (const item of raw) {
      if (!item || typeof item !== 'object') continue;
      const candidate = item as {
        provider?: unknown;
        webhook?: unknown;
        on?: unknown;
      };
      if (candidate.provider !== 'slack' && candidate.provider !== 'discord') {
        continue;
      }
      if (typeof candidate.webhook !== 'string' || candidate.webhook.trim() === '') {
        continue;
      }
      if (!Array.isArray(candidate.on) || candidate.on.length === 0) {
        continue;
      }

      const on: NotificationEvent[] = [];
      for (const value of candidate.on) {
        if (value === 'SUCCEEDED' || value === 'FAILED') {
          on.push(value);
        }
      }
      if (on.length === 0) continue;

      notifications.push({
        provider: candidate.provider,
        webhook: candidate.webhook,
        on,
      });
    }

    return notifications;
  }

  private applyHttpOverrides(
    stepType: string,
    baseRequest: unknown,
    requestOverride: StepRunJobPayload['requestOverride'],
  ): unknown {
    if (stepType !== 'http' || !requestOverride) return baseRequest;

    if (typeof baseRequest !== 'object' || baseRequest === null) return baseRequest;

    const base = baseRequest as Record<string, unknown>;
    const merged: Record<string, unknown> = { ...base };

    if (requestOverride.query && typeof merged.query === 'object' && merged.query !== null) {
      merged.query = {
        ...(merged.query as Record<string, unknown>),
        ...requestOverride.query,
      };
    } else if (requestOverride.query) {
      merged.query = requestOverride.query;
    }

    if (requestOverride.body !== undefined) {
      merged.body = requestOverride.body;
    }

    return merged;
  }
}

function unwrapExecutorOutput(value: unknown): unknown {
  if (!value || typeof value !== 'object') return value;
  // Unwrap HttpExecutor body helper wrapper.
  if ('_lunieHttp' in (value as any) && 'data' in (value as any)) {
    return (value as any).data;
  }
  return value;
}
