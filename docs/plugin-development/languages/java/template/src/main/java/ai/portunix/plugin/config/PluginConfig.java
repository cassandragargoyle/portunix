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

    public String getName() {
        return name;
    }

    public String getVersion() {
        return version;
    }

    public String getDescription() {
        return description;
    }

    public String getLogLevel() {
        return logLevel;
    }
}
