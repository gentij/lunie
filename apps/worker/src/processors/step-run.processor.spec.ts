import { StepRunProcessor } from './step-run.processor';
import {
  type PrismaService,
  SecretRepository,
  StepRunRepository,
  WorkflowRunRepository,
  WorkflowVersionRepository,
} from '@lunie/db-access';
import type { Job } from 'bullmq';
import type { StepRunJobPayload } from '@lunie/contracts';
import { ExecutorRegistry } from '../executors/executor-registry';
import { createPrismaServiceMock } from 'test/prisma.mocks';
import { CryptoService } from '../crypto/crypto.service';

describe('StepRunProcessor', () => {
  let executeMock: jest.Mock;
  let registryMock: jest.Mock;
  let cacheMock: {
    getWorkflowVersion: jest.Mock;
    getWorkflowRun: jest.Mock;
    getSecret: jest.Mock;
    setSecret: jest.Mock;
  };
  let runNotifierMock: {
    notifyRunCompletion: jest.Mock;
  };

  beforeEach(() => {
    jest.useFakeTimers();
    jest.clearAllMocks();

    executeMock = jest.fn().mockResolvedValue({
      statusCode: 200,
      headers: { 'content-type': 'application/json' },
      body: { ok: true },
    });

    registryMock = jest.fn().mockReturnValue({ stepType: 'http', execute: executeMock });

    cacheMock = {
      getWorkflowVersion: jest.fn((_: string, loader: () => Promise<unknown>) => loader()),
      getWorkflowRun: jest.fn((_: string, loader: () => Promise<unknown>) => loader()),
      getSecret: jest.fn().mockReturnValue(undefined),
      setSecret: jest.fn(),
    };

    runNotifierMock = {
      notifyRunCompletion: jest.fn().mockResolvedValue(undefined),
    };
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('marks step succeeded and workflow succeeded when all steps complete', async () => {
    const prisma = createPrismaServiceMock();

    const startedAt = new Date(Date.now() - 50);

    prisma.workflowRun.updateMany.mockResolvedValue({ count: 1 });
    prisma.workflowVersion.findUnique.mockResolvedValue({
      id: 'wfv_1',
      definition: {
        steps: [
          {
            key: 'step_1',
            type: 'http',
            request: {
              method: 'POST',
              url: 'https://x.test',
              query: { base: '1' },
              body: { content: 'base' },
            },
          },
          { key: 'step_2', type: 'http', request: { method: 'GET', url: 'https://x.test' } },
        ],
      },
    });

    prisma.stepRun.update.mockResolvedValue({});
    prisma.stepRun.findUnique.mockResolvedValue({
      id: 'sr_1',
      workflowRunId: 'wfr_1',
      startedAt,
    });
    prisma.stepRun.findMany.mockResolvedValue([
      { id: 'sr_1', status: 'SUCCEEDED' },
      { id: 'sr_2', status: 'SUCCEEDED' },
    ]);

    prisma.workflowRun.findUnique.mockResolvedValue({
      id: 'wfr_1',
      workflowId: 'wf_1',
    });

    const prismaService = prisma as unknown as PrismaService;
    const stepRunRepo = new StepRunRepository(prismaService);
    const workflowRunRepo = new WorkflowRunRepository(prismaService);
    const workflowVersionRepo = new WorkflowVersionRepository(prismaService);
    const secretRepo = {
      findManyByNames: jest.fn().mockResolvedValue([]),
    } as unknown as SecretRepository;

    process.env.LUNIE_SECRET_KEY = '0'.repeat(64);
    const crypto = new CryptoService();

    const redis = { eval: jest.fn().mockResolvedValue([1, 60]) };

    const processor = new StepRunProcessor(
      stepRunRepo,
      workflowRunRepo,
      workflowVersionRepo,
      secretRepo,
      crypto,
      redis as any,
      { get: registryMock } as unknown as ExecutorRegistry,
      cacheMock as any,
      runNotifierMock as any,
    );

    const job = {
      id: 'job_1',
      data: {
        workflowRunId: 'wfr_1',
        stepRunId: 'sr_1',
        stepKey: 'step_1',
        workflowVersionId: 'wfv_1',
        input: { hello: 'world' },
        requestOverride: {
          query: { override: '2' },
          body: { content: 'override' },
        },
      },
    } as unknown as Job<StepRunJobPayload>;

    const p = processor.process(job);
    await jest.runOnlyPendingTimersAsync();
    await p;

    expect(prisma.workflowRun.updateMany).toHaveBeenCalledWith(
      expect.objectContaining({ where: { id: 'wfr_1', status: 'QUEUED' } }),
    );

    expect(prisma.stepRun.update).toHaveBeenCalledWith(
      expect.objectContaining({ where: { id: 'sr_1' } }),
    );

    expect(executeMock).toHaveBeenCalledTimes(1);
    expect(executeMock).toHaveBeenCalledWith({
      request: {
        method: 'POST',
        url: 'https://x.test',
        query: { base: '1', override: '2' },
        body: { content: 'override' },
      },
      input: {
        input: { hello: 'world' },
        steps: {},
        secret: {},
      },
    });

    expect(prisma.workflowRun.updateMany).toHaveBeenCalledWith(
      expect.objectContaining({
        where: expect.objectContaining({ id: 'wfr_1', finishedAt: null }),
        data: expect.objectContaining({ status: 'SUCCEEDED' }) as unknown,
      }),
    );
    expect(runNotifierMock.notifyRunCompletion).not.toHaveBeenCalled();
  });

  it('continues when rate limit check fails', async () => {
    const prisma = createPrismaServiceMock();

    const startedAt = new Date(Date.now() - 50);

    prisma.workflowRun.updateMany.mockResolvedValue({ count: 1 });
    prisma.workflowVersion.findUnique.mockResolvedValue({
      id: 'wfv_1',
      definition: {
        steps: [
          {
            key: 'step_1',
            type: 'http',
            request: { method: 'GET', url: 'https://x.test' },
          },
        ],
      },
    });

    prisma.stepRun.update.mockResolvedValue({});
    prisma.stepRun.findUnique.mockResolvedValue({
      id: 'sr_1',
      workflowRunId: 'wfr_1',
      startedAt,
    });
    prisma.stepRun.findMany.mockResolvedValue([{ id: 'sr_1', status: 'SUCCEEDED' }]);

    prisma.workflowRun.findUnique.mockResolvedValue({
      id: 'wfr_1',
      workflowId: 'wf_1',
    });

    const prismaService = prisma as unknown as PrismaService;
    const stepRunRepo = new StepRunRepository(prismaService);
    const workflowRunRepo = new WorkflowRunRepository(prismaService);
    const workflowVersionRepo = new WorkflowVersionRepository(prismaService);
    const secretRepo = {
      findManyByNames: jest.fn().mockResolvedValue([]),
    } as unknown as SecretRepository;

    process.env.LUNIE_SECRET_KEY = '0'.repeat(64);
    const crypto = new CryptoService();

    const redis = { eval: jest.fn().mockRejectedValue(new Error('redis down')) };

    const processor = new StepRunProcessor(
      stepRunRepo,
      workflowRunRepo,
      workflowVersionRepo,
      secretRepo,
      crypto,
      redis as any,
      { get: registryMock } as unknown as ExecutorRegistry,
      cacheMock as any,
      runNotifierMock as any,
    );

    const job = {
      id: 'job_1',
      data: {
        workflowRunId: 'wfr_1',
        stepRunId: 'sr_1',
        stepKey: 'step_1',
        workflowVersionId: 'wfv_1',
        input: { hello: 'world' },
      },
    } as unknown as Job<StepRunJobPayload>;

    const p = processor.process(job);
    await jest.runOnlyPendingTimersAsync();
    await p;

    expect(executeMock).toHaveBeenCalledTimes(1);
  });

  it('dispatches run notifications when workflow completes', async () => {
    const prisma = createPrismaServiceMock();

    prisma.workflowRun.updateMany.mockResolvedValue({ count: 1 });
    prisma.workflowVersion.findUnique.mockResolvedValue({
      id: 'wfv_1',
      definition: {
        notifications: [
          {
            provider: 'discord',
            webhook: '{{secret.DISCORD_WEBHOOK_URL}}',
            on: ['SUCCEEDED'],
          },
        ],
        steps: [
          {
            key: 'step_1',
            type: 'http',
            request: { method: 'GET', url: 'https://x.test' },
          },
        ],
      },
    });

    prisma.stepRun.update.mockResolvedValue({});
    prisma.stepRun.findUnique.mockResolvedValue({
      id: 'sr_1',
      workflowRunId: 'wfr_1',
      startedAt: new Date(Date.now() - 100),
    });
    prisma.stepRun.findMany.mockResolvedValue([{ id: 'sr_1', status: 'SUCCEEDED' }]);
    prisma.workflowRun.findUnique.mockResolvedValue({
      id: 'wfr_1',
      workflowId: 'wf_1',
      workflowVersionId: 'wfv_1',
      triggerId: null,
      eventId: null,
      status: 'SUCCEEDED',
      input: {},
      overrides: null,
      output: null,
      startedAt: new Date('2026-01-01T00:00:00.000Z'),
      finishedAt: new Date('2026-01-01T00:00:05.000Z'),
      createdAt: new Date('2026-01-01T00:00:00.000Z'),
      updatedAt: new Date('2026-01-01T00:00:05.000Z'),
    });

    const prismaService = prisma as unknown as PrismaService;
    const stepRunRepo = new StepRunRepository(prismaService);
    const workflowRunRepo = new WorkflowRunRepository(prismaService);
    const workflowVersionRepo = new WorkflowVersionRepository(prismaService);
    const secretRepo = {
      findManyByNames: jest.fn().mockResolvedValue([]),
    } as unknown as SecretRepository;

    process.env.LUNIE_SECRET_KEY = '0'.repeat(64);
    const crypto = new CryptoService();

    const redis = { eval: jest.fn().mockResolvedValue([1, 60]) };

    const processor = new StepRunProcessor(
      stepRunRepo,
      workflowRunRepo,
      workflowVersionRepo,
      secretRepo,
      crypto,
      redis as any,
      { get: registryMock } as unknown as ExecutorRegistry,
      cacheMock as any,
      runNotifierMock as any,
    );

    const job = {
      id: 'job_1',
      data: {
        workflowRunId: 'wfr_1',
        stepRunId: 'sr_1',
        stepKey: 'step_1',
        workflowVersionId: 'wfv_1',
        input: {},
      },
    } as unknown as Job<StepRunJobPayload>;

    const p = processor.process(job);
    await jest.runOnlyPendingTimersAsync();
    await p;

    expect(runNotifierMock.notifyRunCompletion).toHaveBeenCalledWith(
      expect.objectContaining({
        finalStatus: 'SUCCEEDED',
      }),
    );
  });

  it('marks workflow failed when a sibling fails and all steps complete', async () => {
    const prisma = createPrismaServiceMock();

    prisma.workflowRun.updateMany.mockResolvedValue({ count: 1 });

    prisma.workflowVersion.findUnique.mockResolvedValue({
      id: 'wfv_1',
      definition: {
        steps: [{ key: 'step_1', type: 'http', request: { method: 'GET', url: 'https://x.test' } }],
      },
    });

    prisma.stepRun.update.mockResolvedValue({});
    prisma.stepRun.findUnique.mockResolvedValue({
      id: 'sr_1',
      workflowRunId: 'wfr_1',
      startedAt: new Date(),
    });

    prisma.stepRun.findMany.mockResolvedValue([
      { id: 'sr_1', status: 'FAILED' },
      { id: 'sr_2', status: 'SUCCEEDED' },
    ]);

    prisma.workflowRun.findUnique.mockResolvedValue({
      id: 'wfr_1',
      workflowId: 'wf_1',
    });

    const prismaService = prisma as unknown as PrismaService;
    const stepRunRepo = new StepRunRepository(prismaService);
    const workflowRunRepo = new WorkflowRunRepository(prismaService);
    const workflowVersionRepo = new WorkflowVersionRepository(prismaService);
    const secretRepo = {
      findManyByNames: jest.fn().mockResolvedValue([]),
    } as unknown as SecretRepository;

    process.env.LUNIE_SECRET_KEY = '0'.repeat(64);
    const crypto = new CryptoService();

    const executeFailed = jest.fn().mockRejectedValue(new Error('boom'));

    const redis = { eval: jest.fn().mockResolvedValue([1, 60]) };

    const processor = new StepRunProcessor(
      stepRunRepo,
      workflowRunRepo,
      workflowVersionRepo,
      secretRepo,
      crypto,
      redis as any,
      { get: () => ({ stepType: 'http', execute: executeFailed }) } as unknown as ExecutorRegistry,
      cacheMock as any,
      runNotifierMock as any,
    );

    const job = {
      id: 'job_1',
      data: {
        workflowRunId: 'wfr_1',
        stepRunId: 'sr_1',
        stepKey: 'step_1',
        workflowVersionId: 'wfv_1',
        input: {},
      },
    } as unknown as Job<StepRunJobPayload>;

    await expect(processor.process(job)).rejects.toThrow('boom');

    expect(prisma.workflowRun.updateMany).toHaveBeenCalledWith(
      expect.objectContaining({
        where: expect.objectContaining({ id: 'wfr_1', finishedAt: null }),
        data: expect.objectContaining({ status: 'FAILED' }) as unknown,
      }),
    );
    expect(runNotifierMock.notifyRunCompletion).not.toHaveBeenCalled();
  });
});
