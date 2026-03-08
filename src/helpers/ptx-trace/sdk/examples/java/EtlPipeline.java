package ai.portunix.trace.examples;

import ai.portunix.trace.*;
import ai.portunix.trace.cli.CliException;
import ai.portunix.trace.models.*;

import java.util.*;
import java.util.regex.Pattern;

/**
 * PTX-TRACE Java SDK Example - ETL Pipeline
 *
 * This example demonstrates a complete ETL pipeline with tracing.
 * It shows how to:
 * - Create and manage sessions
 * - Trace extraction, transformation, and loading phases
 * - Use source information for data lineage
 * - Handle errors with proper severity
 * - Use fluent API for tracing
 */
public class EtlPipeline {

    // Sample data
    private static final String[][] SAMPLE_DATA = {
        {"1", "John Doe", "john@example.com", "+420 777 123 456"},
        {"2", "Jane Smith", "jane@test", "+1 555 234 5678"},
        {"3", "Bob Wilson", "bob@company.org", "invalid"},
        {"4", "Alice Brown", "alice@example.com", "+44 20 1234 5678"}
    };

    private static final String[] COLUMNS = {"id", "name", "email", "phone"};

    public static void main(String[] args) {
        System.out.println("=".repeat(60));
        System.out.println("PTX-TRACE Java SDK - ETL Pipeline Example");
        System.out.println("=".repeat(60));

        try {
            System.out.println("\n1. Running ETL Pipeline...\n");
            runEtlPipeline();

            System.out.println("\n2. Running Fluent API Example...\n");
            runFluentApiExample();

            System.out.println("\n" + "=".repeat(60));
            System.out.println("All examples completed!");
            System.out.println("=".repeat(60));

        } catch (CliException e) {
            System.err.println("Error: " + e.getMessage());
            System.err.println("Exit code: " + e.getExitCode());
            System.exit(1);
        }
    }

    /**
     * Run the complete ETL pipeline with tracing.
     */
    private static void runEtlPipeline() throws CliException {
        // Create session with PII masking enabled
        try (Session session = Trace.newSession("customer-import")
                .withPIIMasking(true)
                .withTags("etl", "daily")
                .build()) {

            // ============================================
            // PHASE 1: EXTRACT
            // ============================================

            List<Map<String, String>> records;

            try (Operation op = session.start("extract")) {
                op.source(SourceInfo.csv("customers.csv"));
                op.context("format", "csv");

                // Read sample data
                records = readSampleData();

                op.output("record_count", records.size());
                op.output("columns", String.join(",", COLUMNS));
                op.success();
            }

            System.out.println("Extracted " + records.size() + " records");

            // ============================================
            // PHASE 2: TRANSFORM
            // ============================================

            List<Map<String, String>> transformedRecords = new ArrayList<>();
            int errorCount = 0;

            for (int rowNum = 0; rowNum < records.size(); rowNum++) {
                Map<String, String> record = records.get(rowNum);

                try (Operation op = session.start("transform")) {
                    op.source(SourceInfo.csv("customers.csv", rowNum + 1));
                    op.input("id", record.get("id"));
                    op.input("name", record.get("name"));
                    op.input("email", record.get("email"));
                    op.input("phone", record.get("phone"));

                    // Validate email
                    if (!validateEmail(record.get("email"))) {
                        op.error("Invalid email format: " + record.get("email"), Severity.MEDIUM);
                        op.tag("validation_error");
                        errorCount++;
                        continue;
                    }

                    // Transform record
                    try {
                        Map<String, String> transformed = transformRecord(record);

                        // Check phone normalization
                        if (transformed.get("phone") == null) {
                            op.error("Invalid phone number format", Severity.LOW);
                            // Recovery: set to null
                        }

                        op.output("id", transformed.get("id"));
                        op.output("name", transformed.get("name"));
                        op.output("email", transformed.get("email"));
                        op.output("phone", transformed.get("phone") != null ? transformed.get("phone") : "NULL");
                        op.success();

                        transformedRecords.add(transformed);

                    } catch (Exception e) {
                        op.error(e);
                        errorCount++;
                    }
                }
            }

            System.out.println("Transformed " + transformedRecords.size() + " records, " + errorCount + " errors");

            // ============================================
            // PHASE 3: LOAD
            // ============================================

            try (Operation op = session.start("load")) {
                op.context("destination", "postgres://localhost/customers");
                op.context("table", "customers");
                op.input("record_count", transformedRecords.size());

                try {
                    int inserted = loadToDatabase(transformedRecords);

                    op.output("inserted", inserted);
                    op.output("skipped", records.size() - inserted);
                    op.success();

                } catch (Exception e) {
                    op.error(e.getMessage(), Severity.CRITICAL);
                    throw e;
                }
            }

            // ============================================
            // SUMMARY
            // ============================================

            Map<String, Object> stats = session.stats();

            System.out.println("\n" + "=".repeat(50));
            System.out.println("ETL Pipeline Summary");
            System.out.println("=".repeat(50));
            System.out.println("Session ID: " + session.getId());
            System.out.println("Stats: " + stats);
        }
    }

