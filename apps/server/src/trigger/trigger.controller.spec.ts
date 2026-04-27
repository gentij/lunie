import { Test, TestingModule } from '@nestjs/testing';
import { TriggerController } from './trigger.controller';
import { TriggerService } from './trigger.service';
import { OrchestrationService } from 'src/core/orchestration.service';
import { WorkflowService } from 'src/workflow/workflow.service';
import { createTriggerFixture } from 'test/trigger/trigger.fixtures';

describe('TriggerController', () => {
  let controller: TriggerController;
  let service: TriggerService;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [TriggerController],
      providers: [
        {
          provide: TriggerService,
          useValue: {
            create: jest.fn(),
            list: jest.fn(),
            get: jest.fn(),
            update: jest.fn(),
            delete: jest.fn(),
            rotateWebhookKey: jest.fn(),
          },
        },
        {
          provide: OrchestrationService,
          useValue: {
            startWorkflow: jest.fn(),
          },
        },
        {
          provide: WorkflowService,
          useValue: {
            get: jest.fn(),
          },
        },
      ],
    }).compile();

    controller = module.get<TriggerController>(TriggerController);
    service = module.get<TriggerService>(TriggerService);
  });

  it('create() calls TriggerService.create()', async () => {
    const trigger = createTriggerFixture({
      workflowId: 'wf_1',
      type: 'MANUAL',
    });
    const createSpy = jest.spyOn(service, 'create').mockResolvedValue(trigger);

    await expect(controller.create('wf_1', { type: 'MANUAL' })).resolves.toBe(
      trigger,
    );

    expect(createSpy).toHaveBeenCalledWith({
      workflowId: 'wf_1',
      type: 'MANUAL',
      name: undefined,
      isActive: undefined,
      config: undefined,
    });
  });

  it('list() calls TriggerService.list()', async () => {
    const list = [createTriggerFixture({ id: 'tr_1' })];
    const listSpy = jest.spyOn(service, 'list').mockResolvedValue({
      items: list,
      pagination: {
        page: 1,
        pageSize: 25,
        total: 1,
        totalPages: 1,
        hasNext: false,
        hasPrev: false,
      },
    });

    await expect(
      controller.list('wf_1', { page: 1, pageSize: 25 }),
    ).resolves.toEqual({
      items: list,
      pagination: {
        page: 1,
        pageSize: 25,
        total: 1,
        totalPages: 1,
        hasNext: false,
        hasPrev: false,
      },
    });
    expect(listSpy).toHaveBeenCalledWith({
      workflowId: 'wf_1',
      page: 1,
      pageSize: 25,
    });
  });

  it('get() calls TriggerService.get()', async () => {
    const trigger = createTriggerFixture({ id: 'tr_1', workflowId: 'wf_1' });
    const getSpy = jest.spyOn(service, 'get').mockResolvedValue(trigger);

    await expect(controller.get('wf_1', 'tr_1')).resolves.toBe(trigger);
    expect(getSpy).toHaveBeenCalledWith('wf_1', 'tr_1');
  });

  it('update() calls TriggerService.update()', async () => {
    const trigger = createTriggerFixture({
      id: 'tr_1',
      workflowId: 'wf_1',
      name: 'Updated',
    });
    const updateSpy = jest.spyOn(service, 'update').mockResolvedValue(trigger);

    await expect(
      controller.update('wf_1', 'tr_1', { name: 'Updated' }),
    ).resolves.toBe(trigger);

    expect(updateSpy).toHaveBeenCalledWith('wf_1', 'tr_1', {
      name: 'Updated',
      isActive: undefined,
      config: undefined,
    });
  });

  it('delete() calls TriggerService.delete()', async () => {
    const trigger = createTriggerFixture({ id: 'tr_1', isActive: false });
    const deleteSpy = jest.spyOn(service, 'delete').mockResolvedValue(trigger);

    await expect(controller.delete('wf_1', 'tr_1')).resolves.toBe(trigger);
    expect(deleteSpy).toHaveBeenCalledWith('wf_1', 'tr_1');
  });

  it('rotateWebhookKey() calls TriggerService.rotateWebhookKey()', async () => {
    const rotateSpy = jest
      .spyOn(service, 'rotateWebhookKey')
      .mockResolvedValue({ webhookKey: 'lunie_key' });

    await expect(controller.rotateWebhookKey('wf_1', 'tr_1')).resolves.toEqual({
      webhookKey: 'lunie_key',
    });
    expect(rotateSpy).toHaveBeenCalledWith('wf_1', 'tr_1');
  });
});
