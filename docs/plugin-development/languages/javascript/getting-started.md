# JavaScript/Node.js Plugin Development Guide

JavaScript (Node.js) offers fast iteration, a massive ecosystem, and excellent async I/O for Portunix plugins. This guide walks you through
creating Node.js plugins using `@grpc/grpc-js`.

## Prerequisites

- Node.js 18 LTS or later (16 minimum)
- npm 9+ or yarn 1.22+
- Protocol Buffers compiler (`protoc`) — optional, only for AOT codegen
- Portunix development environment

## Quick Start

Create a new JavaScript plugin using the Portunix CLI:

```bash
portunix plugin create my-js-plugin --language=javascript
cd my-js-plugin
npm install
```

This creates a complete Node.js project structure:

```text
my-js-plugin/
├── plugin.json                # Plugin manifest
├── package.json               # npm manifest
├── package-lock.json          # Dependency lockfile
├── src/
│   ├── main.js                # Plugin entry point
│   ├── config.js              # Configuration loading
│   ├── handlers/
│   │   └── plugin.js          # gRPC handlers
│   └── services/
│       └── plugin.js          # Business logic
├── proto/
│   └── plugin.proto           # gRPC service definition
├── test/
│   └── plugin.test.js
└── README.md
```

## Plugin Structure

### Package Manifest (package.json)

```json
{
  "name": "my-js-plugin",
  "version": "1.0.0",
  "description": "Example JavaScript plugin for Portunix",
  "type": "module",
  "main": "src/main.js",
  "scripts": {
    "start": "node src/main.js",
    "test": "node --test test/",
    "test:coverage": "node --test --experimental-test-coverage test/",
    "proto:generate": "echo 'Using @grpc/proto-loader for runtime loading (no codegen needed)'",
    "lint": "eslint src/ test/"
  },
  "engines": {
    "node": ">=18"
  },
  "dependencies": {
    "@grpc/grpc-js": "^1.10.6",
    "@grpc/proto-loader": "^0.7.10",
    "yaml": "^2.4.2"
  },
  "devDependencies": {
    "eslint": "^9.0.0"
  }
}
```

### Main Entry Point (src/main.js)

```javascript
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';
import { parseArgs } from 'node:util';
import grpc from '@grpc/grpc-js';
import protoLoader from '@grpc/proto-loader';

import { loadConfig } from './config.js';
import { createPluginHandler } from './handlers/plugin.js';

const __dirname = dirname(fileURLToPath(import.meta.url));
const PROTO_PATH = resolve(__dirname, '..', 'proto', 'plugin.proto');

async function main() {
    const { values } = parseArgs({
        options: {
            port: { type: 'string', default: '50051' },
            config: { type: 'string', default: 'config.yaml' },
        },
    });

    const config = await loadConfig(values.config);

    const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
        keepCase: true,
        longs: String,
        enums: String,
        defaults: true,
        oneofs: true,
    });
    const proto = grpc.loadPackageDefinition(packageDefinition).js_plugin_template;

    const server = new grpc.Server();
    server.addService(proto.PluginService.service, createPluginHandler(config));

    const bindAddress = `0.0.0.0:${values.port}`;
    server.bindAsync(bindAddress, grpc.ServerCredentials.createInsecure(), (err, port) => {
        if (err) {
            console.error('Failed to bind:', err);
            process.exit(1);
        }
        console.log(`Plugin server listening on port ${port}`);
    });

    const shutdown = () => {
        console.log('Shutting down gracefully...');
        server.tryShutdown(() => process.exit(0));
    };
    process.on('SIGINT', shutdown);
    process.on('SIGTERM', shutdown);
}

main().catch((err) => {
    console.error(err);
    process.exit(1);
});
```

### Configuration (src/config.js)

```javascript
import { readFile } from 'node:fs/promises';
import YAML from 'yaml';

const DEFAULTS = Object.freeze({
    plugin: {
        name: 'my-js-plugin',
        version: '1.0.0',
        description: '',
        log_level: 'info',
    },
    server: {
        port: 50051,
        health_port: 50052,
    },
});

export async function loadConfig(path) {
    try {
        const raw = await readFile(path, 'utf8');
        const parsed = YAML.parse(raw) || {};
        return {
            plugin: { ...DEFAULTS.plugin, ...(parsed.plugin || {}) },
            server: { ...DEFAULTS.server, ...(parsed.server || {}) },
        };
    } catch (err) {
        if (err.code === 'ENOENT') {
            return structuredClone(DEFAULTS);
        }
        throw err;
    }
}
```

### gRPC Handlers (src/handlers/plugin.js)

