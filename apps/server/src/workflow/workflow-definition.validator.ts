import * as jmespathImport from 'jmespath';

const jmespath = jmespathImport as unknown as {
  compile: (expression: string) => unknown;
};
import type { WorkflowDefinition } from '@taskforge/contracts';

export type ValidationIssue = {
  field?: string;
  stepKey?: string;
  message: string;
};

type StepLike = {
  key: string;
  type: string;
  dependsOn?: string[];
  input?: Record<string, unknown>;
  request?: unknown;
};

type TransformRequestLike = { output?: unknown };
type ConditionRequestLike = { expr?: unknown };
type NotificationLike = {
  provider: 'slack' | 'discord';
  webhook: string;
  on: Array<'SUCCEEDED' | 'FAILED'>;
};

export function validateWorkflowDefinitionStrict(
  definition: WorkflowDefinition,
): ValidationIssue[] {
  const issues: ValidationIssue[] = [];

  const steps = (definition.steps ?? []) as unknown as StepLike[];
  const notifications = (definition.notifications ??
    []) as unknown as NotificationLike[];
  const workflowInputKeys = new Set(Object.keys(definition.input ?? {}));

  const stepKeyCounts = new Map<string, number>();
  for (const s of steps) {
    stepKeyCounts.set(s.key, (stepKeyCounts.get(s.key) ?? 0) + 1);
  }
  for (const [key, count] of stepKeyCounts) {
    if (count > 1) {
      issues.push({
        field: 'steps',
        stepKey: key,
        message: `Duplicate step key: "${key}"`,
      });
    }
  }

  const allStepKeys = new Set(steps.map((s) => s.key));

  // Validate dependsOn references
  for (let i = 0; i < steps.length; i++) {
    const step = steps[i];
    const deps = step.dependsOn ?? [];
    for (let j = 0; j < deps.length; j++) {
      const dep = deps[j];
      if (dep === step.key) {
        issues.push({
          field: `steps[${i}].dependsOn[${j}]`,
          stepKey: step.key,
          message: `stepKey=${step.key}: dependsOn cannot reference itself`,
        });
      } else if (!allStepKeys.has(dep)) {
        issues.push({
          field: `steps[${i}].dependsOn[${j}]`,
          stepKey: step.key,
          message: `stepKey=${step.key}: dependsOn references unknown step "${dep}"`,
        });
      }
    }
  }

  // Validate templates and infer dependencies from {{steps.<key>...}} references.
  const inferredDeps = new Map<string, Set<string>>();
  for (const s of steps) inferredDeps.set(s.key, new Set());

  const stepRefPattern = /\{\{\s*steps\.([a-zA-Z0-9_-]+)(?:\.[^}]*)?\s*\}\}/g;
  const inputRefPattern = /\{\{\s*input\.([a-zA-Z0-9_-]+)(?:\.[^}]*)?\s*\}\}/g;
  const rootPattern = /\{\{\s*([a-zA-Z0-9_-]+)(?:\.|\}\})/g;
  const allowedRoots = new Set(['input', 'steps', 'secret']);

  const validateTemplateRoots = (
    value: string,
    path: string,
    stepKey?: string,
  ) => {
    rootPattern.lastIndex = 0;
    let rootMatch: RegExpExecArray | null;
    while ((rootMatch = rootPattern.exec(value)) !== null) {
      const root = rootMatch[1];
      if (!root || allowedRoots.has(root)) continue;
      issues.push({
        field: path,
        stepKey,
        message: `stepKey=${stepKey ?? 'workflow'}: invalid template root "${root}" (allowed: input, steps, secret)`,
      });
    }
  };

  walk(definition.input ?? {}, 'input', (value, path) => {
    if (typeof value !== 'string') return;
    validateTemplateRoots(value, path);
  });

  for (let i = 0; i < notifications.length; i++) {
    const notification = notifications[i];
    const webhook = String(notification.webhook ?? '').trim();
    const field = `notifications[${i}].webhook`;

    validateTemplateRoots(webhook, field);

    const isSecretRef = /^\{\{\s*secret\.[A-Za-z0-9_-]+\s*\}\}$/.test(webhook);
    if (isSecretRef) {
      continue;
    }

    try {
      const parsed = new URL(webhook);
      if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
        throw new Error('unsupported protocol');
      }
    } catch {
      issues.push({
        field,
        message:
          'notification webhook must be an absolute http(s) URL or {{secret.NAME}}',
      });
    }
  }

  for (let i = 0; i < steps.length; i++) {
    const step = steps[i];
    const stepInputKeys = new Set(Object.keys(step.input ?? {}));
    const allowedInputKeys = new Set<string>([
      ...workflowInputKeys,
      ...stepInputKeys,
    ]);

    walk(step.request, `steps[${i}].request`, (value, path) => {
      if (typeof value !== 'string') return;
      validateTemplateRoots(value, path, step.key);

      let m: RegExpExecArray | null;
      stepRefPattern.lastIndex = 0;
      while ((m = stepRefPattern.exec(value)) !== null) {
        const ref = m[1];
        if (!ref) continue;
        if (!allStepKeys.has(ref)) {
          issues.push({
            field: path,
            stepKey: step.key,
            message: `stepKey=${step.key}: references unknown step "${ref}"`,
          });
        } else if (ref !== step.key) {
          inferredDeps.get(step.key)?.add(ref);
        }
      }

      inputRefPattern.lastIndex = 0;
      while ((m = inputRefPattern.exec(value)) !== null) {
        const ref = m[1];
        if (!ref) continue;
        if (!allowedInputKeys.has(ref)) {
          issues.push({
            field: path,
            stepKey: step.key,
            message: `stepKey=${step.key}: input field "${ref}" must be declared in workflow definition.input or step.input`,
          });
        }
      }
    });

    walk(step.input ?? {}, `steps[${i}].input`, (value, path) => {
      if (typeof value !== 'string') return;
      validateTemplateRoots(value, path, step.key);
    });
  }

  // Validate transform $jmes nodes and compile expressions.
  for (let i = 0; i < steps.length; i++) {
    const step = steps[i];
    if (step.type !== 'transform') continue;

    const output = (step.request as TransformRequestLike | undefined)?.output;
    walkJmes(
      output,
      `steps[${i}].request.output`,
      (expr, field) => {
        try {
          jmespath.compile(expr);
        } catch (err) {
          const message = err instanceof Error ? err.message : String(err);
          issues.push({
            field,
            stepKey: step.key,
            message: `stepKey=${step.key}: invalid JMESPath expression: ${message}`,
          });
        }
      },
      issues,
      step.key,
    );
  }

  // Validate condition expressions compile.
  for (let i = 0; i < steps.length; i++) {
    const step = steps[i];
    if (step.type !== 'condition') continue;
    const expr = (step.request as ConditionRequestLike | undefined)?.expr;
    if (typeof expr !== 'string' || expr.trim().length === 0) continue;

    try {
      jmespath.compile(expr);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      issues.push({
        field: `steps[${i}].request.expr`,
        stepKey: step.key,
        message: `stepKey=${step.key}: invalid JMESPath expression: ${message}`,
      });
    }
  }

  // Cycle detection using explicit + inferred dependencies
  const deps = new Map<string, Set<string>>();
  for (const s of steps) {
    const set = new Set<string>();
    for (const d of s.dependsOn ?? []) {
      if (allStepKeys.has(d) && d !== s.key) set.add(d);
    }
    for (const d of inferredDeps.get(s.key) ?? []) {
      if (allStepKeys.has(d) && d !== s.key) set.add(d);
    }
    deps.set(s.key, set);
  }

  const indegree = new Map<string, number>();
  const graph = new Map<string, string[]>();
  for (const k of allStepKeys) {
    indegree.set(k, 0);
    graph.set(k, []);
  }
  for (const [k, ds] of deps) {
    for (const d of ds) {
      graph.get(d)?.push(k);
      indegree.set(k, (indegree.get(k) ?? 0) + 1);
    }
  }

  const queue: string[] = [];
  for (const [k, deg] of indegree) {
    if (deg === 0) queue.push(k);
  }
  let visited = 0;
  while (queue.length > 0) {
    const k = queue.shift()!;
    visited++;
    for (const child of graph.get(k) ?? []) {
      const nd = (indegree.get(child) ?? 0) - 1;
      indegree.set(child, nd);
      if (nd === 0) queue.push(child);
    }
  }

  if (visited !== allStepKeys.size && allStepKeys.size > 0) {
    const remaining = Array.from(indegree.entries())
      .filter(([, deg]) => deg > 0)
      .map(([k]) => k);

    issues.push({
      field: 'steps',
      stepKey: remaining.join(','),
      message: `Dependency cycle detected (explicit or template-based). Steps involved: ${remaining.join(
        ', ',
      )}`,
    });
  }

  return issues;
}

