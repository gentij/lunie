import { Injectable, Logger } from '@nestjs/common';
import { StepExecutor, ExecutorOutput } from '../executor.interface';
import { HttpExecutorInput, HttpRequestSpec } from './http.types';

const DEFAULT_TIMEOUT_MS = 30_000;
const HTTP_SOFT_MAX_BYTES = 256 * 1024;
const HTTP_HARD_MAX_BYTES = 10 * 1024 * 1024;

@Injectable()
export class HttpExecutor implements StepExecutor {
  readonly stepType = 'http';
  private readonly logger = new Logger(HttpExecutor.name);

  async execute(input: unknown): Promise<ExecutorOutput> {
    const validated = input as HttpExecutorInput;
    const { request } = validated;

    this.logger.debug(`Executing HTTP ${request.method} ${request.url}`);

    const startTime = Date.now();

    try {
      const response = await this.makeRequest(request);
      const durationMs = Date.now() - startTime;

      this.logger.log(
        `HTTP ${request.method} ${request.url} - ${response.statusCode} (${durationMs}ms)`,
      );

      return {
        statusCode: response.statusCode,
        headers: response.headers,
        body: response.body,
      };
    } catch (error) {
      const durationMs = Date.now() - startTime;

      this.logger.error(
        `HTTP ${request.method} ${request.url} failed after ${durationMs}ms: ${error instanceof Error ? error.message : 'unknown error'}`,
      );

      throw error;
    }
  }

  private async makeRequest(request: HttpRequestSpec): Promise<{
    statusCode: number;
    headers: Record<string, string>;
    body: unknown;
  }> {
    let url: URL;
    try {
      url = new URL(request.url);
    } catch {
      throw new Error(`Invalid URL after template resolution: "${request.url}"`);
    }

    if (request.query) {
      for (const [key, value] of Object.entries(request.query)) {
        url.searchParams.append(key, String(value));
      }
    }

    const headers: Record<string, string> = request.headers ?? {};

    const fetchOptions: RequestInit = {
      method: request.method,
      headers,
      signal: AbortSignal.timeout(request.timeoutMs ?? DEFAULT_TIMEOUT_MS),
    };

    if (request.body && ['POST', 'PUT', 'PATCH'].includes(request.method)) {
      fetchOptions.body = JSON.stringify(request.body);
      if (!headers['Content-Type']) {
        headers['Content-Type'] = 'application/json';
      }
    }

    const response = await fetch(url.toString(), fetchOptions);

    const responseHeaders: Record<string, string> = {};
    response.headers.forEach((value, key) => {
      responseHeaders[key] = value;
    });

    const { text, bytesRead, truncated } = await this.readTextWithLimits(
      response,
      HTTP_SOFT_MAX_BYTES,
      HTTP_HARD_MAX_BYTES,
    );

    const contentType = response.headers.get('content-type') ?? '';
    const meta = {
      contentType,
      truncated,
      bytesRead,
      softMaxBytes: HTTP_SOFT_MAX_BYTES,
      hardMaxBytes: HTTP_HARD_MAX_BYTES,
    };

    if (truncated) {
      throw new Error(
        `HTTP response truncated (softMaxBytes=${HTTP_SOFT_MAX_BYTES} bytesRead=${bytesRead})`,
      );
    }

    let data: unknown = text;
    if (contentType.includes('application/json')) {
      try {
        data = JSON.parse(text);
      } catch {
        data = text;
      }
    }

    const body = { _lunieHttp: meta, data };

    return {
      statusCode: response.status,
      headers: responseHeaders,
      body,
    };
  }

  private async readTextWithLimits(
    response: Response,
    softMaxBytes: number,
    hardMaxBytes: number,
  ): Promise<{ text: string; bytesRead: number; truncated: boolean }> {
    if (!response.body) {
      const text = await response.text();
      const bytesRead = Buffer.byteLength(text, 'utf8');
      return { text, bytesRead, truncated: bytesRead > softMaxBytes };
    }

    const reader = response.body.getReader();
    const chunks: Uint8Array[] = [];
    let total = 0;

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      if (!value) continue;

      total += value.byteLength;
      if (total > hardMaxBytes) {
        try {
          await reader.cancel();
        } catch {
          // ignore
        }
        throw new Error(`HTTP response body too large (bytes>${hardMaxBytes})`);
      }

      if (total > softMaxBytes) {
        try {
          await reader.cancel();
        } catch {
          // ignore
        }
        return {
          text: Buffer.concat(chunks.map((c) => Buffer.from(c))).toString('utf8'),
          bytesRead: total,
          truncated: true,
        };
      }

      chunks.push(value);
    }

    const buf = Buffer.concat(chunks.map((c) => Buffer.from(c)));
    return { text: buf.toString('utf8'), bytesRead: total, truncated: false };
  }
}