```javascript
import { PluginBusinessService } from '../services/plugin.js';

export function createPluginHandler(config) {
    const service = new PluginBusinessService(config);

    return {
        GetInfo(call, callback) {
            callback(null, {
                info: {
                    name: config.plugin.name,
                    version: config.plugin.version,
                    description: config.plugin.description,
                    capabilities: ['example-capability'],
                },
            });
        },

        HealthCheck(call, callback) {
            callback(null, { status: 'SERVING' });
        },

        Execute(call, callback) {
            try {
                const result = service.execute(call.request.command, call.request.args || []);
                callback(null, { result, status: 'SUCCESS' });
            } catch (err) {
                callback(null, { status: 'ERROR', error_message: err.message });
            }
        },

        Shutdown(call, callback) {
            service.cleanup();
            callback(null, { status: 'SUCCESS' });
        },

        ListTools(call, callback) {
            callback(null, { tools: [] });
        },

        CallTool(call, callback) {
            callback(null, { status: 'NOT_FOUND', error_message: 'No tools registered' });
        },
    };
}
```

### Business Logic (src/services/plugin.js)

```javascript
export class PluginBusinessService {
    constructor(config) {
        this.config = config;
    }

    execute(command, args) {
        switch (command) {
            case 'hello':
                return this.#handleHello(args);
            case 'process':
                return this.#handleProcess(args);
            default:
                throw new Error(`unknown command: ${command}`);
        }
    }

    #handleHello(args) {
        const name = args[0] ?? 'World';
        return `Hello, ${name}!`;
    }

    #handleProcess(args) {
        if (args.length === 0) {
            throw new Error('process command requires at least one argument');
        }
        return `Processed: ${args.join(', ')}`;
    }

    cleanup() {
        console.log('Performing cleanup operations...');
    }
}
```

## Building and Testing

### Build / Package

There is no compilation step — Node.js runs the source directly. For distribution use `npm pack`:

```bash
npm install
npm pack                # creates my-js-plugin-1.0.0.tgz
```

### Tests (Node built-in test runner)

```bash
# Unit tests
npm test

# With coverage (Node 20+)
npm run test:coverage
```

#### Test Example (test/plugin.test.js)

```javascript
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { PluginBusinessService } from '../src/services/plugin.js';

const config = { plugin: { name: 't', version: '0.0.1', description: '', log_level: 'info' } };

test('hello returns default greeting', () => {
    const service = new PluginBusinessService(config);
    assert.equal(service.execute('hello', []), 'Hello, World!');
});

test('hello returns custom greeting', () => {
    const service = new PluginBusinessService(config);
    assert.equal(service.execute('hello', ['Portunix']), 'Hello, Portunix!');
});

test('unknown command throws', () => {
    const service = new PluginBusinessService(config);
    assert.throws(() => service.execute('unknown', []), /unknown command/);
});

test('process requires arguments', () => {
    const service = new PluginBusinessService(config);
    assert.throws(() => service.execute('process', []), /at least one argument/);
});
```

## Development Workflow

```bash
# Install dependencies
npm install

# Run in dev mode
node src/main.js --config config.yaml

# Install into Portunix
portunix plugin install .
portunix plugin enable my-js-plugin
portunix plugin start my-js-plugin
```

## Advanced Features

### MCP Tool Exposure

```javascript
ListTools(call, callback) {
    callback(null, {
        tools: [
            {
                name: 'process_file',
                description: 'Process a file with custom logic',
                schema: JSON.stringify({
                    type: 'object',
                    properties: { file_path: { type: 'string' } },
                    required: ['file_path'],
                }),
            },
        ],
    });
},

CallTool(call, callback) {
    if (call.request.tool_name !== 'process_file') {
        callback(null, { status: 'NOT_FOUND' });
        return;
    }
    const args = JSON.parse(call.request.arguments || '{}');
    callback(null, { status: 'SUCCESS', result: `Processed file: ${args.file_path}` });
},
```

### Logging

The template uses `console`. For production logging use `pino`:

```bash
npm install pino
```

```javascript
import pino from 'pino';
const logger = pino({ level: config.plugin.log_level });
logger.info({ command }, 'Executing command');
```

## Deployment

### Dockerfile

```dockerfile
FROM node:20-alpine AS base
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev

COPY src/ ./src/
COPY proto/ ./proto/
COPY plugin.json ./

EXPOSE 50051
CMD ["node", "src/main.js"]
```

### Plugin Manifest

See `template/plugin.json` for the complete manifest.

## Troubleshooting

1. **`@grpc/grpc-js` install fails on Alpine** — install `libc6-compat`:
   `apk add --no-cache libc6-compat`.
2. **ESM import errors** — ensure `"type": "module"` is set in `package.json`.
3. **TLS handshake errors** — for local dev use `grpc.ServerCredentials.createInsecure()`;
   for production load real certs.

## Next Steps

- Study the [template code](template/) for a complete example
- Learn about [MCP integration](../../mcp-integration/exposing-tools.md) for AI agents
- Review the [plugin architecture](../../architecture.md)

## Resources

- [@grpc/grpc-js documentation](https://github.com/grpc/grpc-node/tree/master/packages/grpc-js)
- [Node.js built-in test runner](https://nodejs.org/api/test.html)
- [Protocol Buffers for JavaScript](https://protobuf.dev/getting-started/javascripttutorial/)
