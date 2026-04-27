import { Test } from '@nestjs/testing';
import { EventService } from './event.service';
import {
  EventRepository,
  WorkflowRepository,
  TriggerRepository,
} from '@lunie/db-access';
import {
  createEventRepositoryMock,
  type EventRepositoryMock,
} from 'test/event/event.repository.mock';
import {
  createWorkflowRepositoryMock,
  type WorkflowRepositoryMock,
} from 'test/workflow/workflow.repository.mock';
import {
  createTriggerRepositoryMock,
  type TriggerRepositoryMock,
} from 'test/trigger/trigger.repository.mock';
import { createWorkflowFixture } from 'test/workflow/workflow.fixtures';
import { createTriggerFixture } from 'test/trigger/trigger.fixtures';
import {
  createEventFixture,
  createEventListFixture,
} from 'test/event/event.fixtures';
import { AppError } from 'src/common/http/errors/app-error';

describe('EventService', () => {
  let service: EventService;
  let repo: EventRepositoryMock;
  let workflowRepo: WorkflowRepositoryMock;
  let triggerRepo: TriggerRepositoryMock;

  beforeEach(async () => {
    repo = createEventRepositoryMock();
    workflowRepo = createWorkflowRepositoryMock();
    triggerRepo = createTriggerRepositoryMock();

    const moduleRef = await Test.createTestingModule({
      providers: [
        EventService,
        { provide: EventRepository, useValue: repo },
        { provide: WorkflowRepository, useValue: workflowRepo },
        { provide: TriggerRepository, useValue: triggerRepo },
      ],
    }).compile();

    service = moduleRef.get(EventService);
  });

  it('create() creates an event for trigger', async () => {
    const created = createEventFixture({
      triggerId: 'tr_1',
      type: 'WEBHOOK',
      externalId: 'ext_1',
    });

    repo.create.mockResolvedValue(created);

    await expect(
      service.create({
        triggerId: 'tr_1',
        type: 'WEBHOOK',
        externalId: 'ext_1',
        payload: { hello: 'world' },
      }),
    ).resolves.toBe(created);

    expect(repo.create).toHaveBeenCalledWith({
      trigger: { connect: { id: 'tr_1' } },
      type: 'WEBHOOK',
      externalId: 'ext_1',
      payload: { hello: 'world' },
      receivedAt: undefined,
    });
  });

  it('list() returns events for trigger', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const trigger = createTriggerFixture({ id: 'tr_1', workflowId: 'wf_1' });
    const list = createEventListFixture(2);

    workflowRepo.findById.mockResolvedValue(wf);
    triggerRepo.findById.mockResolvedValue(trigger);
    repo.findPageByTrigger.mockResolvedValue({ items: list, total: 2 });

    await expect(
      service.list({
        workflowId: 'wf_1',
        triggerId: 'tr_1',
        page: 1,
        pageSize: 25,
      }),
    ).resolves.toEqual({
      items: list,
      pagination: {
        page: 1,
        pageSize: 25,
        total: 2,
        totalPages: 1,
        hasNext: false,
        hasPrev: false,
      },
    });
    expect(repo.findPageByTrigger).toHaveBeenCalledWith({
      triggerId: 'tr_1',
      page: 1,
      pageSize: 25,
    });
  });

  it('list() throws notFound when workflow missing', async () => {
    workflowRepo.findById.mockResolvedValue(null);

    await expect(
      service.list({
        workflowId: 'missing',
        triggerId: 'tr_1',
        page: 1,
        pageSize: 25,
      }),
    ).rejects.toBeInstanceOf(AppError);
  });

  it('list() throws notFound when trigger belongs to another workflow', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const trigger = createTriggerFixture({ id: 'tr_1', workflowId: 'wf_2' });

    workflowRepo.findById.mockResolvedValue(wf);
    triggerRepo.findById.mockResolvedValue(trigger);

    await expect(
      service.list({
        workflowId: 'wf_1',
        triggerId: 'tr_1',
        page: 1,
        pageSize: 25,
      }),
    ).rejects.toBeInstanceOf(AppError);
  });

  it('get() returns event when found', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const trigger = createTriggerFixture({ id: 'tr_1', workflowId: 'wf_1' });
    const event = createEventFixture({ id: 'ev_1', triggerId: 'tr_1' });

    workflowRepo.findById.mockResolvedValue(wf);
    triggerRepo.findById.mockResolvedValue(trigger);
    repo.findById.mockResolvedValue(event);

    await expect(service.get('wf_1', 'tr_1', 'ev_1')).resolves.toBe(event);
    expect(repo.findById).toHaveBeenCalledWith('ev_1');
  });

  it('get() throws notFound when event missing', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const trigger = createTriggerFixture({ id: 'tr_1', workflowId: 'wf_1' });

    workflowRepo.findById.mockResolvedValue(wf);
    triggerRepo.findById.mockResolvedValue(trigger);
    repo.findById.mockResolvedValue(null);

    await expect(service.get('wf_1', 'tr_1', 'missing')).rejects.toBeInstanceOf(
      AppError,
    );
  });

  it('get() throws notFound when event belongs to another trigger', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const trigger = createTriggerFixture({ id: 'tr_1', workflowId: 'wf_1' });
    const event = createEventFixture({ id: 'ev_1', triggerId: 'tr_2' });

    workflowRepo.findById.mockResolvedValue(wf);
    triggerRepo.findById.mockResolvedValue(trigger);
    repo.findById.mockResolvedValue(event);

    await expect(service.get('wf_1', 'tr_1', 'ev_1')).rejects.toBeInstanceOf(
      AppError,
    );
  });
});
