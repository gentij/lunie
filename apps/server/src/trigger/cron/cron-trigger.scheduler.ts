import { Injectable, Logger } from '@nestjs/common';
import { Interval } from '@nestjs/schedule';
import type { Trigger, Workflow } from '@prisma/client';
import { PrismaService } from '@lunie/db-access';
import { OrchestrationService } from 'src/core/orchestration.service';
import { CronTriggerConfigSchema } from './cron-trigger.types';
import { parseExpression } from 'cron-parser';

const TICK_MS = 60_000;

@Injectable()
export class CronTriggerScheduler {
  private readonly logger = new Logger(CronTriggerScheduler.name);

  constructor(
    private readonly prisma: PrismaService,
    private readonly orchestration: OrchestrationService,
  ) {}

  @Interval(TICK_MS)
  async tick(): Promise<void> {
    const now = new Date();

    const triggers = await this.prisma.trigger.findMany({
      where: {
        type: 'CRON',
        isActive: true,
        workflow: { isActive: true, latestVersionId: { not: null } },
      },
      include: {
        workflow: { select: { id: true, latestVersionId: true } },
      },
    });

    for (const trigger of triggers) {
      const typedTrigger = trigger as Trigger & {
        workflow: Pick<Workflow, 'id' | 'latestVersionId'>;
      };

      await this.handleTrigger(now, typedTrigger).catch((err) => {
        const msg = err instanceof Error ? err.message : String(err);
        this.logger.error(
          `Cron trigger tick failed triggerId=${trigger.id}: ${msg}`,
        );
      });
    }
  }

  private async handleTrigger(
    now: Date,
    trigger: Trigger & { workflow: Pick<Workflow, 'id' | 'latestVersionId'> },
  ): Promise<void> {
    const parsed = CronTriggerConfigSchema.safeParse(trigger.config);
    if (!parsed.success) {
      this.logger.warn(
        `Invalid CRON trigger config triggerId=${trigger.id}: ${parsed.error.message}`,
      );
      return;
    }

    const cfg = parsed.data;
    const timezone = cfg.timezone ?? 'UTC';

    // Determine if the cron schedule is due in the current minute window.
    // No catch-up: we only consider the most recent scheduled occurrence and
    // fire if it falls within [minuteStart, minuteEnd).
    const minuteStart = floorToMinute(now);
    const minuteEnd = new Date(minuteStart.getTime() + 60_000);

    const interval = parseExpression(cfg.cron, {
      tz: timezone,
      currentDate: minuteEnd,
    });
    const prev = interval.prev().toDate();

    if (prev < minuteStart || prev >= minuteEnd) {
      return;
    }

    const externalId = `${trigger.id}:${formatMinuteKey(prev)}`;
    const workflowVersionId = trigger.workflow.latestVersionId;
    if (!workflowVersionId) return;

    try {
      await this.orchestration.startWorkflow({
        workflowId: trigger.workflow.id,
        workflowVersionId,
        triggerId: trigger.id,
        eventType: 'CRON',
        eventExternalId: externalId,
        eventPayload: {
          cron: cfg.cron,
          timezone,
          scheduledAt: prev.toISOString(),
          ...cfg.input,
        },
        input: cfg.input,
      });

      this.logger.log(
        `Cron fired triggerId=${trigger.id} workflowId=${trigger.workflow.id} externalId=${externalId}`,
      );
    } catch (err) {
      // Dedupe using Event(triggerId, externalId) unique constraint.
      if (isUniqueEventConflict(err)) {
        return;
      }
      throw err;
    }
  }
}

function floorToMinute(d: Date): Date {
  const out = new Date(d);
  out.setSeconds(0, 0);
  return out;
}

function formatMinuteKey(d: Date): string {
  // v1: store dedupe key in UTC minute ISO without seconds.
  // Example: 2026-02-10T15:18
  return d.toISOString().slice(0, 16);
}

function isUniqueEventConflict(err: unknown): boolean {
  if (!err || typeof err !== 'object') return false;
  // Prisma unique constraint error.
  return (
    'code' in err &&
    typeof (err as { code?: unknown }).code === 'string' &&
    (err as { code: string }).code === 'P2002'
  );
}
