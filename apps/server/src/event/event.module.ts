import { Module } from '@nestjs/common';
import { PrismaModule } from 'src/prisma/prisma.module';
import { WorkflowModule } from 'src/workflow/workflow.module';
import { TriggerModule } from 'src/trigger/trigger.module';
import { EventController } from './event.controller';
import { EventRepository } from '@lunie/db-access';
import { EventService } from './event.service';

@Module({
  imports: [PrismaModule, WorkflowModule, TriggerModule],
  controllers: [EventController],
  providers: [EventService, EventRepository],
  exports: [EventService],
})
export class EventModule {}
