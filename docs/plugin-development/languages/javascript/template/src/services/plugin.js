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
