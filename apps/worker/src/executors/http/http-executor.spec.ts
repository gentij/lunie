/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access */

import { HttpExecutor } from './http-executor';

describe('HttpExecutor', () => {
  const originalFetch = global.fetch;

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it('parses JSON responses', async () => {
    global.fetch = jest.fn().mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { 'content-type': 'application/json' },
      }),
    ) as any;

    const exec = new HttpExecutor();
    const out = await exec.execute({
      request: { method: 'GET', url: 'https://example.com' },
      input: {},
    });

    expect(out.statusCode).toBe(200);
    expect(out.body).toEqual({
      _lunieHttp: expect.objectContaining({
        contentType: expect.any(String),
        truncated: false,
        softMaxBytes: expect.any(Number),
        hardMaxBytes: expect.any(Number),
        bytesRead: expect.any(Number),
      }),
      data: { ok: true },
    });
  });

  it('sends JSON body for POST and sets Content-Type if missing', async () => {
    global.fetch = jest.fn().mockResolvedValue(
      new Response('ok', {
        status: 200,
        headers: { 'content-type': 'text/plain' },
      }),
    ) as any;

    const exec = new HttpExecutor();
    await exec.execute({
      request: {
        method: 'POST',
        url: 'https://example.com',
        body: { a: 1 },
      },
      input: {},
    });

    const call = (global.fetch as any).mock.calls[0];
    const opts = call[1];
    expect(opts.method).toBe('POST');
    expect(opts.body).toBe(JSON.stringify({ a: 1 }));
    expect(opts.headers['Content-Type']).toBe('application/json');
  });

  it('appends query params', async () => {
    global.fetch = jest.fn().mockResolvedValue(
      new Response('ok', {
        status: 200,
        headers: { 'content-type': 'text/plain' },
      }),
    ) as any;

    const exec = new HttpExecutor();
    await exec.execute({
      request: {
        method: 'GET',
        url: 'https://example.com/path',
        query: { a: '1', b: 2, c: true },
      },
      input: {},
    });

    const calledUrl = (global.fetch as any).mock.calls[0][0] as string;
    expect(calledUrl).toContain('a=1');
    expect(calledUrl).toContain('b=2');
    expect(calledUrl).toContain('c=true');
  });
});
