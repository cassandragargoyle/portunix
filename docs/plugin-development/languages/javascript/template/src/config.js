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
