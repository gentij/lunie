import { Controller, Get, Param, ParseIntPipe, Query } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import {
  ApiEnvelope,
  ApiPaginatedEnvelope,
} from 'src/common/swagger/envelope/api-envelope.decorator';
import { StepRunService } from './step-run.service';
import { StepRunListQueryDto, StepRunResDto } from './dto/step-run.dto';
import { WorkflowService } from 'src/workflow/workflow.service';
import { WorkflowRunService } from 'src/workflow-run/workflow-run.service';

@ApiTags('Step Runs')
@ApiBearerAuth('bearer')
@Controller('workflows/by-key/:workflowKey/runs/:runNumber/steps')
export class StepRunKeyController {
  constructor(
    private readonly service: StepRunService,
    private readonly workflowService: WorkflowService,
    private readonly workflowRunService: WorkflowRunService,
  ) {}

  @ApiPaginatedEnvelope(StepRunResDto, {
    description: 'List step runs by workflow key and run number',
    errors: [401, 404, 500],
  })
  @Get()
  async list(
    @Param('workflowKey') workflowKey: string,
    @Param('runNumber', ParseIntPipe) runNumber: number,
    @Query() query: StepRunListQueryDto,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    const run = await this.workflowRunService.getByNumber(
      workflow.id,
      runNumber,
    );
    return this.service.list({
      workflowId: workflow.id,
      workflowRunId: run.id,
      ...query,
    });
  }

  @ApiEnvelope(StepRunResDto, {
    description: 'Get step run by workflow key, run number, and step key',
    errors: [401, 404, 500],
  })
  @Get(':stepKey')
  async get(
    @Param('workflowKey') workflowKey: string,
    @Param('runNumber', ParseIntPipe) runNumber: number,
    @Param('stepKey') stepKey: string,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    const run = await this.workflowRunService.getByNumber(
      workflow.id,
      runNumber,
    );
    return this.service.getByStepKey(workflow.id, run.id, stepKey);
  }
}
