import type { StepRun } from '@prisma/client';

export type StepRunRepositoryMock = {
  create: jest.Mock<Promise<StepRun>, [any]>;
  findManyByWorkflowRun: jest.Mock<Promise<StepRun[]>, [string]>;
  findPageByWorkflowRun: jest.Mock<
    Promise<{ items: StepRun[]; total: number }>,
    [
      {
        workflowRunId: string;
        page: number;
        pageSize: number;
        sortBy: 'createdAt' | 'updatedAt';
        sortOrder: 'asc' | 'desc';
      },
    ]
  >;
  findById: jest.Mock<Promise<StepRun | null>, [string]>;
  findByWorkflowRunAndStepKey: jest.Mock<
    Promise<StepRun | null>,
    [string, string]
  >;
  update: jest.Mock<Promise<StepRun>, [string, any]>;
};

export const createStepRunRepositoryMock = (): StepRunRepositoryMock => ({
  create: jest.fn<Promise<StepRun>, [any]>(),
  findManyByWorkflowRun: jest.fn<Promise<StepRun[]>, [string]>(),
  findPageByWorkflowRun: jest.fn<
    Promise<{ items: StepRun[]; total: number }>,
    [
      {
        workflowRunId: string;
        page: number;
        pageSize: number;
        sortBy: 'createdAt' | 'updatedAt';
        sortOrder: 'asc' | 'desc';
      },
    ]
  >(),
  findById: jest.fn<Promise<StepRun | null>, [string]>(),
  findByWorkflowRunAndStepKey: jest.fn<
    Promise<StepRun | null>,
    [string, string]
  >(),
  update: jest.fn<Promise<StepRun>, [string, any]>(),
});
