"""PTX-TRACE SDK Operation - Traced operation classes."""

from contextlib import contextmanager
import time
from typing import TYPE_CHECKING, Any, Callable, Dict, List, Optional

from .models import Severity, SourceInfo, RecoveryInfo

if TYPE_CHECKING:
    from .session import Session


class OperationContext:
    """Context passed to traced functions for recording output and metadata."""

    def __init__(self, operation: "Operation"):
        self._operation = operation

    def output(self, key: str, value: Any) -> "OperationContext":
        """Set an output field."""
        self._operation.output(key, value)
        return self

    def outputs(self, **kwargs: Any) -> "OperationContext":
        """Set multiple output fields."""
        for key, value in kwargs.items():
            self._operation.output(key, value)
        return self

    def tag(self, tag: str) -> "OperationContext":
        """Add a tag."""
        self._operation.tag(tag)
        return self

    def context(self, key: str, value: Any) -> "OperationContext":
        """Add context information."""
        self._operation.context(key, value)
        return self


class Operation:
    """
    Represents a traced operation.

    Use as a context manager for automatic timing and recording:

        with session.trace("normalize_phone") as op:
            op.input(phone=raw_phone)
            result = normalize(raw_phone)
            op.output(phone=result)
            op.success()
    """

    def __init__(
        self,
        session: "Session",
        name: str,
        op_type: str = "transform",
        tags: Optional[List[str]] = None,
    ):
        self._session = session
        self._name = name
        self._type = op_type
        self._tags = tags or []
        self._input_data: Dict[str, Any] = {}
        self._output_data: Dict[str, Any] = {}
        self._context_data: Dict[str, Any] = {}
        self._source: Optional[SourceInfo] = None
        self._error_msg: Optional[str] = None
        self._error_severity: Severity = Severity.MEDIUM
        self._recovery: Optional[RecoveryInfo] = None
        self._status: str = ""
        self._start_time: float = 0
        self._duration_us: Optional[int] = None
        self._is_ended: bool = False

    @property
    def name(self) -> str:
        """Operation name."""
        return self._name

    def __enter__(self) -> "Operation":
        self._start_time = time.time()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if exc_type is not None:
            # Exception occurred - record as error
            self.error(str(exc_val), Severity.HIGH)

        self.end()
        return False  # Don't suppress exceptions

    def input(self, key: str = None, value: Any = None, **kwargs) -> "Operation":
        """
        Set input data.

        Can be called as:
            op.input("phone", raw_phone)
            op.input(phone=raw_phone, email=raw_email)
        """
        if key is not None and value is not None:
            self._input_data[key] = value
        for k, v in kwargs.items():
            self._input_data[k] = v
        return self

    def inputs(self, data: Dict[str, Any]) -> "Operation":
        """Set multiple input fields from a dictionary."""
        self._input_data.update(data)
        return self

    def output(self, key: str = None, value: Any = None, **kwargs) -> "Operation":
        """
        Set output data.

        Can be called as:
            op.output("phone", normalized_phone)
            op.output(phone=normalized_phone, valid=True)
        """
        if key is not None and value is not None:
            self._output_data[key] = value
        for k, v in kwargs.items():
            self._output_data[k] = v
        return self

    def outputs(self, data: Dict[str, Any]) -> "Operation":
        """Set multiple output fields from a dictionary."""
        self._output_data.update(data)
        return self

    def source(self, source: SourceInfo) -> "Operation":
        """Set the data source."""
        self._source = source
        return self

    def tag(self, *tags: str) -> "Operation":
        """Add one or more tags."""
        self._tags.extend(tags)
        return self

    def context(self, key: str, value: Any) -> "Operation":
        """Add context information."""
        self._context_data[key] = value
        return self

    def error(self, message: str, severity: Severity = Severity.MEDIUM) -> "Operation":
        """Record an error."""
        self._error_msg = message
        self._error_severity = severity
        return self

    def error_with_code(self, code: str, message: str, severity: Severity = Severity.MEDIUM) -> "Operation":
        """Record an error with a specific code."""
        self._error_msg = f"[{code}] {message}"
        self._error_severity = severity
        return self

    def recovery(self, strategy: str, success: bool) -> "Operation":
        """Record a recovery attempt."""
        self._recovery = RecoveryInfo(attempted=True, strategy=strategy, success=success)
        return self

    def success(self) -> "Operation":
        """Mark the operation as successful."""
        self._status = "success"
        return self

    def set_duration(self, duration_us: int) -> "Operation":
        """Set duration manually in microseconds."""
        self._duration_us = duration_us
        return self

    def end(self) -> None:
        """End the operation and record it."""
        if self._is_ended:
            return

        self._is_ended = True

        # Calculate duration if not set manually
        if self._duration_us is None and self._start_time > 0:
            self._duration_us = int((time.time() - self._start_time) * 1_000_000)

        # Record event via CLI
        self._session._cli.add_event(
            operation=self._name,
            input_data=self._input_data if self._input_data else None,
            output_data=self._output_data if self._output_data else None,
            status=self._status if not self._error_msg else None,
            error=self._error_msg,
            tags=self._tags if self._tags else None,
            duration=self._duration_us,
        )


