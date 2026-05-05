import { Test } from '@nestjs/testing';
import { WorkflowRunService } from './workflow-run.service';
import { WorkflowRunRepository, WorkflowRepository } from '@lunie/db-access';
import {
  createWorkflowRunRepositoryMock,
  type WorkflowRunRepositoryMock,
} from 'test/workflow-run/workflow-run.repository.mock';
import {
  createWorkflowRepositoryMock,
  type WorkflowRepositoryMock,
} from 'test/workflow/workflow.repository.mock';
import { createWorkflowFixture } from 'test/workflow/workflow.fixtures';
import {
  createWorkflowRunFixture,
  createWorkflowRunListFixture,
} from 'test/workflow-run/workflow-run.fixtures';
import { AppError } from 'src/common/http/errors/app-error';

describe('WorkflowRunService', () => {
  let service: WorkflowRunService;
  let repo: WorkflowRunRepositoryMock;
  let workflowRepo: WorkflowRepositoryMock;

  beforeEach(async () => {
    repo = createWorkflowRunRepositoryMock();
    workflowRepo = createWorkflowRepositoryMock();

    const moduleRef = await Test.createTestingModule({
      providers: [
        WorkflowRunService,
        { provide: WorkflowRunRepository, useValue: repo },
        { provide: WorkflowRepository, useValue: workflowRepo },
      ],
    }).compile();

    service = moduleRef.get(WorkflowRunService);
  });

  it('create() creates a workflow run', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const created = createWorkflowRunFixture({
      workflowId: 'wf_1',
      workflowVersionId: 'wfv_1',
    });

    workflowRepo.findById.mockResolvedValue(wf);
    repo.create.mockResolvedValue(created);

    await expect(
      service.create({
        workflowId: 'wf_1',
        workflowVersionId: 'wfv_1',
        number: 1,
        status: 'QUEUED',
        input: { foo: 'bar' },
      }),
    ).resolves.toBe(created);

    expect(repo.create).toHaveBeenCalledWith({
      workflow: { connect: { id: 'wf_1' } },
      workflowVersion: { connect: { id: 'wfv_1' } },
      number: 1,
      trigger: undefined,
      event: undefined,
      status: 'QUEUED',
      input: { foo: 'bar' },
      output: undefined,
      startedAt: undefined,
      finishedAt: undefined,
    });
  });

  it('create() throws notFound when workflow missing', async () => {
    workflowRepo.findById.mockResolvedValue(null);

    await expect(
      service.create({
        workflowId: 'missing',
        workflowVersionId: 'wfv_1',
        number: 1,
      }),
    ).rejects.toBeInstanceOf(AppError);
  });

  it('list() returns runs for workflow', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const list = createWorkflowRunListFixture(2);

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findPageByWorkflow.mockResolvedValue({ items: list, total: 2 });

    await expect(
      service.list({ workflowId: 'wf_1', page: 1, pageSize: 25 }),
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
    expect(repo.findPageByWorkflow).toHaveBeenCalledWith({
      workflowId: 'wf_1',
      page: 1,
      pageSize: 25,
    });
  });

  it('get() returns run when found', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const run = createWorkflowRunFixture({ id: 'wfr_1', workflowId: 'wf_1' });

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findById.mockResolvedValue(run);

    await expect(service.get('wf_1', 'wfr_1')).resolves.toBe(run);
    expect(repo.findById).toHaveBeenCalledWith('wfr_1');
  });

  it('get() throws notFound when run missing', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    workflowRepo.findById.mockResolvedValue(wf);
    repo.findById.mockResolvedValue(null);

    await expect(service.get('wf_1', 'missing')).rejects.toBeInstanceOf(
      AppError,
    );
  });

  it('get() throws notFound when run belongs to another workflow', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const run = createWorkflowRunFixture({ id: 'wfr_1', workflowId: 'wf_2' });

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findById.mockResolvedValue(run);

    await expect(service.get('wf_1', 'wfr_1')).rejects.toBeInstanceOf(AppError);
  });

  it('update() updates run after existence check', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const run = createWorkflowRunFixture({ id: 'wfr_1', workflowId: 'wf_1' });
    const updated = createWorkflowRunFixture({
      id: 'wfr_1',
      workflowId: 'wf_1',
      status: 'SUCCEEDED',
    });

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findById.mockResolvedValue(run);
    repo.update.mockResolvedValue(updated);

    await expect(
      service.update('wf_1', 'wfr_1', { status: 'SUCCEEDED' }),
    ).resolves.toBe(updated);

    expect(repo.update).toHaveBeenCalledWith('wfr_1', { status: 'SUCCEEDED' });
  });
});
