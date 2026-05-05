import { Injectable } from "@nestjs/common";
import { PrismaService } from "../prisma.service";
import type { Prisma, Workflow } from "@prisma/client";

@Injectable()
export class WorkflowRepository {
  constructor(private readonly prisma: PrismaService) {}

  create(data: Prisma.WorkflowCreateInput): Promise<Workflow> {
    return this.prisma.workflow.create({ data });
  }

  findMany(): Promise<Workflow[]> {
    return this.prisma.workflow.findMany({ orderBy: { createdAt: "desc" } });
  }

  async findPage(params: {
    page: number;
    pageSize: number;
    sortBy: "createdAt" | "updatedAt";
    sortOrder: "asc" | "desc";
  }): Promise<{ items: Workflow[]; total: number }> {
    const skip = (params.page - 1) * params.pageSize;
    const orderBy =
      params.sortBy === "updatedAt"
        ? [{ updatedAt: params.sortOrder }, { id: params.sortOrder }]
        : [{ createdAt: params.sortOrder }, { id: params.sortOrder }];
    const [items, total] = await Promise.all([
      this.prisma.workflow.findMany({
        orderBy,
        skip,
        take: params.pageSize,
      }),
      this.prisma.workflow.count(),
    ]);

    return { items, total };
  }

  findById(id: string): Promise<Workflow | null> {
    return this.prisma.workflow.findUnique({ where: { id } });
  }

  findByKey(key: string): Promise<Workflow | null> {
    return this.prisma.workflow.findUnique({ where: { key } });
  }

  update(id: string, data: Prisma.WorkflowUpdateInput): Promise<Workflow> {
    return this.prisma.workflow.update({ where: { id }, data });
  }

  softDelete(id: string): Promise<Workflow> {
    return this.prisma.workflow.update({
      where: { id },
      data: { isActive: false },
    });
  }
}
