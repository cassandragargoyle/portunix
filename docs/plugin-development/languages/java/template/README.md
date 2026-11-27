# Java Plugin Template

Template for Java-based Portunix plugins using Maven, gRPC, and Protocol Buffers.

## Quick Start

```bash
# Build the plugin (generates proto classes, runs tests, builds the shaded JAR)
mvn clean package

# Run as a gRPC service
java -jar target/my-java-plugin-1.0.0.jar --port 50051 --config src/main/resources/config.yaml

# Run a command once (CLI mode)
java -jar target/my-java-plugin-1.0.0.jar hello Portunix

# Install into Portunix from this directory (build the JAR first)
portunix plugin install .
portunix plugin enable java-plugin-template
portunix java-plugin-template hello Portunix
portunix plugin start java-plugin-template
```

The manifest uses `"runtime": "java"` with `"binary": "target/my-java-plugin-1.0.0.jar"`;
Portunix runs it as `java -jar target/my-java-plugin-1.0.0.jar` from the install
directory. The plugin implements the Portunix contract `portunix.plugins.PluginService`
(`src/main/proto/plugin.proto`); keep that file unchanged.

## Structure

```text
.
├── plugin.json                 # Plugin manifest consumed by Portunix
├── pom.xml                     # Maven build descriptor
└── src/
    ├── main/
    │   ├── java/ai/portunix/plugin/
    │   │   ├── Main.java                       # Entry point — CLI command or gRPC server
    │   │   ├── config/PluginConfig.java        # YAML config loader
    │   │   ├── manifest/PluginManifest.java    # plugin.json -> PluginInfo / PluginCommand
    │   │   ├── handlers/PluginServiceImpl.java # gRPC method handlers
    │   │   └── services/PluginBusinessService.java  # Business logic
    │   ├── proto/plugin.proto                  # Copy of the Portunix plugin contract
    │   └── resources/config.yaml               # Default runtime configuration
    └── test/
        └── java/ai/portunix/plugin/
            ├── PluginBusinessServiceTest.java
            └── PluginServiceImplTest.java
```

See `../getting-started.md` for the full walkthrough.
