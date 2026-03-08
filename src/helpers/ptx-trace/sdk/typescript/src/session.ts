/**
 * PTX-TRACE SDK Session - Manages tracing session lifecycle.
 */

import { CliExecutor } from './cli';
import {
  SessionInfo,
  SessionOptions,
  SessionStatus,
  TraceEvent,
  EventQueryOptions,
} from './models';
import { Operation, OperationBuilder } from './operation';

/** Manages a PTX-TRACE session */
export class Session {
  private cli: CliExecutor;
  private _name: string;
  private source?: string;
  private destination?: string;
  private tags: string[];
  private sampling: number;
  private piiMasking: boolean;
  private _sessionId?: string;
  private started: boolean;
  private closed: boolean = false;

  private constructor(
    cli: CliExecutor,
    name: string,
    options: SessionOptions = {},
    sessionId?: string
  ) {
    this.cli = cli;
    this._name = name;
    this.source = options.source;
    this.destination = options.destination;
    this.tags = options.tags ?? [];
    this.sampling = options.sampling ?? 1.0;
    this.piiMasking = options.piiMasking ?? false;
    this._sessionId = sessionId;
    this.started = sessionId !== undefined;
  }

  /**
   * Create and start a new session.
   *
   * @param name Session name
   * @param options Session options
   */
  static async create(
    name: string,
    options: SessionOptions = {}
  ): Promise<Session> {
    const binaryPath = options.binaryPath ?? 'portunix';
    const cli = new CliExecutor(binaryPath);

    const session = new Session(cli, name, options);
    await session.start();
    return session;
  }

  /**
   * Load an existing session by ID.
   *
   * @param sessionId Session ID
   * @param binaryPath Optional path to portunix binary
   */
  static async load(
    sessionId: string,
    binaryPath: string = 'portunix'
  ): Promise<Session> {
    const cli = new CliExecutor(binaryPath);
    const info = await cli.getSessionStats(sessionId);

    return new Session(
      cli,
      info.name,
      { tags: info.tags },
      sessionId
    );
  }

  /**
   * List all sessions.
   *
   * @param options Query options
   * @param binaryPath Optional path to portunix binary
   */
  static async list(
    options: { limit?: number; status?: string } = {},
    binaryPath: string = 'portunix'
  ): Promise<SessionInfo[]> {
    const cli = new CliExecutor(binaryPath);
    return cli.listSessions(options.limit, options.status);
  }

  /** Session ID */
  get id(): string | undefined {
    return this._sessionId;
  }

  /** Session name */
  get name(): string {
    return this._name;
  }

  /** Whether session is active */
  get isActive(): boolean {
    return this.started && !this.closed;
  }

  /** Start the session */
  async start(): Promise<string> {
    if (this.started && this._sessionId) {
      return this._sessionId;
    }

    this._sessionId = await this.cli.startSession(
      this._name,
      this.source,
      this.destination,
      this.tags.length > 0 ? this.tags : undefined,
      this.sampling,
      this.piiMasking
    );
    this.started = true;
    return this._sessionId;
  }

  /** End the session with a specific status */
  async end(status: SessionStatus = SessionStatus.COMPLETED): Promise<void> {
    if (this.closed) {
      return;
    }

    await this.cli.endSession(status);
    this.closed = true;
  }

  /** End the session successfully */
  async close(): Promise<void> {
    await this.end(SessionStatus.COMPLETED);
  }

  /** End the session as failed */
  async fail(): Promise<void> {
    await this.end(SessionStatus.FAILED);
  }

  /** End the session as cancelled */
  async cancel(): Promise<void> {
    await this.end(SessionStatus.CANCELLED);
  }

  /**
   * Create a traced operation builder.
   *
   * @param operationName Name of the operation
   */
  trace(operationName: string): OperationBuilder {
    return new OperationBuilder(this.cli, operationName);
  }

  /**
   * Start a traced operation directly.
   *
   * @param operationName Name of the operation
   * @param type Operation type
   */
  startOperation(
    operationName: string,
    type: string = 'transform'
  ): Operation {
    return new Operation(this.cli, operationName, type);
  }

  /**
   * Execute a traced operation with a function.
   *
   * @param name Operation name
   * @param fn Function to execute
   */
  async withOperation<T>(
    name: string,
    fn: (op: Operation) => T | Promise<T>
  ): Promise<T> {
    const op = this.startOperation(name);

    try {
      const result = await fn(op);
      op.success();
      return result;
    } catch (e) {
      op.error(String(e));
      throw e;
    } finally {
      await op.end();
    }
  }

  /** Get session statistics */
  async stats(): Promise<SessionInfo> {
    return this.cli.getSessionStats(this._sessionId);
  }

  /**
   * Get events from this session.
   *
   * @param options Query options
   */
  async events(options: EventQueryOptions = {}): Promise<TraceEvent[]> {
    return this.cli.viewEvents(
      this._sessionId,
      options.operation,
      options.status,
      options.level,
      options.tag,
      options.limit ?? 100
    );
  }

  /**
   * Query events using SQL-like syntax.
   *
   * @param query SQL-like query string
   * @param limit Maximum number of results
   */
  async query(
    query: string,
    limit: number = 100
  ): Promise<Record<string, unknown>[]> {
    return this.cli.queryEvents(query, this._sessionId, limit);
  }
}
