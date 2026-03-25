import { z } from 'zod';

export const HttpMethodSchema = z.enum([
  'GET',
  'POST',
  'PUT',
  'PATCH',
  'DELETE',
]);

export const HttpRequestSpecSchema = z.object({
  method: HttpMethodSchema,
  url: z
    .string()
    .refine((value) => {
      // Allow template patterns like {{input.apiUrl}}/path
      if (value.includes('{{') && value.includes('}}')) {
        return true;
      }
      // Must be valid URL
      try {
        new URL(value);
        return value.startsWith('http://') || value.startsWith('https://');
      } catch {
        return false;
      }
    }, {
      message: 'url must be an absolute http(s) URL or a template pattern',
    }),
  headers: z.record(z.string(), z.string()).optional(),
  query: z
    .record(z.string(), z.union([z.string(), z.number(), z.boolean()]))
    .optional(),
  body: z.unknown().optional(),
  timeoutMs: z.number().int().positive().optional(),
});

export const BaseStepDefinitionSchema = z.object({
  key: z
    .string()
    .min(1)
    .regex(/^[a-zA-Z0-9_-]+$/, {
      message: 'key must contain only letters, numbers, underscore, or hyphen',
    }),
  dependsOn: z.array(z.string()).optional(),
  input: z.record(z.string(), z.unknown()).optional(),
  outputPolicy: z
    .object({
      truncate: z.boolean().optional(),
      maxBytes: z.number().int().positive().optional(),
    })
    .strict()
    .optional(),

  rateLimit: z
    .object({
      key: z
        .string()
        .min(1)
        .regex(/^[A-Za-z0-9_]+$/, {
          message: 'rateLimit.key must contain only letters, numbers, underscore',
        }),
      max: z.number().int().positive(),
      perSeconds: z.number().int().positive(),
    })
    .strict()
    .optional(),
});

export const HttpStepDefinitionSchema = BaseStepDefinitionSchema.extend({
  type: z.literal('http'),
  request: HttpRequestSpecSchema,
});

export const TransformRequestSpecSchema = z
  .object({
    source: z.record(z.string(), z.unknown()).optional(),
    output: z.unknown(),
  })
  .strict();

export const TransformStepDefinitionSchema = BaseStepDefinitionSchema.extend({
  type: z.literal('transform'),
  request: TransformRequestSpecSchema,
});

export const ConditionRequestSpecSchema = z
  .object({
    expr: z.string().min(1),
    assert: z.boolean().optional(),
    message: z.string().min(1).optional(),
    source: z.record(z.string(), z.unknown()).optional(),
  })
  .strict();

export const ConditionStepDefinitionSchema = BaseStepDefinitionSchema.extend({
  type: z.literal('condition'),
  request: ConditionRequestSpecSchema,
});

export const NotificationProviderSchema = z.enum(['slack', 'discord']);
export const NotificationEventSchema = z.enum(['SUCCEEDED', 'FAILED']);

export const WorkflowNotificationSchema = z
  .object({
    provider: NotificationProviderSchema,
    webhook: z
      .string()
      .min(1)
      .refine((value) => {
        const trimmed = value.trim();
        if (/^\{\{\s*secret\.[A-Za-z0-9_-]+\s*\}\}$/.test(trimmed)) {
          return true;
        }
        try {
          const parsed = new URL(trimmed);
          return parsed.protocol === 'http:' || parsed.protocol === 'https:';
        } catch {
          return false;
        }
      }, {
        message:
          'webhook must be an absolute http(s) URL or {{secret.NAME}} reference',
      }),
    on: z.array(NotificationEventSchema).min(1),
  })
  .strict();

export const StepDefinitionSchema = z.discriminatedUnion('type', [
  HttpStepDefinitionSchema,
  TransformStepDefinitionSchema,
  ConditionStepDefinitionSchema,
]);

export const WorkflowDefinitionSchema = z.object({
  input: z.record(z.string(), z.unknown()).optional(),
  notifications: z.array(WorkflowNotificationSchema).optional(),
  steps: z.array(StepDefinitionSchema).default([]),
});

export type HttpMethod = z.infer<typeof HttpMethodSchema>;
export type HttpRequestSpec = z.infer<typeof HttpRequestSpecSchema>;
export type HttpStepDefinition = z.infer<typeof HttpStepDefinitionSchema>;
export type TransformRequestSpec = z.infer<typeof TransformRequestSpecSchema>;
export type TransformStepDefinition = z.infer<typeof TransformStepDefinitionSchema>;
export type ConditionRequestSpec = z.infer<typeof ConditionRequestSpecSchema>;
export type ConditionStepDefinition = z.infer<typeof ConditionStepDefinitionSchema>;
export type NotificationProvider = z.infer<typeof NotificationProviderSchema>;
export type NotificationEvent = z.infer<typeof NotificationEventSchema>;
export type WorkflowNotification = z.infer<typeof WorkflowNotificationSchema>;
export type StepDefinition = z.infer<typeof StepDefinitionSchema>;
export type WorkflowDefinition = z.infer<typeof WorkflowDefinitionSchema>;
