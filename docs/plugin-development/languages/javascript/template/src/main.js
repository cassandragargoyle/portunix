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
