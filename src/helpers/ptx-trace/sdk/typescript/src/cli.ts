/**
 * PTX-TRACE SDK CLI Wrapper - Subprocess calls to portunix trace commands.
 */

import { spawn } from 'child_process';
import {
  SessionInfo,
  TraceEvent,
  parseSessionInfo,
  parseTraceEvent,
} from './models';

/** Exception thrown when CLI command fails */
export class CliError extends Error {
  constructor(
    message: string,
    public exitCode: number = 1,
    public stderr: string = ''
  ) {
    super(message);
    this.name = 'CliError';
  }
}

/** Wrapper for executing portunix trace CLI commands */
export class CliExecutor {
  private binaryPath: string;
  private timeoutMs: number;

  constructor(binaryPath: string = 'portunix', timeoutMs: number = 30000) {
    this.binaryPath = binaryPath;
    this.timeoutMs = timeoutMs;
  }

  /** Execute a CLI command and return raw output */
  private async execute(...args: string[]): Promise<string> {
    return new Promise((resolve, reject) => {
      const cmd = [this.binaryPath, 'trace', ...args];

      const proc = spawn(this.binaryPath, ['trace', ...args], {
        timeout: this.timeoutMs,
      });

      let stdout = '';
      let stderr = '';

      proc.stdout.on('data', (data: Buffer) => {
        stdout += data.toString();
      });

      proc.stderr.on('data', (data: Buffer) => {
        stderr += data.toString();
      });

      proc.on('error', (err: Error) => {
        reject(new CliError(`Failed to execute command: ${err.message}`, -1));
      });

      proc.on('close', (code: number | null) => {
        if (code !== 0) {
          reject(
            new CliError(
              `Command failed: ${stderr.trim()}`,
              code ?? 1,
              stderr
            )
          );
        } else {
          resolve(stdout.trim());
        }
      });
    });
  }

  /** Execute a CLI command and parse JSON output */
  private async executeJson<T>(...args: string[]): Promise<T | null> {
    const output = await this.execute(...args, '--format', 'json');

    if (!output) {
      return null;
    }

    try {
      return JSON.parse(output) as T;
    } catch (e) {
      throw new CliError(`Failed to parse JSON: ${e}`);
    }
  }

  /** Start a new trace session */
  async startSession(
    name: string,
    source?: string,
    destination?: string,
    tags?: string[],
    sampling: number = 1.0,
    piiMask: boolean = false
  ): Promise<string> {
    const args: string[] = ['start', name];

    if (source) {
      args.push('--source', source);
    }

    if (destination) {
      args.push('--destination', destination);
    }

    if (tags) {
      for (const tag of tags) {
        args.push('--tag', tag);
      }
    }

    if (sampling < 1.0) {
      args.push('--sampling', String(sampling));
    }

    if (piiMask) {
      args.push('--pii-mask');
    }

    const output = await this.execute(...args);

    // Parse session ID from output like "Session started: ses_2026-01-27_import"
    for (const line of output.split('\n')) {
      if (line.startsWith('Session started:')) {
        return line.split(':')[1].trim();
      }
    }

    throw new CliError('Failed to get session ID from output');
  }

  /** End the active trace session */
  async endSession(
    status: string = 'completed',
    summary: boolean = false
  ): Promise<void> {
    const args: string[] = ['end', '--status', status];

    if (summary) {
      args.push('--summary');
    }

    await this.execute(...args);
  }

  /** Add a trace event */
  async addEvent(
    operation: string,
    inputData?: Record<string, unknown>,
    outputData?: Record<string, unknown>,
    status?: string,
    error?: string,
    tags?: string[],
    duration?: number
  ): Promise<void> {
    const args: string[] = ['event', operation];

    if (inputData && Object.keys(inputData).length > 0) {
      const inputStr = Object.entries(inputData)
        .map(([k, v]) => `${k}=${v}`)
        .join(',');
      args.push('--input', inputStr);
    }

    if (outputData && Object.keys(outputData).length > 0) {
      const outputStr = Object.entries(outputData)
        .map(([k, v]) => `${k}=${v}`)
        .join(',');
      args.push('--output', outputStr);
    }

    if (error) {
      args.push('--error', error);
    } else if (status) {
      args.push('--status', status);
    }

    if (tags) {
      for (const tag of tags) {
        args.push('--tag', tag);
      }
    }

    if (duration !== undefined) {
      args.push('--duration', String(duration));
    }

    await this.execute(...args);
  }

  /** List sessions */
  async listSessions(
    limit?: number,
    status?: string
  ): Promise<SessionInfo[]> {
    const args: string[] = ['sessions'];

    if (limit) {
      args.push('--limit', String(limit));
    }

    if (status) {
      args.push('--status', status);
    }

    const data = await this.executeJson<Record<string, unknown>[]>(...args);

    if (!data) {
      return [];
    }

    return data.map(parseSessionInfo);
  }

  /** Get session statistics */
  async getSessionStats(sessionId?: string): Promise<SessionInfo> {
    const args: string[] = ['stats'];

    if (sessionId) {
      args.push(sessionId);
    }

    const data = await this.executeJson<Record<string, unknown>>(...args);

    if (!data) {
      throw new CliError('No session found');
    }

    return parseSessionInfo(data);
  }

  /** View events */
  async viewEvents(
    sessionId?: string,
    operation?: string,
    status?: string,
    level?: string,
    tag?: string,
    limit: number = 100
  ): Promise<TraceEvent[]> {
    const args: string[] = ['view'];

    if (sessionId) {
      args.push(sessionId);
    }

    if (operation) {
      args.push('--operation', operation);
    }

    if (status) {
      args.push('--status', status);
    }

    if (level) {
      args.push('--level', level);
    }

    if (tag) {
      args.push('--tag', tag);
    }

    args.push('--limit', String(limit));

    const data = await this.executeJson<Record<string, unknown>[]>(...args);

    if (!data) {
      return [];
    }

    return data.map(parseTraceEvent);
  }

  /** Query events using SQL-like syntax */
  async queryEvents(
    query: string,
    sessionId?: string,
    limit: number = 100
  ): Promise<Record<string, unknown>[]> {
    const args: string[] = ['query', query];

    if (sessionId) {
      args.push('--session', sessionId);
    }

    args.push('--limit', String(limit));

    const data = await this.executeJson<Record<string, unknown>[]>(...args);
    return data ?? [];
  }
}
