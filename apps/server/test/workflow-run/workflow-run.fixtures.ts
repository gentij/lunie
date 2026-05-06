import type { WorkflowRun, WorkflowRunStatus } from '@prisma/client';

type WorkflowRunOverrides = Partial<WorkflowRun>;

export const createWorkflowRunFixture = (
  overrides: WorkflowRunOverrides = {},
): WorkflowRun => {
  const now = new Date('2026-01-23T10:00:00.000Z');

  return {
    id: 'wfr_1',
    workflowId: 'wf_1',
    workflowVersionId: 'wfv_1',
    number: 1,
    triggerId: 'tr_1',
    eventId: 'ev_1',
    status: 'QUEUED' as WorkflowRunStatus,
    input: {},
    overrides: null,
    output: null,
    startedAt: null,
    finishedAt: null,
    createdAt: now,
    updatedAt: now,
    ...overrides,
  };
};

export const createWorkflowRunListFixture = (count = 3): WorkflowRun[] =>
  Array.from({ length: count }, (_, i) =>
    createWorkflowRunFixture({
      id: `wfr_${i + 1}`,
    }),
  );
