import { Test } from '@nestjs/testing';
import { SecretService } from './secret.service';
import { SecretRepository } from '@lunie/db-access';
import {
  createSecretRepositoryMock,
  type SecretRepositoryMock,
} from 'test/secret/secret.repository.mock';
import {
  createSecretFixture,
  createSecretListFixture,
} from 'test/secret/secret.fixtures';
import { AppError } from 'src/common/http/errors/app-error';
import { CryptoService } from 'src/crypto/crypto.service';
import { ConfigService } from '@nestjs/config';

function createConfigMock() {
  return { get: jest.fn().mockReturnValue('0'.repeat(64)) };
}

describe('SecretService', () => {
  let service: SecretService;
  let repo: SecretRepositoryMock;

  beforeEach(async () => {
    repo = createSecretRepositoryMock();
    const config = createConfigMock();
    const crypto = new CryptoService(config as unknown as ConfigService);

    const moduleRef = await Test.createTestingModule({
      providers: [
        SecretService,
        { provide: SecretRepository, useValue: repo },
        { provide: CryptoService, useValue: crypto },
        {
          provide: ConfigService,
          useValue: config as unknown as ConfigService,
        },
      ],
    }).compile();

    service = moduleRef.get(SecretService);
  });

  it('create() creates a secret', async () => {
    const created = createSecretFixture({ name: 'API_KEY' });
    repo.create.mockResolvedValue(created);

    await service.create({ name: 'API_KEY', value: 'secret-value' });

    expect(repo.create).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'API_KEY',
        description: undefined,
        value: expect.stringMatching(/^luniesec:v1:/) as unknown as string,
      }),
    );
  });

  it('list() returns secrets', async () => {
    const list = createSecretListFixture(2);
    repo.findPage.mockResolvedValue({ items: list, total: 2 });

    await expect(service.list({ page: 1, pageSize: 25 })).resolves.toEqual({
      items: list,
      pagination: {
        page: 1,
        pageSize: 25,
        total: 2,
        totalPages: 1,
        hasNext: false,
        hasPrev: false,
      },
    });
    expect(repo.findPage).toHaveBeenCalledWith({ page: 1, pageSize: 25 });
  });

  it('get() returns secret when found', async () => {
    const crypto = new CryptoService(
      createConfigMock() as unknown as ConfigService,
    );
    const secret = createSecretFixture({ id: 'sec_1' });
    const encrypted = crypto.encryptSecret(secret.value);
    repo.findById.mockResolvedValue({ ...secret, value: encrypted });

    await expect(service.get('sec_1')).resolves.toStrictEqual(secret);
    expect(repo.findById).toHaveBeenCalledWith('sec_1');
  });

  it('get() throws notFound when secret missing', async () => {
    repo.findById.mockResolvedValue(null);

    await expect(service.get('missing')).rejects.toBeInstanceOf(AppError);
  });

  it('update() updates secret after existence check', async () => {
    const crypto = new CryptoService(
      createConfigMock() as unknown as ConfigService,
    );
    const secret = createSecretFixture({ id: 'sec_1' });
    const encrypted = crypto.encryptSecret(secret.value);
    const updated = createSecretFixture({ id: 'sec_1', name: 'UPDATED' });

    repo.findById.mockResolvedValue({ ...secret, value: encrypted });
    repo.update.mockResolvedValue({ ...updated, value: encrypted });

    await expect(
      service.update('sec_1', { name: 'UPDATED' }),
    ).resolves.toStrictEqual(updated);

    expect(repo.update).toHaveBeenCalledWith(
      'sec_1',
      expect.objectContaining({ name: 'UPDATED' }),
    );
  });

  it('delete() deletes secret after existence check', async () => {
    const crypto = new CryptoService(
      createConfigMock() as unknown as ConfigService,
    );
    const secret = createSecretFixture({ id: 'sec_1' });
    const encrypted = crypto.encryptSecret(secret.value);
    repo.findById.mockResolvedValue({ ...secret, value: encrypted });
    repo.delete.mockResolvedValue({ ...secret, value: encrypted });

    await expect(service.delete('sec_1')).resolves.toStrictEqual(secret);
    expect(repo.delete).toHaveBeenCalledWith('sec_1');
  });
});
