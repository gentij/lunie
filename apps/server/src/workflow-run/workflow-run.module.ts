import { Module } from '@nestjs/common';
import { PrismaModule } from 'src/prisma/prisma.module';
import { WorkflowModule } from 'src/workflow/workflow.module';
import { WorkflowRunController } from './workflow-run.controller';
import { WorkflowRunKeyController } from './workflow-run-key.controller';
import { WorkflowRunRepository } from '@lunie/db-access';
import { WorkflowRunService } from './workflow-run.service';

@Module({
  imports: [PrismaModule, WorkflowModule],
  controllers: [WorkflowRunController, WorkflowRunKeyController],
  providers: [WorkflowRunService, WorkflowRunRepository],
  exports: [WorkflowRunService, WorkflowRunRepository],
})
export class WorkflowRunModule {}
