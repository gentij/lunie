import { Injectable } from '@nestjs/common';
import { ApiTokenRepository } from '@lunie/db-access';
import { ApiToken } from '@prisma/client';
import { AppError } from 'src/common/http/errors/app-error';
import { ErrorDefinitions } from 'src/common/http/errors/error-codes';

@Injectable()
export class ApiTokenService {
  constructor(private readonly repo: ApiTokenRepository) {}

  async hasAnyActiveToken(): Promise<boolean> {
    const tokens = await this.repo.findActive();
    return tokens.length > 0;
  }

  async createAdminToken(params: {
    name: string;
    tokenHash: string;
  }): Promise<ApiToken> {
    return this.repo.create({
      name: params.name,
      tokenHash: params.tokenHash,
      scopes: [], // empty = full access
    });
  }

  async validateTokenHash(tokenHash: string): Promise<ApiToken | null> {
    const token = await this.repo.findByHash(tokenHash);

    if (!token) return null;

    if (token.revokedAt)
      throw AppError.unauthorized(ErrorDefinitions.AUTH.REVOKED_TOKEN);

    void this.repo.updateLastUsed(token.id).catch(() => undefined);

    return token;
  }

  async revokeToken(id: string): Promise<ApiToken> {
    return this.repo.revoke(id);
  }
}
