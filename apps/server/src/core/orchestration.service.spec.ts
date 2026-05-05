import { Test } from '@nestjs/testing';

import { PrismaService } from '@lunie/db-access';
import { Prisma } from '@prisma/client';
import { StepRunQueueService } from 'src/queue/step-run-queue.service';
import { OrchestrationService } from './orchestration.service';
import {
  createPrismaServiceMock,
  type PrismaServiceMock,
  type PrismaTxMock,
} from 'test/prisma/prisma.mocks';

describe('OrchestrationService', () => {
  let service: OrchestrationService;

  it('creates event, workflowRun, stepRuns and enqueues', async () => {
    const enqueueStepRun = jest
      .fn()
      .mockResolvedValueOnce({ id: 'job_1' })
      .mockResolvedValueOnce({ id: 'job_2' });

    const tx: PrismaTxMock = {
      trigger: {
        findFirst: jest.fn().mockResolvedValue({ id: 'tr_manual' }),
        create: jest.fn(),
        findMany: jest.fn().mockResolvedValue([]),
      },
      event: {
        create: jest.fn().mockResolvedValue({ id: 'ev_1' }),
      },
      workflowRun: {
        create: jest.fn().mockResolvedValue({ id: 'wfr_1' }),
        update: jest.fn(),
      },
      workflowVersion: {
        create: jest.fn(),
        findFirst: jest.fn(),
        findUniqueOrThrow: jest.fn().mockResolvedValue({
          definition: {
            steps: [
              { key: 'step_1', type: 'http', request: { method: 'GET' } },
              { key: 'step_2', type: 'http', request: { method: 'GET' } },
            ],
          },
        }),
      },
      stepRun: {
        create: jest
          .fn()
          .mockImplementationOnce((args: unknown) => {
            const createArgs = args as { data: { stepKey: string } };
            return {
              id: 'sr_1',
              stepKey: createArgs.data.stepKey,
            };
          })
          .mockImplementationOnce((args: unknown) => {
            const createArgs = args as { data: { stepKey: string } };
            return {
              id: 'sr_2',
              stepKey: createArgs.data.stepKey,
            };
          }),
        updateMany: jest.fn(),
      },
      workflow: {
        create: jest.fn(),
        update: jest.fn().mockResolvedValue({ runSequence: 1 }),
      },
    };

    const prismaMock: PrismaServiceMock = createPrismaServiceMock();
    prismaMock.$transaction.mockImplementation((cb) => Promise.resolve(cb(tx)));

    const moduleRef = await Test.createTestingModule({
      providers: [
        OrchestrationService,
        {
          provide: PrismaService,
          useValue: prismaMock as unknown as PrismaService,
        },
        { provide: StepRunQueueService, useValue: { enqueueStepRun } },
      ],
    }).compile();

    service = moduleRef.get(OrchestrationService);

    const result = await service.startWorkflow({
      workflowId: 'wf_1',
      workflowVersionId: 'wfv_1',
      eventType: 'MANUAL',
      input: { hello: 'world' },
      overrides: {
        step_1: { body: { content: 'dynamic' } },
      },
    });

    expect(result).toEqual({
      workflowRunId: 'wfr_1',
      stepRunIds: ['sr_1', 'sr_2'],
    });

    expect(tx.trigger?.findFirst).toHaveBeenCalledWith({
      where: { workflowId: 'wf_1', type: 'MANUAL' },
      select: { id: true },
    });

    const eventCreate = tx.event?.create as jest.MockedFunction<
      (args: { data: { triggerId: string; type?: string } }) => unknown
    >;
    const eventCreateArgs = eventCreate.mock.calls[0]?.[0];
    if (!eventCreateArgs) throw new Error('Expected event.create to be called');
    expect(eventCreateArgs.data.triggerId).toBe('tr_manual');
    expect(eventCreateArgs.data.type).toBe('MANUAL');

    const workflowRunCreate = tx.workflowRun?.create as jest.MockedFunction<
      (args: {
        data: {
          workflowId: string;
          workflowVersionId: string;
          number?: number;
          triggerId?: string;
          eventId?: string;
          overrides?: unknown;
        };
      }) => unknown
    >;
    const workflowRunCreateArgs = workflowRunCreate.mock.calls[0]?.[0];
    if (!workflowRunCreateArgs)
      throw new Error('Expected workflowRun.create to be called');
    expect(workflowRunCreateArgs.data.workflowId).toBe('wf_1');
    expect(workflowRunCreateArgs.data.workflowVersionId).toBe('wfv_1');
    expect(workflowRunCreateArgs.data.number).toBe(1);
    expect(workflowRunCreateArgs.data.triggerId).toBe('tr_manual');
    expect(workflowRunCreateArgs.data.eventId).toBe('ev_1');
    expect(workflowRunCreateArgs.data.overrides).toEqual({
      step_1: { body: { content: 'dynamic' } },
    });
    expect(tx.workflow.update).toHaveBeenCalledWith({
      where: { id: 'wf_1' },
      data: { runSequence: { increment: 1 } },
      select: { runSequence: true },
    });

    expect(tx.stepRun?.create).toHaveBeenCalledTimes(2);

    expect(tx.stepRun?.create).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({
        data: expect.objectContaining({
          stepKey: 'step_1',
          requestOverride: { body: { content: 'dynamic' } },
        }) as unknown,
      }),
    );
    expect(tx.stepRun?.create).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({
        data: expect.objectContaining({
          stepKey: 'step_2',
          requestOverride: Prisma.DbNull,
        }) as unknown,
      }),
    );

    expect(enqueueStepRun).toHaveBeenCalledTimes(2);
    expect(enqueueStepRun).toHaveBeenNthCalledWith(
      1,
      'http',
      expect.objectContaining({
        stepRunId: 'sr_1',
        stepKey: 'step_1',
        requestOverride: {
          body: { content: 'dynamic' },
        },
      }),
      undefined,
    );
    expect(enqueueStepRun).toHaveBeenNthCalledWith(
      2,
      'http',
      expect.objectContaining({
        stepRunId: 'sr_2',
        stepKey: 'step_2',
        requestOverride: undefined,
      }),
      undefined,
    );
  });

  it('infers dependsOn from step template references', async () => {
    const enqueueStepRun = jest
      .fn()
      .mockResolvedValueOnce({ id: 'job_1' })
      .mockResolvedValueOnce({ id: 'job_2' });

    const tx: PrismaTxMock = {
      trigger: {
        findFirst: jest.fn().mockResolvedValue({ id: 'tr_manual' }),
        create: jest.fn(),
        findMany: jest.fn().mockResolvedValue([]),
      },
      event: {
        create: jest.fn().mockResolvedValue({ id: 'ev_1' }),
      },
      workflowRun: {
        create: jest.fn().mockResolvedValue({ id: 'wfr_1' }),
        update: jest.fn(),
      },
      workflowVersion: {
        create: jest.fn(),
        findFirst: jest.fn(),
        findUniqueOrThrow: jest.fn().mockResolvedValue({
          definition: {
            steps: [
              { key: 'step_1', type: 'http', request: { method: 'GET' } },
              {
                key: 'step_2',
                type: 'http',
                request: {
                  method: 'POST',
                  url: 'https://example.com',
                  body: {
                    a: '{{steps.step_1.output.statusCode}}',
                  },
                },
              },
            ],
          },
        }),
      },
      stepRun: {
        create: jest
          .fn()
          .mockResolvedValueOnce({ id: 'sr_1', stepKey: 'step_1' })
          .mockResolvedValueOnce({ id: 'sr_2', stepKey: 'step_2' }),
      },
      workflow: {
        create: jest.fn(),
        update: jest.fn().mockResolvedValue({ runSequence: 1 }),
      },
    };

    const prismaMock: PrismaServiceMock = createPrismaServiceMock();
    prismaMock.$transaction.mockImplementation((cb) => Promise.resolve(cb(tx)));

    const moduleRef = await Test.createTestingModule({
      providers: [
        OrchestrationService,
        {
          provide: PrismaService,
          useValue: prismaMock as unknown as PrismaService,
        },
        { provide: StepRunQueueService, useValue: { enqueueStepRun } },
      ],
    }).compile();

    service = moduleRef.get(OrchestrationService);

    await service.startWorkflow({
      workflowId: 'wf_1',
      workflowVersionId: 'wfv_1',
      eventType: 'MANUAL',
    });

    expect(enqueueStepRun).toHaveBeenNthCalledWith(
      2,
      'http',
      expect.objectContaining({
        stepKey: 'step_2',
        dependsOn: ['step_1'],
      }),
      expect.objectContaining({ dependsOn: ['job_1'] }),
    );
  });

  it('lets manual/webhook input override workflow definition input defaults', async () => {
    const enqueueStepRun = jest.fn();

    const tx: PrismaTxMock = {
      trigger: {
        findFirst: jest.fn().mockResolvedValue({ id: 'tr_manual' }),
        create: jest.fn(),
        findMany: jest.fn().mockResolvedValue([]),
      },
      event: {
        create: jest.fn().mockResolvedValue({ id: 'ev_1' }),
      },
      workflowRun: {
        create: jest.fn().mockResolvedValue({ id: 'wfr_1' }),
        update: jest.fn().mockResolvedValue({ id: 'wfr_1' }),
      },
      workflowVersion: {
        create: jest.fn(),
        findFirst: jest.fn(),
        findUniqueOrThrow: jest.fn().mockResolvedValue({
          definition: {
            input: {
              shouldFail: false,
              source: 'workflow-default',
            },
            steps: [],
          },
        }),
      },
      workflow: {
        create: jest.fn(),
        update: jest.fn().mockResolvedValue({ runSequence: 1 }),
      },
    };

    const prismaMock: PrismaServiceMock = createPrismaServiceMock();
    prismaMock.$transaction.mockImplementation((cb) => Promise.resolve(cb(tx)));

    const moduleRef = await Test.createTestingModule({
      providers: [
        OrchestrationService,
        {
          provide: PrismaService,
          useValue: prismaMock as unknown as PrismaService,
        },
        { provide: StepRunQueueService, useValue: { enqueueStepRun } },
      ],
    }).compile();

    service = moduleRef.get(OrchestrationService);

    await service.startWorkflow({
      workflowId: 'wf_1',
      workflowVersionId: 'wfv_1',
      eventType: 'MANUAL',
      input: {
        shouldFail: true,
        branch: 'manual-input',
      },
    });

    const workflowRunCreate = tx.workflowRun?.create as jest.MockedFunction<
      (args: {
        data: {
          input?: unknown;
        };
      }) => unknown
    >;
    const workflowRunCreateArgs = workflowRunCreate.mock.calls[0]?.[0];
    if (!workflowRunCreateArgs)
      throw new Error('Expected workflowRun.create to be called');

    expect(workflowRunCreateArgs.data.input).toEqual({
      shouldFail: true,
      source: 'workflow-default',
      branch: 'manual-input',
    });
  });

  it('completes immediately when no steps exist', async () => {
    const enqueueStepRun = jest.fn();

    const tx: PrismaTxMock = {
      trigger: {
        findFirst: jest.fn().mockResolvedValue({ id: 'tr_manual' }),
        create: jest.fn(),
        findMany: jest.fn().mockResolvedValue([]),
      },
      event: {
        create: jest.fn().mockResolvedValue({ id: 'ev_1' }),
      },
      workflowRun: {
        create: jest.fn().mockResolvedValue({ id: 'wfr_1' }),
        update: jest.fn().mockResolvedValue({ id: 'wfr_1' }),
      },
      workflowVersion: {
        create: jest.fn(),
        findFirst: jest.fn(),
        findUniqueOrThrow: jest
          .fn()
          .mockResolvedValue({ definition: { steps: [] } }),
      },
      workflow: {
        create: jest.fn(),
        update: jest.fn().mockResolvedValue({ runSequence: 1 }),
      },
    };

    const prismaMock: PrismaServiceMock = createPrismaServiceMock();
    prismaMock.$transaction.mockImplementation((cb) => Promise.resolve(cb(tx)));

    const moduleRef = await Test.createTestingModule({
      providers: [
        OrchestrationService,
        {
          provide: PrismaService,
          useValue: prismaMock as unknown as PrismaService,
        },
        { provide: StepRunQueueService, useValue: { enqueueStepRun } },
      ],
    }).compile();

    service = moduleRef.get(OrchestrationService);

    const result = await service.startWorkflow({
      workflowId: 'wf_1',
      workflowVersionId: 'wfv_1',
      eventType: 'MANUAL',
    });

    expect(result).toEqual({ workflowRunId: 'wfr_1', stepRunIds: [] });
    expect(enqueueStepRun).not.toHaveBeenCalled();
    const workflowRunUpdate = tx.workflowRun?.update as jest.MockedFunction<
      (args: { where: { id: string }; data: { status?: string } }) => unknown
    >;
    const workflowRunUpdateArgs = workflowRunUpdate.mock.calls[0]?.[0];
    if (!workflowRunUpdateArgs)
      throw new Error('Expected workflowRun.update to be called');
    expect(workflowRunUpdateArgs.where.id).toBe('wfr_1');
    expect(workflowRunUpdateArgs.data.status).toBe('SUCCEEDED');
  });

  it('marks workflow failed when enqueue fails', async () => {
    const enqueueStepRun = jest
      .fn()
      .mockResolvedValueOnce({ id: 'job_1' })
      .mockRejectedValueOnce(new Error('redis down'));

    const tx1: PrismaTxMock = {
      trigger: {
        findFirst: jest.fn().mockResolvedValue({ id: 'tr_manual' }),
        create: jest.fn(),
        findMany: jest.fn().mockResolvedValue([]),
      },
      event: {
        create: jest.fn().mockResolvedValue({ id: 'ev_1' }),
      },
      workflowRun: {
        create: jest.fn().mockResolvedValue({ id: 'wfr_1' }),
        update: jest.fn(),
      },
      workflowVersion: {
        create: jest.fn(),
        findFirst: jest.fn(),
        findUniqueOrThrow: jest.fn().mockResolvedValue({
          definition: {
            steps: [
              { key: 'step_1', type: 'http', request: {} },
              { key: 'step_2', type: 'http', request: {} },
            ],
          },
        }),
      },
      stepRun: {
        create: jest
          .fn()
          .mockResolvedValueOnce({ id: 'sr_1', stepKey: 'step_1' })
          .mockResolvedValueOnce({ id: 'sr_2', stepKey: 'step_2' }),
      },
      workflow: {
        create: jest.fn(),
        update: jest.fn().mockResolvedValue({ runSequence: 1 }),
      },
    };

    const tx2: PrismaTxMock = {
      stepRun: {
        updateMany: jest.fn().mockResolvedValue({ count: 1 }),
      },
      workflowRun: {
        update: jest.fn().mockResolvedValue({ id: 'wfr_1' }),
        create: jest.fn(),
      },
      workflow: {
        create: jest.fn(),
        update: jest.fn(),
      },
      workflowVersion: {
        create: jest.fn(),
        findFirst: jest.fn(),
      },
    };

    const prismaMock: PrismaServiceMock = createPrismaServiceMock();
    prismaMock.$transaction
      .mockImplementationOnce((cb) => Promise.resolve(cb(tx1)))
      .mockImplementationOnce((cb) => Promise.resolve(cb(tx2)));

    const moduleRef = await Test.createTestingModule({
      providers: [
        OrchestrationService,
        {
          provide: PrismaService,
          useValue: prismaMock as unknown as PrismaService,
        },
        { provide: StepRunQueueService, useValue: { enqueueStepRun } },
      ],
    }).compile();

    service = moduleRef.get(OrchestrationService);

    await expect(
      service.startWorkflow({
        workflowId: 'wf_1',
        workflowVersionId: 'wfv_1',
        eventType: 'MANUAL',
      }),
    ).rejects.toThrow('redis down');

    const stepRunUpdateMany = tx2.stepRun?.updateMany as jest.MockedFunction<
      (args: {
        where: { id: { in: string[] }; status: string };
        data: { status?: string };
      }) => unknown
    >;
    const stepRunUpdateManyArgs = stepRunUpdateMany.mock.calls[0]?.[0];
    if (!stepRunUpdateManyArgs)
      throw new Error('Expected stepRun.updateMany to be called');
    expect(stepRunUpdateManyArgs.where.id.in).toEqual(['sr_2']);
    expect(stepRunUpdateManyArgs.where.status).toBe('QUEUED');
    expect(stepRunUpdateManyArgs.data.status).toBe('FAILED');

    const workflowRunUpdate = tx2.workflowRun?.update as jest.MockedFunction<
      (args: { where: { id: string }; data: { status?: string } }) => unknown
    >;
    const workflowRunUpdateArgs = workflowRunUpdate.mock.calls[0]?.[0];
    if (!workflowRunUpdateArgs)
      throw new Error('Expected workflowRun.update to be called');
    expect(workflowRunUpdateArgs.where.id).toBe('wfr_1');
    expect(workflowRunUpdateArgs.data.status).toBe('FAILED');
  });

  it('creates a keyed manual trigger when no manual trigger exists', async () => {
    const enqueueStepRun = jest.fn();

    const tx: PrismaTxMock = {
      trigger: {
        findFirst: jest.fn().mockResolvedValue(null),
        findMany: jest
          .fn()
          .mockResolvedValue([
            { key: null, name: 'Webhook Trigger', type: 'WEBHOOK' },
          ]),
        create: jest.fn().mockResolvedValue({ id: 'tr_manual_created' }),
      },
      event: {
        create: jest.fn().mockResolvedValue({ id: 'ev_1' }),
      },
      workflowRun: {
        create: jest.fn().mockResolvedValue({ id: 'wfr_1' }),
        update: jest.fn().mockResolvedValue({ id: 'wfr_1' }),
      },
      workflowVersion: {
        create: jest.fn(),
        findFirst: jest.fn(),
        findUniqueOrThrow: jest
          .fn()
          .mockResolvedValue({ definition: { steps: [] } }),
      },
      workflow: {
        create: jest.fn(),
        update: jest.fn().mockResolvedValue({ runSequence: 1 }),
      },
    };

    const prismaMock: PrismaServiceMock = createPrismaServiceMock();
    prismaMock.$transaction.mockImplementation((cb) => Promise.resolve(cb(tx)));

    const moduleRef = await Test.createTestingModule({
      providers: [
        OrchestrationService,
        {
          provide: PrismaService,
          useValue: prismaMock as unknown as PrismaService,
        },
        { provide: StepRunQueueService, useValue: { enqueueStepRun } },
      ],
    }).compile();

    service = moduleRef.get(OrchestrationService);

    await service.startWorkflow({
      workflowId: 'wf_1',
      workflowVersionId: 'wfv_1',
      eventType: 'MANUAL',
    });

    expect(tx.trigger.create).toHaveBeenCalledWith({
      data: {
        workflowId: 'wf_1',
        key: 'manual',
        type: 'MANUAL',
        name: 'Manual',
        isActive: true,
        config: {},
      },
      select: { id: true },
    });
  });
});
