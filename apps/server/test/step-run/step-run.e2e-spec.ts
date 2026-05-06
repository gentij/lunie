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

import { StepRunController } from 'src/step-run/step-run.controller';
import { StepRunKeyController } from 'src/step-run/step-run-key.controller';
import { StepRunService } from 'src/step-run/step-run.service';
import { WorkflowRunService } from 'src/workflow-run/workflow-run.service';
import { WorkflowService } from 'src/workflow/workflow.service';

import { AllExceptionsFilter } from 'src/common/http/filters/all-exceptions.filter';
import { ResponseInterceptor } from 'src/common/http/interceptors/response.interceptor';
import { AllowAuthGuard } from 'test/utils/allow-auth.guard';

import {
  createStepRunRepositoryMock,
  type StepRunRepositoryMock,
} from 'test/step-run/step-run.repository.mock';
import {
  createStepRunFixture,
  createStepRunListFixture,
} from 'test/step-run/step-run.fixtures';
import {
  createWorkflowRunRepositoryMock,
  type WorkflowRunRepositoryMock,
} from 'test/workflow-run/workflow-run.repository.mock';
import { createWorkflowRunFixture } from 'test/workflow-run/workflow-run.fixtures';
import {
  StepRunRepository,
  WorkflowRepository,
  WorkflowRunRepository,
} from '@lunie/db-access';
import {
  createWorkflowRepositoryMock,
  type WorkflowRepositoryMock,
} from 'test/workflow/workflow.repository.mock';
import { createWorkflowFixture } from 'test/workflow/workflow.fixtures';

describe('StepRun (e2e)', () => {
  let app: NestFastifyApplication;
  let repo: StepRunRepositoryMock;
  let runRepo: WorkflowRunRepositoryMock;
  let workflowRepo: WorkflowRepositoryMock;
  let workflowService: { getByKey: jest.Mock };

  beforeEach(async () => {
    repo = createStepRunRepositoryMock();
    runRepo = createWorkflowRunRepositoryMock();
    workflowRepo = createWorkflowRepositoryMock();
    workflowService = { getByKey: jest.fn() };

    const moduleRef = await Test.createTestingModule({
      controllers: [StepRunController, StepRunKeyController],
      providers: [
        StepRunService,
        WorkflowRunService,
        { provide: StepRunRepository, useValue: repo },
        { provide: WorkflowRunRepository, useValue: runRepo },
        { provide: WorkflowRepository, useValue: workflowRepo },
        { provide: WorkflowService, useValue: workflowService },

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

  it('GET /workflows/:workflowId/runs/:runId/steps -> 200 + data array', async () => {
    const run = createWorkflowRunFixture({ id: 'wfr_1' });
    const list = createStepRunListFixture(2);

    runRepo.findById.mockResolvedValue(run);
    repo.findPageByWorkflowRun.mockResolvedValue({ items: list, total: 2 });

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/runs/wfr_1/steps',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(Array.isArray(body.data.items)).toBe(true);
    expect(body.data.items).toHaveLength(2);
    expect(body.data.pagination.total).toBe(2);
  });

  it('GET /workflows/:workflowId/runs/:runId/steps/:id -> 200 when found', async () => {
    const run = createWorkflowRunFixture({ id: 'wfr_1' });
    const step = createStepRunFixture({ id: 'sr_1', workflowRunId: 'wfr_1' });

    runRepo.findById.mockResolvedValue(run);
    repo.findById.mockResolvedValue(step);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/runs/wfr_1/steps/sr_1',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.id).toBe('sr_1');
  });

  it('GET /workflows/by-key/:workflowKey/runs/:runNumber/steps/:stepKey -> 200 when found', async () => {
    const workflow = createWorkflowFixture({ id: 'wf_1', key: 'deploy-api' });
    const run = createWorkflowRunFixture({
      id: 'wfr_1',
      workflowId: 'wf_1',
      number: 42,
    });
    const step = createStepRunFixture({
      id: 'sr_1',
      workflowRunId: 'wfr_1',
      stepKey: 'fetch_post',
    });

    workflowService.getByKey.mockResolvedValue(workflow);
    workflowRepo.findById.mockResolvedValue(workflow);
    runRepo.findByWorkflowAndNumber.mockResolvedValue(run);
    runRepo.findById.mockResolvedValue(run);
    repo.findByWorkflowRunAndStepKey.mockResolvedValue(step);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/by-key/deploy-api/runs/42/steps/fetch_post',
    });

    expect(res.statusCode).toBe(200);
    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.id).toBe('sr_1');
    expect(body.data.stepKey).toBe('fetch_post');
  });

  it('GET /workflows/:workflowId/runs/:runId/steps/:id -> 404 when missing', async () => {
    const run = createWorkflowRunFixture({ id: 'wfr_1' });
    runRepo.findById.mockResolvedValue(run);
    repo.findById.mockResolvedValue(null);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/runs/wfr_1/steps/missing',
    });

    expect(res.statusCode).toBe(404);

    const body = res.json();
    expect(body.ok).toBe(false);
    expect(body.error).toBeDefined();
  });
});
