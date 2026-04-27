import { APP_PIPE, APP_INTERCEPTOR, APP_GUARD } from '@nestjs/core';
import { Module } from '@nestjs/common';
import { CacheModule } from '@nestjs/cache-manager';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ZodValidationPipe, ZodSerializerInterceptor } from 'nestjs-zod';
import { LoggerModule } from 'nestjs-pino';
import { ScheduleModule } from '@nestjs/schedule';
import { ThrottlerModule } from '@nestjs/throttler';
import { ThrottlerBehindProxyGuard } from './common/http/guards/throttler-behind-proxy.guard';
import Keyv from 'keyv';
import KeyvRedis from '@keyv/redis';

import { validateEnv } from './config/env';
import { HealthModule } from './health/health.module';
import { CoreModule } from './core/core.module';
import { PrismaModule } from './prisma/prisma.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      validate: validateEnv,
    }),
    CacheModule.registerAsync({
      isGlobal: true,
      inject: [ConfigService],
      useFactory: (config: ConfigService) => ({
        stores: new Keyv({
          store: new KeyvRedis(config.get<string>('REDIS_URL') ?? ''),
          ttl: (config.get<number>('CACHE_TTL_SECONDS') ?? 60) * 1000,
          namespace:
            config.get<string>('CACHE_REDIS_PREFIX') ?? 'lunie:server:',
        }),
        ttl: (config.get<number>('CACHE_TTL_SECONDS') ?? 60) * 1000,
        namespace: config.get<string>('CACHE_REDIS_PREFIX') ?? 'lunie:server:',
      }),
    }),
    ScheduleModule.forRoot(),
    ThrottlerModule.forRoot([
      {
        ttl: 60_000,
        limit: 60,
      },
    ]),
    LoggerModule.forRoot(),
    CoreModule,
    PrismaModule,
    HealthModule,
  ],
  providers: [
    {
      provide: APP_PIPE,
      useClass: ZodValidationPipe,
    },
    {
      provide: APP_INTERCEPTOR,
      useClass: ZodSerializerInterceptor,
    },
    {
      provide: APP_GUARD,
      useClass: ThrottlerBehindProxyGuard,
    },
  ],
})
export class AppModule {}
