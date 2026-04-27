import { createZodDto } from 'nestjs-zod';
import { z } from 'zod';
import { WorkflowDefinitionSchema } from '@lunie/contracts';
import {
  PaginationQuerySchema,
  SortOrderSchema,
} from 'src/common/dto/pagination.dto';

export const WorkflowVersionListSortBySchema = z.enum(['version', 'createdAt']);
export const WorkflowVersionListQuerySchema = PaginationQuerySchema.extend({
  sortBy: WorkflowVersionListSortBySchema.default('version'),
  sortOrder: SortOrderSchema.default('desc'),
});
export class WorkflowVersionListQueryDto extends createZodDto(
  WorkflowVersionListQuerySchema,
) {}

export const WorkflowVersionResSchema = z.object({
  id: z.string(),
  workflowId: z.string(),
  version: z.number().int().positive(),
  definition: WorkflowDefinitionSchema,
  createdAt: z.string().datetime(),
});
export class WorkflowVersionResDto extends createZodDto(
  WorkflowVersionResSchema,
) {}

export const CreateWorkflowVersionReqSchema = z.object({
  definition: WorkflowDefinitionSchema,
});
export class CreateWorkflowVersionReqDto extends createZodDto(
  CreateWorkflowVersionReqSchema,
) {}
