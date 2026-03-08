"""PTX-TRACE SDK Session - Manages tracing session lifecycle."""

from contextlib import contextmanager
import functools
from typing import Any, Callable, Dict, List, Optional, TypeVar

from .cli import CLIExecutor
from .models import SessionInfo, SessionStatus, TraceEvent
from .operation import Operation, OperationBuilder

T = TypeVar("T")


class Session:
    """
    Manages a PTX-TRACE session.

    Can be used as a context manager for automatic session lifecycle:

        with Session("import-customers", pii_masking=True) as session:
            with session.trace("normalize_phone") as op:
                op.input(phone=raw_phone)
                result = normalize(raw_phone)
                op.output(phone=result)
                op.success()

    Or manually controlled:

        session = Session.create("import-customers")
        try:
            # ... operations ...
        finally:
            session.close()
    """

    def __init__(
        self,
        name: str,
        source: Optional[str] = None,
        destination: Optional[str] = None,
        tags: Optional[List[str]] = None,
        sampling: float = 1.0,
        pii_masking: bool = False,
        binary_path: Optional[str] = None,
        _session_id: Optional[str] = None,
    ):
        """
        Initialize a session.

        Args:
            name: Session name
            source: Source data file or URL
            destination: Destination connection string
            tags: List of tags
            sampling: Sampling rate (0.0-1.0)
            pii_masking: Enable PII masking
            binary_path: Path to portunix binary
            _session_id: Internal - existing session ID
        """
        self._cli = CLIExecutor(binary_path)
        self._name = name
        self._source = source
        self._destination = destination
        self._tags = tags or []
        self._sampling = sampling
        self._pii_masking = pii_masking
        self._session_id: Optional[str] = _session_id
        self._is_started = _session_id is not None
        self._is_closed = False

    @classmethod
    def create(
        cls,
        name: str,
        source: Optional[str] = None,
        destination: Optional[str] = None,
        tags: Optional[List[str]] = None,
        sampling: float = 1.0,
        pii_masking: bool = False,
        binary_path: Optional[str] = None,
    ) -> "Session":
        """
        Create and start a new session.

        This is the preferred way to create a session when not using
        the context manager.

        Args:
            name: Session name
            source: Source data file or URL
            destination: Destination connection string
            tags: List of tags
            sampling: Sampling rate (0.0-1.0)
            pii_masking: Enable PII masking
            binary_path: Path to portunix binary

        Returns:
            Started Session instance
        """
        session = cls(
            name=name,
            source=source,
            destination=destination,
            tags=tags,
            sampling=sampling,
            pii_masking=pii_masking,
            binary_path=binary_path,
        )
        session.start()
        return session

    @classmethod
    def load(cls, session_id: str, binary_path: Optional[str] = None) -> "Session":
        """
        Load an existing session by ID.

        Args:
            session_id: Session ID to load
            binary_path: Path to portunix binary

        Returns:
            Session instance
        """
        cli = CLIExecutor(binary_path)
        info = cli.get_session_stats(session_id)

        return cls(
            name=info.name,
            tags=info.tags,
            binary_path=binary_path,
            _session_id=session_id,
        )

    @classmethod
    def list(cls, limit: Optional[int] = None, status: Optional[str] = None, binary_path: Optional[str] = None) -> List[SessionInfo]:
        """
        List all sessions.

        Args:
            limit: Maximum number of sessions
            status: Filter by status
            binary_path: Path to portunix binary

        Returns:
            List of SessionInfo objects
        """
        cli = CLIExecutor(binary_path)
        return cli.list_sessions(limit=limit, status=status)

    @property
    def id(self) -> Optional[str]:
        """Session ID."""
        return self._session_id

    @property
    def name(self) -> str:
        """Session name."""
        return self._name

    @property
    def is_active(self) -> bool:
        """Whether session is active."""
        return self._is_started and not self._is_closed

    def __enter__(self) -> "Session":
        if not self._is_started:
            self.start()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if exc_type is not None:
            self.end(SessionStatus.FAILED)
        else:
            self.close()
        return False

    def start(self) -> str:
        """
        Start the session.

        Returns:
            Session ID
        """
        if self._is_started:
            return self._session_id

        self._session_id = self._cli.start_session(
            name=self._name,
            source=self._source,
            destination=self._destination,
            tags=self._tags,
            sampling=self._sampling,
            pii_mask=self._pii_masking,
        )
        self._is_started = True
        return self._session_id

    def end(self, status: SessionStatus = SessionStatus.COMPLETED) -> None:
        """
        End the session with a specific status.

        Args:
            status: Session status
        """
        if self._is_closed:
            return

        self._cli.end_session(status=status.value)
        self._is_closed = True

    def close(self) -> None:
        """End the session successfully."""
        self.end(SessionStatus.COMPLETED)

    def fail(self) -> None:
        """End the session as failed."""
        self.end(SessionStatus.FAILED)

    def cancel(self) -> None:
        """End the session as cancelled."""
        self.end(SessionStatus.CANCELLED)

    def trace(self, operation_name: str) -> OperationBuilder:
        """
        Create a traced operation builder.

        Use the fluent API or as a context manager:

            # Fluent API
            session.trace("validate") \\
                .input("email", email) \\
                .execute(lambda ctx: ctx.output("valid", True))

            # Context manager
            with session.trace("validate").start() as op:
                op.input("email", email)
                op.output("valid", True)
                op.success()

        Args:
            operation_name: Name of the operation

        Returns:
            OperationBuilder for fluent configuration
        """
        return OperationBuilder(session=self, name=operation_name)

    def start_operation(self, operation_name: str, op_type: str = "transform") -> Operation:
        """
        Start a traced operation directly.

        Args:
            operation_name: Name of the operation
            op_type: Operation type

        Returns:
            Operation instance for context manager use
        """
        return Operation(session=self, name=operation_name, op_type=op_type)

    @contextmanager
    def operation(self, name: str, op_type: str = "transform"):
        """
        Context manager for a traced operation.

        Args:
            name: Operation name
            op_type: Operation type

        Yields:
            Operation instance
        """
        op = Operation(session=self, name=name, op_type=op_type)
        with op:
            yield op

    def traced(
        self,
        operation_name: Optional[str] = None,
        tags: Optional[List[str]] = None,
    ) -> Callable[[Callable[..., T]], Callable[..., T]]:
        """
        Decorator for tracing a function.

        Args:
            operation_name: Name of the operation (defaults to function name)
            tags: Tags to add to the operation

        Returns:
            Decorator function

        Example:
            @session.traced("validate_email", tags=["validation"])
            def validate_email(email: str) -> dict:
                return {"valid": True}
        """
        def decorator(fn: Callable[..., T]) -> Callable[..., T]:
            @functools.wraps(fn)
            def wrapper(*args, **kwargs) -> T:
                op_name = operation_name or fn.__name__

                op = Operation(
                    session=self,
                    name=op_name,
                    tags=tags or [],
                )

                # Record args as input (excluding self for methods)
                input_data = {}
                arg_names = fn.__code__.co_varnames[:fn.__code__.co_argcount]

                for i, arg in enumerate(args):
                    if i < len(arg_names):
                        name = arg_names[i]
                        if name != "self":
                            input_data[name] = _serialize_value(arg)

                for key, value in kwargs.items():
                    input_data[key] = _serialize_value(value)

                if input_data:
                    op.inputs(input_data)

                with op:
                    try:
                        result = fn(*args, **kwargs)
                        op.output("result", _serialize_value(result))
                        op.success()
                        return result
                    except Exception as e:
                        op.error(str(e))
                        raise

            return wrapper
        return decorator

    def stats(self) -> SessionInfo:
        """
        Get session statistics.

        Returns:
            SessionInfo with stats
        """
        return self._cli.get_session_stats(self._session_id)

    def events(
        self,
        operation: Optional[str] = None,
        status: Optional[str] = None,
        level: Optional[str] = None,
        tag: Optional[str] = None,
        limit: int = 100,
    ) -> List[TraceEvent]:
        """
        Get events from this session.

        Args:
            operation: Filter by operation name
            status: Filter by status
            level: Filter by level
            tag: Filter by tag
            limit: Maximum number of events

        Returns:
            List of TraceEvent objects
        """
        return self._cli.view_events(
            session_id=self._session_id,
            operation=operation,
            status=status,
            level=level,
            tag=tag,
            limit=limit,
        )

    def query(self, query: str, limit: int = 100) -> List[Dict[str, Any]]:
        """
        Query events using SQL-like syntax.

        Args:
            query: SQL-like query string
            limit: Maximum number of results

        Returns:
            List of query results
        """
        return self._cli.query_events(
            query=query,
            session_id=self._session_id,
            limit=limit,
        )


def _serialize_value(value: Any) -> Any:
    """Serialize a value for JSON."""
    if value is None:
        return None
    if isinstance(value, (str, int, float, bool)):
        return value
    if isinstance(value, (list, tuple)):
        return [_serialize_value(v) for v in value]
    if isinstance(value, dict):
        return {k: _serialize_value(v) for k, v in value.items()}
    # For complex objects, use string representation
    return str(value)
