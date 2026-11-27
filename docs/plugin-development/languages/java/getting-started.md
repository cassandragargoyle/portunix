# Java Plugin Development Guide

Java offers enterprise-grade tooling, mature gRPC support, and a rich ecosystem for Portunix plugins. This guide walks you through creating
Java plugins using Maven and gRPC.

## Prerequisites

- Java 17 or later (`portunix install java-21`)
- Maven 3.6+ (`portunix install maven`)
- Protocol Buffers compiler (`protoc`) — not installed separately, the Maven
  `protobuf-maven-plugin` downloads it for the build platform
- Portunix development environment

## Quick Start

Start from a copy of the [template](template/):

```bash
cp -r docs/plugin-development/languages/java/template my-java-plugin
cd my-java-plugin
mvn clean package
```

The template has this structure:

```text
my-java-plugin/
├── plugin.json                 # Plugin manifest
├── pom.xml                     # Maven build descriptor
├── src/
│   ├── main/
│   │   ├── java/
│   │   │   └── ai/portunix/plugin/
│   │   │       ├── Main.java                   # Entry point - CLI command or gRPC server
│   │   │       ├── config/
│   │   │       │   └── PluginConfig.java       # Configuration loading
│   │   │       ├── manifest/
│   │   │       │   └── PluginManifest.java     # plugin.json -> PluginInfo / PluginCommand
│   │   │       ├── handlers/
│   │   │       │   └── PluginServiceImpl.java  # gRPC handler
│   │   │       └── services/
│   │   │           └── PluginBusinessService.java  # Business logic
│   │   ├── proto/
│   │   │   └── plugin.proto    # Copy of the Portunix plugin contract
│   │   └── resources/
│   │       └── config.yaml     # Default configuration
│   └── test/
│       └── java/
│           └── ai/portunix/plugin/
│               ├── PluginBusinessServiceTest.java
│               └── PluginServiceImplTest.java
└── README.md
```

## Plugin Structure

### Maven Build Descriptor (pom.xml)

The build generates the Java classes from `src/main/proto/plugin.proto` and packages a
shaded (self-contained) JAR with `ai.portunix.plugin.Main` as main class. The relevant
parts of the template `pom.xml`:

```xml
<properties>
    <maven.compiler.source>17</maven.compiler.source>
    <maven.compiler.target>17</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
    <grpc.version>1.63.0</grpc.version>
    <protobuf.version>3.25.3</protobuf.version>
</properties>

<dependencies>
    <!-- grpc-netty-shaded, grpc-protobuf, grpc-stub, grpc-services ${grpc.version} -->
    <!-- javax.annotation-api 1.3.2 -->
    <dependency>
        <groupId>com.google.code.gson</groupId>
        <artifactId>gson</artifactId>
        <version>2.10.1</version>
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
            <!-- Surefire 3 runs JUnit 5 tests; the Maven default version finds none -->
            <artifactId>maven-surefire-plugin</artifactId>
            <version>3.2.5</version>
        </plugin>
        <plugin>
            <artifactId>maven-shade-plugin</artifactId>
            <version>3.5.2</version>
            <executions>
                <execution>
                    <phase>package</phase>
                    <goals>
                        <goal>shade</goal>
                    </goals>
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
```

Keep the `maven-surefire-plugin` pin: the Surefire version Maven uses by default does
not detect JUnit 5 tests, so `mvn test` would pass without running a single test.

### Contract (src/main/proto/plugin.proto)

`plugin.proto` is a copy of the `PluginService` part of the Portunix contract
(`src/app/plugins/proto/plugin.proto`). Keep its `package portunix.plugins`, the
service and the messages unchanged, otherwise Portunix fails with `Unimplemented` on
`Initialize`. Only the Java options differ; they put the generated classes into
`ai.portunix.plugin.proto` (`PluginProto` messages, `PluginServiceGrpc` stubs):

