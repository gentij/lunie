import { Module } from '@nestjs/common';
import { PrismaModule } from 'src/prisma/prisma.module';
import { WorkflowModule } from 'src/workflow/workflow.module';
import { OrchestrationModule } from 'src/core/orchestration.module';
import { CryptoModule } from 'src/crypto/crypto.module';
import { TriggerController } from './trigger.controller';
import { TriggerKeyController } from './trigger-key.controller';
import { TriggerWebhookPublicController } from './trigger-webhook-public.controller';
import { TriggerRepository } from '@lunie/db-access';
import { TriggerService } from './trigger.service';
import { CronTriggerScheduler } from './cron/cron-trigger.scheduler';

@Module({
  imports: [PrismaModule, WorkflowModule, OrchestrationModule, CryptoModule],
  controllers: [
    TriggerController,
    TriggerKeyController,
    TriggerWebhookPublicController,
  ],
  providers: [TriggerService, TriggerRepository, CronTriggerScheduler],
  exports: [TriggerService, TriggerRepository],
})
export class TriggerModule {}
