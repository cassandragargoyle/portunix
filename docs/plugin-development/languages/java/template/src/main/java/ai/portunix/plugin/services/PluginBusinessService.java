package ai.portunix.plugin.services;

import ai.portunix.plugin.config.PluginConfig;

import java.util.List;
import java.util.logging.Logger;

public class PluginBusinessService {
    private static final Logger logger = Logger.getLogger(PluginBusinessService.class.getName());
    private final PluginConfig config;

    public PluginBusinessService(PluginConfig config) {
        this.config = config;
    }

    public String execute(String command, List<String> args) {
        return switch (command) {
            case "hello" -> handleHello(args);
            case "process" -> handleProcess(args);
            default -> throw new IllegalArgumentException("unknown command: " + command);
        };
    }

    private String handleHello(List<String> args) {
        String name = args.isEmpty() ? "World" : args.get(0);
        return "Hello, " + name + "!";
    }

    private String handleProcess(List<String> args) {
        if (args.isEmpty()) {
            throw new IllegalArgumentException("process command requires at least one argument");
        }
        return "Processed: " + String.join(", ", args);
    }

    public void cleanup() {
        logger.info("Performing cleanup operations...");
    }
}
