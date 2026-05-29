# Java Plugin Template

Template for Java-based Portunix plugins using Maven, gRPC, and Protocol Buffers.

## Quick Start

```bash
# Build the plugin (generates proto + shaded JAR)
mvn clean package

# Run locally
java -jar target/my-java-plugin-1.0.0.jar --port 50051 --config src/main/resources/config.yaml

# Install into Portunix
portunix plugin install target/my-java-plugin-1.0.0.jar
portunix plugin enable my-java-plugin
portunix plugin start my-java-plugin
```

## Structure

```text
.
├── plugin.json                 # Plugin manifest consumed by Portunix
├── pom.xml                     # Maven build descriptor
└── src/
    ├── main/
    │   ├── java/ai/portunix/plugin/
    │   │   ├── Main.java
    │   │   ├── config/PluginConfig.java
    │   │   ├── handlers/PluginServiceImpl.java
    │   │   └── services/PluginBusinessService.java
    │   ├── proto/plugin.proto
    │   └── resources/config.yaml
    └── test/
        └── java/ai/portunix/plugin/PluginBusinessServiceTest.java
```

See `../getting-started.md` for the full walkthrough.
