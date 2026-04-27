import { createDecipheriv } from 'crypto';
import { Injectable } from '@nestjs/common';

@Injectable()
export class CryptoService {
  private readonly secretKey: Buffer;

  constructor() {
    const raw = process.env.LUNIE_SECRET_KEY;
    if (!raw) {
      throw new Error('LUNIE_SECRET_KEY is required (32-byte base64 or 64-char hex)');
    }

    this.secretKey = decodeKey(raw);
    if (this.secretKey.length !== 32) {
      throw new Error('LUNIE_SECRET_KEY must decode to 32 bytes');
    }
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
    const plaintext = Buffer.concat([decipher.update(ciphertext), decipher.final()]);
    return plaintext.toString('utf8');
  }
}

function decodeKey(raw: string): Buffer {
  if (/^[0-9a-fA-F]{64}$/.test(raw)) {
    return Buffer.from(raw, 'hex');
  }
  return Buffer.from(raw, 'base64');
}
