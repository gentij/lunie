import type { WorkflowDefinition } from '@lunie/contracts';
import {
  getReferencedSecrets,
  validateWorkflowDefinitionStrict,
} from './workflow-definition.validator';

describe('validateWorkflowDefinitionStrict', () => {
  it('rejects invalid template roots in step requests', () => {
    const definition = {
      steps: [
        {
          key: 'discord',
          type: 'http',
          request: {
            method: 'POST',
            url: '{{secrets.DISCORD_WEBHOOK_URL}}',
          },
        },
      ],
    };

    const issues = validateWorkflowDefinitionStrict(
      definition as WorkflowDefinition,
    );

    const issue = issues.find(
      (item) =>
        item.field === 'steps[0].request.url' && item.stepKey === 'discord',
    );

    expect(issue).toBeDefined();
    expect(issue?.message).toContain('invalid template root "secrets"');
  });

  it('rejects invalid template roots in workflow input', () => {
    const definition = {
      input: {
        webhook: '{{secrets.DISCORD_WEBHOOK_URL}}',
      },
      steps: [],
    };

    const issues = validateWorkflowDefinitionStrict(
      definition as WorkflowDefinition,
    );

    const issue = issues.find((item) => item.field === 'input.webhook');

    expect(issue).toBeDefined();
    expect(issue?.message).toContain('invalid template root "secrets"');
  });

  it('allows valid template roots', () => {
    const definition = {
      input: {
        webhook: '{{secret.DISCORD_WEBHOOK_URL}}',
      },
      steps: [
        {
          key: 'discord',
          type: 'http',
          request: {
            method: 'POST',
            url: '{{secret.DISCORD_WEBHOOK_URL}}',
          },
        },
      ],
    };

    const issues = validateWorkflowDefinitionStrict(
      definition as WorkflowDefinition,
    );

    expect(issues).toEqual([]);
  });

  it('allows notification webhook secret template', () => {
    const definition = {
      notifications: [
        {
          provider: 'discord',
          webhook: '{{secret.DISCORD_WEBHOOK_URL}}',
          on: ['FAILED'],
        },
      ],
      steps: [],
    };

    const issues = validateWorkflowDefinitionStrict(
      definition as WorkflowDefinition,
    );

    expect(issues).toEqual([]);
  });

  it('rejects invalid notification webhook values', () => {
    const definition = {
      notifications: [
        {
          provider: 'slack',
          webhook: 'not-a-url',
          on: ['SUCCEEDED'],
        },
      ],
      steps: [],
    };

    const issues = validateWorkflowDefinitionStrict(
      definition as WorkflowDefinition,
    );

    const issue = issues.find(
      (item) => item.field === 'notifications[0].webhook',
    );

    expect(issue).toBeDefined();
    expect(issue?.message).toContain('notification webhook must be');
  });

  it('extracts referenced secrets from notifications', () => {
    const definition = {
      notifications: [
        {
          provider: 'discord',
          webhook: '{{secret.DISCORD_WEBHOOK_URL}}',
          on: ['FAILED'],
        },
      ],
      steps: [],
    };

    const refs = getReferencedSecrets(definition as WorkflowDefinition);
    expect(refs).toEqual([
      {
        name: 'DISCORD_WEBHOOK_URL',
        field: 'notifications[0].webhook',
      },
    ]);
  });
});
