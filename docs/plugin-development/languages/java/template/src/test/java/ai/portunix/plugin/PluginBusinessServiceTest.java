package ai.portunix.plugin;

import ai.portunix.plugin.config.PluginConfig;
import ai.portunix.plugin.services.PluginBusinessService;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;

class PluginBusinessServiceTest {

    @Test
    void helloReturnsDefaultGreeting() {
        PluginBusinessService service = new PluginBusinessService(new PluginConfig());
        assertEquals("Hello, World!", service.execute("hello", List.of()));
    }

    @Test
    void helloReturnsCustomGreeting() {
        PluginBusinessService service = new PluginBusinessService(new PluginConfig());
        assertEquals("Hello, Portunix!", service.execute("hello", List.of("Portunix")));
    }

    @Test
    void unknownCommandThrows() {
        PluginBusinessService service = new PluginBusinessService(new PluginConfig());
        assertThrows(IllegalArgumentException.class,
            () -> service.execute("unknown", List.of()));
    }

    @Test
    void processRequiresArguments() {
        PluginBusinessService service = new PluginBusinessService(new PluginConfig());
        assertThrows(IllegalArgumentException.class,
            () -> service.execute("process", List.of()));
    }
}
