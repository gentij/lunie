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

import { WorkflowVersionController } from 'src/workflow-version/workflow-version.controller';
import { WorkflowVersionKeyController } from 'src/workflow-version/workflow-version-key.controller';
import { WorkflowVersionService } from 'src/workflow-version/workflow-version.service';
import { WorkflowService } from 'src/workflow/workflow.service';

import { AllExceptionsFilter } from 'src/common/http/filters/all-exceptions.filter';
import { ResponseInterceptor } from 'src/common/http/interceptors/response.interceptor';
import { AllowAuthGuard } from 'test/utils/allow-auth.guard';

import {
  createWorkflowVersionRepositoryMock,
  type WorkflowVersionRepositoryMock,
} from 'test/workflow-version/workflow-version.repository.mock';
import {
  createWorkflowVersionFixture,
  createWorkflowVersionListFixture,
} from 'test/workflow-version/workflow-version.fixtures';
import {
  createWorkflowRepositoryMock,
  type WorkflowRepositoryMock,
} from 'test/workflow/workflow.repository.mock';
import { createWorkflowFixture } from 'test/workflow/workflow.fixtures';
import {
  WorkflowRepository,
  WorkflowVersionRepository,
} from '@lunie/db-access';
import { createCacheManagerMock } from 'test/utils/cache-manager.mock';

describe('WorkflowVersion (e2e)', () => {
  let app: NestFastifyApplication;
  let repo: WorkflowVersionRepositoryMock;
  let workflowRepo: WorkflowRepositoryMock;
  let workflowService: { getByKey: jest.Mock };

  beforeEach(async () => {
    repo = createWorkflowVersionRepositoryMock();
    workflowRepo = createWorkflowRepositoryMock();
    workflowService = { getByKey: jest.fn() };

    const moduleRef = await Test.createTestingModule({
      controllers: [WorkflowVersionController, WorkflowVersionKeyController],
      providers: [
        WorkflowVersionService,
        { provide: WorkflowVersionRepository, useValue: repo },
        { provide: WorkflowRepository, useValue: workflowRepo },
        { provide: WorkflowService, useValue: workflowService },
        { provide: CACHE_MANAGER, useValue: createCacheManagerMock() },

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

  it('GET /workflows/:workflowId/versions -> 200 + data array', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const list = createWorkflowVersionListFixture(2);

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findPageByWorkflow.mockResolvedValue({ items: list, total: 2 });

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/versions',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(Array.isArray(body.data.items)).toBe(true);
    expect(body.data.items).toHaveLength(2);
    expect(body.data.pagination.total).toBe(2);
  });

  it('GET /workflows/:workflowId/versions/:version -> 200 when found', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const version = createWorkflowVersionFixture({
      workflowId: 'wf_1',
      version: 1,
    });

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findByWorkflowAndVersion.mockResolvedValue(version);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/versions/1',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.version).toBe(1);
  });

  it('GET /workflows/by-key/:workflowKey/versions/:version -> 200 when found', async () => {
    const workflow = createWorkflowFixture({ id: 'wf_1', key: 'deploy-api' });
    const version = createWorkflowVersionFixture({
      workflowId: 'wf_1',
      version: 1,
    });

    workflowService.getByKey.mockResolvedValue(workflow);
    workflowRepo.findById.mockResolvedValue(workflow);
    repo.findByWorkflowAndVersion.mockResolvedValue(version);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/by-key/deploy-api/versions/1',
    });

    expect(res.statusCode).toBe(200);
    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.version).toBe(1);
  });

  it('GET /workflows/:workflowId/versions -> 404 when workflow missing', async () => {
    workflowRepo.findById.mockResolvedValue(null);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/missing/versions',
    });

    expect(res.statusCode).toBe(404);

    const body = res.json();
    expect(body.ok).toBe(false);
    expect(body.error).toBeDefined();
  });

  it('GET /workflows/:workflowId/versions/:version -> 404 when version missing', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    workflowRepo.findById.mockResolvedValue(wf);
    repo.findByWorkflowAndVersion.mockResolvedValue(null);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/versions/9',
    });

    expect(res.statusCode).toBe(404);

    const body = res.json();
    expect(body.ok).toBe(false);
    expect(body.error).toBeDefined();
  });
});
