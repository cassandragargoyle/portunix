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
