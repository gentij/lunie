import { Test } from '@nestjs/testing';
import { WorkflowVersionService } from './workflow-version.service';
import {
  WorkflowVersionRepository,
  WorkflowRepository,
} from '@lunie/db-access';
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
import { AppError } from 'src/common/http/errors/app-error';
import { CACHE_MANAGER } from '@nestjs/cache-manager';

describe('WorkflowVersionService', () => {
  let service: WorkflowVersionService;
  let repo: WorkflowVersionRepositoryMock;
  let workflowRepo: WorkflowRepositoryMock;

  beforeEach(async () => {
    repo = createWorkflowVersionRepositoryMock();
    workflowRepo = createWorkflowRepositoryMock();
    const cacheStore = new Map<string, unknown>();
    const cache = {
      get: jest.fn((key: string) => Promise.resolve(cacheStore.get(key))),
      set: jest.fn((key: string, value: unknown) => {
        cacheStore.set(key, value);
        return Promise.resolve();
      }),
      del: jest.fn((key: string) => {
        cacheStore.delete(key);
        return Promise.resolve();
      }),
    };

    const moduleRef = await Test.createTestingModule({
      providers: [
        WorkflowVersionService,
        { provide: WorkflowVersionRepository, useValue: repo },
        { provide: WorkflowRepository, useValue: workflowRepo },
        { provide: CACHE_MANAGER, useValue: cache },
      ],
    }).compile();

    service = moduleRef.get(WorkflowVersionService);
  });

  it('list() returns versions for workflow', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const list = createWorkflowVersionListFixture(2);

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

  it('list() throws notFound when workflow missing', async () => {
    workflowRepo.findById.mockResolvedValue(null);

    await expect(
      service.list({ workflowId: 'missing', page: 1, pageSize: 25 }),
    ).rejects.toBeInstanceOf(AppError);
    expect(repo.findPageByWorkflow).not.toHaveBeenCalled();
  });

  it('get() returns a version when found', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    const version = createWorkflowVersionFixture({
      workflowId: 'wf_1',
      version: 1,
    });

    workflowRepo.findById.mockResolvedValue(wf);
    repo.findByWorkflowAndVersion.mockResolvedValue(version);

    await expect(service.get('wf_1', 1)).resolves.toBe(version);
    expect(repo.findByWorkflowAndVersion).toHaveBeenCalledWith('wf_1', 1);
  });

  it('get() throws notFound when workflow missing', async () => {
    workflowRepo.findById.mockResolvedValue(null);

    await expect(service.get('missing', 1)).rejects.toBeInstanceOf(AppError);
    expect(repo.findByWorkflowAndVersion).not.toHaveBeenCalled();
  });

  it('get() throws notFound when version missing', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    workflowRepo.findById.mockResolvedValue(wf);
    repo.findByWorkflowAndVersion.mockResolvedValue(null);

    await expect(service.get('wf_1', 42)).rejects.toBeInstanceOf(AppError);
  });
});
