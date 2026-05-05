import { createZodDto } from 'nestjs-zod';
import { z } from 'zod';
import {
  PaginationQuerySchema,
  SortOrderSchema,
} from 'src/common/dto/pagination.dto';

export const TriggerTypeSchema = z.enum(['MANUAL', 'WEBHOOK', 'CRON']);
export const TriggerListSortBySchema = z.enum(['createdAt', 'updatedAt']);
export const TriggerListQuerySchema = PaginationQuerySchema.extend({
  sortBy: TriggerListSortBySchema.default('createdAt'),
  sortOrder: SortOrderSchema.default('desc'),
});

export const TriggerResSchema = z.object({
  id: z.string(),
  workflowId: z.string(),
  key: z.string().nullable(),
  type: TriggerTypeSchema,
  name: z.string().nullable(),
  isActive: z.boolean(),
  config: z.unknown(),
  createdAt: z.iso.datetime(),
  updatedAt: z.iso.datetime(),
});

export class TriggerResDto extends createZodDto(TriggerResSchema) {}
export class TriggerListQueryDto extends createZodDto(TriggerListQuerySchema) {}

export const CreateTriggerReqSchema = z.object({
  type: TriggerTypeSchema,
  name: z.string().min(1).max(120).optional(),
  config: z.unknown().default({}),
  isActive: z.boolean().optional(),
});

export class CreateTriggerReqDto extends createZodDto(CreateTriggerReqSchema) {}

export const UpdateTriggerReqSchema = z.object({
  name: z.string().min(1).max(120).optional(),
  config: z.unknown().optional(),
  isActive: z.boolean().optional(),
});

export class UpdateTriggerReqDto extends createZodDto(UpdateTriggerReqSchema) {}

export const RotateWebhookKeyResSchema = z.object({
  webhookKey: z.string(),
});
export class RotateWebhookKeyResDto extends createZodDto(
  RotateWebhookKeyResSchema,
) {}

export const TriggerWebhookIngressResSchema = z.object({
  status: z.enum(['accepted', 'trigger_inactive']),
});
export class TriggerWebhookIngressResDto extends createZodDto(
  TriggerWebhookIngressResSchema,
) {}
