import { Module } from '@nestjs/common';
import { BullModule } from '@nestjs/bullmq';
import { STEP_RUN_QUEUE_NAME, QueueConfigModule } from '@lunie/queue-config';
import { StepRunQueueService } from './step-run-queue.service';

@Module({
  imports: [
    QueueConfigModule,
    BullModule.registerQueue({ name: STEP_RUN_QUEUE_NAME }),
  ],
  providers: [StepRunQueueService],
  exports: [StepRunQueueService],
})
export class QueueModule {}
