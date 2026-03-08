/**
 * PTX-TRACE SDK Models - Data types and interfaces.
 */

/** Error severity levels */
export enum Severity {
  LOW = 'low',
  MEDIUM = 'medium',
  HIGH = 'high',
  CRITICAL = 'critical',
}

/** Log levels */
export enum Level {
  DEBUG = 'debug',
  INFO = 'info',
  WARNING = 'warning',
  ERROR = 'error',
}

/** Session status values */
export enum SessionStatus {
  ACTIVE = 'active',
  COMPLETED = 'completed',
  FAILED = 'failed',
  CANCELLED = 'cancelled',
}

/** Data source information */
export interface SourceInfo {
  type: string;
  file?: string;
  row?: number;
  column?: string;
  url?: string;
  table?: string;
}

/** Create a CSV file source */
export function csvSource(file: string, row?: number, column?: string): SourceInfo {
  return { type: 'csv', file, row, column };
}

/** Create a database source */
export function dbSource(url: string, table: string, row?: number): SourceInfo {
  return { type: 'database', url, table, row };
}

/** Create a file source */
export function fileSource(file: string, row?: number): SourceInfo {
  return { type: 'file', file, row };
}

/** Create an API source */
export function apiSource(url: string): SourceInfo {
  return { type: 'api', url };
}

/** Error information */
export interface ErrorInfo {
  code: string;
  message: string;
  severity: Severity;
  category?: string;
  details?: Record<string, unknown>;
  suggestion?: string;
}

/** Recovery information */
export interface RecoveryInfo {
  attempted: boolean;
  strategy: string;
  success: boolean;
}

/** Operation statistics */
export interface OperationStats {
  count: number;
  avgDuration: number;
  minDuration?: number;
  maxDuration?: number;
  totalErrors?: number;
}

/** Session statistics */
export interface SessionStats {
  totalEvents: number;
  byStatus: Record<string, number>;
  byOperation?: Record<string, OperationStats>;
  byLevel?: Record<string, number>;
}

/** Session information */
export interface SessionInfo {
  id: string;
  name: string;
  status: SessionStatus;
  startedAt: Date;
  endedAt?: Date;
  tags?: string[];
  stats?: SessionStats;
}

/** Trace event information */
export interface TraceEvent {
  id: string;
  traceId: string;
  sessionId: string;
  timestamp: Date;
  operationType: string;
  operationName: string;
  level: Level;
  durationUs?: number;
  inputFields?: Record<string, unknown>;
  outputFields?: Record<string, unknown>;
  outputStatus?: string;
  error?: ErrorInfo;
  tags?: string[];
  context?: Record<string, unknown>;
  parentId?: string;
}

/** Session creation options */
export interface SessionOptions {
  source?: string;
  destination?: string;
  tags?: string[];
  sampling?: number;
  piiMasking?: boolean;
  binaryPath?: string;
}

/** Event query options */
export interface EventQueryOptions {
  operation?: string;
  status?: string;
  level?: string;
  tag?: string;
  limit?: number;
}

/** Parse session info from CLI JSON */
export function parseSessionInfo(data: Record<string, unknown>): SessionInfo {
  return {
    id: data.id as string,
    name: data.name as string,
    status: (data.status as string) as SessionStatus,
    startedAt: new Date(data.started_at as string),
    endedAt: data.ended_at ? new Date(data.ended_at as string) : undefined,
    tags: data.tags as string[] | undefined,
    stats: data.stats ? parseSessionStats(data.stats as Record<string, unknown>) : undefined,
  };
}

/** Parse session stats from CLI JSON */
function parseSessionStats(data: Record<string, unknown>): SessionStats {
  const byOperation: Record<string, OperationStats> = {};
  if (data.by_operation) {
    for (const [name, stats] of Object.entries(data.by_operation as Record<string, unknown>)) {
      const s = stats as Record<string, unknown>;
      byOperation[name] = {
        count: s.count as number,
        avgDuration: s.avg_us as number,
        minDuration: s.min_us as number | undefined,
        maxDuration: s.max_us as number | undefined,
        totalErrors: s.errors as number | undefined,
      };
    }
  }

  return {
    totalEvents: data.total_events as number,
    byStatus: data.by_status as Record<string, number>,
    byOperation: Object.keys(byOperation).length > 0 ? byOperation : undefined,
    byLevel: data.by_level as Record<string, number> | undefined,
  };
}

/** Parse trace event from CLI JSON */
export function parseTraceEvent(data: Record<string, unknown>): TraceEvent {
  const operation = data.operation as Record<string, unknown>;
  const input = data.input as Record<string, unknown> | undefined;
  const output = data.output as Record<string, unknown> | undefined;
  const errorData = data.error as Record<string, unknown> | undefined;

  return {
    id: data.id as string,
    traceId: data.trace_id as string,
    sessionId: data.session_id as string,
    timestamp: new Date(data.timestamp as string),
    operationType: operation.type as string,
    operationName: operation.name as string,
    level: (data.level as string) as Level,
    durationUs: data.duration_us as number | undefined,
    inputFields: input?.fields as Record<string, unknown> | undefined,
    outputFields: output?.fields as Record<string, unknown> | undefined,
    outputStatus: output?.status as string | undefined,
    error: errorData
      ? {
          code: errorData.code as string,
          message: errorData.message as string,
          severity: (errorData.severity as string) as Severity,
          category: errorData.category as string | undefined,
          details: errorData.details as Record<string, unknown> | undefined,
          suggestion: errorData.suggestion as string | undefined,
        }
      : undefined,
    tags: data.tags as string[] | undefined,
    context: data.context as Record<string, unknown> | undefined,
    parentId: data.parent_id as string | undefined,
  };
}