class OperationBuilder:
    """
    Fluent API for building and executing traced operations.

    Example:
        session.trace("validate_email") \\
            .input("email", email) \\
            .tag("validation") \\
            .execute(lambda ctx: ctx.output("valid", validate(email)))
    """

    def __init__(
        self,
        session: "Session",
        name: str,
        op_type: str = "transform",
    ):
        self._session = session
        self._name = name
        self._type = op_type
        self._tags: List[str] = []
        self._input_data: Dict[str, Any] = {}
        self._source: Optional[SourceInfo] = None
        self._context_data: Dict[str, Any] = {}
        self._rule_id: Optional[str] = None
        self._rule_version: Optional[str] = None

    def with_type(self, op_type: str) -> "OperationBuilder":
        """Set operation type."""
        self._type = op_type
        return self

    def input(self, key: str = None, value: Any = None, **kwargs) -> "OperationBuilder":
        """Set input data."""
        if key is not None and value is not None:
            self._input_data[key] = value
        for k, v in kwargs.items():
            self._input_data[k] = v
        return self

    def inputs(self, data: Dict[str, Any]) -> "OperationBuilder":
        """Set multiple input fields."""
        self._input_data.update(data)
        return self

    def source(self, source: SourceInfo) -> "OperationBuilder":
        """Set the data source."""
        self._source = source
        return self

    def tag(self, *tags: str) -> "OperationBuilder":
        """Add tags."""
        self._tags.extend(tags)
        return self

    def with_rule(self, rule_id: str, version: str = "") -> "OperationBuilder":
        """Set rule information."""
        self._rule_id = rule_id
        self._rule_version = version
        return self

    def context(self, key: str, value: Any) -> "OperationBuilder":
        """Add context information."""
        self._context_data[key] = value
        return self

    def execute(self, fn: Callable[[OperationContext], Any]) -> Any:
        """
        Execute the operation with a function.

        The function receives an OperationContext for recording outputs.

        Args:
            fn: Function to execute, receives OperationContext

        Returns:
            Result of the function
        """
        op = Operation(
            session=self._session,
            name=self._name,
            op_type=self._type,
            tags=self._tags.copy(),
        )

        # Apply builder settings
        for k, v in self._input_data.items():
            op.input(k, v)

        if self._source:
            op.source(self._source)

        for k, v in self._context_data.items():
            op.context(k, v)

        if self._rule_id:
            op.context("rule_id", self._rule_id)
            if self._rule_version:
                op.context("rule_version", self._rule_version)

        ctx = OperationContext(op)

        with op:
            try:
                result = fn(ctx)
                op.success()
                return result
            except Exception as e:
                op.error(str(e), Severity.MEDIUM)
                raise

    def start(self) -> Operation:
        """
        Start the operation manually (for use with context manager).

        Returns:
            Operation instance for context manager use
        """
        op = Operation(
            session=self._session,
            name=self._name,
            op_type=self._type,
            tags=self._tags.copy(),
        )

        # Apply builder settings
        for k, v in self._input_data.items():
            op.input(k, v)

        if self._source:
            op.source(self._source)

        for k, v in self._context_data.items():
            op.context(k, v)

        if self._rule_id:
            op.context("rule_id", self._rule_id)
            if self._rule_version:
                op.context("rule_version", self._rule_version)

        return op
