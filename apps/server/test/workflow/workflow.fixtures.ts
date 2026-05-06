import type { Workflow } from '@prisma/client';

type WorkflowOverrides = Partial<Workflow>;

export const createWorkflowFixture = (
  overrides: WorkflowOverrides = {},
): Workflow => {
  const now = new Date('2026-01-23T10:00:00.000Z');

  return {
    id: 'wf_1',
    key: 'deploy-on-release',
    name: 'Deploy on Release',
    isActive: true,
    runSequence: 0,
    latestVersionId: null,
    createdAt: now,
    updatedAt: now,
    ...overrides,
  };
};

export const createWorkflowListFixture = (count = 3): Workflow[] =>
  Array.from({ length: count }, (_, i) =>
    createWorkflowFixture({
      id: `wf_${i + 1}`,
      name: `Workflow ${i + 1}`,
    }),
  );
