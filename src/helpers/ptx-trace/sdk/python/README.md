# PTX-TRACE Python SDK

Python SDK for the Portunix Trace (PTX-TRACE) system - a universal tracing system for software development, optimized for debugging, AI analysis, and monitoring of data pipelines and workflows.

## Installation

```bash
# From source
cd sdk/python
pip install -e .

# Or install directly
pip install .
```

## Requirements

- Python 3.8+
- `portunix` binary in PATH (or specify path)

## Quick Start

### Basic Usage with Context Manager

```python
from ptx_trace import Session

# Session automatically starts on enter, ends on exit
with Session("import-customers", pii_masking=True) as session:
    # Operations are traced with timing
    with session.operation("normalize_phone") as op:
        op.input(phone="+420 777 123 456")
        result = normalize_phone("+420 777 123 456")
        op.output(phone=result)
        op.success()
```

### Using the Decorator

```python
from ptx_trace import Session

session = Session.create("validation-pipeline")

@session.traced("validate_email", tags=["validation"])
def validate_email(email: str) -> dict:
    is_valid = "@" in email and "." in email
    return {"valid": is_valid, "email": email}

# Function calls are automatically traced
result = validate_email("user@example.com")

session.close()
```

### Fluent API

```python
from ptx_trace import Session, CsvSource

with Session("etl-pipeline") as session:
    # Fluent builder pattern
    session.trace("transform_record") \
        .input(record={"name": "John", "phone": "123"}) \
        .source(CsvSource("customers.csv", row=42)) \
        .tag("transform", "customer") \
        .execute(lambda ctx: ctx.output(
            result=transform(record),
            status="transformed"
        ))
```

## API Reference

### Session

The `Session` class manages the tracing session lifecycle.

```python
# Create and start session (manual control)
session = Session.create(
    name="my-session",
    source="input.csv",           # Optional source info
    destination="postgres://...", # Optional destination
    tags=["production"],          # Optional tags
    sampling=0.5,                 # Sample 50% of events
    pii_masking=True,             # Mask PII data
    binary_path="/usr/bin/portunix"  # Custom binary path
)

# Context manager (auto start/end)
with Session("my-session") as session:
    pass

# List existing sessions
sessions = Session.list(limit=10, status="active")

# Load existing session
session = Session.load("ses_2026-01-27_import")
```

#### Session Methods

- `start()` - Start the session (auto-called with context manager)
- `close()` - End session successfully
- `fail()` - End session as failed
- `cancel()` - End session as cancelled
- `trace(name)` - Create operation builder
- `operation(name)` - Context manager for operation
- `start_operation(name)` - Start operation directly
- `traced(name, tags)` - Decorator for tracing functions
- `stats()` - Get session statistics
- `events(...)` - Query session events
- `query(sql)` - SQL-like query on events

### Operation

Operations represent traced units of work.

```python
with session.operation("process_record") as op:
    # Set input data
    op.input("field", value)
    op.input(field1=value1, field2=value2)
    op.inputs({"field": value})

    # Set output data
    op.output("result", value)
    op.outputs({"result": value})

    # Add metadata
    op.tag("important", "customer")
    op.context("batch_id", 123)
    op.source(CsvSource("data.csv", row=42))

    # Record outcome
    op.success()
    # or
    op.error("Validation failed", Severity.HIGH)
    op.recovery("retry", success=True)
```

### OperationBuilder (Fluent API)

```python
result = session.trace("transform") \
    .with_type("etl")                    # Set operation type
    .input("data", input_data)           # Set input
    .source(DbSource("pg://...", "users")) # Set source
    .tag("critical")                      # Add tags
    .with_rule("R001", "1.0")            # Set rule info
    .context("env", "production")         # Add context
    .execute(lambda ctx: transform(data)) # Execute with tracing
```

### Source Types

```python
from ptx_trace import CsvSource, DbSource, FileSource, ApiSource

# CSV file source
source = CsvSource("data.csv", row=42, column="email")

# Database source
source = DbSource(url="postgres://localhost/db", table="users", row=100)

# Generic file source
source = FileSource("config.json", row=10)

# API source
source = ApiSource(url="https://api.example.com/users")
```

### Severity Levels

```python
from ptx_trace import Severity

op.error("Minor issue", Severity.LOW)
op.error("Validation error", Severity.MEDIUM)
op.error("Data corruption", Severity.HIGH)
op.error("System failure", Severity.CRITICAL)
```

## Complete ETL Example

```python
from ptx_trace import Session, CsvSource, DbSource, Severity
import csv

def run_etl():
    with Session("customer-import", pii_masking=True, tags=["etl", "daily"]) as session:

        # Phase 1: Extract
        with session.operation("extract") as op:
            op.source(CsvSource("customers.csv"))

            with open("customers.csv") as f:
                records = list(csv.DictReader(f))

            op.output(record_count=len(records))
            op.success()

        # Phase 2: Transform each record
        for i, record in enumerate(records):
            with session.operation("transform") as op:
                op.source(CsvSource("customers.csv", row=i+1))
                op.input(**record)

                try:
                    # Normalize phone
                    record["phone"] = normalize_phone(record.get("phone", ""))

                    # Validate email
                    if not validate_email(record.get("email", "")):
                        op.error("Invalid email", Severity.MEDIUM)
                        continue

                    op.output(**record)
                    op.success()

                except Exception as e:
                    op.error(str(e), Severity.HIGH)

        # Phase 3: Load
        with session.operation("load") as op:
            op.context("destination", "postgres://localhost/customers")

            inserted = load_to_database(records)

            op.output(inserted=inserted)
            op.success()

        # Get statistics
        stats = session.stats()
        print(f"Processed {stats.stats.total_events} events")

if __name__ == "__main__":
    run_etl()
```

## Error Handling

```python
from ptx_trace import Session, Severity, CLIError

try:
    with Session("risky-operation") as session:
        with session.operation("critical_step") as op:
            try:
                result = risky_function()
                op.output(result=result)
                op.success()
            except ValueError as e:
                op.error(str(e), Severity.MEDIUM)
                op.recovery("use_default", success=True)
                result = default_value
            except Exception as e:
                op.error(str(e), Severity.CRITICAL)
                raise

except CLIError as e:
    print(f"CLI Error: {e}")
    print(f"Exit code: {e.exit_code}")
```

## Querying Events

```python
from ptx_trace import Session

# Get session events
session = Session.load("ses_2026-01-27_import")

# Filter events
errors = session.events(level="error", limit=50)
transforms = session.events(operation="transform", status="success")

# SQL-like query
slow_ops = session.query("duration_us > 10000 AND operation = 'transform'")
```

## Configuration

The SDK uses the `portunix` binary for all operations. Configure the path if needed:

```python
# Custom binary path
session = Session("my-session", binary_path="/opt/portunix/bin/portunix")

# Or via environment variable
import os
os.environ["PORTUNIX_PATH"] = "/opt/portunix/bin/portunix"
```

## Thread Safety

Sessions are NOT thread-safe. Use separate sessions for concurrent operations:

```python
from concurrent.futures import ThreadPoolExecutor

def process_batch(batch_id, records):
    with Session(f"batch-{batch_id}") as session:
        for record in records:
            with session.operation("process") as op:
                # ... process record
                pass

with ThreadPoolExecutor(max_workers=4) as executor:
    for i, batch in enumerate(batches):
        executor.submit(process_batch, i, batch)
```

## License

MIT License - see LICENSE file for details.
