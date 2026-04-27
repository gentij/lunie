import { createZodDto } from 'nestjs-zod';
import { z } from 'zod';
import { WorkflowDefinitionSchema } from '@lunie/contracts';
import {
  PaginationQuerySchema,
  SortOrderSchema,
} from 'src/common/dto/pagination.dto';

export const WorkflowResSchema = z.object({
  id: z.string(),
  name: z.string(),
  isActive: z.boolean(),
  latestVersionId: z.string().nullable(),
  createdAt: z.iso.datetime(),
  updatedAt: z.iso.datetime(),
});

export class WorkflowResDto extends createZodDto(WorkflowResSchema) {}

export const WorkflowListSortBySchema = z.enum(['createdAt', 'updatedAt']);
export const WorkflowListQuerySchema = PaginationQuerySchema.extend({
  sortBy: WorkflowListSortBySchema.default('createdAt'),
  sortOrder: SortOrderSchema.default('desc'),
});
export class WorkflowListQueryDto extends createZodDto(
  WorkflowListQuerySchema,
) {}

export const CreateWorkflowReqSchema = z.object({
  name: z.string().min(1).max(120),
  definition: WorkflowDefinitionSchema,
});
export class CreateWorkflowReqDto extends createZodDto(
  CreateWorkflowReqSchema,
) {}

export const UpdateWorkflowReqSchema = z.object({
  name: z.string().min(1).max(120).optional(),
  isActive: z.boolean().optional(),
});
export class UpdateWorkflowReqDto extends createZodDto(
  UpdateWorkflowReqSchema,
) {}

export const RunWorkflowResSchema = z.object({
  workflowRunId: z.string(),
  status: z.string(),
});
export class RunWorkflowResDto extends createZodDto(RunWorkflowResSchema) {}

export const RunWorkflowOverrideSchema = z
  .object({
    query: z
      .record(z.string(), z.union([z.string(), z.number(), z.boolean()]))
      .optional(),
    body: z.unknown().optional(),
  })
  .strict();

export const RunWorkflowReqSchema = z.object({
  input: z.record(z.string(), z.unknown()).default({}),
  overrides: z.record(z.string(), RunWorkflowOverrideSchema).default({}),
});

export const parseRunWorkflowReq = (body: unknown) =>
  RunWorkflowReqSchema.parse(body);

export class RunWorkflowReqDto extends createZodDto(RunWorkflowReqSchema) {}
