import { Controller, Get, Param, ParseIntPipe, Query } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import {
  ApiEnvelope,
  ApiPaginatedEnvelope,
} from 'src/common/swagger/envelope/api-envelope.decorator';
import { WorkflowRunService } from './workflow-run.service';
import {
  WorkflowRunListQueryDto,
  WorkflowRunResDto,
} from './dto/workflow-run.dto';
import { WorkflowService } from 'src/workflow/workflow.service';

@ApiTags('Workflow Runs')
@ApiBearerAuth('bearer')
@Controller('workflows/by-key/:workflowKey/runs')
export class WorkflowRunKeyController {
  constructor(
    private readonly service: WorkflowRunService,
    private readonly workflowService: WorkflowService,
  ) {}

  @ApiPaginatedEnvelope(WorkflowRunResDto, {
    description: 'List workflow runs by workflow key',
    errors: [401, 404, 500],
  })
  @Get()
  async list(
    @Param('workflowKey') workflowKey: string,
    @Query() query: WorkflowRunListQueryDto,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    return this.service.list({ workflowId: workflow.id, ...query });
  }

  @ApiEnvelope(WorkflowRunResDto, {
    description: 'Get workflow run by workflow key and run number',
    errors: [401, 404, 500],
  })
  @Get(':runNumber')
  async get(
    @Param('workflowKey') workflowKey: string,
    @Param('runNumber', ParseIntPipe) runNumber: number,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    return this.service.getByNumber(workflow.id, runNumber);
  }
}
