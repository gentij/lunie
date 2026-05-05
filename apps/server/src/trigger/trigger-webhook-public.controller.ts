import { Body, Controller, Param, Post } from '@nestjs/common';
import { ApiTags } from '@nestjs/swagger';
import { AppError } from 'src/common/http/errors/app-error';
import { ErrorDefinitions } from 'src/common/http/errors/error-codes';
import { ApiEnvelope } from 'src/common/swagger/envelope/api-envelope.decorator';
import { OrchestrationService } from 'src/core/orchestration.service';
import { WorkflowService } from 'src/workflow/workflow.service';
import { Public } from 'src/auth/public.decorator';
import { TriggerService } from './trigger.service';
import { TriggerWebhookIngressResDto } from './dto/trigger.dto';

@ApiTags('Webhooks')
@Controller('hooks')
export class TriggerWebhookPublicController {
  constructor(
    private readonly triggerService: TriggerService,
    private readonly orchestrationService: OrchestrationService,
    private readonly workflowService: WorkflowService,
  ) {}

  @ApiEnvelope(TriggerWebhookIngressResDto, {
    description: 'Public webhook ingress',
    errors: [400, 401, 404, 500],
  })
  @Public()
  @Post(':workflowRef/:triggerRef/:webhookKey')
  async handleWebhook(
    @Param('workflowRef') workflowRef: string,
    @Param('triggerRef') triggerRef: string,
    @Param('webhookKey') webhookKey: string,
    @Body() body: unknown,
  ): Promise<{ status: 'accepted' | 'trigger_inactive' }> {
    const { workflow, trigger } = await this.resolveWorkflowAndTrigger(
      workflowRef,
      triggerRef,
    );
    this.triggerService.assertWebhookTriggerType(trigger);

    if (!this.triggerService.hasWebhookKey(trigger)) {
      throw AppError.badRequest(
        ErrorDefinitions.TRIGGER.WEBHOOK_KEY_NOT_CONFIGURED,
      );
    }

    if (!this.triggerService.hasValidWebhookKey(trigger, webhookKey)) {
      throw AppError.unauthorized([
        { field: 'webhookKey', message: 'Invalid webhook key' },
      ]);
    }

    if (!trigger.isActive) {
      return { status: 'trigger_inactive' };
    }

    if (!workflow.latestVersionId) {
      throw new Error('Workflow has no versions');
    }

    const input = normalizeWebhookInput(body);

    await this.orchestrationService.startWorkflow({
      workflowId: workflow.id,
      workflowVersionId: workflow.latestVersionId,
      triggerId: trigger.id,
      eventType: 'WEBHOOK',
      eventPayload: input,
      input,
      overrides: {},
    });

    return { status: 'accepted' };
  }

  private async resolveWorkflowAndTrigger(
    workflowRef: string,
    triggerRef: string,
  ) {
    try {
      const workflow = await this.workflowService.getByKey(workflowRef);
      if (!workflow) {
        throw AppError.notFound(ErrorDefinitions.WORKFLOW.NOT_FOUND);
      }
      const trigger = await this.triggerService.getByKey(
        workflow.id,
        triggerRef,
      );
      if (!trigger) {
        throw AppError.notFound(ErrorDefinitions.TRIGGER.NOT_FOUND);
      }
      return { workflow, trigger };
    } catch (error) {
      if (
        !(error instanceof AppError) ||
        error.code !== ErrorDefinitions.WORKFLOW.NOT_FOUND.code
      ) {
        throw error;
      }
    }

    const workflow = await this.workflowService.get(workflowRef);
    if (!workflow) {
      throw AppError.notFound(ErrorDefinitions.WORKFLOW.NOT_FOUND);
    }
    const trigger = await this.triggerService.get(workflow.id, triggerRef);
    return { workflow, trigger };
  }
}

function normalizeWebhookInput(payload: unknown): Record<string, unknown> {
  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) {
    return payload === undefined ? {} : { payload };
  }

  return payload as Record<string, unknown>;
}
