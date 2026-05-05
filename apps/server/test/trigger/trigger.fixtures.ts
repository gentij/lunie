import type { Trigger, TriggerType } from '@prisma/client';

type TriggerOverrides = Partial<Trigger>;

export const createTriggerFixture = (
  overrides: TriggerOverrides = {},
): Trigger => {
  const now = new Date('2026-01-23T10:00:00.000Z');

  return {
    id: 'tr_1',
    workflowId: 'wf_1',
    key: null,
    type: 'MANUAL' as TriggerType,
    name: 'Manual Trigger',
    isActive: true,
    config: {},
    createdAt: now,
    updatedAt: now,
    ...overrides,
  };
};

export const createTriggerListFixture = (count = 3): Trigger[] =>
  Array.from({ length: count }, (_, i) =>
    createTriggerFixture({
      id: `tr_${i + 1}`,
      name: `Trigger ${i + 1}`,
    }),
  );
