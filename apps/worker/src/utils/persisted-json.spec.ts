import { estimateBytes, wrapForDb } from './persisted-json';

describe('persisted-json', () => {
  it('wraps without truncation when under maxBytes', () => {
    const out = wrapForDb({ ok: true }, { maxBytes: 1000, truncate: true, reason: 'x' });
    expect(out._lunie.truncated).toBe(false);
    expect(out.data).toEqual({ ok: true });
  });

  it('truncates when over maxBytes and truncate=true', () => {
    const big = { text: 'a'.repeat(10000) };
    const out = wrapForDb(big, { maxBytes: 100, truncate: true, reason: 'x' });
    expect(out._lunie.truncated).toBe(true);
    expect(estimateBytes(out.data)).toBeLessThanOrEqual(1000);
  });

  it('throws when over hardMaxBytes', () => {
    const big = { text: 'a'.repeat(1024 * 1024) };
    expect(() =>
      wrapForDb(big, {
        maxBytes: 10,
        truncate: false,
        hardMaxBytes: 100,
        reason: 'x',
      }),
    ).toThrow('payload too large');
  });
});
