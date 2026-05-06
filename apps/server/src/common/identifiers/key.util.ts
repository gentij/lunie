export function slugifyKey(value: string, fallback = 'item'): string {
  const asciiOnly = Array.from(value.normalize('NFKD'))
    .filter((char) => char.charCodeAt(0) <= 0x7f)
    .join('');

  const normalized = asciiOnly
    .toLowerCase()
    .trim()
    .replace(/[\s_]+/g, '-')
    .replace(/[^a-z0-9-]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-+|-+$/g, '');

  if (normalized !== '') {
    return normalized;
  }

  const fallbackKey = fallback.trim();
  if (fallbackKey === '') {
    return 'item';
  }

  if (fallbackKey === value) {
    return 'item';
  }

  return slugifyKey(fallbackKey, 'item');
}

export function buildUniqueKey(
  source: string | null | undefined,
  fallback: string,
  existingKeys: Iterable<string>,
): string {
  const base = slugifyKey(source ?? '', fallback);
  const reserved = new Set(
    Array.from(existingKeys)
      .map((key) => key.trim().toLowerCase())
      .filter((key) => key !== ''),
  );

  if (!reserved.has(base)) {
    return base;
  }

  let suffix = 2;
  let candidate = `${base}-${suffix}`;
  while (reserved.has(candidate)) {
    suffix += 1;
    candidate = `${base}-${suffix}`;
  }

  return candidate;
}

export function deriveReservedKey(params: {
  key?: string | null;
  source?: string | null;
  fallback: string;
}): string {
  if (typeof params.key === 'string' && params.key.trim() !== '') {
    return slugifyKey(params.key, params.fallback);
  }

  return slugifyKey(params.source ?? '', params.fallback);
}
