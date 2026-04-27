interface ResolutionContext {
  input: Record<string, unknown>;
  steps: Record<string, unknown>;
  secret?: Record<string, string>;
}

interface ResolutionResult {
  resolved: unknown;
  referencedSteps: string[];
}

export class TemplateResolver {
  resolve(value: unknown, context: ResolutionContext): ResolutionResult {
    if (typeof value === 'string') {
      return this.resolveString(value, context);
    }

    if (Array.isArray(value)) {
      const results: unknown[] = [];
      const allReferencedSteps: Set<string> = new Set();

      for (const item of value) {
        const result = this.resolve(item, context);
        results.push(result.resolved);
        result.referencedSteps.forEach((s) => allReferencedSteps.add(s));
      }

      return {
        resolved: results,
        referencedSteps: Array.from(allReferencedSteps),
      };
    }

    if (value && typeof value === 'object') {
      const resolved: Record<string, unknown> = {};
      const allReferencedSteps: Set<string> = new Set();

      for (const [key, val] of Object.entries(value as Record<string, unknown>)) {
        const result = this.resolve(val, context);
        resolved[key] = result.resolved;
        result.referencedSteps.forEach((s) => allReferencedSteps.add(s));
      }

      return {
        resolved: resolved,
        referencedSteps: Array.from(allReferencedSteps),
      };
    }

    return { resolved: value, referencedSteps: [] };
  }

  private resolveString(template: string, context: ResolutionContext): ResolutionResult {
    const referencedSteps: Set<string> = new Set();

    const fullStepMatch = template.match(/^\{\{\s*steps\.([a-zA-Z0-9_-]+)(\.[^}]*)?\s*\}\}$/);
    if (fullStepMatch) {
      const stepKey = fullStepMatch[1];
      const path = fullStepMatch[2];
      referencedSteps.add(stepKey);
      const value = this.resolveStepReference(stepKey, path, context);
      return {
        resolved: this.coerceSingleValue(value),
        referencedSteps: Array.from(referencedSteps),
      };
    }

    const fullInputMatch = template.match(/^\{\{\s*input\.([a-zA-Z0-9_-]+)(\.[^}]*)?\s*\}\}$/);
    if (fullInputMatch) {
      const key = fullInputMatch[1];
      const path = fullInputMatch[2];
      const value = this.resolveInputReference(key, path, context);
      return {
        resolved: this.coerceSingleValue(value),
        referencedSteps: [],
      };
    }

    const fullSecretMatch = template.match(/^\{\{\s*secret\.([a-zA-Z0-9_-]+)\s*\}\}$/);
    if (fullSecretMatch) {
      const key = fullSecretMatch[1];
      const value = this.resolveSecretReference(key, context);
      return {
        resolved: this.coerceSingleValue(value),
        referencedSteps: [],
      };
    }

    const stepPattern = /\{\{steps\.([a-zA-Z0-9_-]+)(\.[^}]*)?\}\}/g;
    const inputPattern = /\{\{input\.([a-zA-Z0-9_-]+)(\.[^}]*)?\}\}/g;
    const secretPattern = /\{\{secret\.([a-zA-Z0-9_-]+)\}\}/g;

    let result = template.replace(stepPattern, (_, stepKey, path) => {
      referencedSteps.add(stepKey);

      const value = this.resolveStepReference(stepKey, path, context);
      return this.coerceInterpolatedValue(value);
    });

    result = result.replace(inputPattern, (_, key, path) => {
      const value = this.resolveInputReference(key, path, context);
      return this.coerceInterpolatedValue(value);
    });

    result = result.replace(secretPattern, (_, key) => {
      const value = this.resolveSecretReference(key, context);
      return this.coerceInterpolatedValue(value);
    });

    return {
      resolved: result,
      referencedSteps: Array.from(referencedSteps),
    };
  }

  private resolveStepReference(
    stepKey: string,
    path: string | undefined,
    context: ResolutionContext,
  ): unknown {
    const stepOutput = context.steps[stepKey];
    if (stepOutput === undefined || stepOutput === null) {
      throw new Error(`Referenced step "${stepKey}" does not exist or has not completed`);
    }

    const output = stepOutput as unknown as Record<string, unknown>;
    const stepData = unwrapEnvelope(stepOutput);

    const outputObj =
      stepData && typeof stepData === 'object' ? (stepData as Record<string, unknown>) : undefined;

    const maybeBody = outputObj?.body;
    const bodyData = unwrapHttpBody(maybeBody);

    const data = bodyData !== undefined ? bodyData : stepData;

    if (!path) return data;

    const cleanPath = this.cleanReferencePath(path, 'output');
    if (!cleanPath) return data;

    return this.getByPath(
      data,
      cleanPath,
      () => new Error(`Path "${path}" not found in step "${stepKey}" output`),
    );
  }

  private resolveInputReference(
    key: string,
    path: string | undefined,
    context: ResolutionContext,
  ): unknown {
    if (!(key in context.input)) {
      throw new Error(`Input field "${key}" not found in workflow input`);
    }

    const data = context.input[key];
    if (!path) return data;

    const cleanPath = this.cleanReferencePath(path);
    if (!cleanPath) return data;

    const value = this.getByPath(
      data,
      cleanPath,
      () => new Error(`Path "${path}" not found in workflow input field "${key}"`),
    );

    if (value === undefined) {
      throw new Error(`Path "${cleanPath}" not found in workflow input field "${key}"`);
    }

    return value;
  }

  private resolveSecretReference(key: string, context: ResolutionContext): unknown {
    const secrets = context.secret ?? {};
    if (!(key in secrets)) {
      throw new Error(`Secret "${key}" not found`);
    }
    return secrets[key];
  }

  private cleanReferencePath(path: string, prefixToStrip?: string): string {
    let clean = path.replace(/^\./, '').trim();

    if (prefixToStrip && clean.startsWith(prefixToStrip)) {
      clean = clean.slice(prefixToStrip.length);
      clean = clean.replace(/^\./, '').trim();
    }

    return clean;
  }

  private getByPath(base: unknown, cleanPath: string, makeError: () => Error): unknown {
    const parts = cleanPath.split('.').filter(Boolean);
    let value: unknown = base;

    for (const part of parts) {
      if (value === null || value === undefined) {
        throw makeError();
      }

      value = (value as any)[part];
    }

    if (value === undefined) {
      throw makeError();
    }

    return value;
  }

  private coerceInterpolatedValue(value: unknown): string {
    if (value === null || value === undefined) return '';
    if (typeof value === 'object') return JSON.stringify(value);
    return String(value);
  }

  private coerceSingleValue(value: unknown): unknown {
    if (value === null || value === undefined) return '';
    if (typeof value === 'object') return value;
    return String(value);
  }
}

function unwrapEnvelope(value: unknown): unknown {
  if (!value || typeof value !== 'object') return value;
  if (!('data' in (value as any)) || !('_lunie' in (value as any))) return value;
  const data = (value as any).data;
  return data;
}

function unwrapHttpBody(value: unknown): unknown {
  if (!value || typeof value !== 'object') return undefined;
  if (!('data' in (value as any)) || !('_lunieHttp' in (value as any))) return undefined;
  return (value as any).data;
}
