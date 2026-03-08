/**
 * PTX-TRACE SDK Operation - Traced operation classes.
 */

import { CliExecutor } from './cli';
import { Severity, SourceInfo } from './models';

/** Context passed to traced functions for recording output */
export class OperationContext {
  constructor(private operation: Operation) {}

  /** Set an output field */
  output(key: string, value: unknown): OperationContext {
    this.operation.output(key, value);
    return this;
  }

  /** Set multiple output fields */
  outputs(data: Record<string, unknown>): OperationContext {
    this.operation.outputs(data);
    return this;
  }

  /** Add a tag */
  tag(tag: string): OperationContext {
    this.operation.tag(tag);
    return this;
  }

  /** Add context information */
  context(key: string, value: unknown): OperationContext {
    this.operation.context(key, value);
    return this;
  }
}

/** Represents a traced operation */
export class Operation {
  private inputData: Record<string, unknown> = {};
  private outputData: Record<string, unknown> = {};
  private contextData: Record<string, unknown> = {};
  private _tags: string[] = [];
  private _source?: SourceInfo;
  private _errorMsg?: string;
  private _errorSeverity: Severity = Severity.MEDIUM;
  private _status: string = '';
  private startTime: number = 0;
  private _durationUs?: number;
  private ended: boolean = false;

  constructor(
    private cli: CliExecutor,
    private _name: string,
    private type: string = 'transform',
    tags?: string[]
  ) {
    this._tags = tags ? [...tags] : [];
    this.startTime = process.hrtime.bigint ? Number(process.hrtime.bigint()) : Date.now() * 1e6;
  }

  /** Operation name */
  get name(): string {
    return this._name;
  }

  /** Set input data field */
  input(key: string, value: unknown): Operation {
    this.inputData[key] = value;
    return this;
  }

  /** Set multiple input fields */
  inputs(data: Record<string, unknown>): Operation {
    Object.assign(this.inputData, data);
    return this;
  }

  /** Set output data field */
  output(key: string, value: unknown): Operation {
    this.outputData[key] = value;
    return this;
  }

  /** Set multiple output fields */
  outputs(data: Record<string, unknown>): Operation {
    Object.assign(this.outputData, data);
    return this;
  }

  /** Set the data source */
  source(source: SourceInfo): Operation {
    this._source = source;
    return this;
  }

  /** Add tags */
  tag(...tags: string[]): Operation {
    this._tags.push(...tags);
    return this;
  }

  /** Add context information */
  context(key: string, value: unknown): Operation {
    this.contextData[key] = value;
    return this;
  }

  /** Record an error */
  error(message: string, severity: Severity = Severity.MEDIUM): Operation {
    this._errorMsg = message;
    this._errorSeverity = severity;
    return this;
  }

  /** Record an error with code */
  errorWithCode(
    code: string,
    message: string,
    severity: Severity = Severity.MEDIUM
  ): Operation {
    this._errorMsg = `[${code}] ${message}`;
    this._errorSeverity = severity;
    return this;
  }

  /** Mark the operation as successful */
  success(): Operation {
    this._status = 'success';
    return this;
  }

  /** Set duration manually in microseconds */
  setDuration(durationUs: number): Operation {
    this._durationUs = durationUs;
    return this;
  }

  /** End the operation and record it */
  async end(): Promise<void> {
    if (this.ended) {
      return;
    }
    this.ended = true;

    // Calculate duration if not set manually
    if (this._durationUs === undefined) {
      const now = process.hrtime.bigint
        ? Number(process.hrtime.bigint())
        : Date.now() * 1e6;
      this._durationUs = Math.floor((now - this.startTime) / 1000);
    }

    // Record event via CLI
    await this.cli.addEvent(
      this._name,
      Object.keys(this.inputData).length > 0 ? this.inputData : undefined,
      Object.keys(this.outputData).length > 0 ? this.outputData : undefined,
      this._errorMsg ? undefined : this._status || undefined,
      this._errorMsg,
      this._tags.length > 0 ? this._tags : undefined,
      this._durationUs
    );
  }
}

/** Fluent API for building and executing traced operations */
export class OperationBuilder {
  private type: string = 'transform';
  private tags: string[] = [];
  private inputData: Record<string, unknown> = {};
  private contextData: Record<string, unknown> = {};
  private _source?: SourceInfo;
  private ruleId?: string;
  private ruleVersion?: string;

  constructor(
    private cli: CliExecutor,
    private _name: string
  ) {}

  /** Set operation type */
  withType(type: string): OperationBuilder {
    this.type = type;
    return this;
  }

  /** Set input data field */
  input(key: string, value: unknown): OperationBuilder {
    this.inputData[key] = value;
    return this;
  }

  /** Set multiple input fields */
  inputs(data: Record<string, unknown>): OperationBuilder {
    Object.assign(this.inputData, data);
    return this;
  }

  /** Set the data source */
  source(source: SourceInfo): OperationBuilder {
    this._source = source;
    return this;
  }

  /** Add tags */
  tag(...tags: string[]): OperationBuilder {
    this.tags.push(...tags);
    return this;
  }

  /** Set rule information */
  withRule(ruleId: string, version?: string): OperationBuilder {
    this.ruleId = ruleId;
    this.ruleVersion = version;
    return this;
  }

  /** Add context information */
  context(key: string, value: unknown): OperationBuilder {
    this.contextData[key] = value;
    return this;
  }

  /** Execute the operation with an async function */
  async execute<T>(fn: (ctx: OperationContext) => T | Promise<T>): Promise<T> {
    const op = this.build();
    const ctx = new OperationContext(op);

    try {
      const result = await fn(ctx);
      op.success();
      return result;
    } catch (e) {
      op.error(String(e), Severity.MEDIUM);
      throw e;
    } finally {
      await op.end();
    }
  }

  /** Build the operation for manual control */
  build(): Operation {
    const op = new Operation(this.cli, this._name, this.type, [...this.tags]);

    // Apply builder settings
    for (const [key, value] of Object.entries(this.inputData)) {
      op.input(key, value);
    }

    if (this._source) {
      op.source(this._source);
    }

    for (const [key, value] of Object.entries(this.contextData)) {
      op.context(key, value);
    }

    if (this.ruleId) {
      op.context('rule_id', this.ruleId);
      if (this.ruleVersion) {
        op.context('rule_version', this.ruleVersion);
      }
    }

    return op;
  }
}
