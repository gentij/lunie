import { createZodDto } from 'nestjs-zod';
import { z } from 'zod';
import {
  PaginationQuerySchema,
  SortOrderSchema,
} from 'src/common/dto/pagination.dto';

export const WorkflowRunListSortBySchema = z.enum(['createdAt', 'updatedAt']);
export const WorkflowRunListQuerySchema = PaginationQuerySchema.extend({
  sortBy: WorkflowRunListSortBySchema.default('createdAt'),
  sortOrder: SortOrderSchema.default('desc'),
});
export class WorkflowRunListQueryDto extends createZodDto(
  WorkflowRunListQuerySchema,
) {}

export const WorkflowRunStatusSchema = z.enum([
  'QUEUED',
  'RUNNING',
  'SUCCEEDED',
  'FAILED',
]);

export const WorkflowRunResSchema = z.object({
  id: z.string(),
  workflowId: z.string(),
  workflowVersionId: z.string(),
  number: z.number().int().nullable(),
  triggerId: z.string().nullable(),
  eventId: z.string().nullable(),
  status: WorkflowRunStatusSchema,
  input: z.unknown(),
  overrides: z.unknown().nullable(),
  output: z.unknown().nullable(),
  startedAt: z.iso.datetime().nullable(),
  finishedAt: z.iso.datetime().nullable(),
  createdAt: z.iso.datetime(),
  updatedAt: z.iso.datetime(),
});

export class WorkflowRunResDto extends createZodDto(WorkflowRunResSchema) {}
