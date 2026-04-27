import { Module } from '@nestjs/common';
import { PrismaModule } from 'src/prisma/prisma.module';
import { WorkflowModule } from 'src/workflow/workflow.module';
import { WorkflowVersionController } from './workflow-version.controller';
import { WorkflowVersionRepository } from '@lunie/db-access';
import { WorkflowVersionService } from './workflow-version.service';

@Module({
  imports: [PrismaModule, WorkflowModule],
  controllers: [WorkflowVersionController],
  providers: [WorkflowVersionService, WorkflowVersionRepository],
  exports: [WorkflowVersionService],
})
export class WorkflowVersionModule {}
