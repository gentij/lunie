import { Test } from '@nestjs/testing';
import { ApiTokenService } from './api-token.service';
import { ApiTokenRepository } from '@lunie/db-access';
import {
  createApiTokenRepositoryMock,
  ApiTokenRepositoryMock,
} from 'test/api-token/api-token.repository.mock';

import {
  createApiTokenFixture,
  createRevokedApiTokenFixture,
  createApiTokenListFixture,
} from 'test/api-token/api-token.service.mock';
import { AppError } from 'src/common/http/errors/app-error';

describe('ApiTokenService', () => {
  let service: ApiTokenService;
  let repo: ApiTokenRepositoryMock;

  beforeEach(async () => {
    repo = createApiTokenRepositoryMock();

    const moduleRef = await Test.createTestingModule({
      providers: [
        ApiTokenService,
        { provide: ApiTokenRepository, useValue: repo },
      ],
    }).compile();

    service = moduleRef.get(ApiTokenService);
  });

  describe('hasAnyActiveToken', () => {
    it('returns false when no active tokens exist', async () => {
      repo.findActive.mockResolvedValue([]);

      await expect(service.hasAnyActiveToken()).resolves.toBe(false);
      expect(repo.findActive).toHaveBeenCalledTimes(1);
    });

    it('returns true when active tokens exist', async () => {
      repo.findActive.mockResolvedValue(createApiTokenListFixture(1));

      await expect(service.hasAnyActiveToken()).resolves.toBe(true);
      expect(repo.findActive).toHaveBeenCalledTimes(1);
    });
  });

  describe('createAdminToken', () => {
    it('creates token with empty scopes (admin)', async () => {
      const created = createApiTokenFixture({
        id: 't1',
        name: 'initial-admin',
        tokenHash: 'hashed',
        scopes: [],
      });

      repo.create.mockResolvedValue(created);

      const result = await service.createAdminToken({
        name: 'initial-admin',
        tokenHash: 'hashed',
      });

      expect(repo.create).toHaveBeenCalledWith({
        name: 'initial-admin',
        tokenHash: 'hashed',
        scopes: [],
      });
      expect(result).toBe(created);
    });
  });

  describe('validateTokenHash', () => {
    it('returns null if token is not found', async () => {
      repo.findByHash.mockResolvedValue(null);

      const result = await service.validateTokenHash('hash');
      expect(result).toBeNull();
      expect(repo.updateLastUsed).not.toHaveBeenCalled();
    });

    it('throws unauthorized if token is revoked', async () => {
      repo.findByHash.mockResolvedValue(
        createRevokedApiTokenFixture({
          id: 't1',
          tokenHash: 'hash',
        }),
      );

      await expect(service.validateTokenHash('hash')).rejects.toBeInstanceOf(
        AppError,
      );
      expect(repo.updateLastUsed).not.toHaveBeenCalled();
    });

    it('returns token and updates lastUsedAt if token is active', async () => {
      const token = createApiTokenFixture({
        id: 't1',
        name: 'active',
        tokenHash: 'hash',
      });

      repo.findByHash.mockResolvedValue(token);
      repo.updateLastUsed.mockResolvedValue(
        createApiTokenFixture({
          ...token,
          lastUsedAt: new Date(),
        }),
      );

      const result = await service.validateTokenHash('hash');

      expect(result).toBe(token);
      expect(repo.updateLastUsed).toHaveBeenCalledWith('t1');
    });

    it('does not throw if updateLastUsed fails (fire-and-forget)', async () => {
      const token = createApiTokenFixture({
        id: 't1',
        name: 'active',
        tokenHash: 'hash',
      });

      repo.findByHash.mockResolvedValue(token);
      repo.updateLastUsed.mockRejectedValue(new Error('db down'));

      await expect(service.validateTokenHash('hash')).resolves.toBe(token);
      expect(repo.updateLastUsed).toHaveBeenCalledWith('t1');
    });
  });

  describe('revokeToken', () => {
    it('delegates revoke to repository', async () => {
      const revoked = createRevokedApiTokenFixture({
        id: 't1',
        tokenHash: 'hash',
      });

      repo.revoke.mockResolvedValue(revoked);

      const result = await service.revokeToken('t1');

      expect(repo.revoke).toHaveBeenCalledWith('t1');
      expect(result).toBe(revoked);
    });
  });
});
