import {
  Body,
  Controller,
  Delete,
  Get,
  Param,
  Patch,
  Post,
  Query,
} from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import {
  ApiEnvelope,
  ApiPaginatedEnvelope,
} from 'src/common/swagger/envelope/api-envelope.decorator';
import { WorkflowService } from './workflow.service';
import {
  CreateWorkflowReqDto,
  UpdateWorkflowReqDto,
  WorkflowResDto,
  RunWorkflowReqDto,
  RunWorkflowResDto,
  WorkflowListQueryDto,
  parseRunWorkflowReq,
} from './dto/workflow.dto';
import {
  CreateWorkflowVersionReqDto,
  WorkflowVersionResDto,
} from 'src/workflow-version/dto/workflow-version.dto';
import { OrchestrationService } from 'src/core/orchestration.service';
import {
  ValidateWorkflowDefinitionReqDto,
  ValidateWorkflowDefinitionResDto,
} from './dto/workflow-validation.dto';

@ApiTags('Workflows')
@ApiBearerAuth('bearer')
@Controller('workflows')
export class WorkflowController {
  constructor(
    private readonly service: WorkflowService,
    private readonly orchestrationService: OrchestrationService,
  ) {}

  @ApiEnvelope(WorkflowResDto, {
    description: 'Create workflow',
    errors: [401, 500],
  })
  @Post()
  create(@Body() body: CreateWorkflowReqDto) {
    return this.service.create({
      name: body.name,
      definition: body.definition,
    });
  }

  @ApiPaginatedEnvelope(WorkflowResDto, {
    description: 'List workflows',
    errors: [401, 500],
  })
  @Get()
  list(@Query() query: WorkflowListQueryDto) {
    return this.service.list(query);
  }

  @ApiEnvelope(WorkflowResDto, {
    description: 'Get workflow',
    errors: [401, 404, 500],
  })
  @Get(':id')
  get(@Param('id') id: string) {
    return this.service.get(id);
  }

  @ApiEnvelope(WorkflowResDto, {
    description: 'Update workflow',
    errors: [401, 404, 500],
  })
  @Patch(':id')
  update(@Param('id') id: string, @Body() body: UpdateWorkflowReqDto) {
    return this.service.update(id, body);
  }

  @ApiEnvelope(WorkflowResDto, {
    description: 'Delete workflow (soft)',
    errors: [401, 404, 500],
  })
  @Delete(':id')
  delete(@Param('id') id: string) {
    return this.service.delete(id);
  }

  @ApiEnvelope(WorkflowVersionResDto, {
    description: 'Create workflow version',
    errors: [401, 404, 500],
  })
  @Post(':id/versions')
  createVersion(
    @Param('id') id: string,
    @Body() body: CreateWorkflowVersionReqDto,
  ) {
    return this.service.createVersion(id, body.definition);
  }

  @ApiEnvelope(ValidateWorkflowDefinitionResDto, {
    description: 'Validate workflow definition (no persist)',
    errors: [401, 404, 500],
  })
  @Post(':id/versions/validate')
  async validateVersionDefinition(
    @Param('id') id: string,
    @Body() body: ValidateWorkflowDefinitionReqDto,
  ) {
    // Ensure workflow exists and the caller has access.
    await this.service.get(id);

    const result = this.service.validateDefinition(body.definition);
    if (result.referencedSecrets.length > 0) {
      await this.service.validateDefinitionOrThrow(body.definition);
    }
    return {
      valid: result.valid,
      issues: result.issues,
      inferredDependencies: result.inferredDependencies,
      executionBatches: result.executionBatches,
      referencedSecrets: result.referencedSecrets,
    };
  }

  @ApiEnvelope(RunWorkflowResDto, {
    description: 'Start workflow manually',
    errors: [401, 404, 500],
  })
  @Post(':id/run')
  async runManual(
    @Param('id') workflowId: string,
    @Body() body: RunWorkflowReqDto,
  ) {
    const { input, overrides } = parseRunWorkflowReq(body);

    const workflow = await this.service.get(workflowId);

    if (!workflow.latestVersionId) {
      throw new Error('Workflow has no versions');
    }

    const { workflowRunId, workflowRunNumber } =
      await this.orchestrationService.startWorkflow({
        workflowId,
        workflowVersionId: workflow.latestVersionId,
        eventType: 'MANUAL',
        input,
        overrides,
      });

    return { workflowRunId, workflowRunNumber, status: 'QUEUED' };
  }
}
