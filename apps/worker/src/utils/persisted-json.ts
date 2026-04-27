export const HARD_MAX_BYTES = 10 * 1024 * 1024;

export type PersistPolicy = {
  maxBytes: number;
  truncate: boolean;
  hardMaxBytes?: number;
  reason: string;
};

export type PersistEnvelope = {
  _lunie: {
    truncated: boolean;
    bytesEstimate: number;
    originalBytesEstimate?: number;
    maxBytes: number;
    hardMaxBytes: number;
    reason?: string;
  };
  data: unknown;
};

export function wrapForDb(value: unknown, policy: PersistPolicy): PersistEnvelope {
  const hardMaxBytes = policy.hardMaxBytes ?? HARD_MAX_BYTES;
  const bytes = estimateBytes(value);

  if (bytes > hardMaxBytes) {
    throw new Error(
      `${policy.reason}: payload too large (bytes=${bytes}, hardMaxBytes=${hardMaxBytes})`,
    );
  }

  if (!policy.truncate || bytes <= policy.maxBytes) {
    return {
      _lunie: {
        truncated: false,
        bytesEstimate: bytes,
        maxBytes: policy.maxBytes,
        hardMaxBytes,
        reason: policy.reason,
      },
      data: value,
    };
  }

  const truncatedData = truncateJson(value, policy.maxBytes);
  const truncatedBytes = estimateBytes(truncatedData);

  return {
    _lunie: {
      truncated: true,
      bytesEstimate: truncatedBytes,
      originalBytesEstimate: bytes,
      maxBytes: policy.maxBytes,
      hardMaxBytes,
      reason: policy.reason,
    },
    data: truncatedData,
  };
}

export function estimateBytes(value: unknown): number {
  if (value === undefined) return 0;
  try {
    return Buffer.byteLength(JSON.stringify(value), 'utf8');
  } catch {
    return Buffer.byteLength(String(value), 'utf8');
  }
}

function truncateJson(value: unknown, maxBytes: number): unknown {
  let v: unknown = value;

  if (estimateBytes(v) <= maxBytes) return v;

  v = simplify(value, 0, 20);
  if (estimateBytes(v) <= maxBytes) return v;

  v = simplify(value, 0, 8, { maxArrayLen: 50, maxObjectKeys: 50, maxStringLen: 2000 });
  if (estimateBytes(v) <= maxBytes) return v;

  v = simplify(value, 0, 6, { maxArrayLen: 10, maxObjectKeys: 10, maxStringLen: 500 });
  if (estimateBytes(v) <= maxBytes) return v;

  const str = safeToString(value);
  return truncateStringToBytes(str, maxBytes);
}

function simplify(
  node: unknown,
  depth: number,
  maxDepth: number,
  opts?: { maxArrayLen?: number; maxObjectKeys?: number; maxStringLen?: number },
): unknown {
  if (node === null || node === undefined) return node;
  if (depth >= maxDepth) return '[Truncated: maxDepth]';

  if (typeof node === 'string') {
    const maxStringLen = opts?.maxStringLen;
    return typeof maxStringLen === 'number' && node.length > maxStringLen
      ? node.slice(0, maxStringLen) + '...[truncated]'
      : node;
  }

  if (typeof node === 'number' || typeof node === 'boolean') return node;

  if (Array.isArray(node)) {
    const maxArrayLen = opts?.maxArrayLen;
    const arr = typeof maxArrayLen === 'number' ? node.slice(0, maxArrayLen) : node;
    return arr.map((x) => simplify(x, depth + 1, maxDepth, opts));
  }

  if (typeof node === 'object') {
    const obj = node as Record<string, unknown>;
    const keys = Object.keys(obj);
    const maxObjectKeys = opts?.maxObjectKeys;
    const usedKeys = typeof maxObjectKeys === 'number' ? keys.slice(0, maxObjectKeys) : keys;
    const out: Record<string, unknown> = {};
    for (const k of usedKeys) {
      out[k] = simplify(obj[k], depth + 1, maxDepth, opts);
    }
    if (typeof maxObjectKeys === 'number' && keys.length > maxObjectKeys) {
      out._lunie_truncatedKeys = keys.length - maxObjectKeys;
    }
    return out;
  }

  return safeToString(node);
}

function safeToString(value: unknown): string {
  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}

function truncateStringToBytes(input: string, maxBytes: number): string {
  if (Buffer.byteLength(input, 'utf8') <= maxBytes) return input;
  let lo = 0;
  let hi = input.length;

  while (lo < hi) {
    const mid = Math.ceil((lo + hi) / 2);
    const slice = input.slice(0, mid);
    const bytes = Buffer.byteLength(slice, 'utf8');
    if (bytes <= maxBytes) lo = mid;
    else hi = mid - 1;
  }

  return input.slice(0, lo) + '...[truncated]';
}
