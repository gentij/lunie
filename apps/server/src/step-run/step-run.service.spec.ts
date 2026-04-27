import { Test } from '@nestjs/testing';
import { StepRunService } from './step-run.service';
import { StepRunRepository, WorkflowRunRepository } from '@lunie/db-access';
import {
  createStepRunRepositoryMock,
  type StepRunRepositoryMock,
} from 'test/step-run/step-run.repository.mock';
import {
  createWorkflowRunRepositoryMock,
  type WorkflowRunRepositoryMock,
} from 'test/workflow-run/workflow-run.repository.mock';
import { createWorkflowRunFixture } from 'test/workflow-run/workflow-run.fixtures';
import {
  createStepRunFixture,
  createStepRunListFixture,
} from 'test/step-run/step-run.fixtures';
import { AppError } from 'src/common/http/errors/app-error';

describe('StepRunService', () => {
  let service: StepRunService;
  let repo: StepRunRepositoryMock;
  let runRepo: WorkflowRunRepositoryMock;

  beforeEach(async () => {
    repo = createStepRunRepositoryMock();
    runRepo = createWorkflowRunRepositoryMock();

    const moduleRef = await Test.createTestingModule({
      providers: [
        StepRunService,
        { provide: StepRunRepository, useValue: repo },
        { provide: WorkflowRunRepository, useValue: runRepo },
      ],
    }).compile();

    service = moduleRef.get(StepRunService);
  });

  it('create() creates a step run', async () => {
    const run = createWorkflowRunFixture({ id: 'wfr_1' });
    const created = createStepRunFixture({
      workflowRunId: 'wfr_1',
      stepKey: 'step_1',
    });

    runRepo.findById.mockResolvedValue(run);
    repo.create.mockResolvedValue(created);

    await expect(
      service.create({
        workflowRunId: 'wfr_1',
        stepKey: 'step_1',
        status: 'QUEUED',
        input: { foo: 'bar' },
      }),
    ).resolves.toBe(created);

    expect(repo.create).toHaveBeenCalledWith({
      workflowRun: { connect: { id: 'wfr_1' } },
      stepKey: 'step_1',
      status: 'QUEUED',
      attempt: 0,
      input: { foo: 'bar' },
      output: undefined,
      error: undefined,
      logs: undefined,
      lastErrorAt: undefined,
      durationMs: undefined,
      startedAt: undefined,
      finishedAt: undefined,
    });
  });

  it('create() throws notFound when workflow run missing', async () => {
    runRepo.findById.mockResolvedValue(null);

    await expect(
      service.create({ workflowRunId: 'missing', stepKey: 'step_1' }),
    ).rejects.toBeInstanceOf(AppError);
  });

  it('list() returns step runs for workflow run', async () => {
    const run = createWorkflowRunFixture({ id: 'wfr_1' });
    const list = createStepRunListFixture(2);

    runRepo.findById.mockResolvedValue(run);
    repo.findPageByWorkflowRun.mockResolvedValue({ items: list, total: 2 });

    await expect(
      service.list({ workflowRunId: 'wfr_1', page: 1, pageSize: 25 }),
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
    expect(repo.findPageByWorkflowRun).toHaveBeenCalledWith({
      workflowRunId: 'wfr_1',
      page: 1,
      pageSize: 25,
    });
  });

  it('get() returns step run when found', async () => {
    const run = createWorkflowRunFixture({ id: 'wfr_1' });
    const step = createStepRunFixture({ id: 'sr_1', workflowRunId: 'wfr_1' });

    runRepo.findById.mockResolvedValue(run);
    repo.findById.mockResolvedValue(step);

    await expect(service.get('wfr_1', 'sr_1')).resolves.toBe(step);
    expect(repo.findById).toHaveBeenCalledWith('sr_1');
  });

  it('get() throws notFound when step run missing', async () => {
    const run = createWorkflowRunFixture({ id: 'wfr_1' });
    runRepo.findById.mockResolvedValue(run);
    repo.findById.mockResolvedValue(null);

    await expect(service.get('wfr_1', 'missing')).rejects.toBeInstanceOf(
      AppError,
    );
  });

  it('get() throws notFound when step run belongs to another run', async () => {
    const run = createWorkflowRunFixture({ id: 'wfr_1' });
    const step = createStepRunFixture({ id: 'sr_1', workflowRunId: 'wfr_2' });

    runRepo.findById.mockResolvedValue(run);
    repo.findById.mockResolvedValue(step);

    await expect(service.get('wfr_1', 'sr_1')).rejects.toBeInstanceOf(AppError);
  });

  it('update() updates step run after existence check', async () => {
    const run = createWorkflowRunFixture({ id: 'wfr_1' });
    const step = createStepRunFixture({ id: 'sr_1', workflowRunId: 'wfr_1' });
    const updated = createStepRunFixture({
      id: 'sr_1',
      workflowRunId: 'wfr_1',
      status: 'SUCCEEDED',
    });

    runRepo.findById.mockResolvedValue(run);
    repo.findById.mockResolvedValue(step);
    repo.update.mockResolvedValue(updated);

    await expect(
      service.update('wfr_1', 'sr_1', { status: 'SUCCEEDED' }),
    ).resolves.toBe(updated);

    expect(repo.update).toHaveBeenCalledWith('sr_1', { status: 'SUCCEEDED' });
  });
});
