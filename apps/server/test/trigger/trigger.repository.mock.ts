import type { Trigger } from '@prisma/client';

export type TriggerRepositoryMock = {
  create: jest.Mock<Promise<Trigger>, [any]>;
  findManyByWorkflow: jest.Mock<Promise<Trigger[]>, [string]>;
  findPageByWorkflow: jest.Mock<
    Promise<{ items: Trigger[]; total: number }>,
    [
      {
        workflowId: string;
        page: number;
        pageSize: number;
        sortBy: 'createdAt' | 'updatedAt';
        sortOrder: 'asc' | 'desc';
      },
    ]
  >;
  findById: jest.Mock<Promise<Trigger | null>, [string]>;
  findByWorkflowAndKey: jest.Mock<Promise<Trigger | null>, [string, string]>;
  update: jest.Mock<Promise<Trigger>, [string, any]>;
  softDelete: jest.Mock<Promise<Trigger>, [string]>;
};

export const createTriggerRepositoryMock = (): TriggerRepositoryMock => ({
  create: jest.fn<Promise<Trigger>, [any]>(),
  findManyByWorkflow: jest.fn<Promise<Trigger[]>, [string]>(),
  findPageByWorkflow: jest.fn<
    Promise<{ items: Trigger[]; total: number }>,
    [
      {
        workflowId: string;
        page: number;
        pageSize: number;
        sortBy: 'createdAt' | 'updatedAt';
        sortOrder: 'asc' | 'desc';
      },
    ]
  >(),
  findById: jest.fn<Promise<Trigger | null>, [string]>(),
  findByWorkflowAndKey: jest.fn<Promise<Trigger | null>, [string, string]>(),
  update: jest.fn<Promise<Trigger>, [string, any]>(),
  softDelete: jest.fn<Promise<Trigger>, [string]>(),
});
