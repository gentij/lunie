import { ApiTokenRepository } from '@lunie/db-access';

export type ApiTokenRepositoryMock = jest.Mocked<
  Pick<
    ApiTokenRepository,
    'findActive' | 'findByHash' | 'create' | 'updateLastUsed' | 'revoke'
  >
>;

export const createApiTokenRepositoryMock = (): ApiTokenRepositoryMock => ({
  findActive: jest.fn(),
  findByHash: jest.fn(),
  create: jest.fn(),
  updateLastUsed: jest.fn(),
  revoke: jest.fn(),
});
