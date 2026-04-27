/* eslint-disable
  @typescript-eslint/no-unsafe-assignment
*/

import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { APP_FILTER, APP_GUARD, APP_INTERCEPTOR, APP_PIPE } from '@nestjs/core';
import { Test } from '@nestjs/testing';
import {
  FastifyAdapter,
  type NestFastifyApplication,
} from '@nestjs/platform-fastify';
import { ZodSerializerInterceptor, ZodValidationPipe } from 'nestjs-zod';

import {
  PrismaService,
  SecretRepository,
  WorkflowRepository,
} from '@lunie/db-access';
import { ResponseInterceptor } from 'src/common/http/interceptors/response.interceptor';
import { AllExceptionsFilter } from 'src/common/http/filters/all-exceptions.filter';
import { OrchestrationService } from 'src/core/orchestration.service';
import { StepRunQueueService } from 'src/queue/step-run-queue.service';
import { WorkflowController } from 'src/workflow/workflow.controller';
import { WorkflowService } from 'src/workflow/workflow.service';
import { createPrismaServiceMock } from 'test/prisma/prisma.mocks';
import { createSecretRepositoryMock } from 'test/secret/secret.repository.mock';
import { AllowAuthGuard } from 'test/utils/allow-auth.guard';
import { createCacheManagerMock } from 'test/utils/cache-manager.mock';
import {
  createWorkflowRepositoryMock,
  type WorkflowRepositoryMock,
} from 'test/workflow/workflow.repository.mock';
import { createWorkflowFixture } from 'test/workflow/workflow.fixtures';

describe('Workflow Run Input (integration e2e)', () => {
  let app: NestFastifyApplication;
  let repo: WorkflowRepositoryMock;
  let prisma: ReturnType<typeof createPrismaServiceMock>;

  beforeEach(async () => {
    repo = createWorkflowRepositoryMock();
    prisma = createPrismaServiceMock();
    const secretRepo = createSecretRepositoryMock();
    const cache = createCacheManagerMock();

    const moduleRef = await Test.createTestingModule({
      controllers: [WorkflowController],
      providers: [
        WorkflowService,
        OrchestrationService,
        { provide: WorkflowRepository, useValue: repo },
        { provide: PrismaService, useValue: prisma },
        { provide: SecretRepository, useValue: secretRepo },
        { provide: CACHE_MANAGER, useValue: cache },
        {
          provide: StepRunQueueService,
          useValue: { enqueueStepRun: jest.fn() },
        },

        { provide: APP_PIPE, useClass: ZodValidationPipe },
        { provide: APP_INTERCEPTOR, useClass: ZodSerializerInterceptor },
        { provide: APP_FILTER, useClass: AllExceptionsFilter },

        { provide: APP_GUARD, useClass: AllowAuthGuard },
        { provide: APP_INTERCEPTOR, useClass: ResponseInterceptor },
      ],
    }).compile();

    app = moduleRef.createNestApplication<NestFastifyApplication>(
      new FastifyAdapter(),
    );

    await app.init();
    await app.getHttpAdapter().getInstance().ready();
  });

  afterEach(async () => {
    await app?.close();
  });

  it('POST /workflows/:id/run lets input override workflow defaults', async () => {
    repo.findById.mockResolvedValue(
      createWorkflowFixture({ id: 'wf_1', latestVersionId: 'wfv_1' }),
    );

    const tx = {
      trigger: {
        findFirst: jest.fn().mockResolvedValue({ id: 'tr_manual' }),
        create: jest.fn(),
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
              shouldPass: true,
            },
            steps: [],
          },
        }),
      },
      workflow: {
        create: jest.fn(),
        update: jest.fn(),
      },
    };

    prisma.$transaction.mockImplementation((cb) => Promise.resolve(cb(tx)));

    const res = await app.inject({
      method: 'POST',
      url: '/workflows/wf_1/run',
      payload: {
        input: { shouldPass: false },
        overrides: {},
      },
    });

    expect(res.statusCode).toBe(201);

    expect(tx.event.create).toHaveBeenCalledWith(
      expect.objectContaining({
        data: expect.objectContaining({
          payload: { shouldPass: false },
        }),
      }),
    );

    expect(tx.workflowRun.create).toHaveBeenCalledWith(
      expect.objectContaining({
        data: expect.objectContaining({
          input: { shouldPass: false },
        }),
      }),
    );
  });
});
