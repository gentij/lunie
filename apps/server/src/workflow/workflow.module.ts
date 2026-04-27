import { Module } from '@nestjs/common';
import { SecretRepository, WorkflowRepository } from '@lunie/db-access';
import { PrismaModule } from 'src/prisma/prisma.module';
import { OrchestrationModule } from 'src/core/orchestration.module';
import { WorkflowController } from './workflow.controller';
import { WorkflowService } from './workflow.service';

@Module({
  imports: [PrismaModule, OrchestrationModule],
  controllers: [WorkflowController],
  providers: [WorkflowService, WorkflowRepository, SecretRepository],
  exports: [WorkflowService, WorkflowRepository],
})
export class WorkflowModule {}
