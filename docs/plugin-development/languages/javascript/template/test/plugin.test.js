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
