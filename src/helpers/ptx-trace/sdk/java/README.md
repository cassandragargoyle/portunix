# PTX-TRACE Java SDK

Java SDK for the Portunix Trace (PTX-TRACE) system - a universal tracing system for software development, optimized for debugging, AI analysis, and monitoring of data pipelines and workflows.

## Requirements

- Java 21+
- `portunix` binary in PATH (or specify path)
- Maven 3.6+ (for building)

## Installation

### Maven

```xml
<dependency>
    <groupId>ai.portunix</groupId>
    <artifactId>ptx-trace</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Building from Source

```bash
cd sdk/java
mvn clean package
```

## Quick Start

### Basic Usage with Try-with-Resources

```java
import ai.portunix.trace.*;
import ai.portunix.trace.models.*;

// Session automatically starts on creation, ends on close
try (Session session = Trace.newSession("import-customers")
        .withPIIMasking(true)
        .build()) {

    // Operations are traced with timing
    try (Operation op = session.start("normalize_phone")) {
        op.input("phone", "+420 777 123 456");
        String result = normalizePhone("+420 777 123 456");
        op.output("phone", result);
        op.success();
    }
}
```

### Fluent API with Lambda

```java
import ai.portunix.trace.*;
import ai.portunix.trace.models.*;

try (Session session = Trace.createSession("validation-pipeline")) {

    session.trace("validate_email")
        .input("email", email)
        .tag("validation")
        .execute(ctx -> {
            boolean valid = validateEmail(email);
            ctx.output("valid", valid);
        });
}
```

### With Source Information

```java
import ai.portunix.trace.*;
import ai.portunix.trace.models.*;

try (Session session = Trace.newSession("etl-pipeline")
        .withSource("customers.csv")
        .withTags("production", "daily")
        .build()) {

    for (int row = 0; row < records.size(); row++) {
        Record record = records.get(row);

        session.trace("transform_record")
            .input("record", record.toString())
            .source(SourceInfo.csv("customers.csv", row + 1))
            .tag("transform", "customer")
            .execute(ctx -> {
                Record transformed = transform(record);
                ctx.output("result", transformed.toString());
            });
    }
}
```

## API Reference

### Trace (Factory)

The `Trace` class provides factory methods for creating sessions.

```java
// Configure defaults
Trace.setDefaultBinaryPath("/opt/portunix/bin/portunix");
Trace.setDefaultTimeout(60);

// Create session with builder
Session session = Trace.newSession("my-session")
    .withSource("input.csv")
    .withDestination("postgres://localhost/db")
    .withTags("production", "batch")
    .withSampling(0.5)       // Sample 50% of events
    .withPIIMasking(true)    // Mask PII data
    .withBinaryPath("/usr/bin/portunix")
    .withTimeout(30)
    .build();

// Quick session creation
Session session = Trace.createSession("my-session");

// Load existing session
Session session = Trace.loadSession("ses_2026-01-27_import");

// List sessions
List<Map<String, Object>> sessions = Trace.listSessions(10, "active");
```

### Session

The `Session` class manages the tracing session lifecycle.

```java
try (Session session = Trace.createSession("my-session")) {
    // Session is automatically started

    // Create operations
    session.start("operation_name");       // Direct operation
    session.trace("operation_name");       // Builder pattern

    // End with specific status
    session.fail();     // Mark as failed
    session.cancel();   // Mark as cancelled
    // close() marks as completed (default)

    // Query data
    Map<String, Object> stats = session.stats();
    List<Map<String, Object>> events = session.events(100);
    List<Map<String, Object>> errors = session.events(null, null, "error", null, 50);
}
```

### Operation

Operations represent traced units of work.

```java
try (Operation op = session.start("process_record")) {
    // Set input data
    op.input("field", value);
    op.inputs(Map.of("field1", value1, "field2", value2));

    // Set output data
    op.output("result", value);
    op.outputs(Map.of("result1", value1, "result2", value2));

    // Add metadata
    op.tag("important", "customer");
    op.context("batch_id", 123);
    op.source(SourceInfo.csv("data.csv", 42));

    // Record outcome
    op.success();
    // or
    op.error("Validation failed", Severity.HIGH);
    op.error(exception);
}
```

### OperationBuilder (Fluent API)

```java
session.trace("transform")
    .withType("etl")                          // Set operation type
    .input("data", inputData)                 // Set input
    .source(SourceInfo.database("pg://...", "users")) // Set source
    .tag("critical")                          // Add tags
    .withRule("R001", "1.0")                  // Set rule info
    .context("env", "production")             // Add context
    .execute(ctx -> {                         // Execute with tracing
        Object result = transform(inputData);
        ctx.output("result", result);
    });
```

### Source Types

```java
import ai.portunix.trace.models.SourceInfo;

// CSV file source
SourceInfo source = SourceInfo.csv("data.csv", 42, "email");

// Database source
SourceInfo source = SourceInfo.database("postgres://localhost/db", "users", 100);

