/* eslint-disable
  @typescript-eslint/no-unsafe-assignment,
  @typescript-eslint/no-unsafe-member-access
*/

import { Test } from '@nestjs/testing';
import { APP_FILTER, APP_GUARD, APP_INTERCEPTOR, APP_PIPE } from '@nestjs/core';
import {
  FastifyAdapter,
  type NestFastifyApplication,
} from '@nestjs/platform-fastify';
import { ZodSerializerInterceptor, ZodValidationPipe } from 'nestjs-zod';
import { CACHE_MANAGER } from '@nestjs/cache-manager';

import { WorkflowController } from 'src/workflow/workflow.controller';
import { WorkflowService } from 'src/workflow/workflow.service';
import { OrchestrationService } from 'src/core/orchestration.service';
import { AllExceptionsFilter } from 'src/common/http/filters/all-exceptions.filter';
import { ResponseInterceptor } from 'src/common/http/interceptors/response.interceptor';
import { AllowAuthGuard } from 'test/utils/allow-auth.guard';
import { WorkflowRepository } from '@lunie/db-access';
import { PrismaService } from '@lunie/db-access';
import { SecretRepository } from '@lunie/db-access';
import {
  createWorkflowRepositoryMock,
  type WorkflowRepositoryMock,
} from 'test/workflow/workflow.repository.mock';
import { createWorkflowFixture } from 'test/workflow/workflow.fixtures';
import { createSecretRepositoryMock } from 'test/secret/secret.repository.mock';
import { createCacheManagerMock } from 'test/utils/cache-manager.mock';

describe('Workflow Run (integration e2e)', () => {
  let app: NestFastifyApplication;
  let repo: WorkflowRepositoryMock;
  let orchestration: { startWorkflow: jest.Mock };

  beforeEach(async () => {
    repo = createWorkflowRepositoryMock();
    orchestration = { startWorkflow: jest.fn() };
    const secretRepo = createSecretRepositoryMock();
    const cache = createCacheManagerMock();

    const moduleRef = await Test.createTestingModule({
      controllers: [WorkflowController],
      providers: [
        WorkflowService,
        { provide: WorkflowRepository, useValue: repo },
        { provide: PrismaService, useValue: { $transaction: jest.fn() } },
        { provide: SecretRepository, useValue: secretRepo },
        { provide: CACHE_MANAGER, useValue: cache },
        { provide: OrchestrationService, useValue: orchestration },

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

  it('POST /workflows/:id/run calls orchestration with MANUAL event', async () => {
    repo.findById.mockResolvedValue(
      createWorkflowFixture({ id: 'wf_1', latestVersionId: 'wfv_1' }),
    );
    orchestration.startWorkflow.mockResolvedValue({
      workflowRunId: 'wfr_1',
      workflowRunNumber: 5,
      stepRunIds: [],
    });

    const res = await app.inject({
      method: 'POST',
      url: '/workflows/wf_1/run',
      payload: {
        input: { hello: 'world' },
        overrides: {
          step_1: {
            body: { content: 'dynamic' },
          },
        },
      },
    });

    expect(res.statusCode).toBe(201);
    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.workflowRunId).toBe('wfr_1');
    expect(body.data.workflowRunNumber).toBe(5);
    expect(body.data.status).toBe('QUEUED');

    expect(orchestration.startWorkflow).toHaveBeenCalledWith(
      expect.objectContaining({
        workflowId: 'wf_1',
        workflowVersionId: 'wfv_1',
        eventType: 'MANUAL',
        input: { hello: 'world' },
        overrides: {
          step_1: {
            body: { content: 'dynamic' },
          },
        },
      }),
    );
  });

  it('POST /workflows/by-key/:workflowKey/run resolves workflow by key', async () => {
    const workflow = createWorkflowFixture({
      id: 'wf_1',
      key: 'deploy-api',
      latestVersionId: 'wfv_1',
    });
    repo.findByKey.mockResolvedValue(workflow);
    repo.findById.mockResolvedValue(workflow);
    orchestration.startWorkflow.mockResolvedValue({
      workflowRunId: 'wfr_2',
      workflowRunNumber: 9,
      stepRunIds: [],
    });

    const res = await app.inject({
      method: 'POST',
      url: '/workflows/by-key/deploy-api/run',
      payload: { input: { hello: 'world' }, overrides: {} },
    });

    expect(res.statusCode).toBe(201);
    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.workflowRunId).toBe('wfr_2');
    expect(body.data.workflowRunNumber).toBe(9);
    expect(orchestration.startWorkflow).toHaveBeenCalledWith(
      expect.objectContaining({
        workflowId: 'wf_1',
        workflowVersionId: 'wfv_1',
        eventType: 'MANUAL',
      }),
    );
  });
});
