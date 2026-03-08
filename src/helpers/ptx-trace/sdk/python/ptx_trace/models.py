"""PTX-TRACE SDK Models - Data types and enums."""

from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional


class Severity(Enum):
    """Error severity levels."""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class Level(Enum):
    """Log levels."""
    DEBUG = "debug"
    INFO = "info"
    WARNING = "warning"
    ERROR = "error"


class SessionStatus(Enum):
    """Session status values."""
    ACTIVE = "active"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


@dataclass
class SourceInfo:
    """Data source information."""
    type: str
    file: Optional[str] = None
    row: Optional[int] = None
    column: Optional[str] = None
    url: Optional[str] = None
    table: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        result = {"type": self.type}
        if self.file:
            result["file"] = self.file
        if self.row is not None:
            result["row"] = self.row
        if self.column:
            result["column"] = self.column
        if self.url:
            result["url"] = self.url
        if self.table:
            result["table"] = self.table
        return result


class CsvSource(SourceInfo):
    """CSV file data source."""
    def __init__(self, file: str, row: Optional[int] = None, column: Optional[str] = None):
        super().__init__(type="csv", file=file, row=row, column=column)


class DbSource(SourceInfo):
    """Database data source."""
    def __init__(self, url: str, table: str, row: Optional[int] = None):
        super().__init__(type="database", url=url, table=table, row=row)


class FileSource(SourceInfo):
    """Generic file data source."""
    def __init__(self, file: str, row: Optional[int] = None):
        super().__init__(type="file", file=file, row=row)


class ApiSource(SourceInfo):
    """API endpoint data source."""
    def __init__(self, url: str):
        super().__init__(type="api", url=url)


@dataclass
class ErrorInfo:
    """Error information."""
    code: str
    message: str
    severity: Severity = Severity.MEDIUM
    category: Optional[str] = None
    details: Optional[Dict[str, Any]] = None
    suggestion: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        result = {
            "code": self.code,
            "message": self.message,
            "severity": self.severity.value,
        }
        if self.category:
            result["category"] = self.category
        if self.details:
            result["details"] = self.details
        if self.suggestion:
            result["suggestion"] = self.suggestion
        return result


@dataclass
class RecoveryInfo:
    """Recovery attempt information."""
    attempted: bool = True
    strategy: str = ""
    success: bool = False

    def to_dict(self) -> Dict[str, Any]:
        return {
            "attempted": self.attempted,
            "strategy": self.strategy,
            "success": self.success,
        }


@dataclass
class OperationStats:
    """Statistics for a single operation type."""
    count: int = 0
    avg_duration: float = 0.0
    min_duration: int = 0
    max_duration: int = 0
    total_errors: int = 0

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "OperationStats":
        return cls(
            count=data.get("count", 0),
            avg_duration=data.get("avg_us", 0.0),
            min_duration=data.get("min_us", 0),
            max_duration=data.get("max_us", 0),
            total_errors=data.get("errors", 0),
        )


@dataclass
class SessionStats:
    """Session statistics."""
    total_events: int = 0
    by_status: Dict[str, int] = field(default_factory=dict)
    by_operation: Dict[str, OperationStats] = field(default_factory=dict)
    by_level: Dict[str, int] = field(default_factory=dict)

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "SessionStats":
        by_operation = {}
        if data.get("by_operation"):
            for op_name, op_data in data["by_operation"].items():
                by_operation[op_name] = OperationStats.from_dict(op_data)

        return cls(
            total_events=data.get("total_events", 0),
            by_status=data.get("by_status", {}),
            by_operation=by_operation,
            by_level=data.get("by_level", {}),
        )


@dataclass
class SessionInfo:
    """Session information returned from CLI."""
    id: str
    name: str
    status: SessionStatus
    started_at: datetime
    ended_at: Optional[datetime] = None
    tags: List[str] = field(default_factory=list)
    stats: Optional[SessionStats] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "SessionInfo":
        status = SessionStatus(data.get("status", "active"))
        started_at = datetime.fromisoformat(data["started_at"].replace("Z", "+00:00"))
        ended_at = None
        if data.get("ended_at"):
            ended_at = datetime.fromisoformat(data["ended_at"].replace("Z", "+00:00"))

        stats = None
        if data.get("stats"):
            stats = SessionStats.from_dict(data["stats"])

        return cls(
            id=data["id"],
            name=data["name"],
            status=status,
            started_at=started_at,
            ended_at=ended_at,
            tags=data.get("tags", []),
            stats=stats,
        )


@dataclass
class TraceEvent:
    """Trace event information."""
    id: str
    trace_id: str
    session_id: str
    timestamp: datetime
    operation_type: str
    operation_name: str
    level: Level = Level.INFO
    duration_us: int = 0
    input_fields: Dict[str, Any] = field(default_factory=dict)
    output_fields: Dict[str, Any] = field(default_factory=dict)
    output_status: Optional[str] = None
    error: Optional[ErrorInfo] = None
    tags: List[str] = field(default_factory=list)
    context: Dict[str, Any] = field(default_factory=dict)
    parent_id: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "TraceEvent":
        timestamp = datetime.fromisoformat(data["timestamp"].replace("Z", "+00:00"))
        level = Level(data.get("level", "info"))

        error = None
        if data.get("error"):
            err_data = data["error"]
            error = ErrorInfo(
                code=err_data.get("code", ""),
                message=err_data.get("message", ""),
                severity=Severity(err_data.get("severity", "medium")),
                category=err_data.get("category"),
                details=err_data.get("details"),
                suggestion=err_data.get("suggestion"),
            )

        input_fields = {}
        output_fields = {}
        output_status = None

        if data.get("input") and data["input"].get("fields"):
            input_fields = data["input"]["fields"]
        if data.get("output"):
            if data["output"].get("fields"):
                output_fields = data["output"]["fields"]
            output_status = data["output"].get("status")

        return cls(
            id=data["id"],
            trace_id=data["trace_id"],
            session_id=data["session_id"],
            timestamp=timestamp,
            operation_type=data["operation"]["type"],
            operation_name=data["operation"]["name"],
            level=level,
            duration_us=data.get("duration_us", 0),
            input_fields=input_fields,
            output_fields=output_fields,
            output_status=output_status,
            error=error,
            tags=data.get("tags", []),
            context=data.get("context", {}),
            parent_id=data.get("parent_id"),
        )