```protobuf
syntax = "proto3";

package portunix.plugins;

option java_package = "ai.portunix.plugin.proto";
option java_outer_classname = "PluginProto";
option java_multiple_files = false;

service PluginService {
    rpc Initialize(InitializeRequest) returns (InitializeResponse);
    rpc Execute(ExecuteRequest) returns (ExecuteResponse);
    rpc GetInfo(GetInfoRequest) returns (GetInfoResponse);
    rpc Health(HealthRequest) returns (HealthResponse);
    rpc Shutdown(ShutdownRequest) returns (ShutdownResponse);
    rpc ListCommands(ListCommandsRequest) returns (ListCommandsResponse);
}
```

### Main Entry Point (Main.java)

The entry point has two modes. With a command (`portunix java-plugin-template hello
Portunix`) it runs the command once and exits (CLI mode, no gRPC); without a command
(`portunix plugin start`) it runs the gRPC server. Options are `--port`, `--config`,
`--manifest` and `--version`; everything after the command is passed to the command.

```java
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
```

`--version` is also what `portunix plugin health` runs (`java -jar <binary> --version`)
to check that the plugin starts.

### Configuration (config/PluginConfig.java)

`PluginConfig` reads the `plugin` section of a YAML file (`--config`, default
`config.yaml`). A missing file is not an error, the defaults are used instead. The
template ships a sample in `src/main/resources/config.yaml`.

```java
package ai.portunix.plugin.config;

import org.yaml.snakeyaml.Yaml;

import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.util.Map;

public class PluginConfig {
    private String name = "my-java-plugin";
    private String version = "1.0.0";
    private String description = "";
    private String logLevel = "info";

    public static PluginConfig load(String path) throws IOException {
        Yaml yaml = new Yaml();
        try (FileInputStream input = new FileInputStream(path)) {
            Map<String, Object> data = yaml.load(input);
            return fromMap(data);
        } catch (FileNotFoundException notFound) {
            return new PluginConfig();
        }
    }

    @SuppressWarnings("unchecked")
    private static PluginConfig fromMap(Map<String, Object> data) {
        PluginConfig config = new PluginConfig();
        if (data == null) {
            return config;
        }
        Map<String, Object> plugin = (Map<String, Object>) data.getOrDefault("plugin", Map.of());
        config.name = (String) plugin.getOrDefault("name", config.name);
        config.version = (String) plugin.getOrDefault("version", config.version);
        config.description = (String) plugin.getOrDefault("description", config.description);
        config.logLevel = (String) plugin.getOrDefault("log_level", config.logLevel);
        return config;
    }

    // getName, getVersion, getDescription, getLogLevel
}
```

### Manifest (manifest/PluginManifest.java)

`plugin.json` is the single source of the plugin information and commands.
`PluginManifest` reads it with Gson (`--manifest`, default `plugin.json` in the working
directory) and maps it onto the contract messages:

- `toPluginInfo()` — `PluginInfo` with name, version, description, author, license,
  `supported_os` (from `dependencies.os_support`), commands, capabilities and
  `required_permissions` (from `permissions`)
- `toPluginCommands()` — the `commands` array as `PluginCommand` / `PluginParameter`
  messages
- `toCapabilities()` — `PluginCapabilities`; the access flags are derived from the
  non-empty `permissions` lists and `mcp_tools` holds the names from
  `ai_integration.mcp_tools`

`load()` fails when the manifest has no `name`, so a service plugin does not start with
a broken manifest.

```java
public static PluginManifest load(Path path) throws IOException {
    try (Reader reader = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
        PluginManifest manifest = new Gson().fromJson(reader, PluginManifest.class);
        if (manifest == null || manifest.name == null || manifest.name.isEmpty()) {
            throw new IOException("manifest " + path + " has no plugin name");
        }
        return manifest;
    }
}

public PluginCapabilities toCapabilities() {
    Permissions p = permissions == null ? new Permissions() : permissions;
    PluginCapabilities.Builder builder = PluginCapabilities.newBuilder()
        .setFilesystemAccess(!orEmpty(p.filesystem).isEmpty())
        .setNetworkAccess(!orEmpty(p.network).isEmpty())
        .setDatabaseAccess(!orEmpty(p.database).isEmpty())
        .setContainerAccess(false)
        .setSystemCommands(!orEmpty(p.system).isEmpty());
    if (aiIntegration != null) {
        for (NamedItem tool : orEmpty(aiIntegration.mcpTools)) {
            builder.addMcpTools(orEmpty(tool.name));
        }
    }
    return builder.build();
}
```

