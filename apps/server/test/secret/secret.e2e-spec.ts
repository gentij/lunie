/* eslint-disable
  @typescript-eslint/no-unsafe-assignment,
  @typescript-eslint/no-unsafe-member-access
*/

import { Test } from '@nestjs/testing';
import { APP_FILTER, APP_GUARD, APP_INTERCEPTOR, APP_PIPE } from '@nestjs/core';
import {
  FastifyAdapter,
  type NestFastifyApplication,
} from '@nestjs/platform-fastify';
import { ZodSerializerInterceptor, ZodValidationPipe } from 'nestjs-zod';

import { SecretController } from 'src/secret/secret.controller';
import { SecretService } from 'src/secret/secret.service';

import { AllExceptionsFilter } from 'src/common/http/filters/all-exceptions.filter';
import { ResponseInterceptor } from 'src/common/http/interceptors/response.interceptor';
import { AllowAuthGuard } from 'test/utils/allow-auth.guard';

import {
  createSecretRepositoryMock,
  type SecretRepositoryMock,
} from 'test/secret/secret.repository.mock';
import { createCryptoServiceMock } from 'test/crypto/crypto.service.mock';
import {
  createSecretFixture,
  createSecretListFixture,
} from 'test/secret/secret.fixtures';
import { SecretRepository } from '@lunie/db-access';
import { CryptoService } from 'src/crypto/crypto.service';

describe('Secret (e2e)', () => {
  let app: NestFastifyApplication;
  let repo: SecretRepositoryMock;

  beforeEach(async () => {
    repo = createSecretRepositoryMock();

    const moduleRef = await Test.createTestingModule({
      controllers: [SecretController],
      providers: [
        SecretService,
        { provide: SecretRepository, useValue: repo },
        { provide: CryptoService, useValue: createCryptoServiceMock() },

        { provide: APP_PIPE, useClass: ZodValidationPipe },
        { provide: APP_INTERCEPTOR, useClass: ZodSerializerInterceptor },
        { provide: APP_FILTER, useClass: AllExceptionsFilter },

        { provide: APP_GUARD, useClass: AllowAuthGuard },
        { provide: APP_INTERCEPTOR, useClass: ResponseInterceptor },
      ],
    }).compile();

    app = moduleRef.createNestApplication<NestFastifyApplication>(
      new FastifyAdapter(),
    );

    await app.init();
    await app.getHttpAdapter().getInstance().ready();
  });

  afterEach(async () => {
    await app.close();
  });

  it('POST /secrets -> 201 creates secret', async () => {
    const created = createSecretFixture({ name: 'API_KEY' });
    repo.create.mockResolvedValue(created);

    const res = await app.inject({
      method: 'POST',
      url: '/secrets',
      payload: { name: 'API_KEY', value: 'secret-value' },
    });

    expect(res.statusCode).toBe(201);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.name).toBe('API_KEY');
  });

  it('GET /secrets -> 200 + data array', async () => {
    const list = createSecretListFixture(2);
    repo.findPage.mockResolvedValue({ items: list, total: 2 });

    const res = await app.inject({ method: 'GET', url: '/secrets' });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(Array.isArray(body.data.items)).toBe(true);
    expect(body.data.items).toHaveLength(2);
    expect(body.data.pagination.total).toBe(2);
  });

  it('GET /secrets/:id -> 200 when found', async () => {
    const secret = createSecretFixture({ id: 'sec_1' });
    repo.findById.mockResolvedValue(secret);

    const res = await app.inject({ method: 'GET', url: '/secrets/sec_1' });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.id).toBe('sec_1');
  });

  it('PATCH /secrets/:id -> 200 updates secret', async () => {
    const secret = createSecretFixture({ id: 'sec_1' });
    const updated = createSecretFixture({ id: 'sec_1', name: 'UPDATED' });

    repo.findById.mockResolvedValue(secret);
    repo.update.mockResolvedValue(updated);

    const res = await app.inject({
      method: 'PATCH',
      url: '/secrets/sec_1',
      payload: { name: 'UPDATED' },
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.name).toBe('UPDATED');
  });

  it('DELETE /secrets/:id -> 200 deletes secret', async () => {
    const secret = createSecretFixture({ id: 'sec_1' });
    repo.findById.mockResolvedValue(secret);
    repo.delete.mockResolvedValue(secret);

    const res = await app.inject({
      method: 'DELETE',
      url: '/secrets/sec_1',
    });

    expect(res.statusCode).toBe(200);

    const body = res.json();
    expect(body.ok).toBe(true);
    expect(body.data.id).toBe('sec_1');
  });

  it('GET /secrets/:id -> 404 when missing', async () => {
    repo.findById.mockResolvedValue(null);

    const res = await app.inject({ method: 'GET', url: '/secrets/missing' });

    expect(res.statusCode).toBe(404);

    const body = res.json();
    expect(body.ok).toBe(false);
    expect(body.error).toBeDefined();
  });
});
