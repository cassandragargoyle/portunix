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

The entry point has two modes. With a command (`portunix my-js-plugin hello Portunix`)
it runs the command once and exits; without a command (`portunix plugin start`) it
boots the gRPC server. The gRPC modules are imported only in service mode, so CLI
commands start fast.

```javascript
import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';
import { parseArgs } from 'node:util';

import { loadConfig } from './config.js';
import { PluginBusinessService } from './services/plugin.js';

const __dirname = dirname(fileURLToPath(import.meta.url));
const PROTO_PATH = resolve(__dirname, '..', 'proto', 'plugin.proto');
const MANIFEST_PATH = resolve(__dirname, '..', 'plugin.json');

// Without a command the plugin runs as a gRPC service (portunix plugin start);
// with a command it runs once as a CLI (portunix <plugin> <command> [args...])
async function main() {
    const { values, positionals } = parseArgs({
        allowPositionals: true,
        options: {
            port: { type: 'string', default: '50051' },
            config: { type: 'string', default: 'config.yaml' },
            version: { type: 'boolean', default: false },
        },
    });

    const config = await loadConfig(values.config);

    if (values.version) {
        console.log(`${config.plugin.name} ${config.plugin.version}`);
        return;
    }

    if (positionals.length > 0) {
        const [command, ...args] = positionals;
        const service = new PluginBusinessService(config);
        console.log(service.execute(command, args));
        return;
    }

    await serve(config, values.port);
}

// serve boots the gRPC server implementing portunix.plugins.PluginService;
// gRPC modules load only in service mode
async function serve(config, port) {
    const { default: grpc } = await import('@grpc/grpc-js');
    const { default: protoLoader } = await import('@grpc/proto-loader');
    const { createPluginHandler } = await import('./handlers/plugin.js');

    const manifest = JSON.parse(await readFile(MANIFEST_PATH, 'utf8'));

    const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
        keepCase: true,
        longs: String,
        enums: String,
        defaults: true,
        oneofs: true,
    });
    const proto = grpc.loadPackageDefinition(packageDefinition).portunix.plugins;

    const server = new grpc.Server();
    const shutdown = () => {
        console.log('Shutting down gracefully...');
        server.tryShutdown(() => process.exit(0));
    };
    server.addService(proto.PluginService.service, createPluginHandler(config, manifest));

    // Portunix connects over localhost; the plugin is not exposed to the network
    const bindAddress = `127.0.0.1:${port}`;
    server.bindAsync(bindAddress, grpc.ServerCredentials.createInsecure(), (err, boundPort) => {
        if (err) {
            console.error('Failed to bind:', err);
            process.exit(1);
        }
        console.log(`Plugin server listening on port ${boundPort}`);
    });

    process.on('SIGINT', shutdown);
    process.on('SIGTERM', shutdown);
}

main().catch((err) => {
    console.error(`Error: ${err.message}`);
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

The handlers implement `portunix.plugins.PluginService` — the service Portunix
calls when it starts the plugin as a service (`portunix plugin start`): `Initialize`,
`GetInfo`, `Health`, `Execute`, `ListCommands` and `Shutdown`. `proto/plugin.proto`
is a copy of the `PluginService` part of the Portunix contract
(`src/app/plugins/proto/plugin.proto`); keep its package, service and messages
unchanged, otherwise Portunix fails with `Unimplemented` on `Initialize`. Plugin
info and commands are read from `plugin.json`, and the server listens on
`127.0.0.1` only, because Portunix connects over localhost. `Shutdown` only cleans
up and answers: Portunix ends the process right after it (SIGTERM, handled in
`main.js`), and a plugin that exits inside `Shutdown` drops the response.

```javascript
import { PluginBusinessService } from '../services/plugin.js';

// toPluginCommand maps a plugin.json command onto the PluginCommand message
function toPluginCommand(command) {
    return {
        name: command.name,
        description: command.description ?? '',
        subcommands: command.subcommands ?? [],
        parameters: (command.parameters ?? []).map((p) => ({
            name: p.name,
            type: p.type ?? 'string',
            description: p.description ?? '',
            required: Boolean(p.required),
            default_value: p.default_value ?? '',
        })),
        examples: command.examples ?? [],
    };
}