### gRPC Handler (handlers/PluginServiceImpl.java)

The handler implements `portunix.plugins.PluginService` — the service Portunix calls
when it starts the plugin as a service. `portunix plugin start` launches
`java -jar <binary> --port N` with the plugin install directory as working directory
and then calls `Initialize`, `Health` and `Execute`; `portunix plugin stop` calls
`Shutdown`. The server listens on `127.0.0.1` only, because Portunix connects over
localhost.

Two rules matter:

- **Execute errors** — a failed command is reported in the response
  (`success=false`, `error`, `exit_code=1`), not as a gRPC error, so Portunix can show
  the message to the user
- **Shutdown** — `shutdown` only cleans up and answers. Portunix sends SIGTERM right
  after the call, and the JVM shutdown hook in `Main` stops the server. A plugin that
  calls `System.exit` inside `shutdown` drops the response and `portunix plugin stop`
  fails with `EOF`

```java
package ai.portunix.plugin.handlers;

import ai.portunix.plugin.config.PluginConfig;
import ai.portunix.plugin.manifest.PluginManifest;
import ai.portunix.plugin.proto.PluginProto.ExecuteRequest;
import ai.portunix.plugin.proto.PluginProto.ExecuteResponse;
import ai.portunix.plugin.proto.PluginProto.GetInfoRequest;
import ai.portunix.plugin.proto.PluginProto.GetInfoResponse;
import ai.portunix.plugin.proto.PluginProto.HealthRequest;
import ai.portunix.plugin.proto.PluginProto.HealthResponse;
import ai.portunix.plugin.proto.PluginProto.InitializeRequest;
import ai.portunix.plugin.proto.PluginProto.InitializeResponse;
import ai.portunix.plugin.proto.PluginProto.ListCommandsRequest;
import ai.portunix.plugin.proto.PluginProto.ListCommandsResponse;
import ai.portunix.plugin.proto.PluginProto.ShutdownRequest;
import ai.portunix.plugin.proto.PluginProto.ShutdownResponse;
import ai.portunix.plugin.proto.PluginServiceGrpc;
import ai.portunix.plugin.services.PluginBusinessService;
import io.grpc.stub.StreamObserver;

import java.util.logging.Logger;

/**
 * Implements portunix.plugins.PluginService; plugin info and commands come
 * from plugin.json, so the manifest stays the single source
 */
public class PluginServiceImpl extends PluginServiceGrpc.PluginServiceImplBase {
    private static final Logger logger = Logger.getLogger(PluginServiceImpl.class.getName());

    private final PluginManifest manifest;
    private final PluginBusinessService service;
    private final long startedAt = System.currentTimeMillis();

    public PluginServiceImpl(PluginConfig config, PluginManifest manifest) {
        this.manifest = manifest;
        this.service = new PluginBusinessService(config);
    }

    @Override
    public void initialize(InitializeRequest request, StreamObserver<InitializeResponse> responseObserver) {
        logger.info("Initialize called by " + request.getPluginName());
        responseObserver.onNext(InitializeResponse.newBuilder()
            .setSuccess(true)
            .setMessage("initialized")
            .setPluginInfo(manifest.toPluginInfo())
            .build());
        responseObserver.onCompleted();
    }

    @Override
    public void getInfo(GetInfoRequest request, StreamObserver<GetInfoResponse> responseObserver) {
        responseObserver.onNext(GetInfoResponse.newBuilder().setPluginInfo(manifest.toPluginInfo()).build());
        responseObserver.onCompleted();
    }

    @Override
    public void health(HealthRequest request, StreamObserver<HealthResponse> responseObserver) {
        responseObserver.onNext(HealthResponse.newBuilder()
            .setHealthy(true)
            .setStatus("running")
            .setMessage("ok")
            .setUptimeSeconds((System.currentTimeMillis() - startedAt) / 1000)
            .build());
        responseObserver.onCompleted();
    }

    // Command errors are reported in the response, not as gRPC errors, so
    // Portunix can show them to the user
    @Override
    public void execute(ExecuteRequest request, StreamObserver<ExecuteResponse> responseObserver) {
        try {
            String output = service.execute(request.getCommand(), request.getArgsList());
            responseObserver.onNext(ExecuteResponse.newBuilder()
                .setSuccess(true)
                .setOutput(output)
                .setExitCode(0)
                .build());
        } catch (IllegalArgumentException e) {
            responseObserver.onNext(ExecuteResponse.newBuilder()
                .setSuccess(false)
                .setMessage(e.getMessage())
                .setError(e.getMessage())
                .setExitCode(1)
                .build());
        }
        responseObserver.onCompleted();
    }

    @Override
    public void listCommands(ListCommandsRequest request, StreamObserver<ListCommandsResponse> responseObserver) {
        responseObserver.onNext(ListCommandsResponse.newBuilder()
            .addAllCommands(manifest.toPluginCommands())
            .setCapabilities(manifest.toCapabilities())
            .build());
        responseObserver.onCompleted();
    }

    // Portunix stops the process right after this call (SIGTERM, handled by
    // the shutdown hook in Main); exiting here would drop the response
    @Override
    public void shutdown(ShutdownRequest request, StreamObserver<ShutdownResponse> responseObserver) {
        service.cleanup();
        responseObserver.onNext(ShutdownResponse.newBuilder()
            .setSuccess(true)
            .setMessage("cleanup done")
            .build());
        responseObserver.onCompleted();
    }
}
```

