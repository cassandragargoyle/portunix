# Java Plugin Development Guide

Java offers enterprise-grade tooling, mature gRPC support, and a rich ecosystem for Portunix plugins. This guide walks you through creating
robust Java plugins using Maven and gRPC.

## Prerequisites

- Java 17 or later (Java 11 minimum)
- Maven 3.6+ (Gradle 7+ also supported)
- Protocol Buffers compiler (`protoc`) — auto-managed by the Maven `protobuf-maven-plugin`
- Portunix development environment

## Quick Start

Create a new Java plugin using the Portunix CLI:

```bash
portunix plugin create my-java-plugin --language=java
cd my-java-plugin
```

This creates a complete Java project structure:

```text
my-java-plugin/
├── plugin.json                 # Plugin manifest
├── pom.xml                     # Maven build descriptor
├── src/
│   ├── main/
│   │   ├── java/
│   │   │   └── ai/portunix/plugin/
│   │   │       ├── Main.java               # Plugin entry point
│   │   │       ├── config/
│   │   │       │   └── PluginConfig.java   # Configuration loading
│   │   │       ├── handlers/
│   │   │       │   └── PluginServiceImpl.java  # gRPC handler
│   │   │       └── services/
│   │   │           └── PluginService.java  # Business logic
│   │   ├── proto/
│   │   │   └── plugin.proto    # gRPC service definition
│   │   └── resources/
│   │       └── config.yaml     # Default configuration
│   └── test/
│       └── java/
│           └── ai/portunix/plugin/
│               └── PluginServiceImplTest.java
└── README.md
```

## Plugin Structure

### Maven Build Descriptor (pom.xml)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>ai.portunix.plugin</groupId>
    <artifactId>my-java-plugin</artifactId>
    <version>1.0.0</version>
    <packaging>jar</packaging>

    <properties>
        <maven.compiler.source>17</maven.compiler.source>
        <maven.compiler.target>17</maven.compiler.target>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <grpc.version>1.63.0</grpc.version>
        <protobuf.version>3.25.3</protobuf.version>
    </properties>

    <dependencies>
        <dependency>
            <groupId>io.grpc</groupId>
            <artifactId>grpc-netty-shaded</artifactId>
            <version>${grpc.version}</version>
        </dependency>
        <dependency>
            <groupId>io.grpc</groupId>
            <artifactId>grpc-protobuf</artifactId>
            <version>${grpc.version}</version>
        </dependency>
        <dependency>
            <groupId>io.grpc</groupId>
            <artifactId>grpc-stub</artifactId>
            <version>${grpc.version}</version>
        </dependency>
        <dependency>
            <groupId>io.grpc</groupId>
            <artifactId>grpc-services</artifactId>
            <version>${grpc.version}</version>
        </dependency>
        <dependency>
            <groupId>javax.annotation</groupId>
            <artifactId>javax.annotation-api</artifactId>
            <version>1.3.2</version>
        </dependency>
        <dependency>
            <groupId>org.yaml</groupId>
            <artifactId>snakeyaml</artifactId>
            <version>2.2</version>
        </dependency>
        <dependency>
            <groupId>org.junit.jupiter</groupId>
            <artifactId>junit-jupiter</artifactId>
            <version>5.10.2</version>
            <scope>test</scope>
        </dependency>
    </dependencies>

    <build>
        <extensions>
            <extension>
                <groupId>kr.motd.maven</groupId>
                <artifactId>os-maven-plugin</artifactId>
                <version>1.7.1</version>
            </extension>
        </extensions>
        <plugins>
            <plugin>
                <groupId>org.xolstice.maven.plugins</groupId>
                <artifactId>protobuf-maven-plugin</artifactId>
                <version>0.6.1</version>
                <configuration>
                    <protocArtifact>com.google.protobuf:protoc:${protobuf.version}:exe:${os.detected.classifier}</protocArtifact>
                    <pluginId>grpc-java</pluginId>
                    <pluginArtifact>io.grpc:protoc-gen-grpc-java:${grpc.version}:exe:${os.detected.classifier}</pluginArtifact>
                </configuration>
                <executions>
                    <execution>
                        <goals>
                            <goal>compile</goal>
                            <goal>compile-custom</goal>
                        </goals>
                    </execution>
                </executions>
            </plugin>
            <plugin>
                <artifactId>maven-shade-plugin</artifactId>
                <version>3.5.2</version>
                <executions>
                    <execution>
                        <phase>package</phase>
                        <goals><goal>shade</goal></goals>
                        <configuration>
                            <transformers>
                                <transformer implementation="org.apache.maven.plugins.shade.resource.ManifestResourceTransformer">
                                    <mainClass>ai.portunix.plugin.Main</mainClass>
                                </transformer>
                            </transformers>
                        </configuration>
                    </execution>
                </executions>
            </plugin>
        </plugins>
    </build>
