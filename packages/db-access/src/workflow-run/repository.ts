import { Injectable } from '@nestjs/common';
import type { Prisma, WorkflowRun } from '@prisma/client';
import { PrismaService } from '../prisma.service';

@Injectable()
export class WorkflowRunRepository {
  constructor(private readonly prisma: PrismaService) {}

  create(data: Prisma.WorkflowRunCreateInput): Promise<WorkflowRun> {
    return this.prisma.workflowRun.create({ data });
  }

  findManyByWorkflow(workflowId: string): Promise<WorkflowRun[]> {
    return this.prisma.workflowRun.findMany({
      where: { workflowId },
      orderBy: { createdAt: 'desc' },
    });
  }

  async findPageByWorkflow(params: {
    workflowId: string;
    page: number;
    pageSize: number;
    sortBy: 'createdAt' | 'updatedAt';
    sortOrder: 'asc' | 'desc';
  }): Promise<{ items: WorkflowRun[]; total: number }> {
    const skip = (params.page - 1) * params.pageSize;
    const orderBy =
      params.sortBy === 'updatedAt'
        ? [{ updatedAt: params.sortOrder }, { id: params.sortOrder }]
        : [{ createdAt: params.sortOrder }, { id: params.sortOrder }];
    const [items, total] = await Promise.all([
      this.prisma.workflowRun.findMany({
        where: { workflowId: params.workflowId },
        orderBy,
        skip,
        take: params.pageSize,
      }),
      this.prisma.workflowRun.count({
        where: { workflowId: params.workflowId },
      }),
    ]);

    return { items, total };
  }

  findById(id: string): Promise<WorkflowRun | null> {
    return this.prisma.workflowRun.findUnique({ where: { id } });
  }

  update(
    id: string,
    data: Prisma.WorkflowRunUpdateInput,
  ): Promise<WorkflowRun> {
    return this.prisma.workflowRun.update({ where: { id }, data });
  }

  async markRunningIfQueued(id: string): Promise<boolean> {
    const result = await this.prisma.workflowRun.updateMany({
      where: {
        id,
        status: 'QUEUED',
      },
      data: {
        status: 'RUNNING',
        startedAt: new Date(),
      },
    });

    return result.count > 0;
  }

  async finalizeIfOpen(
    id: string,
    status: 'SUCCEEDED' | 'FAILED',
  ): Promise<boolean> {
    const result = await this.prisma.workflowRun.updateMany({
      where: {
        id,
        finishedAt: null,
        status: {
          in: ['QUEUED', 'RUNNING'],
        },
      },
      data: {
        status,
        finishedAt: new Date(),
      },
    });

    return result.count > 0;
  }
}
