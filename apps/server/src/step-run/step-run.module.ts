import { Module } from '@nestjs/common';
import { PrismaModule } from 'src/prisma/prisma.module';
import { WorkflowRunModule } from 'src/workflow-run/workflow-run.module';
import { StepRunController } from './step-run.controller';
import { WorkflowModule } from 'src/workflow/workflow.module';
import { StepRunRepository } from '@lunie/db-access';
import { StepRunService } from './step-run.service';
import { StepRunKeyController } from './step-run-key.controller';

@Module({
  imports: [PrismaModule, WorkflowRunModule, WorkflowModule],
  controllers: [StepRunController, StepRunKeyController],
  providers: [StepRunService, StepRunRepository],
  exports: [StepRunService],
})
export class StepRunModule {}
