#!/usr/bin/env python3
"""
PTX-TRACE Python SDK Example - ETL Pipeline

This example demonstrates a complete ETL pipeline with tracing.
It shows how to:
- Create and manage sessions
- Trace extraction, transformation, and loading phases
- Use source information for data lineage
- Handle errors with proper severity
- Use decorators for function tracing
"""

import csv
import io
from typing import Dict, List, Optional

from ptx_trace import Session, CsvSource, DbSource, Severity


# Sample data
SAMPLE_CSV = """id,name,email,phone
1,John Doe,john@example.com,+420 777 123 456
2,Jane Smith,jane@test,+1 555 234 5678
3,Bob Wilson,bob@company.org,invalid
4,Alice Brown,alice@example.com,+44 20 1234 5678
"""


def normalize_phone(phone: str) -> Optional[str]:
    """Normalize phone number by removing spaces and validating format."""
    if not phone:
        return None

    # Remove spaces
    normalized = phone.replace(" ", "")

    # Basic validation - must start with + and have digits
    if not normalized.startswith("+"):
        return None
    if not normalized[1:].replace("-", "").isdigit():
        return None

    return normalized


def validate_email(email: str) -> bool:
    """Validate email format."""
    return "@" in email and "." in email.split("@")[1]


def transform_record(record: Dict[str, str]) -> Dict[str, str]:
    """Transform a single record."""
    return {
        "id": record["id"],
        "name": record["name"].strip().title(),
        "email": record["email"].lower().strip(),
        "phone": normalize_phone(record["phone"]),
    }


def load_to_database(records: List[Dict[str, str]]) -> int:
    """Simulate loading records to database."""
    # In real implementation, this would insert into database
    print(f"Loading {len(records)} records to database...")
    return len(records)


def run_etl_pipeline():
    """Run the complete ETL pipeline with tracing."""

    # Create session with PII masking enabled
    with Session("customer-import", pii_masking=True, tags=["etl", "daily"]) as session:

        # ============================================
        # PHASE 1: EXTRACT
        # ============================================

        with session.operation("extract") as op:
            op.source(CsvSource("customers.csv"))
            op.context("format", "csv")

            # Read CSV data
            reader = csv.DictReader(io.StringIO(SAMPLE_CSV))
            records = list(reader)

            op.output("record_count", len(records))
            op.output("columns", list(records[0].keys()) if records else [])
            op.success()

        print(f"Extracted {len(records)} records")

        # ============================================
        # PHASE 2: TRANSFORM
        # ============================================

        transformed_records = []
        error_count = 0

        for row_num, record in enumerate(records, start=1):
            with session.operation("transform") as op:
                op.source(CsvSource("customers.csv", row=row_num))
                op.input("id", record["id"])
                op.input("name", record["name"])
                op.input("email", record["email"])
                op.input("phone", record["phone"])

                # Validate email
                if not validate_email(record["email"]):
                    op.error(f"Invalid email format: {record['email']}", Severity.MEDIUM)
                    op.tag("validation_error")
                    error_count += 1
                    continue

                # Transform record
                try:
                    transformed = transform_record(record)

                    # Check phone normalization
                    if transformed["phone"] is None:
                        op.error("Invalid phone number format", Severity.LOW)
                        op.recovery("set_null", success=True)

                    op.output("id", transformed["id"])
                    op.output("name", transformed["name"])
                    op.output("email", transformed["email"])
                    op.output("phone", transformed["phone"] or "NULL")
                    op.success()

                    transformed_records.append(transformed)

                except Exception as e:
                    op.error(str(e), Severity.HIGH)
                    error_count += 1

        print(f"Transformed {len(transformed_records)} records, {error_count} errors")

        # ============================================
        # PHASE 3: LOAD
        # ============================================

        with session.operation("load") as op:
            op.context("destination", "postgres://localhost/customers")
            op.context("table", "customers")
            op.input("record_count", len(transformed_records))

            try:
                inserted = load_to_database(transformed_records)

                op.output("inserted", inserted)
                op.output("skipped", len(records) - inserted)
                op.success()

            except Exception as e:
                op.error(str(e), Severity.CRITICAL)
                raise

        # ============================================
        # SUMMARY
        # ============================================

        # Get and display session statistics
        stats = session.stats()

        print("\n" + "=" * 50)
        print("ETL Pipeline Summary")
        print("=" * 50)
        print(f"Session ID: {session.id}")
        print(f"Total Events: {stats.stats.total_events}")
        print(f"By Status: {stats.stats.by_status}")
        print(f"By Level: {stats.stats.by_level}")

        if stats.stats.by_operation:
            print("\nOperation Statistics:")
            for op_name, op_stats in stats.stats.by_operation.items():
                print(f"  {op_name}:")
                print(f"    Count: {op_stats.count}")
                print(f"    Avg Duration: {op_stats.avg_duration:.2f} us")
                if op_stats.total_errors > 0:
                    print(f"    Errors: {op_stats.total_errors}")


def run_with_decorator():
    """Example using the @traced decorator."""

    session = Session.create("validation-pipeline", tags=["decorator-example"])

    @session.traced("validate_email", tags=["validation"])
    def validate_email_traced(email: str) -> dict:
        """Validate an email address."""
        valid = "@" in email and "." in email.split("@")[-1]
        return {"email": email, "valid": valid}

    @session.traced("validate_phone", tags=["validation"])
    def validate_phone_traced(phone: str) -> dict:
        """Validate a phone number."""
        normalized = normalize_phone(phone)
        return {"phone": phone, "normalized": normalized, "valid": normalized is not None}

    try:
        # These function calls are automatically traced
        result1 = validate_email_traced("user@example.com")
        result2 = validate_email_traced("invalid-email")
        result3 = validate_phone_traced("+420 777 123 456")
        result4 = validate_phone_traced("not a phone")

        print("\nValidation Results:")
        print(f"  Email 1: {result1}")
        print(f"  Email 2: {result2}")
        print(f"  Phone 1: {result3}")
        print(f"  Phone 2: {result4}")

    finally:
        session.close()


def run_with_fluent_api():
    """Example using the fluent API."""

    with Session("fluent-example") as session:
        # Fluent API with execute
        session.trace("process_data") \
            .with_type("etl") \
            .input("source", "api") \
            .input("count", 100) \
            .source(DbSource("postgres://localhost/db", "users")) \
            .tag("production", "batch") \
            .with_rule("R001", "1.0") \
            .execute(lambda ctx: (
                ctx.output("processed", 100),
                ctx.output("status", "complete"),
                ctx.tag("success")
            ))

        print("Fluent API example completed")


if __name__ == "__main__":
    print("=" * 60)
    print("PTX-TRACE Python SDK - ETL Pipeline Example")
    print("=" * 60)

    print("\n1. Running ETL Pipeline...\n")
    run_etl_pipeline()

    print("\n2. Running Decorator Example...\n")
    run_with_decorator()

    print("\n3. Running Fluent API Example...\n")
    run_with_fluent_api()

    print("\n" + "=" * 60)
    print("All examples completed!")
    print("=" * 60)