export function getInferredDependencies(
  definition: WorkflowDefinition,
): Record<string, string[]> {
  const steps = (definition.steps ?? []) as unknown as StepLike[];
  const allStepKeys = new Set(steps.map((s) => s.key));
  const templatePattern = /\{\{\s*steps\.([a-zA-Z0-9_-]+)(?:\.[^}]*)?\s*\}\}/g;

  const inferred: Record<string, Set<string>> = {};
  for (const s of steps) inferred[s.key] = new Set();

  for (const step of steps) {
    const requestStr = JSON.stringify(step.request ?? {});
    templatePattern.lastIndex = 0;
    let match: RegExpExecArray | null;
    while ((match = templatePattern.exec(requestStr)) !== null) {
      const ref = match[1];
      if (!ref) continue;
      if (ref === step.key) continue;
      if (!allStepKeys.has(ref)) continue;
      inferred[step.key].add(ref);
    }
  }

  const out: Record<string, string[]> = {};
  for (const [k, set] of Object.entries(inferred)) {
    out[k] = Array.from(set);
  }

  return out;
}

export function getReferencedSecrets(
  definition: WorkflowDefinition,
): Array<{ name: string; field: string; stepKey?: string }> {
  const steps = (definition.steps ?? []) as unknown as StepLike[];
  const secretPattern = /\{\{\s*secret\.([a-zA-Z0-9_-]+)\s*\}\}/g;

  const refs: Array<{ name: string; field: string; stepKey?: string }> = [];

  // definition.input can also contain templates
  walk(definition.input ?? {}, 'input', (value, path) => {
    if (typeof value !== 'string') return;
    secretPattern.lastIndex = 0;
    let m: RegExpExecArray | null;
    while ((m = secretPattern.exec(value)) !== null) {
      const name = m[1];
      if (!name) continue;
      refs.push({ name, field: path });
    }
  });

  for (let i = 0; i < steps.length; i++) {
    const step = steps[i];
    walk(step.request, `steps[${i}].request`, (value, path) => {
      if (typeof value !== 'string') return;
      secretPattern.lastIndex = 0;
      let m: RegExpExecArray | null;
      while ((m = secretPattern.exec(value)) !== null) {
        const name = m[1];
        if (!name) continue;
        refs.push({ name, field: path, stepKey: step.key });
      }
    });
  }

  const notifications = (definition.notifications ?? []) as Array<{
    webhook?: unknown;
  }>;
  for (let i = 0; i < notifications.length; i++) {
    const webhook = notifications[i]?.webhook;
    if (typeof webhook !== 'string') continue;
    secretPattern.lastIndex = 0;
    let m: RegExpExecArray | null;
    while ((m = secretPattern.exec(webhook)) !== null) {
      const name = m[1];
      if (!name) continue;
      refs.push({ name, field: `notifications[${i}].webhook` });
    }
  }

  return refs;
}

