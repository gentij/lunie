import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { BullModule } from '@nestjs/bullmq';
import { STEP_RUN_QUEUE_NAME, QueueConfigModule } from '@taskforge/queue-config';
import { PrismaModule } from './prisma/prisma.module';
import { CryptoModule } from './crypto/crypto.module';
import { RedisModule } from './redis/redis.module';
import {
  StepRunRepository,
  SecretRepository,
  WorkflowRunRepository,
  WorkflowVersionRepository,
} from '@taskforge/db-access';
import { ExecutorRegistry } from './executors/executor-registry';
import { HttpExecutorModule } from './executors/http/http-executor.module';
import { TransformExecutorModule } from './executors/transform/transform-executor.module';
import { ConditionExecutorModule } from './executors/condition/condition-executor.module';
import { StepRunProcessor } from './processors/step-run.processor';
import { WorkerCacheService } from './cache/worker-cache.service';
import { RunNotifierService } from './notifications/run-notifier.service';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: '.env',
    }),
    PrismaModule,
    CryptoModule,
    RedisModule,
    QueueConfigModule,
    BullModule.registerQueue({ name: STEP_RUN_QUEUE_NAME }),
    HttpExecutorModule,
    TransformExecutorModule,
    ConditionExecutorModule,
  ],
  providers: [
    StepRunProcessor,
    StepRunRepository,
    SecretRepository,
    WorkflowRunRepository,
    WorkflowVersionRepository,
    ExecutorRegistry,
    WorkerCacheService,
    RunNotifierService,
  ],
  exports: [],
})
export class WorkerModule {}
