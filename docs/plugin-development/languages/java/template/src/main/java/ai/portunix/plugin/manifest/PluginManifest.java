package ai.portunix.plugin.manifest;

import ai.portunix.plugin.proto.PluginProto.PluginCapabilities;
import ai.portunix.plugin.proto.PluginProto.PluginCommand;
import ai.portunix.plugin.proto.PluginProto.PluginInfo;
import ai.portunix.plugin.proto.PluginProto.PluginParameter;
import ai.portunix.plugin.proto.PluginProto.PluginPermissions;
import com.google.gson.Gson;
import com.google.gson.annotations.SerializedName;

import java.io.IOException;
import java.io.Reader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

/**
 * Part of plugin.json the plugin reports to Portunix; plugin.json is the
 * single source of the plugin information and commands
 */
public class PluginManifest {
    String name = "";
    String version = "";
    String description = "";
    String author = "";
    String license = "";
    List<Command> commands = List.of();
    Dependencies dependencies = new Dependencies();
    @SerializedName("ai_integration")
    AiIntegration aiIntegration = new AiIntegration();
    Permissions permissions = new Permissions();

    static class Command {
        String name = "";
        String description = "";
        List<String> subcommands = List.of();
        List<Parameter> parameters = List.of();
        List<String> examples = List.of();
    }

    static class Parameter {
        String name = "";
        String type = "string";
        String description = "";
        boolean required;
        @SerializedName("default_value")
        String defaultValue = "";
    }

    static class Dependencies {
        @SerializedName("os_support")
        List<String> osSupport = List.of();
    }

    static class AiIntegration {
        @SerializedName("mcp_tools")
        List<NamedItem> mcpTools = List.of();
    }

    static class NamedItem {
        String name = "";
    }

    static class Permissions {
        List<String> filesystem = List.of();
        List<String> network = List.of();
        List<String> database = List.of();
        List<String> system = List.of();
        String level = "limited";
    }

    public static PluginManifest load(Path path) throws IOException {
        try (Reader reader = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            PluginManifest manifest = new Gson().fromJson(reader, PluginManifest.class);
            if (manifest == null || manifest.name == null || manifest.name.isEmpty()) {
                throw new IOException("manifest " + path + " has no plugin name");
            }
            return manifest;
        }
    }

    public String getName() {
        return name;
    }

    public String getVersion() {
        return orEmpty(version);
    }

    public PluginInfo toPluginInfo() {
        return PluginInfo.newBuilder()
            .setName(name)
            .setVersion(orEmpty(version))
            .setDescription(orEmpty(description))
            .setAuthor(orEmpty(author))
            .setLicense(orEmpty(license))
            .addAllSupportedOs(orEmpty(dependencies == null ? null : dependencies.osSupport))
            .addAllCommands(toPluginCommands())
            .setCapabilities(toCapabilities())
            .setRequiredPermissions(toPermissions())
            .build();
    }

    public List<PluginCommand> toPluginCommands() {
        List<PluginCommand> result = new ArrayList<>();
        for (Command command : orEmpty(commands)) {
            PluginCommand.Builder builder = PluginCommand.newBuilder()
                .setName(orEmpty(command.name))
                .setDescription(orEmpty(command.description))
                .addAllSubcommands(orEmpty(command.subcommands))
                .addAllExamples(orEmpty(command.examples));
            for (Parameter parameter : orEmpty(command.parameters)) {
                builder.addParameters(PluginParameter.newBuilder()
                    .setName(orEmpty(parameter.name))
                    .setType(parameter.type == null || parameter.type.isEmpty() ? "string" : parameter.type)
                    .setDescription(orEmpty(parameter.description))
                    .setRequired(parameter.required)
                    .setDefaultValue(orEmpty(parameter.defaultValue)));
            }
            result.add(builder.build());
        }
        return result;
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

    private PluginPermissions toPermissions() {
        Permissions p = permissions == null ? new Permissions() : permissions;
        return PluginPermissions.newBuilder()
            .addAllFilesystem(orEmpty(p.filesystem))
            .addAllNetwork(orEmpty(p.network))
            .addAllDatabase(orEmpty(p.database))
            .addAllSystem(orEmpty(p.system))
            .setLevel(p.level == null || p.level.isEmpty() ? "limited" : p.level)
            .build();
    }

    private static String orEmpty(String value) {
        return value == null ? "" : value;
    }

    private static <T> List<T> orEmpty(List<T> value) {
        return value == null ? List.of() : value;
    }
}