</project>
```

### Main Entry Point (Main.java)

```java
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
```

### Configuration (config/PluginConfig.java)

```java
package ai.portunix.plugin.config;

import org.yaml.snakeyaml.Yaml;

import java.io.FileInputStream;
import java.io.IOException;
import java.util.Map;

public class PluginConfig {
    private String name;
    private String version;
    private String description;
    private String logLevel = "info";

    public static PluginConfig load(String path) throws IOException {
        Yaml yaml = new Yaml();
        try (FileInputStream input = new FileInputStream(path)) {
            Map<String, Object> data = yaml.load(input);
            return fromMap(data);
        }
    }

    @SuppressWarnings("unchecked")
    private static PluginConfig fromMap(Map<String, Object> data) {
        PluginConfig config = new PluginConfig();
        Map<String, Object> plugin = (Map<String, Object>) data.getOrDefault("plugin", Map.of());
        config.name = (String) plugin.getOrDefault("name", "my-java-plugin");
        config.version = (String) plugin.getOrDefault("version", "1.0.0");
        config.description = (String) plugin.getOrDefault("description", "");
        config.logLevel = (String) plugin.getOrDefault("log_level", "info");
        return config;
    }

    public String getName() { return name; }
    public String getVersion() { return version; }
    public String getDescription() { return description; }
    public String getLogLevel() { return logLevel; }
}
```

### gRPC Handler (handlers/PluginServiceImpl.java)

```java
package ai.portunix.plugin.handlers;

import ai.portunix.plugin.config.PluginConfig;
import ai.portunix.plugin.proto.PluginProto.*;
import ai.portunix.plugin.proto.PluginServiceGrpc;
import ai.portunix.plugin.services.PluginBusinessService;
import io.grpc.stub.StreamObserver;

public class PluginServiceImpl extends PluginServiceGrpc.PluginServiceImplBase {
    private final PluginConfig config;
    private final PluginBusinessService service;

    public PluginServiceImpl(PluginConfig config) {
        this.config = config;
        this.service = new PluginBusinessService(config);
    }

    @Override
    public void getInfo(GetInfoRequest request, StreamObserver<GetInfoResponse> responseObserver) {
        PluginInfo info = PluginInfo.newBuilder()
            .setName(config.getName())
            .setVersion(config.getVersion())
            .setDescription(config.getDescription())
            .addCapabilities("example-capability")
            .build();
        responseObserver.onNext(GetInfoResponse.newBuilder().setInfo(info).build());
        responseObserver.onCompleted();
    }

    @Override
    public void healthCheck(HealthCheckRequest request, StreamObserver<HealthCheckResponse> responseObserver) {
        HealthCheckResponse response = HealthCheckResponse.newBuilder()
            .setStatus(HealthCheckResponse.Status.SERVING)
            .build();
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void execute(ExecuteRequest request, StreamObserver<ExecuteResponse> responseObserver) {
        String result = service.execute(request.getCommand(), request.getArgsList());
        responseObserver.onNext(ExecuteResponse.newBuilder()
            .setResult(result)
            .setStatus(ExecuteResponse.Status.SUCCESS)
            .build());
        responseObserver.onCompleted();
    }

    @Override
    public void shutdown(ShutdownRequest request, StreamObserver<ShutdownResponse> responseObserver) {
        service.cleanup();
        responseObserver.onNext(ShutdownResponse.newBuilder()
            .setStatus(ShutdownResponse.Status.SUCCESS)
            .build());
        responseObserver.onCompleted();
    }
}
```

### Business Logic (services/PluginBusinessService.java)

```java
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
```

## Building and Testing

### Build

```bash
# Generate proto + compile
mvn clean package

# Skip tests
mvn clean package -DskipTests

# Run tests only
mvn test
```

The shaded JAR is produced at `target/my-java-plugin-1.0.0.jar` and is runnable with:

```bash
java -jar target/my-java-plugin-1.0.0.jar --port 50051 --config config.yaml
```

### Unit Test Example (src/test/java/.../PluginServiceImplTest.java)

```java
package ai.portunix.plugin;

import ai.portunix.plugin.services.PluginBusinessService;
import ai.portunix.plugin.config.PluginConfig;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

class PluginServiceImplTest {
    @Test
    void helloReturnsGreeting() {
        PluginConfig config = new PluginConfig();
        PluginBusinessService service = new PluginBusinessService(config);
        assertEquals("Hello, Portunix!", service.execute("hello", List.of("Portunix")));
    }