    /**
     * Example using the fluent API.
     */
    private static void runFluentApiExample() throws CliException {
        try (Session session = Trace.createSession("fluent-example")) {

            // Fluent API with execute
            session.trace("process_data")
                .withType("etl")
                .input("source", "api")
                .input("count", 100)
                .source(SourceInfo.database("postgres://localhost/db", "users"))
                .tag("production", "batch")
                .withRule("R001", "1.0")
                .execute(ctx -> {
                    ctx.output("processed", 100);
                    ctx.output("status", "complete");
                    ctx.tag("success");
                });

            System.out.println("Fluent API example completed");
        }
    }

    /**
     * Read sample data into list of maps.
     */
    private static List<Map<String, String>> readSampleData() {
        List<Map<String, String>> records = new ArrayList<>();

        for (String[] row : SAMPLE_DATA) {
            Map<String, String> record = new LinkedHashMap<>();
            for (int i = 0; i < COLUMNS.length; i++) {
                record.put(COLUMNS[i], row[i]);
            }
            records.add(record);
        }

        return records;
    }

    /**
     * Normalize phone number by removing spaces and validating format.
     */
    private static String normalizePhone(String phone) {
        if (phone == null || phone.isEmpty()) {
            return null;
        }

        // Remove spaces
        String normalized = phone.replaceAll("\\s+", "");

        // Basic validation - must start with + and have digits
        if (!normalized.startsWith("+")) {
            return null;
        }

        String digits = normalized.substring(1).replaceAll("-", "");
        if (!digits.matches("\\d+")) {
            return null;
        }

        return normalized;
    }

    /**
     * Validate email format.
     */
    private static boolean validateEmail(String email) {
        if (email == null || email.isEmpty()) {
            return false;
        }

        int atIndex = email.indexOf("@");
        if (atIndex < 1) {
            return false;
        }

        String domain = email.substring(atIndex + 1);
        return domain.contains(".");
    }

    /**
     * Transform a single record.
     */
    private static Map<String, String> transformRecord(Map<String, String> record) {
        Map<String, String> transformed = new LinkedHashMap<>();

        transformed.put("id", record.get("id"));

        // Normalize name
        String name = record.get("name");
        if (name != null) {
            name = name.trim();
            // Title case
            String[] words = name.split("\\s+");
            StringBuilder sb = new StringBuilder();
            for (String word : words) {
                if (sb.length() > 0) sb.append(" ");
                sb.append(Character.toUpperCase(word.charAt(0)));
                if (word.length() > 1) {
                    sb.append(word.substring(1).toLowerCase());
                }
            }
            name = sb.toString();
        }
        transformed.put("name", name);

        // Normalize email
        String email = record.get("email");
        if (email != null) {
            email = email.toLowerCase().trim();
        }
        transformed.put("email", email);

        // Normalize phone
        transformed.put("phone", normalizePhone(record.get("phone")));

        return transformed;
    }

    /**
     * Simulate loading records to database.
     */
    private static int loadToDatabase(List<Map<String, String>> records) {
        System.out.println("Loading " + records.size() + " records to database...");
        // In real implementation, this would insert into database
        return records.size();
    }
}
