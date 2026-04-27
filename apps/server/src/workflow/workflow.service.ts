import { Injectable } from '@nestjs/common';
import type { Prisma, Workflow, WorkflowVersion } from '@prisma/client';
import { WorkflowDefinitionSchema } from '@lunie/contracts';
import {
  SecretRepository,
  WorkflowRepository,
  PrismaService,
} from '@lunie/db-access';
import { ErrorDefinitions } from 'src/common/http/errors/error-codes';
import { AppError } from 'src/common/http/errors/app-error';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { Inject } from '@nestjs/common';
import type { Cache } from 'cache-manager';
import {
  getInferredDependencies,
  getExecutionBatchesFromDependencies,
  getReferencedSecrets,
  validateWorkflowDefinitionStrict,
} from './workflow-definition.validator';
import { cacheKeys } from 'src/cache/cache-keys';
import { buildPaginationMeta } from 'src/common/pagination/pagination';

@Injectable()
export class WorkflowService {
  constructor(
    private readonly repo: WorkflowRepository,
    private readonly secretRepo: SecretRepository,
    private readonly prisma: PrismaService,
    @Inject(CACHE_MANAGER) private readonly cache: Cache,
  ) {}

  async create(params: {
    name: string;
    definition: unknown;
  }): Promise<Workflow> {
    const normalizedDefinition = WorkflowDefinitionSchema.parse(
      params.definition,
    );
    await this.validateDefinitionOrThrow(normalizedDefinition);

    const created = await this.prisma.$transaction(async (tx) => {
      const workflow = await tx.workflow.create({
        data: { name: params.name },
      });

      await tx.trigger.create({
        data: {
          workflowId: workflow.id,
          type: 'MANUAL',
          name: 'Manual',
          isActive: true,
          config: {},
        },
      });

      const v1 = await tx.workflowVersion.create({
        data: {
          workflowId: workflow.id,
          version: 1,
          definition: normalizedDefinition as Prisma.InputJsonValue,
        },
      });

      const updatedWorkflow = await tx.workflow.update({
        where: { id: workflow.id },
        data: { latestVersionId: v1.id },
      });

      return updatedWorkflow;
    });

    return created;
  }

  async createVersion(
    workflowId: string,
    definition: unknown,
  ): Promise<WorkflowVersion> {
    await this.get(workflowId);

    const normalizedDefinition = WorkflowDefinitionSchema.parse(definition);
    await this.validateDefinitionOrThrow(normalizedDefinition);

    const created = await this.prisma.$transaction(async (tx) => {
      const latest = await tx.workflowVersion.findFirst({
        where: { workflowId },
        orderBy: { version: 'desc' },
        select: { version: true },
      });

      const nextVersion = (latest?.version ?? 0) + 1;

      const created = await tx.workflowVersion.create({
        data: {
          workflowId,
          version: nextVersion,
          definition: normalizedDefinition as Prisma.InputJsonValue,
        },
      });

      await tx.workflow.update({
        where: { id: workflowId },
        data: { latestVersionId: created.id },
      });

      return created;
    });

    await this.cache.del(cacheKeys.workflowGet(workflowId));
    return created;
  }

  validateDefinition(definition: unknown) {
    const normalizedDefinition = WorkflowDefinitionSchema.parse(definition);
    const issues = validateWorkflowDefinitionStrict(normalizedDefinition);

    const secretRefs = getReferencedSecrets(normalizedDefinition);
    const referencedSecrets = Array.from(
      new Set(secretRefs.map((r) => r.name)),
    );

    const inferredDependencies = getInferredDependencies(normalizedDefinition);
    let executionBatches: string[][] = [];
    if (issues.length === 0) {
      executionBatches =
        getExecutionBatchesFromDependencies(inferredDependencies);
    }

    return {
      valid: issues.length === 0,
      issues,
      inferredDependencies,
      executionBatches,
      referencedSecrets,
    };
  }

  async validateDefinitionOrThrow(definition: unknown) {
    const result = this.validateDefinition(definition);
    if (!result.valid) {
      throw AppError.badRequest(
        ErrorDefinitions.COMMON.VALIDATION_ERROR,
        result.issues,
      );
    }

    const normalizedDefinition = WorkflowDefinitionSchema.parse(definition);
    const secretRefs = getReferencedSecrets(normalizedDefinition);

    const secrets = (await (
      this.secretRepo as unknown as {
        findManyByNames: (names: string[]) => Promise<Array<{ name: string }>>;
      }
    ).findManyByNames(result.referencedSecrets)) as Array<{ name: string }>;

    const existing = new Set<string>(secrets.map((s) => s.name));
    const missing = result.referencedSecrets.filter((n) => !existing.has(n));
    if (missing.length > 0) {
      const issues = secretRefs
        .filter((r) => missing.includes(r.name))
        .map((r) => ({
          field: r.field,
          stepKey: r.stepKey,
          message: `secret "${r.name}" not found`,
        }));

      throw AppError.badRequest(
        ErrorDefinitions.COMMON.VALIDATION_ERROR,
        issues,
      );
    }

    return result;
  }

  async list(params: {
    page: number;
    pageSize: number;
    sortBy: 'createdAt' | 'updatedAt';
    sortOrder: 'asc' | 'desc';
  }): Promise<{
    items: Workflow[];
    pagination: ReturnType<typeof buildPaginationMeta>;
  }> {
    const { items, total } = await this.repo.findPage(params);
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

  async get(id: string): Promise<Workflow> {
    const key = cacheKeys.workflowGet(id);
    try {
      const cached = await this.cache.get<Workflow>(key);
      if (cached) return cached;
    } catch {
      // fail-open: cache errors should not break API
    }

    const wf = await this.repo.findById(id);

    if (!wf) throw AppError.notFound(ErrorDefinitions.WORKFLOW.NOT_FOUND);

    try {
      await this.cache.set(key, wf);
    } catch {
      // fail-open: cache errors should not break API
    }
    return wf;
  }

  async update(
    id: string,
    patch: { name?: string; isActive?: boolean },
  ): Promise<Workflow> {
    await this.get(id);

    const updated = await this.repo.update(id, patch);
    await this.cache.del(cacheKeys.workflowGet(id));
    return updated;
  }

  async delete(id: string): Promise<Workflow> {
    await this.get(id);

    const deleted = await this.repo.softDelete(id);
    await this.cache.del(cacheKeys.workflowGet(id));
    return deleted;
  }
}
