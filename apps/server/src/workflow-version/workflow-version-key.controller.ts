import { Controller, Get, Param, ParseIntPipe, Query } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import {
  ApiEnvelope,
  ApiPaginatedEnvelope,
} from 'src/common/swagger/envelope/api-envelope.decorator';
import { WorkflowVersionService } from './workflow-version.service';
import {
  WorkflowVersionListQueryDto,
  WorkflowVersionResDto,
} from './dto/workflow-version.dto';
import { WorkflowService } from 'src/workflow/workflow.service';

@ApiTags('Workflow Versions')
@ApiBearerAuth('bearer')
@Controller('workflows/by-key/:workflowKey/versions')
export class WorkflowVersionKeyController {
  constructor(
    private readonly service: WorkflowVersionService,
    private readonly workflowService: WorkflowService,
  ) {}

  @ApiPaginatedEnvelope(WorkflowVersionResDto, {
    description: 'List workflow versions by workflow key',
    errors: [401, 404, 500],
  })
  @Get()
  async list(
    @Param('workflowKey') workflowKey: string,
    @Query() query: WorkflowVersionListQueryDto,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    return this.service.list({ workflowId: workflow.id, ...query });
  }

  @ApiEnvelope(WorkflowVersionResDto, {
    description: 'Get workflow version by workflow key',
    errors: [401, 404, 500],
  })
  @Get(':version')
  async get(
    @Param('workflowKey') workflowKey: string,
    @Param('version', ParseIntPipe) version: number,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    return this.service.get(workflow.id, version);
  }
}
