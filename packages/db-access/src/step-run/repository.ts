import { Injectable } from "@nestjs/common";
import type { Prisma, StepRun } from "@prisma/client";
import { PrismaService } from "../prisma.service";

@Injectable()
export class StepRunRepository {
  constructor(private readonly prisma: PrismaService) {}

  create(data: Prisma.StepRunCreateInput): Promise<StepRun> {
    return this.prisma.stepRun.create({ data });
  }

  findManyByWorkflowRun(workflowRunId: string): Promise<StepRun[]> {
    return this.prisma.stepRun.findMany({
      where: { workflowRunId },
      orderBy: { createdAt: "asc" },
    });
  }

  async findPageByWorkflowRun(params: {
    workflowRunId: string;
    page: number;
    pageSize: number;
    sortBy: "createdAt" | "updatedAt";
    sortOrder: "asc" | "desc";
  }): Promise<{ items: StepRun[]; total: number }> {
    const skip = (params.page - 1) * params.pageSize;
    const orderBy =
      params.sortBy === "updatedAt"
        ? [{ updatedAt: params.sortOrder }, { id: params.sortOrder }]
        : [{ createdAt: params.sortOrder }, { id: params.sortOrder }];
    const [items, total] = await Promise.all([
      this.prisma.stepRun.findMany({
        where: { workflowRunId: params.workflowRunId },
        orderBy,
        skip,
        take: params.pageSize,
      }),
      this.prisma.stepRun.count({
        where: { workflowRunId: params.workflowRunId },
      }),
    ]);

    return { items, total };
  }

  findById(id: string): Promise<StepRun | null> {
    return this.prisma.stepRun.findUnique({ where: { id } });
  }

  findByWorkflowRunAndStepKey(
    workflowRunId: string,
    stepKey: string,
  ): Promise<StepRun | null> {
    return this.prisma.stepRun.findUnique({
      where: { workflowRunId_stepKey: { workflowRunId, stepKey } },
    });
  }

  findFirst(args: Prisma.StepRunFindFirstArgs): Promise<StepRun | null> {
    return this.prisma.stepRun.findFirst(args);
  }

  update(id: string, data: Prisma.StepRunUpdateInput): Promise<StepRun> {
    return this.prisma.stepRun.update({ where: { id }, data });
  }
}
