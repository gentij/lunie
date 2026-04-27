import type { SecretRepository } from '@lunie/db-access';
import type { WorkflowNotification } from '@lunie/contracts';
import type { WorkflowRun } from '@prisma/client';
import { CryptoService } from '../crypto/crypto.service';
import { WorkerCacheService } from '../cache/worker-cache.service';
import { RunNotifierService } from './run-notifier.service';

describe('RunNotifierService', () => {
  let secretRepo: { findManyByNames: jest.Mock };
  let cache: {
    getSecret: jest.Mock;
    setSecret: jest.Mock;
  };

  const makeRun = (): WorkflowRun => ({
    id: 'wfr_1',
    workflowId: 'wf_1',
    workflowVersionId: 'wfv_1',
    triggerId: null,
    eventId: null,
    status: 'FAILED',
    input: {},
    overrides: null,
    output: null,
    startedAt: new Date('2026-01-01T00:00:00.000Z'),
    finishedAt: new Date('2026-01-01T00:00:10.000Z'),
    createdAt: new Date('2026-01-01T00:00:00.000Z'),
    updatedAt: new Date('2026-01-01T00:00:10.000Z'),
  });

  beforeEach(() => {
    process.env.LUNIE_SECRET_KEY = '0'.repeat(64);
    secretRepo = {
      findManyByNames: jest.fn(),
    };
    cache = {
      getSecret: jest.fn().mockReturnValue(undefined),
      setSecret: jest.fn(),
    };
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      status: 200,
    } as never);
  });

  it('sends provider payload when event matches', async () => {
    secretRepo.findManyByNames.mockResolvedValue([
      { name: 'DISCORD_WEBHOOK_URL', value: 'https://discord.example/webhook' },
    ]);

    const notifier = new RunNotifierService(
      secretRepo as unknown as SecretRepository,
      new CryptoService(),
      cache as unknown as WorkerCacheService,
    );

    const notifications: WorkflowNotification[] = [
      {
        provider: 'discord',
        webhook: '{{secret.DISCORD_WEBHOOK_URL}}',
        on: ['FAILED'],
      },
    ];

    await notifier.notifyRunCompletion({
      workflowRun: makeRun(),
      notifications,
      finalStatus: 'FAILED',
    });

    expect(secretRepo.findManyByNames).toHaveBeenCalledWith([
      'DISCORD_WEBHOOK_URL',
    ]);
    expect(cache.setSecret).toHaveBeenCalledWith(
      'DISCORD_WEBHOOK_URL',
      'https://discord.example/webhook',
    );
    expect(global.fetch).toHaveBeenCalledTimes(1);
    expect(global.fetch).toHaveBeenCalledWith(
      'https://discord.example/webhook',
      expect.objectContaining({
        method: 'POST',
        headers: { 'content-type': 'application/json' },
      }),
    );
    const [, options] = (global.fetch as jest.Mock).mock.calls[0] as [
      string,
      RequestInit,
    ];
    const payload = JSON.parse(String(options.body)) as {
      content?: string;
      embeds?: Array<{
        title?: string;
        fields?: Array<{ name?: string; value?: string }>;
      }>;
    };
    expect(payload.content).toContain('Lunie run FAILED');
    expect(Array.isArray(payload.embeds)).toBe(true);
    expect(payload.embeds?.[0]?.title).toContain('Workflow Run FAILED');
    const fields = payload.embeds?.[0]?.fields ?? [];
    expect(fields.some((field) => field.name === 'Workflow Run ID')).toBe(true);
    expect(fields.some((field) => field.name === 'Duration')).toBe(true);
  });

  it('does not send when event is not selected', async () => {
    const notifier = new RunNotifierService(
      secretRepo as unknown as SecretRepository,
      new CryptoService(),
      cache as unknown as WorkerCacheService,
    );

    await notifier.notifyRunCompletion({
      workflowRun: makeRun(),
      notifications: [
        {
          provider: 'slack',
          webhook: 'https://hooks.slack.test/services/1',
          on: ['SUCCEEDED'],
        },
      ],
      finalStatus: 'FAILED',
    });

    expect(global.fetch).not.toHaveBeenCalled();
  });

  it('does not throw when webhook request fails', async () => {
    (global.fetch as jest.Mock).mockRejectedValue(new Error('network down'));

    const notifier = new RunNotifierService(
      secretRepo as unknown as SecretRepository,
      new CryptoService(),
      cache as unknown as WorkerCacheService,
    );

    await expect(
      notifier.notifyRunCompletion({
        workflowRun: makeRun(),
        notifications: [
          {
            provider: 'slack',
            webhook: 'https://hooks.slack.test/services/1',
            on: ['FAILED'],
          },
        ],
        finalStatus: 'FAILED',
      }),
    ).resolves.toBeUndefined();
  });
});
