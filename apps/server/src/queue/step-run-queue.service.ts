import { Injectable } from '@nestjs/common';
import { InjectQueue } from '@nestjs/bullmq';
import type { JobsOptions, Queue } from 'bullmq';
import { StepRunJobPayload, StepRunJobPayloadSchema } from '@lunie/contracts';
import { STEP_RUN_QUEUE_NAME } from './queue.constants';

const DEFAULT_ATTEMPTS = 3;
const DEFAULT_BACKOFF_DELAY_MS = 5000;
const DEFAULT_REMOVE_ON_COMPLETE = 1000;
const DEFAULT_REMOVE_ON_FAIL = 1000;

@Injectable()
export class StepRunQueueService {
  constructor(
    @InjectQueue(STEP_RUN_QUEUE_NAME)
    private readonly queue: Queue,
  ) {}

  async enqueueHttpStepRun(payload: StepRunJobPayload, options?: JobsOptions) {
    return this.enqueueStepRun('http', payload, options);
  }

  async enqueueStepRun(
    stepType: string,
    payload: StepRunJobPayload,
    options?: JobsOptions,
  ) {
    const normalized = StepRunJobPayloadSchema.parse(payload);

    return this.queue.add(stepType, normalized, {
      jobId: normalized.stepRunId,
      attempts: DEFAULT_ATTEMPTS,
      backoff: { type: 'exponential', delay: DEFAULT_BACKOFF_DELAY_MS },
      removeOnComplete: DEFAULT_REMOVE_ON_COMPLETE,
      removeOnFail: DEFAULT_REMOVE_ON_FAIL,
      ...options,
    });
  }
}