// createPluginHandler implements portunix.plugins.PluginService; plugin info
// and commands come from plugin.json, so the manifest stays the single source
export function createPluginHandler(config, manifest) {
    const service = new PluginBusinessService(config);
    const startedAt = Date.now();
    const permissions = manifest.permissions ?? {};

    const pluginInfo = () => ({
        name: manifest.name,
        version: manifest.version,
        description: manifest.description ?? '',
        author: manifest.author ?? '',
        license: manifest.license ?? '',
        supported_os: manifest.dependencies?.os_support ?? [],
        commands: (manifest.commands ?? []).map(toPluginCommand),
        capabilities: capabilities(),
        required_permissions: {
            filesystem: permissions.filesystem ?? [],
            network: permissions.network ?? [],
            database: permissions.database ?? [],
            system: permissions.system ?? [],
            level: permissions.level ?? 'limited',
        },
    });

    const capabilities = () => ({
        filesystem_access: (permissions.filesystem ?? []).length > 0,
        network_access: (permissions.network ?? []).length > 0,
        database_access: (permissions.database ?? []).length > 0,
        container_access: false,
        system_commands: (permissions.system ?? []).length > 0,
        mcp_tools: (manifest.ai_integration?.mcp_tools ?? []).map((t) => t.name),
    });

    return {
        Initialize(call, callback) {
            callback(null, { success: true, message: 'initialized', plugin_info: pluginInfo() });
        },

        GetInfo(call, callback) {
            callback(null, { plugin_info: pluginInfo() });
        },

        Health(call, callback) {
            callback(null, {
                healthy: true,
                status: 'running',
                message: 'ok',
                uptime_seconds: Math.floor((Date.now() - startedAt) / 1000),
                metrics: {},
            });
        },

        Execute(call, callback) {
            try {
                const output = service.execute(call.request.command, call.request.args ?? []);
                callback(null, { success: true, output, exit_code: 0 });
            } catch (err) {
                callback(null, { success: false, message: err.message, error: err.message, exit_code: 1 });
            }
        },

        ListCommands(call, callback) {
            callback(null, { commands: (manifest.commands ?? []).map(toPluginCommand), capabilities: capabilities() });
        },

        // Portunix stops the process after this call (SIGTERM, handled in
        // main.js); exiting here would drop the response before it is sent
        Shutdown(call, callback) {
            service.cleanup();
            callback(null, { success: true, message: 'cleanup done' });
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

The Portunix contract has no separate tool RPCs. Declare MCP tools in `plugin.json`
under `ai_integration.mcp_tools`; the template reports their names in
`capabilities.mcp_tools` of `Initialize`, `GetInfo` and `ListCommands`, and a tool
call reaches the plugin as an `Execute` request with the tool name as `command`:

```json
{
  "ai_integration": {
    "mcp_tools": [
      { "name": "process_file", "description": "Process a file with custom logic" }
    ]
  }
}
```

```javascript
// src/services/plugin.js — handle the tool like any other command
case 'process_file':
    return `Processed file: ${args[0]}`;
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

JavaScript plugins use the `node` runtime. `binary` is the path of the entry script
relative to the plugin directory, without the `node` command:

```json
{
  "plugin": {
    "type": "grpc",
    "binary": "src/main.js",
    "runtime": "node",
    "runtime_version": ">=18",
    "port": 9001
  }
}
```

- Portunix runs the plugin as `node <install_path>/src/main.js [args...]`, with the
  plugin install directory as working directory and `PORTUNIX_CWD` set to the caller
  directory (see the [plugin architecture](../../architecture.md))
- `runtime_version` is optional; `portunix plugin install` and `portunix plugin check`
  compare it with `node --version`
- A missing Node.js runtime is reported with the fix `portunix install nodejs`;
  `plugin health` reports the plugin as unhealthy (`runtime_missing`)
- A Node.js older than `runtime_version` is reported with an upgrade hint: on Windows
  `portunix install nodejs --variant=latest`; on Linux the `nodejs` package comes from
  the distribution and may be older, so the hint points to https://nodejs.org or nvm

See `template/plugin.json` for the complete manifest.

### npm Dependencies

`portunix plugin install` copies the plugin directory and then handles the
`dependencies` from `package.json`:

- **`node_modules` shipped with the plugin** — used as is, npm is not run. Use this
  for offline installation or native modules built for the target platform
- **No `node_modules`** — Portunix runs npm in the install directory:
  `npm ci --omit=dev` when `package-lock.json` (or `npm-shrinkwrap.json`) is present,
  otherwise `npm install --omit=dev`. Commit the lockfile for reproducible installs
- **No `dependencies`** — nothing is installed

npm must be available when dependencies are installed; otherwise the installation
stops with the hint `portunix install nodejs` and leaves nothing behind. An installed
plugin whose `node_modules` is missing is reported as broken and fixed with
`portunix plugin repair <name>`. `devDependencies` are never installed.

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
