import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import {
  randomBytes,
  createHash,
  createHmac,
  timingSafeEqual,
  createCipheriv,
  createDecipheriv,
} from 'crypto';

@Injectable()
export class CryptoService {
  private readonly secretKey: Buffer;

  constructor(private readonly config: ConfigService) {
    const raw = this.config.get<string>('LUNIE_SECRET_KEY');
    if (!raw) {
      throw new Error(
        'LUNIE_SECRET_KEY is required (32-byte base64 or 64-char hex)',
      );
    }

    this.secretKey = decodeKey(raw);
    if (this.secretKey.length !== 32) {
      throw new Error('LUNIE_SECRET_KEY must decode to 32 bytes');
    }
  }

  generateApiToken(): string {
    return `lunie_${randomBytes(32).toString('hex')}`;
  }

  hashApiToken(token: string): string {
    return createHash('sha256').update(token).digest('hex');
  }

  verifyHmac(
    payload: string | Buffer,
    secret: string,
    signature: string,
  ): boolean {
    const computed = createHmac('sha256', secret).update(payload).digest('hex');

    return this.secureCompare(computed, signature);
  }

  secureCompare(a: string, b: string): boolean {
    if (a.length !== b.length) return false;
    return timingSafeEqual(Buffer.from(a), Buffer.from(b));
  }

  public generateId(): string {
    const bytes = new Uint8Array(16);
    crypto.getRandomValues(bytes);
    return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
  }

  encryptSecret(plaintext: string): string {
    const iv = randomBytes(12);
    const cipher = createCipheriv('aes-256-gcm', this.secretKey, iv);
    const ciphertext = Buffer.concat([
      cipher.update(plaintext, 'utf8'),
      cipher.final(),
    ]);
    const tag = cipher.getAuthTag();

    // Format: luniesec:v1:<ivb64>:<tagb64>:<ctb64>
    return `luniesec:v1:${iv.toString('base64')}:${tag.toString('base64')}:${ciphertext.toString('base64')}`;
  }

  decryptSecret(value: string): string {
    if (!value.startsWith('luniesec:v1:')) {
      return value;
    }

    const parts = value.split(':');
    if (parts.length !== 5) {
      throw new Error('Invalid secret ciphertext format');
    }

    const iv = Buffer.from(parts[2], 'base64');
    const tag = Buffer.from(parts[3], 'base64');
    const ciphertext = Buffer.from(parts[4], 'base64');

    const decipher = createDecipheriv('aes-256-gcm', this.secretKey, iv);
    decipher.setAuthTag(tag);
    const plaintext = Buffer.concat([
      decipher.update(ciphertext),
      decipher.final(),
    ]);
    return plaintext.toString('utf8');
  }
}

function decodeKey(raw: string): Buffer {
  if (/^[0-9a-fA-F]{64}$/.test(raw)) {
    return Buffer.from(raw, 'hex');
  }

  return Buffer.from(raw, 'base64');
}
