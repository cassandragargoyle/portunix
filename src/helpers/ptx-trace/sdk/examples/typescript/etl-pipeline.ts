/**
 * PTX-TRACE TypeScript SDK Example - ETL Pipeline
 *
 * This example demonstrates a complete ETL pipeline with tracing.
 * It shows how to:
 * - Create and manage sessions
 * - Trace extraction, transformation, and loading phases
 * - Use source information for data lineage
 * - Handle errors with proper severity
 * - Use fluent API for tracing
 */

import {
  Session,
  csvSource,
  dbSource,
  Severity,
} from '../src'; // Use 'ptx-trace' when installed from npm

// Sample data
const SAMPLE_DATA = [
  { id: '1', name: 'John Doe', email: 'john@example.com', phone: '+420 777 123 456' },
  { id: '2', name: 'Jane Smith', email: 'jane@test', phone: '+1 555 234 5678' },
  { id: '3', name: 'Bob Wilson', email: 'bob@company.org', phone: 'invalid' },
  { id: '4', name: 'Alice Brown', email: 'alice@example.com', phone: '+44 20 1234 5678' },
];

interface Record {
  id: string;
  name: string;
  email: string;
  phone: string | null;
}

/**
 * Normalize phone number by removing spaces and validating format.
 */
function normalizePhone(phone: string): string | null {
  if (!phone) {
    return null;
  }

  // Remove spaces
  const normalized = phone.replace(/\s+/g, '');

  // Basic validation - must start with + and have digits
  if (!normalized.startsWith('+')) {
    return null;
  }

  const digits = normalized.substring(1).replace(/-/g, '');
  if (!/^\d+$/.test(digits)) {
    return null;
  }

  return normalized;
}

/**
 * Validate email format.
 */
function validateEmail(email: string): boolean {
  if (!email) {
    return false;
  }

  const atIndex = email.indexOf('@');
  if (atIndex < 1) {
    return false;
  }

  const domain = email.substring(atIndex + 1);
  return domain.includes('.');
}

/**
 * Transform a single record.
 */
function transformRecord(record: typeof SAMPLE_DATA[0]): Record {
  // Normalize name (title case)
  const name = record.name
    .trim()
    .split(/\s+/)
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ');

  return {
    id: record.id,
    name,
    email: record.email.toLowerCase().trim(),
    phone: normalizePhone(record.phone),
  };
}

/**
 * Simulate loading records to database.
 */
async function loadToDatabase(records: Record[]): Promise<number> {
  console.log(`Loading ${records.length} records to database...`);
  // Simulate async operation
  await new Promise(resolve => setTimeout(resolve, 100));
  return records.length;
}

/**
 * Run the complete ETL pipeline with tracing.
 */
