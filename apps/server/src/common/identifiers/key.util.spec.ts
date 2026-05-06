import { buildUniqueKey, deriveReservedKey, slugifyKey } from './key.util';

describe('key util', () => {
  it('slugifies workflow-like names into lowercase kebab-case', () => {
    expect(slugifyKey(' Demo Workflow ', 'workflow')).toBe('demo-workflow');
    expect(slugifyKey('nightly_sync', 'workflow')).toBe('nightly-sync');
    expect(slugifyKey('GitHub / Release', 'workflow')).toBe('github-release');
  });

  it('falls back when the source cannot produce a key', () => {
    expect(slugifyKey('!!!', 'manual')).toBe('manual');
  });

  it('appends numeric suffixes for collisions', () => {
    expect(
      buildUniqueKey('Demo Workflow', 'workflow', [
        'demo-workflow',
        'demo-workflow-2',
      ]),
    ).toBe('demo-workflow-3');
  });

  it('prefers stored keys when deriving reserved identifiers', () => {
    expect(
      deriveReservedKey({
        key: 'Existing-Key',
        source: 'Ignored Name',
        fallback: 'workflow',
      }),
    ).toBe('existing-key');
  });
});
