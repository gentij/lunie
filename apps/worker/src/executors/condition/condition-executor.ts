import { Injectable, Logger } from '@nestjs/common';
import * as jmespath from 'jmespath';
import type { ExecutorOutput, StepExecutor } from '../executor.interface';
import { ConditionExecutorInputSchema } from './condition.types';

type RootContext = {
  input: unknown;
  source: Record<string, unknown>;
  steps: Record<string, unknown>;
  stepResponses: Record<string, unknown>;
};

@Injectable()
export class ConditionExecutor implements StepExecutor {
  readonly stepType = 'condition';
  private readonly logger = new Logger(ConditionExecutor.name);

  async execute(input: unknown): Promise<ExecutorOutput> {
    const validated = ConditionExecutorInputSchema.parse(input);
    const request = validated.request;
    const assert = request.assert ?? true;

    const ctx = validated.input as any;
    const stepResponses = (ctx?.steps ?? {}) as Record<string, unknown>;
    const steps = unwrapStepBodies(stepResponses);
    const source = request.source ?? {};

    const root: RootContext = {
      input: ctx?.input ?? {},
      source,
      steps,
      stepResponses,
    };

    let value: unknown;
    try {
      value = jmespath.search(root as any, request.expr);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      this.logger.error(`JMESPath evaluation failed: ${message}`);
      // Recommended behavior: treat missing/malformed as falsy.
      value = null;
    }

    const passed = isJmesTruthy(value);

    if (assert && !passed) {
      const msg = request.message ? `: ${request.message}` : '';
      throw new Error(`Condition failed${msg}`);
    }

    return {
      statusCode: 200,
      headers: undefined,
      body: {
        passed,
        value,
      },
    };
  }
}

function isJmesTruthy(value: unknown): boolean {
  if (value === null || value === undefined) return false;
  if (value === false) return false;
  if (typeof value === 'string') return value.length > 0;
  if (Array.isArray(value)) return value.length > 0;
  if (typeof value === 'object') {
    return Object.keys(value as Record<string, unknown>).length > 0;
  }
  return true;
}

function unwrapStepBodies(stepResponses: Record<string, unknown>): Record<string, unknown> {
  const out: Record<string, unknown> = {};

  for (const [key, value] of Object.entries(stepResponses)) {
    if (value && typeof value === 'object' && 'body' in (value as any)) {
      const body = (value as any).body;
      // Unwrap HttpExecutor body wrapper if present.
      if (body && typeof body === 'object' && '_lunieHttp' in body && 'data' in body) {
        out[key] = body.data;
      } else {
        out[key] = body;
      }
    } else {
      out[key] = value;
    }
  }

  return out;
}
