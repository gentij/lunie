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

import { TriggerController } from 'src/trigger/trigger.controller';
import { TriggerService } from 'src/trigger/trigger.service';
import { OrchestrationService } from 'src/core/orchestration.service';
import { WorkflowService } from 'src/workflow/workflow.service';

import { AllExceptionsFilter } from 'src/common/http/filters/all-exceptions.filter';
import { ResponseInterceptor } from 'src/common/http/interceptors/response.interceptor';
import { AllowAuthGuard } from 'test/utils/allow-auth.guard';

import {
  createTriggerRepositoryMock,
  type TriggerRepositoryMock,
} from 'test/trigger/trigger.repository.mock';
import {
  createTriggerFixture,
  createTriggerListFixture,
} from 'test/trigger/trigger.fixtures';
import {
  createWorkflowRepositoryMock,
  type WorkflowRepositoryMock,
} from 'test/workflow/workflow.repository.mock';
import { createWorkflowFixture } from 'test/workflow/workflow.fixtures';
import { TriggerRepository, WorkflowRepository } from '@lunie/db-access';
import { CryptoService } from 'src/crypto/crypto.service';

describe('Trigger (e2e)', () => {
  let app: NestFastifyApplication;
  let repo: TriggerRepositoryMock;
  let workflowRepo: WorkflowRepositoryMock;

  beforeEach(async () => {
    repo = createTriggerRepositoryMock();
    workflowRepo = createWorkflowRepositoryMock();

    const moduleRef = await Test.createTestingModule({
      controllers: [TriggerController],
      providers: [
        TriggerService,
        {
          provide: OrchestrationService,
          useValue: { startWorkflow: jest.fn() },
        },
        {
          provide: WorkflowService,
          useValue: { get: jest.fn() },
        },
        {
          provide: CryptoService,
          useValue: {
            generateApiToken: jest.fn(),
            hashApiToken: jest.fn((value: string) => `hash:${value}`),
            secureCompare: jest.fn((a: string, b: string) => a === b),
          },
        },
        { provide: TriggerRepository, useValue: repo },
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
    await app?.close();
  });

  it('POST /workflows/:workflowId/triggers -> 201 creates trigger', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const created = createTriggerFixture({
      workflowId: 'wf_1',
      type: 'WEBHOOK',
      config: { url: 'https://example.com' },
    });

    workflowRepo.findById.mockResolvedValue(wf);
    repo.create.mockResolvedValue(created);

    const res = await app.inject({
      method: 'POST',
      url: '/workflows/wf_1/triggers',
      payload: { type: 'WEBHOOK', config: { url: 'https://example.com' } },
    });

    expect(res.statusCode).toBe(201);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.type).toBe('WEBHOOK');
    expect(body.data.workflowId).toBe('wf_1');
  });

  it('GET /workflows/:workflowId/triggers -> 200 + data array', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const list = createTriggerListFixture(2);

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findPageByWorkflow.mockResolvedValue({ items: list, total: 2 });

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/triggers',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(Array.isArray(body.data.items)).toBe(true);
    expect(body.data.items).toHaveLength(2);
    expect(body.data.pagination.total).toBe(2);
  });

  it('GET /workflows/:workflowId/triggers/:id -> 200 when found', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const trigger = createTriggerFixture({ id: 'tr_1', workflowId: 'wf_1' });

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findById.mockResolvedValue(trigger);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/triggers/tr_1',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.id).toBe('tr_1');
  });

  it('PATCH /workflows/:workflowId/triggers/:id -> 200 updates trigger', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const trigger = createTriggerFixture({ id: 'tr_1', workflowId: 'wf_1' });
    const updated = createTriggerFixture({
      id: 'tr_1',
      workflowId: 'wf_1',
      name: 'Updated',
    });

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findById.mockResolvedValue(trigger);
    repo.update.mockResolvedValue(updated);

    const res = await app.inject({
      method: 'PATCH',
      url: '/workflows/wf_1/triggers/tr_1',
      payload: { name: 'Updated' },
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.name).toBe('Updated');
  });

  it('DELETE /workflows/:workflowId/triggers/:id -> 200 soft deletes trigger', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const trigger = createTriggerFixture({
      id: 'tr_1',
      workflowId: 'wf_1',
      isActive: true,
    });
    const deleted = createTriggerFixture({
      id: 'tr_1',
      workflowId: 'wf_1',
      isActive: false,
    });

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findById.mockResolvedValue(trigger);
    repo.softDelete.mockResolvedValue(deleted);

    const res = await app.inject({
      method: 'DELETE',
      url: '/workflows/wf_1/triggers/tr_1',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.isActive).toBe(false);
  });

  it('GET /workflows/:workflowId/triggers/:id -> 404 when missing', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    workflowRepo.findById.mockResolvedValue(wf);
    repo.findById.mockResolvedValue(null);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/triggers/missing',
    });

    expect(res.statusCode).toBe(404);

    const body = res.json();
    expect(body.ok).toBe(false);
    expect(body.error).toBeDefined();
  });
});
