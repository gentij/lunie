import type { WorkflowRun } from '@prisma/client';

export type WorkflowRunRepositoryMock = {
  create: jest.Mock<Promise<WorkflowRun>, [any]>;
  findManyByWorkflow: jest.Mock<Promise<WorkflowRun[]>, [string]>;
  findPageByWorkflow: jest.Mock<
    Promise<{ items: WorkflowRun[]; total: number }>,
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
  findById: jest.Mock<Promise<WorkflowRun | null>, [string]>;
  findByWorkflowAndNumber: jest.Mock<
    Promise<WorkflowRun | null>,
    [string, number]
  >;
  update: jest.Mock<Promise<WorkflowRun>, [string, any]>;
};

export const createWorkflowRunRepositoryMock =
  (): WorkflowRunRepositoryMock => ({
    create: jest.fn<Promise<WorkflowRun>, [any]>(),
    findManyByWorkflow: jest.fn<Promise<WorkflowRun[]>, [string]>(),
    findPageByWorkflow: jest.fn<
      Promise<{ items: WorkflowRun[]; total: number }>,
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
    findById: jest.fn<Promise<WorkflowRun | null>, [string]>(),
    findByWorkflowAndNumber: jest.fn<
      Promise<WorkflowRun | null>,
      [string, number]
    >(),
    update: jest.fn<Promise<WorkflowRun>, [string, any]>(),
  });
