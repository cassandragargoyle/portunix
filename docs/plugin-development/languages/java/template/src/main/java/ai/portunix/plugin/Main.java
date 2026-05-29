package ai.portunix.plugin;

import ai.portunix.plugin.config.PluginConfig;
import ai.portunix.plugin.handlers.PluginServiceImpl;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.protobuf.services.HealthStatusManager;

import java.util.concurrent.TimeUnit;
import java.util.logging.Logger;

public class Main {
    private static final Logger logger = Logger.getLogger(Main.class.getName());

    public static void main(String[] args) throws Exception {
        int port = parsePort(args, 50051);
        String configPath = parseConfigPath(args, "config.yaml");

        PluginConfig config = PluginConfig.load(configPath);

        HealthStatusManager healthManager = new HealthStatusManager();

        Server server = ServerBuilder.forPort(port)
            .addService(new PluginServiceImpl(config))
            .addService(healthManager.getHealthService())
            .build()
            .start();

        logger.info("Plugin server listening on port " + port);

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            logger.info("Shutting down gracefully...");
            try {
                server.shutdown().awaitTermination(30, TimeUnit.SECONDS);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }));

        server.awaitTermination();
    }

    private static int parsePort(String[] args, int defaultPort) {
        for (int i = 0; i < args.length - 1; i++) {
            if ("--port".equals(args[i])) {
                return Integer.parseInt(args[i + 1]);
            }
        }
        return defaultPort;
    }

    private static String parseConfigPath(String[] args, String defaultPath) {
        for (int i = 0; i < args.length - 1; i++) {
            if ("--config".equals(args[i])) {
                return args[i + 1];
            }
        }
        return defaultPath;
    }
}