async function runEtlPipeline(): Promise<void> {
  // Create session with PII masking enabled
  const session = await Session.create('customer-import', {
    piiMasking: true,
    tags: ['etl', 'daily'],
  });

  try {
    // ============================================
    // PHASE 1: EXTRACT
    // ============================================

    let records: typeof SAMPLE_DATA;

    await session.withOperation('extract', async (op) => {
      op.source(csvSource('customers.csv'));
      op.context('format', 'csv');

      // Read sample data
      records = [...SAMPLE_DATA];

      op.output('record_count', records.length);
      op.output('columns', Object.keys(records[0]).join(','));
    });

    console.log(`Extracted ${records!.length} records`);

    // ============================================
    // PHASE 2: TRANSFORM
    // ============================================

    const transformedRecords: Record[] = [];
    let errorCount = 0;

    for (let rowNum = 0; rowNum < records!.length; rowNum++) {
      const record = records![rowNum];

      await session.withOperation('transform', async (op) => {
        op.source(csvSource('customers.csv', rowNum + 1));
        op.input('id', record.id);
        op.input('name', record.name);
        op.input('email', record.email);
        op.input('phone', record.phone);

        // Validate email
        if (!validateEmail(record.email)) {
          op.error(`Invalid email format: ${record.email}`, Severity.MEDIUM);
          op.tag('validation_error');
          errorCount++;
          return;
        }

        // Transform record
        try {
          const transformed = transformRecord(record);

          // Check phone normalization
          if (transformed.phone === null) {
            op.error('Invalid phone number format', Severity.LOW);
            // Recovery: set to null
          }

          op.output('id', transformed.id);
          op.output('name', transformed.name);
          op.output('email', transformed.email);
          op.output('phone', transformed.phone ?? 'NULL');

          transformedRecords.push(transformed);

        } catch (e) {
          op.error(String(e), Severity.HIGH);
          errorCount++;
        }
      });
    }

    console.log(`Transformed ${transformedRecords.length} records, ${errorCount} errors`);

    // ============================================
    // PHASE 3: LOAD
    // ============================================

    await session.withOperation('load', async (op) => {
      op.context('destination', 'postgres://localhost/customers');
      op.context('table', 'customers');
      op.input('record_count', transformedRecords.length);

      try {
        const inserted = await loadToDatabase(transformedRecords);

        op.output('inserted', inserted);
        op.output('skipped', records!.length - inserted);

      } catch (e) {
        op.error(String(e), Severity.CRITICAL);
        throw e;
      }
    });

    // ============================================
    // SUMMARY
    // ============================================

    const stats = await session.stats();

    console.log('\n' + '='.repeat(50));
    console.log('ETL Pipeline Summary');
    console.log('='.repeat(50));
    console.log(`Session ID: ${session.id}`);
    console.log(`Total Events: ${stats.stats?.totalEvents}`);
    console.log(`By Status: ${JSON.stringify(stats.stats?.byStatus)}`);
    console.log(`By Level: ${JSON.stringify(stats.stats?.byLevel)}`);

  } finally {
    await session.close();
  }
}

/**
 * Example using the fluent API.
 */
async function runFluentApiExample(): Promise<void> {
  const session = await Session.create('fluent-example');

  try {
    // Fluent API with execute
    await session.trace('process_data')
      .withType('etl')
      .input('source', 'api')
      .input('count', 100)
      .source(dbSource('postgres://localhost/db', 'users'))
      .tag('production', 'batch')
      .withRule('R001', '1.0')
      .execute(async (ctx) => {
        ctx.output('processed', 100);
        ctx.output('status', 'complete');
        ctx.tag('success');
      });

    console.log('Fluent API example completed');

  } finally {
    await session.close();
  }
}

/**
 * Example with manual operation control.
 */
async function runManualOperationExample(): Promise<void> {
  const session = await Session.create('manual-example');

  try {
    // Build operation manually
    const op = session.trace('process_batch')
      .input('batch_id', 123)
      .tag('manual')
      .build();

    try {
      // Do some work
      op.input('items', 50);

      // Simulate processing
      await new Promise(resolve => setTimeout(resolve, 50));

      op.output('processed', 50);
      op.output('failed', 0);
      op.success();

    } catch (e) {
      op.error(String(e), Severity.HIGH);
      throw e;
    } finally {
      await op.end();
    }

    console.log('Manual operation example completed');

  } finally {
    await session.close();
  }
}

// Main execution
async function main(): Promise<void> {
  console.log('='.repeat(60));
  console.log('PTX-TRACE TypeScript SDK - ETL Pipeline Example');
  console.log('='.repeat(60));

  try {
    console.log('\n1. Running ETL Pipeline...\n');
    await runEtlPipeline();

    console.log('\n2. Running Fluent API Example...\n');
    await runFluentApiExample();

    console.log('\n3. Running Manual Operation Example...\n');
    await runManualOperationExample();

    console.log('\n' + '='.repeat(60));
    console.log('All examples completed!');
    console.log('='.repeat(60));

  } catch (e) {
    console.error('Error:', e);
    process.exit(1);
  }
}

main();
