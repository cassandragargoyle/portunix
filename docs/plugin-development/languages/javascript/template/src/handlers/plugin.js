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
