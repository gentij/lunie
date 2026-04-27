import { Injectable } from '@nestjs/common';
import type { Secret } from '@prisma/client';
import { SecretRepository } from '@lunie/db-access';
import { AppError } from 'src/common/http/errors/app-error';
import { ErrorDefinitions } from 'src/common/http/errors/error-codes';
import { CryptoService } from 'src/crypto/crypto.service';
import { buildPaginationMeta } from 'src/common/pagination/pagination';

@Injectable()
export class SecretService {
  constructor(
    private readonly repo: SecretRepository,
    private readonly crypto: CryptoService,
  ) {}

  async create(params: {
    name: string;
    value: string;
    description?: string;
  }): Promise<Secret> {
    const encrypted = this.crypto.encryptSecret(params.value);
    return this.repo.create({
      name: params.name,
      value: encrypted,
      description: params.description,
    });
  }

  async list(params: {
    page: number;
    pageSize: number;
    sortBy: 'createdAt' | 'updatedAt';
    sortOrder: 'asc' | 'desc';
  }): Promise<{
    items: Secret[];
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

  async get(id: string): Promise<Secret> {
    const secret = await this.repo.findById(id);
    if (!secret) throw AppError.notFound(ErrorDefinitions.SECRET.NOT_FOUND);
    return {
      ...secret,
      value: this.crypto.decryptSecret(secret.value),
    };
  }

  async update(
    id: string,
    patch: {
      name?: string;
      value?: string;
      description?: string;
    },
  ): Promise<Secret> {
    await this.get(id);
    const data: Record<string, unknown> = { ...patch };
    if (typeof patch.value === 'string') {
      data.value = this.crypto.encryptSecret(patch.value);
    }

    const updated = await this.repo.update(id, data);
    return {
      ...updated,
      value: this.crypto.decryptSecret(updated.value),
    };
  }

  async delete(id: string): Promise<Secret> {
    await this.get(id);
    const deleted = await this.repo.delete(id);
    return {
      ...deleted,
      value: this.crypto.decryptSecret(deleted.value),
    };
  }
}