### Business Logic (services/PluginBusinessService.java)

The commands are implemented once and shared by both modes: CLI mode calls
`execute` directly, service mode calls it from the `Execute` handler. An unknown command
or invalid arguments throw `IllegalArgumentException`.

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

A new command needs a `case` here and an entry in the `commands` array of
`plugin.json`, so Portunix lists it.

## Building and Testing

### Build

```bash
# Generate proto classes, compile, test and package
mvn clean package

# Skip tests
mvn clean package -DskipTests

# Run tests only
mvn test
```

The shaded JAR is produced at `target/my-java-plugin-1.0.0.jar` — the path `plugin.json`
references as `binary`. Run it from the template root, where `plugin.json` is:

```bash
# Service mode (gRPC server on 127.0.0.1:50051)
java -jar target/my-java-plugin-1.0.0.jar --port 50051 --config src/main/resources/config.yaml

# CLI mode (runs one command and exits)
java -jar target/my-java-plugin-1.0.0.jar hello Portunix
```

### Tests

Maven runs the tests from the template root, so `PluginServiceImplTest` loads the real
`plugin.json`. The template has two test classes:

- `PluginBusinessServiceTest` — the commands (`hello` with and without a name, unknown
  command, `process` without arguments)
- `PluginServiceImplTest` — the gRPC handler called directly with a capturing
  `StreamObserver`, no server needed: `initialize` returns the manifest info, `health`
  is healthy, `execute` reports output and errors in the response, `listCommands`
  returns the commands and `shutdown` answers with success

