import { Test, TestingModule } from '@nestjs/testing';
import { WorkflowController } from './workflow.controller';
import { WorkflowService } from './workflow.service';
import { OrchestrationService } from 'src/core/orchestration.service';
import { createWorkflowFixture } from 'test/workflow/workflow.fixtures';
import { createWorkflowVersionFixture } from 'test/workflow-version/workflow-version.fixtures';

describe('WorkflowController', () => {
  let controller: WorkflowController;
  let service: WorkflowService;
  let orchestrationService: OrchestrationService;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [WorkflowController],
      providers: [
        {
          provide: WorkflowService,
          useValue: {
            create: jest.fn(),
            list: jest.fn(),
            get: jest.fn(),
            update: jest.fn(),
            delete: jest.fn(),
            createVersion: jest.fn(),
            validateDefinition: jest.fn(),
          },
        },
        {
          provide: OrchestrationService,
          useValue: {
            startWorkflow: jest.fn(),
          },
        },
      ],
    }).compile();

    controller = module.get<WorkflowController>(WorkflowController);
    service = module.get<WorkflowService>(WorkflowService);
    orchestrationService =
      module.get<OrchestrationService>(OrchestrationService);
  });

  it('create() calls WorkflowService.create()', async () => {
    const wf = createWorkflowFixture({ name: 'My WF' });
    const createSpy = jest.spyOn(service, 'create').mockResolvedValue(wf);

    await expect(
      controller.create({
        name: 'My WF',
        definition: { steps: [] },
      }),
    ).resolves.toBe(wf);
    expect(createSpy).toHaveBeenCalledWith({
      name: 'My WF',
      definition: { steps: [] },
    });
  });

  it('list() calls WorkflowService.list()', async () => {
    const list = [createWorkflowFixture({ id: 'wf_1' })];
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

    await expect(controller.list({ page: 1, pageSize: 25 })).resolves.toEqual({
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
    expect(listSpy).toHaveBeenCalledWith({ page: 1, pageSize: 25 });
  });

  it('get() calls WorkflowService.get()', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const getSpy = jest.spyOn(service, 'get').mockResolvedValue(wf);

    await expect(controller.get('wf_1')).resolves.toBe(wf);
    expect(getSpy).toHaveBeenCalledWith('wf_1');
  });

  it('update() calls WorkflowService.update()', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1', name: 'New' });
    const updateSpy = jest.spyOn(service, 'update').mockResolvedValue(wf);

    await expect(controller.update('wf_1', { name: 'New' })).resolves.toBe(wf);
    expect(updateSpy).toHaveBeenCalledWith('wf_1', { name: 'New' });
  });

  it('delete() calls WorkflowService.delete()', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1', isActive: false });
    const deleteSpy = jest.spyOn(service, 'delete').mockResolvedValue(wf);

    await expect(controller.delete('wf_1')).resolves.toBe(wf);
    expect(deleteSpy).toHaveBeenCalledWith('wf_1');
  });

  it('createVersion() calls WorkflowService.createVersion()', async () => {
    const version = createWorkflowVersionFixture({
      workflowId: 'wf_1',
      version: 2,
    });
    const createVersionSpy = jest
      .spyOn(service, 'createVersion')
      .mockResolvedValue(version);

    await expect(
      controller.createVersion('wf_1', { definition: { steps: [] } }),
    ).resolves.toBe(version);
    expect(createVersionSpy).toHaveBeenCalledWith('wf_1', {
      steps: [],
    });
  });

  it('validateVersionDefinition() calls WorkflowService.validateDefinition() and WorkflowService.get()', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    jest.spyOn(service, 'get').mockResolvedValue(wf);

    const validateSpy = jest
      .spyOn(service as any, 'validateDefinition')
      .mockReturnValue({
        valid: true,
        issues: [],
        inferredDependencies: {},
        executionBatches: [['a']],
        referencedSecrets: [],
      });

    await expect(
      controller.validateVersionDefinition('wf_1', {
        definition: { steps: [] },
      }),
    ).resolves.toEqual({
      valid: true,
      issues: [],
      inferredDependencies: {},
      executionBatches: [['a']],
      referencedSecrets: [],
    });

    expect(validateSpy).toHaveBeenCalledWith({ steps: [] });
  });

  it('runManual() calls OrchestrationService.startWorkflow() with input + overrides', async () => {
    const wf = createWorkflowFixture({
      id: 'wf_1',
      latestVersionId: 'wfv_1',
    });
    jest.spyOn(service, 'get').mockResolvedValue(wf);

    const startSpy = jest
      .spyOn(orchestrationService, 'startWorkflow')
      .mockResolvedValue({
        workflowRunId: 'wfr_1',
        workflowRunNumber: 7,
        stepRunIds: [],
      });

    await expect(
      controller.runManual('wf_1', {
        input: { hello: 'world' },
        overrides: {
          step_1: { body: { content: 'dynamic' } },
        },
      }),
    ).resolves.toEqual({
      workflowRunId: 'wfr_1',
      workflowRunNumber: 7,
      status: 'QUEUED',
    });

    expect(startSpy).toHaveBeenCalledWith({
      workflowId: 'wf_1',
      workflowVersionId: 'wfv_1',
      eventType: 'MANUAL',
      input: { hello: 'world' },
      overrides: {
        step_1: { body: { content: 'dynamic' } },
      },
    });
  });
});
