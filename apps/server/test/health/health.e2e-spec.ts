import { Test } from '@nestjs/testing';
import { APP_FILTER, APP_GUARD, APP_INTERCEPTOR, APP_PIPE } from '@nestjs/core';
import {
  FastifyAdapter,
  NestFastifyApplication,
} from '@nestjs/platform-fastify';
import { ZodSerializerInterceptor, ZodValidationPipe } from 'nestjs-zod';
import { z } from 'zod';

import { HealthController } from 'src/health/health.controller';
import { HealthService } from 'src/health/health.service';
import { AllExceptionsFilter } from 'src/common/http/filters/all-exceptions.filter';
import { ResponseInterceptor } from 'src/common/http/interceptors/response.interceptor';
import { AllowAuthGuard } from 'test/utils/allow-auth.guard';
import { createPrismaServiceMock } from 'test/prisma/prisma.mocks';
import { mockHealthResponse } from 'test/health/health.mocks';
import { ConfigService } from '@nestjs/config';
import { PrismaService } from '@lunie/db-access';

const HealthEnvelopeSchema = z.object({
  ok: z.literal(true),
  data: z.object({
    status: z.literal('ok'),
    version: z.string(),
    uptime: z.number(),
    db: z.object({
      latencyMs: z.number(),
      ok: z.literal(true),
    }),
  }),
});

describe('Health (e2e)', () => {
  let app: NestFastifyApplication;

  beforeAll(async () => {
    const prisma = createPrismaServiceMock();
    prisma.healthInfo.mockResolvedValue(mockHealthResponse.db);

    const moduleFixture = await Test.createTestingModule({
      controllers: [HealthController],
      providers: [
        HealthService,
        {
          provide: ConfigService,
          useValue: {
            get: jest.fn((key: string, fallback?: string) =>
              key === 'VERSION' ? mockHealthResponse.version : fallback,
            ),
          },
        },
        { provide: PrismaService, useValue: prisma },

        { provide: APP_PIPE, useClass: ZodValidationPipe },
        { provide: APP_INTERCEPTOR, useClass: ZodSerializerInterceptor },
        { provide: APP_FILTER, useClass: AllExceptionsFilter },

        { provide: APP_GUARD, useClass: AllowAuthGuard },
        { provide: APP_INTERCEPTOR, useClass: ResponseInterceptor },
      ],
    }).compile();

    app = moduleFixture.createNestApplication<NestFastifyApplication>(
      new FastifyAdapter(),
    );

    await app.init();
    // Required for Fastify in some setups to ensure the underlying server is ready
    await app.getHttpAdapter().getInstance().ready();
  });

  afterAll(async () => {
    await app.close();
  });

  it('GET /health should return ok + metadata', async () => {
    const res = await app.inject({ method: 'GET', url: '/health' });

    expect(res.statusCode).toBe(200);

    const body = HealthEnvelopeSchema.parse(JSON.parse(res.body));
    expect(body.ok).toBe(true);
    expect(body.data).toEqual(
      expect.objectContaining({
        status: 'ok',
        version: expect.any(String) as string,
        uptime: expect.any(Number) as number,
        db: {
          latencyMs: expect.any(Number) as number,
          ok: true,
        },
      }),
    );
  });
});
