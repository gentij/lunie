import { createZodDto } from 'nestjs-zod';
import { z } from 'zod';
import { WorkflowDefinitionSchema } from '@lunie/contracts';

export const ValidateWorkflowDefinitionReqSchema = z.object({
  definition: WorkflowDefinitionSchema,
});

export class ValidateWorkflowDefinitionReqDto extends createZodDto(
  ValidateWorkflowDefinitionReqSchema,
) {}

export const ValidateWorkflowDefinitionResSchema = z.object({
  valid: z.boolean(),
  issues: z
    .array(
      z.object({
        field: z.string().optional(),
        stepKey: z.string().optional(),
        message: z.string(),
      }),
    )
    .default([]),
  inferredDependencies: z.record(z.string(), z.array(z.string())).default({}),
  executionBatches: z.array(z.array(z.string())).default([]),
  referencedSecrets: z.array(z.string()).default([]),
});

export class ValidateWorkflowDefinitionResDto extends createZodDto(
  ValidateWorkflowDefinitionResSchema,
) {}