// Generic file source
SourceInfo source = SourceInfo.file("config.json");

// API source
SourceInfo source = SourceInfo.api("https://api.example.com/users");
```

### Severity Levels

```java
import ai.portunix.trace.models.Severity;

op.error("Minor issue", Severity.LOW);
op.error("Validation error", Severity.MEDIUM);
op.error("Data corruption", Severity.HIGH);
op.error("System failure", Severity.CRITICAL);
```

## Complete ETL Example

```java
import ai.portunix.trace.*;
import ai.portunix.trace.models.*;

import java.io.*;
import java.util.*;

public class EtlPipeline {
    public static void main(String[] args) throws Exception {
        try (Session session = Trace.newSession("customer-import")
                .withPIIMasking(true)
                .withTags("etl", "daily")
                .build()) {

            List<Map<String, String>> records;

            // Phase 1: Extract
            try (Operation op = session.start("extract")) {
                op.source(SourceInfo.csv("customers.csv"));

                records = readCsv("customers.csv");

                op.output("record_count", records.size());
                op.success();
            }

            // Phase 2: Transform each record
            List<Map<String, String>> transformed = new ArrayList<>();

            for (int i = 0; i < records.size(); i++) {
                Map<String, String> record = records.get(i);
                final int row = i + 1;

                try (Operation op = session.start("transform")) {
                    op.source(SourceInfo.csv("customers.csv", row));
                    record.forEach(op::input);

                    try {
                        // Normalize phone
                        String phone = normalizePhone(record.get("phone"));
                        record.put("phone", phone);

                        // Validate email
                        if (!validateEmail(record.get("email"))) {
                            op.error("Invalid email", Severity.MEDIUM);
                            continue;
                        }

                        record.forEach(op::output);
                        op.success();
                        transformed.add(record);

                    } catch (Exception e) {
                        op.error(e);
                    }
                }
            }

            // Phase 3: Load
            try (Operation op = session.start("load")) {
                op.context("destination", "postgres://localhost/customers");

                int inserted = loadToDatabase(transformed);

                op.output("inserted", inserted);
                op.success();
            }

            // Get statistics
            Map<String, Object> stats = session.stats();
            System.out.println("Session completed: " + session.getId());
        }
    }

    private static List<Map<String, String>> readCsv(String file) {
        // Implementation
        return new ArrayList<>();
    }

    private static String normalizePhone(String phone) {
        // Implementation
        return phone != null ? phone.replaceAll("\\s+", "") : "";
    }

    private static boolean validateEmail(String email) {
        return email != null && email.contains("@");
    }

    private static int loadToDatabase(List<Map<String, String>> records) {
        // Implementation
        return records.size();
    }
}
```

## Error Handling

```java
import ai.portunix.trace.*;
import ai.portunix.trace.cli.CliException;
import ai.portunix.trace.models.*;

try {
    try (Session session = Trace.createSession("risky-operation")) {
        try (Operation op = session.start("critical_step")) {
            try {
                Object result = riskyFunction();
                op.output("result", result);
                op.success();
            } catch (IllegalArgumentException e) {
                op.error(e.getMessage(), Severity.MEDIUM);
                // Handle gracefully
            } catch (Exception e) {
                op.error(e);
                throw e;
            }
        }
    }
} catch (CliException e) {
    System.err.println("CLI Error: " + e.getMessage());
    System.err.println("Exit code: " + e.getExitCode());
}
```

## Configuration

Configure the SDK through the `Trace` class or session builder:

```java
// Global configuration
Trace.setDefaultBinaryPath("/opt/portunix/bin/portunix");
Trace.setDefaultTimeout(60); // seconds

// Per-session configuration
Session session = Trace.newSession("my-session")
    .withBinaryPath("/custom/path/portunix")
    .withTimeout(120)
    .build();
```

## Thread Safety

Sessions are NOT thread-safe. Use separate sessions for concurrent operations:

```java
import java.util.concurrent.*;

ExecutorService executor = Executors.newFixedThreadPool(4);

for (int i = 0; i < batches.size(); i++) {
    final int batchId = i;
    final List<Record> batch = batches.get(i);

    executor.submit(() -> {
        try (Session session = Trace.createSession("batch-" + batchId)) {
            for (Record record : batch) {
                try (Operation op = session.start("process")) {
                    // Process record
                    op.success();
                }
            }
        } catch (CliException e) {
            e.printStackTrace();
        }
    });
}

executor.shutdown();
executor.awaitTermination(1, TimeUnit.HOURS);
```

## Annotation Support (with AspectJ)

The SDK includes a `@Traced` annotation for AOP-based tracing:

```java
import ai.portunix.trace.Traced;

@Traced(operation = "validate_email", tags = {"validation"})
public boolean validateEmail(String email) {
    return email.contains("@");
}
```

Note: This requires AspectJ or similar AOP framework to be configured separately.

## License

MIT License - see LICENSE file for details.
