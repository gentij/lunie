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

import { AllExceptionsFilter } from 'src/common/http/filters/all-exceptions.filter'; // <-- adjust path to your file

import {
  createWorkflowRepositoryMock,
  type WorkflowRepositoryMock,
} from 'test/workflow/workflow.repository.mock';
import {
  createWorkflowFixture,
  createWorkflowListFixture,
} from 'test/workflow/workflow.fixtures';
import { createWorkflowVersionFixture } from 'test/workflow-version/workflow-version.fixtures';
import { AllowAuthGuard } from 'test/utils/allow-auth.guard';
import { ResponseInterceptor } from 'src/common/http/interceptors/response.interceptor';
import {
  PrismaService,
  SecretRepository,
  WorkflowRepository,
} from '@lunie/db-access';
import { createSecretRepositoryMock } from 'test/secret/secret.repository.mock';
import { createCacheManagerMock } from 'test/utils/cache-manager.mock';

describe('Workflow (e2e)', () => {
  let app: NestFastifyApplication;
  let repo: WorkflowRepositoryMock;
  let prisma: {
    $transaction: jest.Mock<Promise<unknown>, [(tx: any) => unknown]>;
  };

  beforeEach(async () => {
    repo = createWorkflowRepositoryMock();
    prisma = {
      $transaction: jest.fn<Promise<unknown>, [(tx: any) => unknown]>(),
    };
    const secretRepo = createSecretRepositoryMock();
    secretRepo.findManyByNames.mockResolvedValue([]);
    const cache = createCacheManagerMock();

    const moduleRef = await Test.createTestingModule({
      controllers: [WorkflowController],
      providers: [
        WorkflowService,
        {
          provide: OrchestrationService,
          useValue: { startWorkflow: jest.fn() },
        },
        {
          provide: SecretRepository,
          useValue: secretRepo,
        },
        { provide: WorkflowRepository, useValue: repo },
        { provide: PrismaService, useValue: prisma },
        { provide: CACHE_MANAGER, useValue: cache },

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

  it('POST /workflows -> 201 + ok:true + data', async () => {
    const created = createWorkflowFixture({
      id: 'wf_new',
      key: 'my-wf',
      name: 'My WF',
    });
    const version = createWorkflowVersionFixture({
      workflowId: created.id,
      version: 1,
      definition: { steps: [] },
    });
    const updated = createWorkflowFixture({
      id: created.id,
      key: 'my-wf',
      name: 'My WF',
      latestVersionId: version.id,
    });

    const tx = {
      trigger: {
        create: jest.fn().mockResolvedValue({
          id: 'tr_manual',
          workflowId: created.id,
          type: 'MANUAL',
          name: 'Manual',
          isActive: true,
          config: {},
          createdAt: new Date(),
          updatedAt: new Date(),
        }),
      },
      workflow: {
        create: jest.fn().mockResolvedValue(created),
        update: jest.fn().mockResolvedValue(updated),
      },
      workflowVersion: {
        create: jest.fn().mockResolvedValue(version),
        findFirst: jest.fn(),
      },
    };

    repo.findMany.mockResolvedValue([]);
    prisma.$transaction.mockImplementation((cb) => Promise.resolve(cb(tx)));

    const res = await app.inject({
      method: 'POST',
      url: '/workflows',
      payload: { name: 'My WF', definition: version.definition },
    });

    expect(res.statusCode).toBe(201);

    const body = res.json();

    expect(body.ok).toBe(true);
    expect(body.data.key).toBe('my-wf');
    expect(body.data.name).toBe('My WF');
    expect(tx.workflow.create).toHaveBeenCalledWith({
      data: { name: 'My WF', key: 'my-wf' },
    });
  });

  it('GET /workflows -> 200 + data array', async () => {
    const list = createWorkflowListFixture(2);
    repo.findPage.mockResolvedValue({ items: list, total: 2 });

    const res = await app.inject({
      method: 'GET',
      url: '/workflows',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(Array.isArray(body.data.items)).toBe(true);
    expect(body.data.items).toHaveLength(2);
    expect(body.data.pagination.total).toBe(2);

    expect(repo.findPage).toHaveBeenCalledTimes(1);
  });

  it('GET /workflows/:id -> 200 when found', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    repo.findById.mockResolvedValue(wf);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.id).toBe('wf_1');
    expect(body.data.key).toBe('deploy-on-release');
  });

  it('GET /workflows/by-key/:workflowKey -> 200 when found', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1', key: 'deploy-on-release' });
    repo.findByKey.mockResolvedValue(wf);
    repo.findById.mockResolvedValue(wf);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/by-key/deploy-on-release',
    });

    expect(res.statusCode).toBe(200);
    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.id).toBe('wf_1');
    expect(body.data.key).toBe('deploy-on-release');
  });

  it('GET /workflows/:id -> 404 with standardized error payload', async () => {
    repo.findById.mockResolvedValue(null);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/missing',
    });

    expect(res.statusCode).toBe(404);

    const body = res.json();
    expect(body.ok).toBe(false);
    expect(body.error).toBeDefined();
    expect(typeof body.error.code).toBe('string');
  });

  it('PATCH /workflows/:id -> 200 updates workflow', async () => {
    const existing = createWorkflowFixture({ id: 'wf_1', name: 'Old' });
    const updated = createWorkflowFixture({ id: 'wf_1', name: 'New' });

    repo.findById.mockResolvedValue(existing);
    repo.update.mockResolvedValue(updated);

    const res = await app.inject({
      method: 'PATCH',
      url: '/workflows/wf_1',
      payload: { name: 'New' },
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.name).toBe('New');

    expect(repo.update).toHaveBeenCalledWith('wf_1', { name: 'New' });
  });

  it('DELETE /workflows/:id -> 200 soft deletes workflow', async () => {
    const existing = createWorkflowFixture({ id: 'wf_1', isActive: true });
    const deleted = createWorkflowFixture({ id: 'wf_1', isActive: false });

    repo.findById.mockResolvedValue(existing);
    repo.softDelete.mockResolvedValue(deleted);

    const res = await app.inject({
      method: 'DELETE',
      url: '/workflows/wf_1',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.isActive).toBe(false);

    expect(repo.softDelete).toHaveBeenCalledWith('wf_1');
  });

  it('POST /workflows/:id/versions -> 201 creates a new version', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const version = createWorkflowVersionFixture({
      id: 'wfv_2',
      workflowId: 'wf_1',
      version: 2,
      definition: {
        steps: [
          {
            key: 's1',
            type: 'http',
            request: { method: 'GET', url: 'https://example.com' },
          },
        ],
      },
    });

    repo.findById.mockResolvedValue(wf);

    const tx = {
      workflow: {
        create: jest.fn(),
        update: jest.fn().mockResolvedValue(wf),
      },
      workflowVersion: {
        findFirst: jest.fn().mockResolvedValue({ version: 1 }),
        create: jest.fn().mockResolvedValue(version),
      },
    };

    prisma.$transaction.mockImplementation((cb) => Promise.resolve(cb(tx)));

    const res = await app.inject({
      method: 'POST',
      url: '/workflows/wf_1/versions',
      payload: { definition: version.definition },
    });

    expect(res.statusCode).toBe(201);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.version).toBe(2);
    expect(body.data.workflowId).toBe('wf_1');
  });

  it('POST /workflows -> 400 validation error when name missing', async () => {
    const res = await app.inject({
      method: 'POST',
      url: '/workflows',
      payload: {}, // invalid
    });

    expect(res.statusCode).toBe(400);

    const body = res.json();
    expect(body.ok).toBe(false);
    expect(body.error.code).toBe('VALIDATION_ERROR');
    expect(Array.isArray(body.error.details)).toBe(true);
  });
});
