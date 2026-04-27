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

import { EventController } from 'src/event/event.controller';
import { EventService } from 'src/event/event.service';

import { AllExceptionsFilter } from 'src/common/http/filters/all-exceptions.filter';
import { ResponseInterceptor } from 'src/common/http/interceptors/response.interceptor';
import { AllowAuthGuard } from 'test/utils/allow-auth.guard';

import {
  createEventRepositoryMock,
  type EventRepositoryMock,
} from 'test/event/event.repository.mock';
import {
  createEventFixture,
  createEventListFixture,
} from 'test/event/event.fixtures';
import {
  createWorkflowRepositoryMock,
  type WorkflowRepositoryMock,
} from 'test/workflow/workflow.repository.mock';
import { createWorkflowFixture } from 'test/workflow/workflow.fixtures';
import {
  createTriggerRepositoryMock,
  type TriggerRepositoryMock,
} from 'test/trigger/trigger.repository.mock';
import { createTriggerFixture } from 'test/trigger/trigger.fixtures';
import {
  EventRepository,
  TriggerRepository,
  WorkflowRepository,
} from '@lunie/db-access';

describe('Event (e2e)', () => {
  let app: NestFastifyApplication;
  let repo: EventRepositoryMock;
  let workflowRepo: WorkflowRepositoryMock;
  let triggerRepo: TriggerRepositoryMock;

  beforeEach(async () => {
    repo = createEventRepositoryMock();
    workflowRepo = createWorkflowRepositoryMock();
    triggerRepo = createTriggerRepositoryMock();

    const moduleRef = await Test.createTestingModule({
      controllers: [EventController],
      providers: [
        EventService,
        { provide: EventRepository, useValue: repo },
        { provide: WorkflowRepository, useValue: workflowRepo },
        { provide: TriggerRepository, useValue: triggerRepo },

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

  it('GET /workflows/:workflowId/triggers/:triggerId/events -> 200 + data array', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const trigger = createTriggerFixture({ id: 'tr_1', workflowId: 'wf_1' });
    const list = createEventListFixture(2);

    workflowRepo.findById.mockResolvedValue(wf);
    triggerRepo.findById.mockResolvedValue(trigger);
    repo.findPageByTrigger.mockResolvedValue({ items: list, total: 2 });

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/triggers/tr_1/events',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(Array.isArray(body.data.items)).toBe(true);
    expect(body.data.items).toHaveLength(2);
    expect(body.data.pagination.total).toBe(2);
  });

  it('GET /workflows/:workflowId/triggers/:triggerId/events/:id -> 200 when found', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const trigger = createTriggerFixture({ id: 'tr_1', workflowId: 'wf_1' });
    const event = createEventFixture({ id: 'ev_1', triggerId: 'tr_1' });

    workflowRepo.findById.mockResolvedValue(wf);
    triggerRepo.findById.mockResolvedValue(trigger);
    repo.findById.mockResolvedValue(event);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/triggers/tr_1/events/ev_1',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.id).toBe('ev_1');
  });

  it('GET /workflows/:workflowId/triggers/:triggerId/events/:id -> 404 when missing', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const trigger = createTriggerFixture({ id: 'tr_1', workflowId: 'wf_1' });

    workflowRepo.findById.mockResolvedValue(wf);
    triggerRepo.findById.mockResolvedValue(trigger);
    repo.findById.mockResolvedValue(null);

    const res = await app.inject({
      method: 'GET',
      url: '/workflows/wf_1/triggers/tr_1/events/missing',
    });

    expect(res.statusCode).toBe(404);

    const body = res.json();
    expect(body.ok).toBe(false);
    expect(body.error).toBeDefined();
  });
});
