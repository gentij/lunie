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

import { WorkflowRunController } from 'src/workflow-run/workflow-run.controller';
import { WorkflowRunService } from 'src/workflow-run/workflow-run.service';

import { AllExceptionsFilter } from 'src/common/http/filters/all-exceptions.filter';
import { ResponseInterceptor } from 'src/common/http/interceptors/response.interceptor';
import { AllowAuthGuard } from 'test/utils/allow-auth.guard';

import {
  createWorkflowRunRepositoryMock,
  type WorkflowRunRepositoryMock,
} from 'test/workflow-run/workflow-run.repository.mock';
import {
  createWorkflowRunFixture,
  createWorkflowRunListFixture,
} from 'test/workflow-run/workflow-run.fixtures';
import {
  createWorkflowRepositoryMock,
  type WorkflowRepositoryMock,
} from 'test/workflow/workflow.repository.mock';
import { createWorkflowFixture } from 'test/workflow/workflow.fixtures';
import { WorkflowRepository, WorkflowRunRepository } from '@lunie/db-access';

describe('WorkflowRun (e2e)', () => {
  let app: NestFastifyApplication;
  let repo: WorkflowRunRepositoryMock;
  let workflowRepo: WorkflowRepositoryMock;

  beforeEach(async () => {
    repo = createWorkflowRunRepositoryMock();
    workflowRepo = createWorkflowRepositoryMock();

    const moduleRef = await Test.createTestingModule({
      controllers: [WorkflowRunController],
      providers: [
        WorkflowRunService,
        { provide: WorkflowRunRepository, useValue: repo },
        { provide: WorkflowRepository, useValue: workflowRepo },

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
    await app.close();
  });

  it('GET /workflows/:workflowId/runs -> 200 + data array', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const list = createWorkflowRunListFixture(2);

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findPageByWorkflow.mockResolvedValue({ items: list, total: 2 });

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/runs',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(Array.isArray(body.data.items)).toBe(true);
    expect(body.data.items).toHaveLength(2);
    expect(body.data.pagination.total).toBe(2);
  });

  it('GET /workflows/:workflowId/runs/:id -> 200 when found', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const run = createWorkflowRunFixture({
      id: 'wfr_1',
      workflowId: 'wf_1',
      number: 42,
    });

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findById.mockResolvedValue(run);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/runs/wfr_1',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.id).toBe('wfr_1');
    expect(body.data.number).toBe(42);
  });

  it('GET /workflows/:workflowId/runs/:id -> 404 when missing', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    workflowRepo.findById.mockResolvedValue(wf);
    repo.findById.mockResolvedValue(null);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/runs/missing',
    });

    expect(res.statusCode).toBe(404);

    const body = res.json();
    expect(body.ok).toBe(false);
    expect(body.error).toBeDefined();
  });
});
