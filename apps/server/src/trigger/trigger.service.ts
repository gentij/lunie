import { Injectable } from '@nestjs/common';
import type { Prisma, Trigger } from '@prisma/client';
import { TriggerRepository, WorkflowRepository } from '@lunie/db-access';
import { AppError } from 'src/common/http/errors/app-error';
import { ErrorDefinitions } from 'src/common/http/errors/error-codes';
import { buildPaginationMeta } from 'src/common/pagination/pagination';
import { CronTriggerConfigSchema } from './cron/cron-trigger.types';
import { parseExpression } from 'cron-parser';
import { CryptoService } from 'src/crypto/crypto.service';
import {
  buildUniqueKey,
  deriveReservedKey,
} from 'src/common/identifiers/key.util';

const CRON_FIVE_FIELD = /^\s*([^\s]+\s+){4}[^\s]+\s*$/;

@Injectable()
export class TriggerService {
  constructor(
    private readonly repo: TriggerRepository,
    private readonly workflowRepo: WorkflowRepository,
    private readonly crypto: CryptoService,
  ) {}

  private async assertWorkflowExists(workflowId: string) {
    const wf = await this.workflowRepo.findById(workflowId);
    if (!wf) throw AppError.notFound(ErrorDefinitions.WORKFLOW.NOT_FOUND);
    return wf;
  }

  async create(params: {
    workflowId: string;
    type: Trigger['type'];
    name?: string;
    isActive?: boolean;
    config?: Prisma.InputJsonValue;
  }): Promise<Trigger> {
    await this.assertWorkflowExists(params.workflowId);

    const existingTriggers = await this.repo.findManyByWorkflow(
      params.workflowId,
    );
    const triggerTypeKey = params.type.toLowerCase();
    const triggerKey = buildUniqueKey(
      params.name ?? triggerTypeKey,
      triggerTypeKey,
      existingTriggers.map((trigger) =>
        deriveReservedKey({
          key: trigger.key,
          source: trigger.name ?? trigger.type,
          fallback: trigger.type.toLowerCase(),
        }),
      ),
    );

    if (params.type === 'CRON') {
      const parsed = CronTriggerConfigSchema.safeParse(params.config ?? {});
      if (!parsed.success) {
        throw AppError.badRequest(ErrorDefinitions.COMMON.VALIDATION_ERROR, [
          { field: 'config', message: parsed.error.message },
        ]);
      }

      // Validate cron expression (5-field only) and timezone.
      try {
        if (!CRON_FIVE_FIELD.test(parsed.data.cron)) {
          throw new Error('Cron expression must have 5 fields');
        }
        parseExpression(parsed.data.cron, { tz: parsed.data.timezone });
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        throw AppError.badRequest(ErrorDefinitions.COMMON.VALIDATION_ERROR, [
          { field: 'config.cron', message: `Invalid cron: ${msg}` },
        ]);
      }
    }

    return this.repo.create({
      workflow: { connect: { id: params.workflowId } },
      key: triggerKey,
      type: params.type,
      name: params.name,
      isActive: params.isActive ?? true,
      config: params.config ?? {},
    });
  }

