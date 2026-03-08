"""
PTX-TRACE Python SDK

A Python SDK for the Portunix Trace system, enabling tracing and debugging
of data pipelines, ETL processes, and software workflows.

Example usage:

    from ptx_trace import Session, Severity, CsvSource

    # Using context manager
    with Session("import-customers", pii_masking=True) as session:
        with session.operation("normalize_phone") as op:
            op.input(phone=raw_phone)
            result = normalize(raw_phone)
            op.output(phone=result)
            op.success()

    # Using decorator
    @session.traced("validate_email", tags=["validation"])
    def validate_email(email: str) -> dict:
        return {"valid": "@" in email}

    # Using fluent API
    session.trace("transform_record") \\
        .input(record=record) \\
        .source(CsvSource("data.csv", row=42)) \\
        .tag("transform") \\
        .execute(lambda ctx: ctx.output(result=transform(record)))
"""

from .models import (
    Severity,
    Level,
    SessionStatus,
    SourceInfo,
    CsvSource,
    DbSource,
    FileSource,
    ApiSource,
    ErrorInfo,
    RecoveryInfo,
    SessionInfo,
    SessionStats,
    TraceEvent,
)

from .session import Session
from .operation import Operation, OperationBuilder, OperationContext
from .cli import CLIExecutor, CLIError

__version__ = "1.0.0"

__all__ = [
    # Main classes
    "Session",
    "Operation",
    "OperationBuilder",
    "OperationContext",
    # Enums
    "Severity",
    "Level",
    "SessionStatus",
    # Source types
    "SourceInfo",
    "CsvSource",
    "DbSource",
    "FileSource",
    "ApiSource",
    # Data models
    "ErrorInfo",
    "RecoveryInfo",
    "SessionInfo",
    "SessionStats",
    "TraceEvent",
    # CLI
    "CLIExecutor",
    "CLIError",
    # Version
    "__version__",
]
