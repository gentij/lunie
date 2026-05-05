import { Module } from '@nestjs/common';
import { PrismaModule } from 'src/prisma/prisma.module';
import { WorkflowModule } from 'src/workflow/workflow.module';
import { WorkflowVersionController } from './workflow-version.controller';
import { WorkflowVersionKeyController } from './workflow-version-key.controller';
import { WorkflowVersionRepository } from '@lunie/db-access';
import { WorkflowVersionService } from './workflow-version.service';

@Module({
  imports: [PrismaModule, WorkflowModule],
  controllers: [WorkflowVersionController, WorkflowVersionKeyController],
  providers: [WorkflowVersionService, WorkflowVersionRepository],
  exports: [WorkflowVersionService],
})
export class WorkflowVersionModule {}
