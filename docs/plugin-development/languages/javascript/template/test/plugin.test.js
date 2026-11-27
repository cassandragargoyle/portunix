import { test } from 'node:test';
import assert from 'node:assert/strict';
import { PluginBusinessService } from '../src/services/plugin.js';

const config = {
    plugin: { name: 't', version: '0.0.1', description: '', log_level: 'info' },
};

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

// Handlers implement portunix.plugins.PluginService; called without a gRPC server
import { createPluginHandler } from '../src/handlers/plugin.js';

const manifest = {
    name: 'js-plugin-template',
    version: '1.0.0',
    description: 'test',
    dependencies: { os_support: ['linux'] },
    permissions: { filesystem: ['read'], network: [], database: [], system: [], level: 'limited' },
    commands: [{ name: 'hello', description: 'Print a greeting', parameters: [{ name: 'name', type: 'string' }] }],
};

function call(handler, method, request = {}) {
    return new Promise((resolve, reject) => {
        handler[method]({ request }, (err, response) => (err ? reject(err) : resolve(response)));
    });
}

test('Initialize reports plugin info from the manifest', async () => {
    const handler = createPluginHandler(config, manifest);
    const response = await call(handler, 'Initialize', { plugin_name: 'js-plugin-template' });
    assert.equal(response.success, true);
    assert.equal(response.plugin_info.name, 'js-plugin-template');
    assert.equal(response.plugin_info.commands[0].name, 'hello');
    assert.equal(response.plugin_info.capabilities.filesystem_access, true);
});

test('Execute runs the business command', async () => {
    const handler = createPluginHandler(config, manifest);
    const response = await call(handler, 'Execute', { command: 'hello', args: ['gRPC'] });
    assert.deepEqual([response.success, response.output, response.exit_code], [true, 'Hello, gRPC!', 0]);
});

test('Execute reports an unknown command as failure', async () => {
    const handler = createPluginHandler(config, manifest);
    const response = await call(handler, 'Execute', { command: 'unknown', args: [] });
    assert.equal(response.success, false);
    assert.match(response.error, /unknown command/);
});

test('Health and Shutdown', async () => {
    const handler = createPluginHandler(config, manifest);
    assert.equal((await call(handler, 'Health')).healthy, true);
    assert.equal((await call(handler, 'Shutdown')).success, true);
});
