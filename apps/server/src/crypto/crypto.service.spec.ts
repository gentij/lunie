import { CryptoService } from './crypto.service';

describe('CryptoService', () => {
  let service: CryptoService;

  beforeEach(() => {
    const config = {
      get: jest.fn().mockReturnValue('0'.repeat(64)),
    } as unknown as ConfigService;
    service = new CryptoService(config);
  });

  it('generateApiToken() returns token with lunie_ prefix and hex payload', () => {
    const token = service.generateApiToken();

    expect(token.startsWith('lunie_')).toBe(true);

    const payload = token.slice('lunie_'.length);
    expect(payload).toMatch(/^[0-9a-f]+$/);
    expect(payload.length).toBe(64);
  });

  it('generateApiToken() produces different values', () => {
    const a = service.generateApiToken();
    const b = service.generateApiToken();
    expect(a).not.toBe(b);
  });

  it('hashApiToken() returns stable sha256 hex digest', () => {
    const hash = service.hashApiToken('lunie_test_token');
    expect(hash).toMatch(/^[0-9a-f]{64}$/);

    const hash2 = service.hashApiToken('lunie_test_token');
    expect(hash).toBe(hash2);
  });
});
import type { ConfigService } from '@nestjs/config';