    @Test
    void unknownCommandThrows() {
        PluginConfig config = new PluginConfig();
        PluginBusinessService service = new PluginBusinessService(config);
        assertThrows(IllegalArgumentException.class,
            () -> service.execute("unknown", List.of()));
    }
}
```

## Development Workflow

```bash
# Compile + run locally
mvn compile exec:java -Dexec.mainClass="ai.portunix.plugin.Main"

# Install into Portunix
portunix plugin install target/my-java-plugin-1.0.0.jar
portunix plugin enable my-java-plugin
portunix plugin start my-java-plugin

# Health check
portunix plugin health my-java-plugin
```

## Advanced Features

### MCP Tool Exposure

Implement the optional `listTools` / `callTool` methods to expose MCP tools to AI agents:

```java
@Override
public void listTools(ListToolsRequest request, StreamObserver<ListToolsResponse> responseObserver) {
    MCPTool tool = MCPTool.newBuilder()
        .setName("process_file")
        .setDescription("Process a file with custom logic")
        .setSchema("""
            {
              "type": "object",
              "properties": {
                "file_path": {"type": "string"}
              },
              "required": ["file_path"]
            }
        """)
        .build();
    responseObserver.onNext(ListToolsResponse.newBuilder().addTools(tool).build());
    responseObserver.onCompleted();
}
```

### Logging

The Java template uses `java.util.logging`. For production-grade logging consider SLF4J + Logback:

```xml
<dependency>
    <groupId>org.slf4j</groupId>
    <artifactId>slf4j-api</artifactId>
    <version>2.0.13</version>
</dependency>
<dependency>
    <groupId>ch.qos.logback</groupId>
    <artifactId>logback-classic</artifactId>
    <version>1.5.6</version>
</dependency>
```

## Deployment

### Dockerfile

```dockerfile
FROM eclipse-temurin:17-jdk AS builder
WORKDIR /app
COPY pom.xml .
COPY src ./src
RUN ./mvnw -B -DskipTests package

FROM eclipse-temurin:17-jre
WORKDIR /app
COPY --from=builder /app/target/*.jar /app/plugin.jar
COPY plugin.json .
EXPOSE 50051
ENTRYPOINT ["java", "-jar", "plugin.jar"]
```

### Plugin Manifest (plugin.json)

See `template/plugin.json` for the complete manifest.

## Troubleshooting

1. **`protoc` fails on Apple Silicon / Linux ARM64** — ensure `os-maven-plugin` produces
   the right classifier; for arm64 add `-Dos.detected.classifier=linux-aarch_64`.
2. **`UnsatisfiedLinkError` for `grpc-netty-shaded`** — use `grpc-netty-shaded` (not
   `grpc-netty`) to avoid native netty conflicts.
3. **Slow startup** — pass `-XX:+UseSerialGC -Xss256k` to reduce footprint for sidecar deployments.

## Next Steps

- Study the [template code](template/) for a complete example
- Learn about [MCP integration](../../mcp-integration/exposing-tools.md) for AI agents
- Review the [plugin architecture](../../architecture.md)

## Resources

- [gRPC Java](https://grpc.io/docs/languages/java/)
- [Protocol Buffers Java Tutorial](https://protobuf.dev/getting-started/javatutorial/)
- [protobuf-maven-plugin](https://www.xolstice.org/protobuf-maven-plugin/)
