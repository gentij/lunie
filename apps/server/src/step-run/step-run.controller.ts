import { Controller, Get, Param, Query } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import {
  ApiEnvelope,
  ApiPaginatedEnvelope,
} from 'src/common/swagger/envelope/api-envelope.decorator';
import { StepRunService } from './step-run.service';
import { StepRunListQueryDto, StepRunResDto } from './dto/step-run.dto';

@ApiTags('Step Runs')
@ApiBearerAuth('bearer')
@Controller('workflows/:workflowId/runs/:runId/steps')
export class StepRunController {
  constructor(private readonly service: StepRunService) {}

  @ApiPaginatedEnvelope(StepRunResDto, {
    description: 'List step runs',
    errors: [401, 404, 500],
  })
  @Get()
  list(
    @Param('workflowId') workflowId: string,
    @Param('runId') runId: string,
    @Query() query: StepRunListQueryDto,
  ) {
    return this.service.list({
      workflowId,
      workflowRunId: runId,
      ...query,
    });
  }

  @ApiEnvelope(StepRunResDto, {
    description: 'Get step run',
    errors: [401, 404, 500],
  })
  @Get(':id')
  get(
    @Param('workflowId') workflowId: string,
    @Param('runId') runId: string,
    @Param('id') id: string,
  ) {
    return this.service.get(workflowId, runId, id);
  }
}