  async list(params: {
    workflowId: string;
    page: number;
    pageSize: number;
    sortBy: 'createdAt' | 'updatedAt';
    sortOrder: 'asc' | 'desc';
  }): Promise<{
    items: Trigger[];
    pagination: ReturnType<typeof buildPaginationMeta>;
  }> {
    await this.assertWorkflowExists(params.workflowId);
    const { items, total } = await this.repo.findPageByWorkflow(params);
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

  async get(workflowId: string, id: string): Promise<Trigger> {
    await this.assertWorkflowExists(workflowId);
    const trigger = await this.repo.findById(id);

    if (!trigger || trigger.workflowId !== workflowId)
      throw AppError.notFound(ErrorDefinitions.TRIGGER.NOT_FOUND);

    return trigger;
  }

  async delete(workflowId: string, id: string): Promise<Trigger> {
    await this.assertWorkflowExists(workflowId);
    const trigger = await this.repo.findById(id);

    if (!trigger || trigger.workflowId !== workflowId)
      throw AppError.notFound(ErrorDefinitions.TRIGGER.NOT_FOUND);

    return this.repo.softDelete(id);
  }

  async update(
    workflowId: string,
    id: string,
    patch: {
      name?: string;
      config?: Prisma.InputJsonValue;
      isActive?: boolean;
    },
  ): Promise<Trigger> {
    await this.get(workflowId, id);

    // If updating config for a CRON trigger, validate it.
    const existing = await this.repo.findById(id);
    if (existing?.type === 'CRON' && patch.config !== undefined) {
      const parsed = CronTriggerConfigSchema.safeParse(patch.config);
      if (!parsed.success) {
        throw AppError.badRequest(ErrorDefinitions.COMMON.VALIDATION_ERROR, [
          { field: 'config', message: parsed.error.message },
        ]);
      }

      try {
        if (!CRON_FIVE_FIELD.test(parsed.data.cron)) {
          throw new Error('Cron expression must have 5 fields');
        }
        parseExpression(parsed.data.cron, { tz: parsed.data.timezone });
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        throw AppError.badRequest(ErrorDefinitions.COMMON.VALIDATION_ERROR, [
          { field: 'config.cron', message: `Invalid cron: ${msg}` },
        ]);
      }
    }

    return this.repo.update(id, patch);
  }

  async rotateWebhookKey(
    workflowId: string,
    id: string,
  ): Promise<{ webhookKey: string }> {
    const trigger = await this.get(workflowId, id);
    this.assertWebhookTriggerType(trigger);

    const webhookKey = this.crypto.generateApiToken();
    const keyHash = this.crypto.hashApiToken(webhookKey);

    const config = toJsonObject(trigger.config);
    config.webhookAuth = {
      mode: 'path-key',
      keyHash,
      rotatedAt: new Date().toISOString(),
    };

    await this.repo.update(id, { config: config as Prisma.InputJsonValue });

    return { webhookKey };
  }

  hasWebhookKey(trigger: Trigger): boolean {
    const webhookAuth = this.getWebhookAuthObject(trigger);
    return (
      webhookAuth?.mode === 'path-key' &&
      typeof webhookAuth.keyHash === 'string'
    );
  }

  hasValidWebhookKey(trigger: Trigger, webhookKey: string): boolean {
    const webhookAuth = this.getWebhookAuthObject(trigger);
    if (!webhookAuth || webhookAuth.mode !== 'path-key') return false;

    const storedHash = webhookAuth.keyHash;
    if (typeof storedHash !== 'string' || storedHash.length === 0) return false;

    const providedHash = this.crypto.hashApiToken(webhookKey.trim());
    return this.crypto.secureCompare(providedHash, storedHash);
  }

  assertWebhookTriggerType(trigger: Trigger): void {
    if (trigger.type !== 'WEBHOOK') {
      throw AppError.badRequest(ErrorDefinitions.TRIGGER.INVALID_TYPE, [
        {
          field: 'type',
          message: 'Webhook operations are only supported for WEBHOOK triggers',
        },
      ]);
    }
  }

  private getWebhookAuthObject(
    trigger: Trigger,
  ): { mode?: unknown; keyHash?: unknown } | null {
    const config = toJsonObject(trigger.config);
    const webhookAuth = config.webhookAuth;

    if (
      !webhookAuth ||
      typeof webhookAuth !== 'object' ||
      Array.isArray(webhookAuth)
    ) {
      return null;
    }

    return webhookAuth as { mode?: unknown; keyHash?: unknown };
  }
}

function toJsonObject(
  value: Prisma.JsonValue | null | undefined,
): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return {};
  }

  return { ...(value as Record<string, unknown>) };
}
