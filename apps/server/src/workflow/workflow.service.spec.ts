import { Test } from '@nestjs/testing';
import { WorkflowService } from './workflow.service';
import { WorkflowRepository, PrismaService } from '@lunie/db-access';
import { SecretRepository } from '@lunie/db-access';
import {
  createWorkflowRepositoryMock,
  type WorkflowRepositoryMock,
} from 'test/workflow/workflow.repository.mock';
import {
  createWorkflowFixture,
  createWorkflowListFixture,
} from 'test/workflow/workflow.fixtures';
import { createWorkflowVersionFixture } from 'test/workflow-version/workflow-version.fixtures';
import { AppError } from 'src/common/http/errors/app-error';
import {
  createPrismaServiceMock,
  PrismaServiceMock,
  PrismaTxMock,
} from 'test/prisma/prisma.mocks';
import { CACHE_MANAGER } from '@nestjs/cache-manager';

describe('WorkflowService', () => {
  let service: WorkflowService;
  let repo: WorkflowRepositoryMock;
  let prisma: PrismaServiceMock;

  beforeEach(async () => {
    repo = createWorkflowRepositoryMock();
    prisma = createPrismaServiceMock();
    const secretRepo = { findManyByNames: jest.fn().mockResolvedValue([]) };
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
        WorkflowService,
        { provide: WorkflowRepository, useValue: repo },
        { provide: SecretRepository, useValue: secretRepo },
        { provide: PrismaService, useValue: prisma },
        { provide: CACHE_MANAGER, useValue: cache },
      ],
    }).compile();

    service = moduleRef.get(WorkflowService);
  });

  it('create() creates a workflow', async () => {
    const existing = createWorkflowFixture({
      id: 'wf_existing',
      key: null,
      name: 'Existing Workflow',
    });
    const wf = createWorkflowFixture({ name: 'My WF', key: 'my-wf' });
    const version = createWorkflowVersionFixture({
      workflowId: wf.id,
      version: 1,
      definition: {
        input: {},
        steps: [
          {
            key: 's1',
            type: 'http',
            request: { method: 'GET', url: 'https://example.com' },
          },
        ],
      },
    });
    const updated = createWorkflowFixture({
      id: wf.id,
      key: 'my-wf',
      name: 'My WF',
      latestVersionId: version.id,
    });

    const tx: PrismaTxMock = {
      trigger: {
        create: jest.fn().mockResolvedValue({
          id: 'tr_manual',
          workflowId: wf.id,
          type: 'MANUAL',
          name: 'Manual',
          isActive: true,
          config: {},
          createdAt: new Date(),
          updatedAt: new Date(),
        }),
      },
      workflow: {
        create: jest.fn().mockResolvedValue(wf),
        update: jest.fn().mockResolvedValue(updated),
      },
      workflowVersion: {
        create: jest.fn().mockResolvedValue(version),
        findFirst: jest.fn(),
      },
    };

    repo.findMany.mockResolvedValue([existing]);
    prisma.$transaction.mockImplementation((cb) => Promise.resolve(cb(tx)));

    await expect(
      service.create({
        name: 'My WF',
        definition: version.definition,
      }),
    ).resolves.toBe(updated);

    expect(prisma.$transaction).toHaveBeenCalledTimes(1);
    expect(tx.workflow.create).toHaveBeenCalledWith({
      data: { name: 'My WF', key: 'my-wf' },
    });

    const triggerCreate = tx.trigger?.create;
    if (!triggerCreate) throw new Error('Expected trigger.create to exist');

    expect(triggerCreate).toHaveBeenCalledWith({
      data: {
        workflowId: wf.id,
        key: 'manual',
        type: 'MANUAL',
        name: 'Manual',
        isActive: true,
        config: {},
      },
    });
    expect(tx.workflowVersion.create).toHaveBeenCalledWith({
      data: {
        workflowId: wf.id,
        version: 1,
        definition: version.definition,
      },
    });
    expect(tx.workflow.update).toHaveBeenCalledWith({
      where: { id: wf.id },
      data: { latestVersionId: version.id },
    });
  });

  it('list() returns workflows', async () => {
    const list = createWorkflowListFixture(2);
    repo.findPage.mockResolvedValue({ items: list, total: 2 });

    await expect(service.list({ page: 1, pageSize: 25 })).resolves.toEqual({
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
    expect(repo.findPage).toHaveBeenCalledWith({ page: 1, pageSize: 25 });
  });

  it('get() returns workflow when found', async () => {
    const wf = createWorkflowFixture({ id: 'wf_x' });
    repo.findById.mockResolvedValue(wf);

    await expect(service.get('wf_x')).resolves.toBe(wf);
    expect(repo.findById).toHaveBeenCalledWith('wf_x');
  });

  it('get() throws AppError.notFound when missing', async () => {
    repo.findById.mockResolvedValue(null);

    await expect(service.get('missing')).rejects.toBeInstanceOf(AppError);
  });

  it('update() updates after existence check', async () => {
    const existing = createWorkflowFixture({ id: 'wf_1', name: 'Old' });
    const updated = createWorkflowFixture({ id: 'wf_1', name: 'New' });

    repo.findById.mockResolvedValue(existing);
    repo.update.mockResolvedValue(updated);

    await expect(service.update('wf_1', { name: 'New' })).resolves.toBe(
      updated,
    );

    expect(repo.findById).toHaveBeenCalledWith('wf_1');
    expect(repo.update).toHaveBeenCalledWith('wf_1', { name: 'New' });
  });

  it('update() throws notFound when workflow missing', async () => {
    repo.findById.mockResolvedValue(null);

    await expect(
      service.update('missing', { isActive: false }),
    ).rejects.toBeInstanceOf(AppError);
    expect(repo.update).not.toHaveBeenCalled();
  });

  it('delete() soft deletes workflow', async () => {
    const existing = createWorkflowFixture({ id: 'wf_1', isActive: true });
    const deleted = createWorkflowFixture({ id: 'wf_1', isActive: false });

    repo.findById.mockResolvedValue(existing);
    repo.softDelete.mockResolvedValue(deleted);

    await expect(service.delete('wf_1')).resolves.toBe(deleted);
    expect(repo.softDelete).toHaveBeenCalledWith('wf_1');
  });

  it('createVersion() creates next version and updates latestVersionId', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    repo.findById.mockResolvedValue(wf);

    const version = createWorkflowVersionFixture({
      id: 'wfv_2',
      workflowId: 'wf_1',
      version: 2,
      definition: {
        steps: [
          {
            key: 'step-1',
            type: 'http',
            request: { method: 'GET', url: 'https://example.com' },
          },
        ],
      },
    });

    const tx: PrismaTxMock = {
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

    await expect(
      service.createVersion('wf_1', version.definition),
    ).resolves.toBe(version);

    expect(tx.workflowVersion.create).toHaveBeenCalledWith({
      data: {
        workflowId: 'wf_1',
        version: 2,
        definition: version.definition,
      },
    });
    expect(tx.workflow.update).toHaveBeenCalledWith({
      where: { id: 'wf_1' },
      data: { latestVersionId: version.id },
    });
  });

  it('createVersion() rejects definitions with unknown input template refs (strict)', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    repo.findById.mockResolvedValue(wf);

    const tx: PrismaTxMock = {
      workflow: {
        create: jest.fn(),
        update: jest.fn().mockResolvedValue(wf),
      },
      workflowVersion: {
        findFirst: jest.fn().mockResolvedValue({ version: 1 }),
        create: jest.fn(),
      },
    };

    prisma.$transaction.mockImplementation((cb) => Promise.resolve(cb(tx)));

    await expect(
      service.createVersion('wf_1', {
        input: { apiUrl: 'https://example.com' },
        steps: [
          {
            key: 's1',
            type: 'http',
            request: {
              method: 'GET',
              url: '{{input.missing}}/x',
            },
          },
        ],
      }),
    ).rejects.toBeInstanceOf(AppError);

    expect(tx.workflowVersion.create).not.toHaveBeenCalled();
  });

  it('createVersion() rejects definitions with invalid $jmes', async () => {
    const wf = createWorkflowFixture({ id: 'wf_1' });
    repo.findById.mockResolvedValue(wf);

    const tx: PrismaTxMock = {
      workflow: {
        create: jest.fn(),
        update: jest.fn().mockResolvedValue(wf),
      },
      workflowVersion: {
        findFirst: jest.fn().mockResolvedValue({ version: 1 }),
        create: jest.fn(),
      },
    };

    prisma.$transaction.mockImplementation((cb) => Promise.resolve(cb(tx)));

    await expect(
      service.createVersion('wf_1', {
        input: {},
        steps: [
          {
            key: 't1',
            type: 'transform',
            request: {
              source: {},
              output: {
                bad: { $jmes: 'length(' },
              },
            },
          },
        ],
      }),
    ).rejects.toBeInstanceOf(AppError);

    expect(tx.workflowVersion.create).not.toHaveBeenCalled();
  });

  // condition step type not enabled yet

  it('createVersion() throws notFound when workflow missing', async () => {
    repo.findById.mockResolvedValue(null);

    await expect(
      service.createVersion('missing', { steps: [] }),
    ).rejects.toBeInstanceOf(AppError);
    expect(prisma.$transaction).not.toHaveBeenCalled();
  });
});