```java
@BeforeEach
void setUp() throws Exception {
    // Maven runs tests from the template root, where plugin.json is
    handler = new PluginServiceImpl(new PluginConfig(), PluginManifest.load(Path.of("plugin.json")));
}

@Test
void executeReportsOutputAndErrors() {
    Capture<ExecuteResponse> ok = new Capture<>();
    handler.execute(ExecuteRequest.newBuilder().setCommand("hello").addArgs("Portunix").build(), ok);
    assertTrue(ok.value.getSuccess());
    assertEquals("Hello, Portunix!", ok.value.getOutput());

    Capture<ExecuteResponse> failed = new Capture<>();
    handler.execute(ExecuteRequest.newBuilder().setCommand("unknown").build(), failed);
    assertFalse(failed.value.getSuccess());
    assertNotEquals(0, failed.value.getExitCode());
}
```

## Development Workflow

```bash
# Build the JAR first; plugin install copies the directory as it is
mvn clean package

# Install into Portunix from the directory containing plugin.json
portunix plugin install .
portunix plugin enable java-plugin-template

# CLI mode
portunix java-plugin-template hello Portunix

# Service mode
portunix plugin start java-plugin-template
portunix plugin health java-plugin-template
portunix plugin stop java-plugin-template
```

## Advanced Features

### MCP Tool Exposure

The Portunix contract has no separate tool RPCs. Declare MCP tools in `plugin.json`
under `ai_integration.mcp_tools`; the template reports their names in
`capabilities.mcp_tools` of `Initialize`, `GetInfo` and `ListCommands`, and a tool
call reaches the plugin as an `Execute` request with the tool name as `command`:

```json
{
  "ai_integration": {
    "mcp_tools": [
      { "name": "process_file", "description": "Process a file with custom logic" }
    ]
  }
}
```

```java
// services/PluginBusinessService.java — handle the tool like any other command
case "process_file" -> "Processed file: " + args.get(0);
```

### Logging

The template uses `java.util.logging`. For production-grade logging consider SLF4J + Logback:

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

### Plugin Manifest

Java plugins use the `java` runtime. `binary` is the path of the JAR relative to the
plugin directory:

```json
{
  "plugin": {
    "type": "grpc",
    "binary": "target/my-java-plugin-1.0.0.jar",
    "runtime": "java",
    "runtime_version": ">=17",
    "port": 9001
  }
}
```

- Portunix runs the plugin as `java <jvm_args> -jar <install_path>/<binary> [args...]`
  with the plugin install directory as working directory; service mode adds
  `--port <port>`
- `jvm_args` is optional; without it Portunix passes `-Xmx256m -Xms64m`
- `runtime_version` is optional; `portunix plugin install` and `portunix plugin check`
  compare it with `java -version`. A missing Java runtime is reported with the fix
  `portunix install java-21`
- Rename the JAR (`artifactId` / `version` in `pom.xml`) and `binary` together

See `template/plugin.json` for the complete manifest.

## Troubleshooting

1. **`Unimplemented` on `Initialize`** — `plugin.proto` was changed; restore the
   `portunix.plugins` package and the `PluginService` definition from the Portunix
   contract.
2. **`portunix plugin stop` fails with `EOF`** — the plugin exits inside `shutdown`;
   answer the call and leave stopping to SIGTERM and the shutdown hook.
3. **`mvn test` runs no tests** — the `maven-surefire-plugin` 3.2.5 pin was removed from
   `pom.xml`.
4. **`protoc` fails on Apple Silicon / Linux ARM64** — ensure `os-maven-plugin` produces
   the right classifier; for arm64 add `-Dos.detected.classifier=linux-aarch_64`.
5. **`UnsatisfiedLinkError` for `grpc-netty-shaded`** — use `grpc-netty-shaded` (not
   `grpc-netty`) to avoid native netty conflicts.

## Next Steps

- Study the [template code](template/) for a complete example
- Learn about [MCP integration](../../mcp-integration/exposing-tools.md) for AI agents
- Review the [plugin architecture](../../architecture.md)

## Resources

- [gRPC Java](https://grpc.io/docs/languages/java/)
- [Protocol Buffers Java Tutorial](https://protobuf.dev/getting-started/javatutorial/)
- [protobuf-maven-plugin](https://www.xolstice.org/protobuf-maven-plugin/)
