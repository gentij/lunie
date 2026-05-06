import type { Workflow } from '@prisma/client';

export type WorkflowRepositoryMock = {
  create: jest.Mock<Promise<Workflow>, [{ name: string }]>;
  findMany: jest.Mock<Promise<Workflow[]>, []>;
  findPage: jest.Mock<
    Promise<{ items: Workflow[]; total: number }>,
    [
      {
        page: number;
        pageSize: number;
        sortBy: 'createdAt' | 'updatedAt';
        sortOrder: 'asc' | 'desc';
      },
    ]
  >;
  findById: jest.Mock<Promise<Workflow | null>, [string]>;
  findByKey: jest.Mock<Promise<Workflow | null>, [string]>;
  update: jest.Mock<
    Promise<Workflow>,
    [string, { name?: string; isActive?: boolean }]
  >;
  softDelete: jest.Mock<Promise<Workflow>, [string]>;
};

export const createWorkflowRepositoryMock = (): WorkflowRepositoryMock => ({
  create: jest.fn<Promise<Workflow>, [{ name: string }]>(),
  findMany: jest.fn<Promise<Workflow[]>, []>(),
  findPage: jest.fn<
    Promise<{ items: Workflow[]; total: number }>,
    [
      {
        page: number;
        pageSize: number;
        sortBy: 'createdAt' | 'updatedAt';
        sortOrder: 'asc' | 'desc';
      },
    ]
  >(),
  findById: jest.fn<Promise<Workflow | null>, [string]>(),
  findByKey: jest.fn<Promise<Workflow | null>, [string]>(),
  update: jest.fn<
    Promise<Workflow>,
    [string, { name?: string; isActive?: boolean }]
  >(),
  softDelete: jest.fn<Promise<Workflow>, [string]>(),
});
