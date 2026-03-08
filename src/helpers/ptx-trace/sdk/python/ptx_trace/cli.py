"""PTX-TRACE SDK CLI Wrapper - Subprocess calls to portunix trace commands."""

import json
import os
import subprocess
from typing import Any, Dict, List, Optional

from .models import SessionInfo, TraceEvent


class CLIError(Exception):
    """Exception raised when CLI command fails."""
    def __init__(self, message: str, exit_code: int = 1, stderr: str = ""):
        super().__init__(message)
        self.exit_code = exit_code
        self.stderr = stderr


class CLIExecutor:
    """Wrapper for executing portunix trace CLI commands."""

    def __init__(self, binary_path: Optional[str] = None):
        """
        Initialize CLI executor.

        Args:
            binary_path: Path to portunix binary. If None, uses 'portunix' from PATH.
        """
        self.binary_path = binary_path or "portunix"

    def _run(self, *args: str, capture_json: bool = False, parse_text: bool = False) -> Any:
        """
        Execute a CLI command.

        Args:
            *args: Command arguments
            capture_json: If True, add --format json and parse output as JSON
            parse_text: If True, return stdout as text

        Returns:
            Parsed JSON object, text output, or None
        """
        cmd = [self.binary_path, "trace", *args]

        if capture_json:
            cmd.extend(["--format", "json"])

        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=30,
            )
        except subprocess.TimeoutExpired:
            raise CLIError("Command timed out", exit_code=-1)
        except FileNotFoundError:
            raise CLIError(f"Portunix binary not found: {self.binary_path}", exit_code=-1)

        if result.returncode != 0:
            raise CLIError(
                f"Command failed: {result.stderr.strip()}",
                exit_code=result.returncode,
                stderr=result.stderr,
            )

        if capture_json and result.stdout.strip():
            try:
                return json.loads(result.stdout)
            except json.JSONDecodeError as e:
                raise CLIError(f"Failed to parse JSON output: {e}")

        if parse_text:
            return result.stdout.strip()

        return None

    def start_session(
        self,
        name: str,
        source: Optional[str] = None,
        destination: Optional[str] = None,
        tags: Optional[List[str]] = None,
        sampling: float = 1.0,
        pii_mask: bool = False,
    ) -> str:
        """
        Start a new trace session.

        Args:
            name: Session name
            source: Source data file or URL
            destination: Destination connection string
            tags: List of tags
            sampling: Sampling rate (0.0-1.0)
            pii_mask: Enable PII masking

        Returns:
            Session ID
        """
        args = ["start", name]

        if source:
            args.extend(["--source", source])
        if destination:
            args.extend(["--destination", destination])
        if tags:
            for tag in tags:
                args.extend(["--tag", tag])
        if sampling < 1.0:
            args.extend(["--sampling", str(sampling)])
        if pii_mask:
            args.append("--pii-mask")

        output = self._run(*args, parse_text=True)

        # Parse session ID from output like "Session started: ses_2026-01-27_import"
        if output:
            for line in output.split("\n"):
                if line.startswith("Session started:"):
                    return line.split(":", 1)[1].strip()

        raise CLIError("Failed to get session ID from output")

    def end_session(self, status: str = "completed", summary: bool = False) -> None:
        """
        End the active trace session.

        Args:
            status: Session status (completed, failed, cancelled)
            summary: Show session summary
        """
        args = ["end", "--status", status]
        if summary:
            args.append("--summary")

        self._run(*args)

    def add_event(
        self,
        operation: str,
        input_data: Optional[Dict[str, Any]] = None,
        output_data: Optional[Dict[str, Any]] = None,
        status: str = "success",
        error: Optional[str] = None,
        tags: Optional[List[str]] = None,
        duration: Optional[int] = None,
    ) -> None:
        """
        Add a trace event to the active session.

        Args:
            operation: Operation name
            input_data: Input data fields
            output_data: Output data fields
            status: Event status
            error: Error message
            tags: List of tags
            duration: Duration in microseconds
        """
        args = ["event", operation]

        if input_data:
            input_str = ",".join(f"{k}={v}" for k, v in input_data.items())
            args.extend(["--input", input_str])

        if output_data:
            output_str = ",".join(f"{k}={v}" for k, v in output_data.items())
            args.extend(["--output", output_str])

        if error:
            args.extend(["--error", error])
        elif status:
            args.extend(["--status", status])

        if tags:
            for tag in tags:
                args.extend(["--tag", tag])

        if duration is not None:
            args.extend(["--duration", str(duration)])

        self._run(*args)

    def list_sessions(
        self,
        limit: Optional[int] = None,
        status: Optional[str] = None,
    ) -> List[SessionInfo]:
        """
        List trace sessions.

        Args:
            limit: Maximum number of sessions
            status: Filter by status

        Returns:
            List of session info objects
        """
        args = ["sessions"]

        if limit:
            args.extend(["--limit", str(limit)])
        if status:
            args.extend(["--status", status])

        data = self._run(*args, capture_json=True)

        if not data:
            return []

        return [SessionInfo.from_dict(s) for s in data]

    def get_session_stats(self, session_id: Optional[str] = None) -> SessionInfo:
        """
        Get session statistics.

        Args:
            session_id: Session ID (uses active/most recent if not specified)

        Returns:
            Session info with stats
        """
        args = ["stats"]
        if session_id:
            args.append(session_id)

        data = self._run(*args, capture_json=True)
        return SessionInfo.from_dict(data)

    def view_events(
        self,
        session_id: Optional[str] = None,
        operation: Optional[str] = None,
        status: Optional[str] = None,
        level: Optional[str] = None,
        tag: Optional[str] = None,
        limit: int = 100,
    ) -> List[TraceEvent]:
        """
        View trace events.

        Args:
            session_id: Session ID (uses active/most recent if not specified)
            operation: Filter by operation name
            status: Filter by status
            level: Filter by level
            tag: Filter by tag
            limit: Maximum number of events

        Returns:
            List of trace events
        """
        args = ["view"]

        if session_id:
            args.append(session_id)
        if operation:
            args.extend(["--operation", operation])
        if status:
            args.extend(["--status", status])
        if level:
            args.extend(["--level", level])
        if tag:
            args.extend(["--tag", tag])
        args.extend(["--limit", str(limit)])

        data = self._run(*args, capture_json=True)

        if not data:
            return []

        return [TraceEvent.from_dict(e) for e in data]

    def query_events(
        self,
        query: str,
        session_id: Optional[str] = None,
        limit: int = 100,
    ) -> List[Dict[str, Any]]:
        """
        Query events using SQL-like syntax.

        Args:
            query: SQL-like query string
            session_id: Session ID
            limit: Maximum number of results

        Returns:
            List of query results
        """
        args = ["query", query]

        if session_id:
            args.extend(["--session", session_id])
        args.extend(["--limit", str(limit)])

        data = self._run(*args, capture_json=True)
        return data or []