export function getExecutionBatchesFromDependencies(
  deps: Record<string, string[]>,
): string[][] {
  const keys = Object.keys(deps);
  const allStepKeys = new Set(keys);
  const inDegree = new Map<string, number>();
  const graph = new Map<string, string[]>();

  for (const k of keys) {
    inDegree.set(k, 0);
    graph.set(k, []);
  }

  for (const [stepKey, depsList] of Object.entries(deps)) {
    for (const dep of depsList) {
      if (!allStepKeys.has(dep)) continue;
      graph.get(dep)?.push(stepKey);
      inDegree.set(stepKey, (inDegree.get(stepKey) ?? 0) + 1);
    }
  }

  const queue: string[] = [];
  for (const [k, deg] of inDegree) {
    if (deg === 0) queue.push(k);
  }

  const batches: string[][] = [];
  let visited = 0;

  while (queue.length > 0) {
    const batch = [...queue];
    batches.push(batch);
    queue.length = 0;

    for (const stepKey of batch) {
      visited++;
      for (const child of graph.get(stepKey) ?? []) {
        const nd = (inDegree.get(child) ?? 0) - 1;
        inDegree.set(child, nd);
        if (nd === 0) queue.push(child);
      }
    }
  }

  if (visited !== keys.length) {
    throw new Error('Dependency cycle detected');
  }

  return batches;
}

function walk(
  node: unknown,
  path: string,
  visit: (value: unknown, path: string) => void,
): void {
  visit(node, path);

  if (Array.isArray(node)) {
    for (let i = 0; i < node.length; i++) {
      walk(node[i], `${path}[${i}]`, visit);
    }
    return;
  }

  if (node && typeof node === 'object') {
    for (const [k, v] of Object.entries(node as Record<string, unknown>)) {
      walk(v, `${path}.${k}`, visit);
    }
  }
}

function walkJmes(
  node: unknown,
  path: string,
  onExpr: (expr: string, field: string) => void,
  issues: ValidationIssue[],
  stepKey: string,
): void {
  if (Array.isArray(node)) {
    for (let i = 0; i < node.length; i++) {
      walkJmes(node[i], `${path}[${i}]`, onExpr, issues, stepKey);
    }
    return;
  }

  if (node && typeof node === 'object') {
    const obj = node as Record<string, unknown>;
    const keys = Object.keys(obj);

    if (keys.includes('$jmes')) {
      if (!(keys.length === 1 && keys[0] === '$jmes')) {
        issues.push({
          field: path,
          stepKey,
          message: `stepKey=${stepKey}: $jmes node must be exactly { "$jmes": "..." }`,
        });
      }

      const expr = obj.$jmes;
      if (typeof expr !== 'string' || expr.trim().length === 0) {
        issues.push({
          field: `${path}.$jmes`,
          stepKey,
          message: `stepKey=${stepKey}: $jmes expression must be a non-empty string`,
        });
      } else {
        onExpr(expr, `${path}.$jmes`);
      }

      return;
    }

    for (const [k, v] of Object.entries(obj)) {
      walkJmes(v, `${path}.${k}`, onExpr, issues, stepKey);
    }
  }
}
