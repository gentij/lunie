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
import type { Prisma } from '@prisma/client';
import { TriggerService } from './trigger.service';
import {
  CreateTriggerReqDto,
  RotateWebhookKeyResDto,
  TriggerListQueryDto,
  TriggerResDto,
  UpdateTriggerReqDto,
} from './dto/trigger.dto';
import { WorkflowService } from 'src/workflow/workflow.service';
import { OrchestrationService } from 'src/core/orchestration.service';
import {
  RunWorkflowReqDto,
  parseRunWorkflowReq,
} from 'src/workflow/dto/workflow.dto';

@ApiTags('Triggers')
@ApiBearerAuth('bearer')
@Controller('workflows/by-key/:workflowKey/triggers')
export class TriggerKeyController {
  constructor(
    private readonly service: TriggerService,
    private readonly workflowService: WorkflowService,
    private readonly orchestrationService: OrchestrationService,
  ) {}

  @ApiEnvelope(TriggerResDto, {
    description: 'Create trigger by workflow key',
    errors: [401, 404, 500],
  })
  @Post()
  async create(
    @Param('workflowKey') workflowKey: string,
    @Body() body: CreateTriggerReqDto,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    return this.service.create({
      workflowId: workflow.id,
      type: body.type,
      name: body.name,
      isActive: body.isActive,
      config: body.config as Prisma.InputJsonValue,
    });
  }

  @ApiPaginatedEnvelope(TriggerResDto, {
    description: 'List triggers by workflow key',
    errors: [401, 404, 500],
  })
  @Get()
  async list(
    @Param('workflowKey') workflowKey: string,
    @Query() query: TriggerListQueryDto,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    return this.service.list({ workflowId: workflow.id, ...query });
  }

  @ApiEnvelope(TriggerResDto, {
    description: 'Get trigger by workflow key and trigger key',
    errors: [401, 404, 500],
  })
  @Get('by-key/:triggerKey')
  async get(
    @Param('workflowKey') workflowKey: string,
    @Param('triggerKey') triggerKey: string,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    return this.service.getByKey(workflow.id, triggerKey);
  }

  @ApiEnvelope(TriggerResDto, {
    description: 'Update trigger by workflow key and trigger key',
    errors: [401, 404, 500],
  })
  @Patch('by-key/:triggerKey')
  async update(
    @Param('workflowKey') workflowKey: string,
    @Param('triggerKey') triggerKey: string,
    @Body() body: UpdateTriggerReqDto,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    const trigger = await this.service.getByKey(workflow.id, triggerKey);
    return this.service.update(workflow.id, trigger.id, {
      name: body.name,
      isActive: body.isActive,
      config:
        body.config === undefined
          ? undefined
          : (body.config as Prisma.InputJsonValue),
    });
  }

  @ApiEnvelope(TriggerResDto, {
    description: 'Delete trigger by workflow key and trigger key',
    errors: [401, 404, 500],
  })
  @Delete('by-key/:triggerKey')
  async delete(
    @Param('workflowKey') workflowKey: string,
    @Param('triggerKey') triggerKey: string,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    const trigger = await this.service.getByKey(workflow.id, triggerKey);
    return this.service.delete(workflow.id, trigger.id);
  }

  @ApiEnvelope(RotateWebhookKeyResDto, {
    description: 'Rotate webhook key by workflow key and trigger key',
    errors: [400, 401, 404, 500],
  })
  @Post('by-key/:triggerKey/webhook-key/rotate')
  async rotateWebhookKey(
    @Param('workflowKey') workflowKey: string,
    @Param('triggerKey') triggerKey: string,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    const trigger = await this.service.getByKey(workflow.id, triggerKey);
    return this.service.rotateWebhookKey(workflow.id, trigger.id);
  }

  @ApiTags('Triggers')
  @Post('by-key/:triggerKey/webhook')
  async handleWebhook(
    @Param('workflowKey') workflowKey: string,
    @Param('triggerKey') triggerKey: string,
    @Body() body: RunWorkflowReqDto,
  ) {
    const workflow = await this.workflowService.getByKey(workflowKey);
    const trigger = await this.service.getByKey(workflow.id, triggerKey);
    const { input, overrides } = parseRunWorkflowReq(body);

    if (!trigger.isActive) {
      return { status: 'trigger_inactive' };
    }

    if (!workflow.latestVersionId) {
      throw new Error('Workflow has no versions');
    }

    await this.orchestrationService.startWorkflow({
      workflowId: workflow.id,
      workflowVersionId: workflow.latestVersionId,
      triggerId: trigger.id,
      eventType: 'WEBHOOK',
      eventPayload: input,
      input,
      overrides,
    });

    return { status: 'accepted' };
  }
}
