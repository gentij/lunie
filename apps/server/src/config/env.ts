import { z } from 'zod';

function isBase64Key32(value: string): boolean {
  try {
    const buf = Buffer.from(value, 'base64');
    return buf.length === 32;
  } catch {
    return false;
  }
}

const envSchema = z.object({
  NODE_ENV: z
    .enum(['development', 'test', 'production'])
    .default('development'),

  PORT: z
    .string()
    .optional()
    .transform((val) => (val ? Number(val) : 3000))
    .refine((val) => Number.isInteger(val) && val > 0, {
      message: 'PORT must be a positive integer',
    }),

  DATABASE_URL: z
    .string()
    .min(1, 'DATABASE_URL is required')
    .url()
    .refine(
      (url) => url.startsWith('postgresql://') || url.startsWith('postgres://'),
      {
        message: 'DATABASE_URL must be a Postgres connection string',
      },
    ),

  REDIS_URL: z
    .string()
    .min(1, 'REDIS_URL is required')
    .url()
    .refine((url) => url.startsWith('redis://'), {
      message: 'REDIS_URL must be a Redis connection string',
    }),

  CACHE_TTL_SECONDS: z
    .string()
    .optional()
    .transform((val) => (val ? Number(val) : 60))
    .refine((val) => Number.isInteger(val) && val > 0, {
      message: 'CACHE_TTL_SECONDS must be a positive integer',
    }),

  CACHE_REDIS_PREFIX: z
    .string()
    .optional()
    .transform((val) => val ?? 'lunie:server:'),

  LUNIE_ADMIN_TOKEN: z
    .string()
    .min(32, 'LUNIE_ADMIN_TOKEN must be at least 32 characters'),

  LUNIE_SECRET_KEY: z
    .string()
    .min(1, 'LUNIE_SECRET_KEY is required')
    .refine((v) => /^[0-9a-fA-F]{64}$/.test(v) || isBase64Key32(v), {
      message: 'LUNIE_SECRET_KEY must be 64-char hex or base64 for 32 bytes',
    }),

  VERSION: z.string().default('1'),
});

export function validateEnv(env: NodeJS.ProcessEnv) {
  const parsed = envSchema.safeParse(env);

  if (!parsed.success) {
    console.error('Invalid environment variables:');
    console.error(z.treeifyError(parsed.error));
    throw new Error('Invalid environment variables');
  }

  return parsed.data;
}

export type Env = z.infer<typeof envSchema>;
