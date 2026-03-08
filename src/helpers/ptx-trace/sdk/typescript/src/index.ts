/**
 * PTX-TRACE TypeScript SDK
 *
 * A TypeScript/JavaScript SDK for the Portunix Trace system, enabling tracing
 * and debugging of data pipelines, ETL processes, and software workflows.
 *
 * @example Basic usage
 * ```typescript
 * import { Session, Severity, csvSource } from 'ptx-trace';
 *
 * const session = await Session.create("import-customers", {
 *   piiMasking: true,
 *   tags: ["production"]
 * });
 *
 * try {
 *   await session.withOperation("normalize_phone", async (op) => {
 *     op.input("phone", rawPhone);
 *     const result = await normalize(rawPhone);
 *     op.output("phone", result);
 *   });
 * } finally {
 *   await session.close();
 * }
 * ```
 *
 * @example Fluent API
 * ```typescript
 * await session.trace("transform_record")
 *   .input("record", record)
 *   .source(csvSource("data.csv", 42))
 *   .tag("transform")
 *   .execute(async (ctx) => {
 *     const result = await transform(record);
 *     ctx.output("result", result);
 *   });
 * ```
 */

// Models and types
export {
  Severity,
  Level,
  SessionStatus,
  SourceInfo,
  csvSource,
  dbSource,
  fileSource,
  apiSource,
  ErrorInfo,
  RecoveryInfo,
  SessionInfo,
  SessionStats,
  OperationStats,
  TraceEvent,
  SessionOptions,
  EventQueryOptions,
} from './models';

// Session
export { Session } from './session';

// Operation
export { Operation, OperationBuilder, OperationContext } from './operation';

// CLI
export { CliExecutor, CliError } from './cli';
