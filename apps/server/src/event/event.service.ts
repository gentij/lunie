import { Injectable } from '@nestjs/common';
import type { Prisma, Event } from '@prisma/client';
import {
  WorkflowRepository,
  TriggerRepository,
  EventRepository,
} from '@lunie/db-access';
import { ErrorDefinitions } from 'src/common/http/errors/error-codes';
import { AppError } from 'src/common/http/errors/app-error';
import { buildPaginationMeta } from 'src/common/pagination/pagination';

@Injectable()
export class EventService {
  constructor(
    private readonly repo: EventRepository,
    private readonly workflowRepo: WorkflowRepository,
    private readonly triggerRepo: TriggerRepository,
  ) {}

  private async assertTriggerExists(workflowId: string, triggerId: string) {
    const wf = await this.workflowRepo.findById(workflowId);
    if (!wf) throw AppError.notFound(ErrorDefinitions.WORKFLOW.NOT_FOUND);

    const trigger = await this.triggerRepo.findById(triggerId);
    if (!trigger || trigger.workflowId !== workflowId)
      throw AppError.notFound(ErrorDefinitions.TRIGGER.NOT_FOUND);

    return trigger;
  }

  async create(params: {
    triggerId: string;
    type?: string;
    externalId?: string;
    payload?: Prisma.InputJsonValue;
    receivedAt?: Date;
  }): Promise<Event> {
    return this.repo.create({
      trigger: { connect: { id: params.triggerId } },
      type: params.type,
      externalId: params.externalId,
      payload: params.payload ?? {},
      receivedAt: params.receivedAt,
    });
  }

  async list(params: {
    workflowId: string;
    triggerId: string;
    page: number;
    pageSize: number;
    sortBy: 'receivedAt' | 'createdAt';
    sortOrder: 'asc' | 'desc';
  }): Promise<{
    items: Event[];
    pagination: ReturnType<typeof buildPaginationMeta>;
  }> {
    await this.assertTriggerExists(params.workflowId, params.triggerId);
    const { items, total } = await this.repo.findPageByTrigger({
      triggerId: params.triggerId,
      page: params.page,
      pageSize: params.pageSize,
      sortBy: params.sortBy,
      sortOrder: params.sortOrder,
    });
    return {
      items,
      pagination: buildPaginationMeta({
        page: params.page,
        pageSize: params.pageSize,
        total,
        sortBy: params.sortBy,
        sortOrder: params.sortOrder,
      }),
    };
  }

  async get(workflowId: string, triggerId: string, id: string): Promise<Event> {
    await this.assertTriggerExists(workflowId, triggerId);
    const event = await this.repo.findById(id);

    if (!event || event.triggerId !== triggerId)
      throw AppError.notFound(ErrorDefinitions.EVENT.NOT_FOUND);

    return event;
  }
}
