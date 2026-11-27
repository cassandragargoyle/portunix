package ai.portunix.plugin;

import ai.portunix.plugin.config.PluginConfig;
import ai.portunix.plugin.handlers.PluginServiceImpl;
import ai.portunix.plugin.manifest.PluginManifest;
import ai.portunix.plugin.services.PluginBusinessService;
import io.grpc.Server;
import io.grpc.netty.shaded.io.grpc.netty.NettyServerBuilder;
import io.grpc.protobuf.services.HealthStatusManager;

import java.net.InetSocketAddress;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.TimeUnit;
import java.util.logging.Logger;

/**
 * Without a command the plugin runs as a gRPC service (portunix plugin start);
 * with a command it runs once as a CLI (portunix &lt;plugin&gt; &lt;command&gt; [args...])
 */
public class Main {
    private static final Logger logger = Logger.getLogger(Main.class.getName());

    public static void main(String[] args) throws Exception {
        int port = 50051;
        String configPath = "config.yaml";
        String manifestPath = "plugin.json";
        boolean showVersion = false;
        List<String> positionals = new ArrayList<>();

        for (int i = 0; i < args.length; i++) {
            String arg = args[i];
            if (!positionals.isEmpty()) {
                // Everything after the command belongs to the command
                positionals.add(arg);
            } else if ("--port".equals(arg) && i + 1 < args.length) {
                port = Integer.parseInt(args[++i]);
            } else if ("--config".equals(arg) && i + 1 < args.length) {
                configPath = args[++i];
            } else if ("--manifest".equals(arg) && i + 1 < args.length) {
                manifestPath = args[++i];
            } else if ("--version".equals(arg)) {
                showVersion = true;
            } else {
                positionals.add(arg);
            }
        }

        PluginConfig config = PluginConfig.load(configPath);

        if (showVersion) {
            System.out.println(config.getName() + " " + config.getVersion());
            return;
        }

        if (!positionals.isEmpty()) {
            runCommand(config, positionals.get(0), positionals.subList(1, positionals.size()));
            return;
        }

        // Portunix runs the plugin with the install directory as working
        // directory, so the default manifest path resolves to its plugin.json
        PluginManifest manifest = PluginManifest.load(Path.of(manifestPath));
        serve(config, manifest, port);
    }

    // runCommand executes one command and prints its output (CLI mode)
    private static void runCommand(PluginConfig config, String command, List<String> args) {
        try {
            System.out.println(new PluginBusinessService(config).execute(command, args));
        } catch (IllegalArgumentException e) {
            System.err.println("Error: " + e.getMessage());
            System.exit(1);
        }
    }

    // serve runs the gRPC server implementing portunix.plugins.PluginService
    // until the JVM receives SIGINT or SIGTERM
    private static void serve(PluginConfig config, PluginManifest manifest, int port) throws Exception {
        HealthStatusManager healthManager = new HealthStatusManager();

        // Portunix connects over localhost; the plugin is not exposed to the network
        Server server = NettyServerBuilder.forAddress(new InetSocketAddress("127.0.0.1", port))
            .addService(new PluginServiceImpl(config, manifest))
            .addService(healthManager.getHealthService())
            .build()
            .start();

        logger.info("Plugin server " + manifest.getName() + " " + manifest.getVersion()
            + " listening on 127.0.0.1:" + port);

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
}
