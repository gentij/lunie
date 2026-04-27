import { Module } from '@nestjs/common';
import { PrismaModule } from 'src/prisma/prisma.module';
import { WorkflowRunModule } from 'src/workflow-run/workflow-run.module';
import { StepRunController } from './step-run.controller';
import { StepRunRepository } from '@lunie/db-access';
import { StepRunService } from './step-run.service';

@Module({
  imports: [PrismaModule, WorkflowRunModule],
  controllers: [StepRunController],
  providers: [StepRunService, StepRunRepository],
  exports: [StepRunService],
})
export class StepRunModule {}
