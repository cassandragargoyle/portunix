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
