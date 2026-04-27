import { Injectable, Logger } from '@nestjs/common';
import { SecretRepository } from '@lunie/db-access';
import type {
  NotificationEvent,
  NotificationProvider,
  WorkflowNotification,
} from '@lunie/contracts';
import type { WorkflowRun } from '@prisma/client';
import { CryptoService } from '../crypto/crypto.service';
import { WorkerCacheService } from '../cache/worker-cache.service';

type NotifyRunCompletionParams = {
  workflowRun: WorkflowRun;
  notifications: WorkflowNotification[];
  finalStatus: NotificationEvent;
};

@Injectable()
export class RunNotifierService {
  private readonly logger = new Logger(RunNotifierService.name);

  constructor(
    private readonly secretRepository: SecretRepository,
    private readonly crypto: CryptoService,
    private readonly cache: WorkerCacheService,
  ) {}

  async notifyRunCompletion(params: NotifyRunCompletionParams): Promise<void> {
    const { workflowRun, notifications, finalStatus } = params;
    if (!Array.isArray(notifications) || notifications.length === 0) {
      return;
    }

    for (const notification of notifications) {
      if (!notification.on?.includes(finalStatus)) {
        continue;
      }

      const webhook = await this.resolveWebhook(notification.webhook);
      if (!webhook) {
        this.logger.warn(
          `Skipping notification for run=${workflowRun.id}: invalid or unresolved webhook (provider=${notification.provider})`,
        );
        continue;
      }

      const payload = this.buildPayload(notification.provider, {
        workflowRun,
        finalStatus,
      });

      try {
        const response = await fetch(webhook, {
          method: 'POST',
          headers: {
            'content-type': 'application/json',
          },
          body: JSON.stringify(payload),
          signal: AbortSignal.timeout(8000),
        });

        if (!response.ok) {
          this.logger.warn(
            `Notification failed for run=${workflowRun.id} provider=${notification.provider} status=${response.status}`,
          );
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        this.logger.warn(
          `Notification error for run=${workflowRun.id} provider=${notification.provider}: ${message}`,
        );
      }
    }
  }

  private buildPayload(
    provider: NotificationProvider,
    params: { workflowRun: WorkflowRun; finalStatus: NotificationEvent },
  ): Record<string, unknown> {
    const { workflowRun, finalStatus } = params;
    const durationMs =
      workflowRun.startedAt && workflowRun.finishedAt
        ? Math.max(
            0,
            workflowRun.finishedAt.getTime() - workflowRun.startedAt.getTime(),
          )
        : null;
    const statusEmoji = finalStatus === 'SUCCEEDED' ? '✅' : '❌';
    const statusColor = finalStatus === 'SUCCEEDED' ? 0x22c55e : 0xef4444;
    const durationLabel =
      durationMs === null ? '-' : `${durationMs} ms (${formatDuration(durationMs)})`;
    const startedAt = workflowRun.startedAt?.toISOString() ?? '-';
    const finishedAt = workflowRun.finishedAt?.toISOString() ?? '-';
    const triggerId = workflowRun.triggerId ?? '-';

    const lines = [
      `Lunie run ${finalStatus}`,
      `workflowRunId: ${workflowRun.id}`,
      `workflowId: ${workflowRun.workflowId}`,
      `workflowVersionId: ${workflowRun.workflowVersionId}`,
      `triggerId: ${triggerId}`,
      `startedAt: ${startedAt}`,
      `finishedAt: ${finishedAt}`,
      `duration: ${durationLabel}`,
    ];
    const message = lines.join('\n');

    if (provider === 'discord') {
      return {
        content: `${statusEmoji} Lunie run ${finalStatus}`,
        embeds: [
          {
            title: `${statusEmoji} Workflow Run ${finalStatus}`,
            color: statusColor,
            fields: [
              {
                name: 'Workflow Run ID',
                value: `\`${workflowRun.id}\``,
                inline: false,
              },
              {
                name: 'Workflow ID',
                value: `\`${workflowRun.workflowId}\``,
                inline: true,
              },
              {
                name: 'Workflow Version ID',
                value: `\`${workflowRun.workflowVersionId}\``,
                inline: true,
              },
              {
                name: 'Trigger ID',
                value: triggerId === '-' ? '-' : `\`${triggerId}\``,
                inline: false,
              },
              {
                name: 'Started At',
                value: startedAt,
                inline: true,
              },
              {
                name: 'Finished At',
                value: finishedAt,
                inline: true,
              },
              {
                name: 'Duration',
                value: durationLabel,
                inline: false,
              },
            ],
            timestamp: workflowRun.finishedAt?.toISOString() ?? new Date().toISOString(),
            footer: {
              text: 'Lunie Notifications',
            },
          },
        ],
      };
    }

    return { text: message };
  }

  private async resolveWebhook(value: string): Promise<string | null> {
    const trimmed = String(value ?? '').trim();
    if (!trimmed) return null;

    const secretRef = /^\{\{\s*secret\.([A-Za-z0-9_-]+)\s*\}\}$/.exec(trimmed);
    if (secretRef?.[1]) {
      const secretName = secretRef[1];
      const cached = this.cache.getSecret(secretName);
      if (cached) {
        return this.normalizeWebhookURL(cached);
      }

      const secrets = await this.secretRepository.findManyByNames([secretName]);
      const match = secrets.find((secret) => secret.name === secretName);
      if (!match) return null;

      const decrypted = this.crypto.decryptSecret(match.value);
      this.cache.setSecret(secretName, decrypted);
      return this.normalizeWebhookURL(decrypted);
    }

    return this.normalizeWebhookURL(trimmed);
  }

  private normalizeWebhookURL(value: string): string | null {
    const trimmed = String(value ?? '').trim();
    if (!trimmed) return null;
    try {
      const parsed = new URL(trimmed);
      if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
        return null;
      }
      return parsed.toString();
    } catch {
      return null;
    }
  }
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;

  const totalSeconds = Math.floor(ms / 1000);
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (hours > 0) return `${hours}h ${minutes}m ${seconds}s`;
  if (minutes > 0) return `${minutes}m ${seconds}s`;
  return `${seconds}s`;
}
