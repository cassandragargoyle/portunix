package ai.portunix.plugin;

import ai.portunix.plugin.config.PluginConfig;
import ai.portunix.plugin.handlers.PluginServiceImpl;
import ai.portunix.plugin.manifest.PluginManifest;
import ai.portunix.plugin.proto.PluginProto.ExecuteRequest;
import ai.portunix.plugin.proto.PluginProto.ExecuteResponse;
import ai.portunix.plugin.proto.PluginProto.HealthRequest;
import ai.portunix.plugin.proto.PluginProto.HealthResponse;
import ai.portunix.plugin.proto.PluginProto.InitializeRequest;
import ai.portunix.plugin.proto.PluginProto.InitializeResponse;
import ai.portunix.plugin.proto.PluginProto.ListCommandsRequest;
import ai.portunix.plugin.proto.PluginProto.ListCommandsResponse;
import ai.portunix.plugin.proto.PluginProto.ShutdownRequest;
import ai.portunix.plugin.proto.PluginProto.ShutdownResponse;
import io.grpc.stub.StreamObserver;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.nio.file.Path;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

class PluginServiceImplTest {
    private PluginServiceImpl handler;

    // Collects the single response of a unary call
    static class Capture<T> implements StreamObserver<T> {
        T value;
        boolean completed;

        @Override
        public void onNext(T value) {
            this.value = value;
        }

        @Override
        public void onError(Throwable t) {
            throw new AssertionError("unexpected gRPC error", t);
        }

        @Override
        public void onCompleted() {
            completed = true;
        }
    }

    @BeforeEach
    void setUp() throws Exception {
        // Maven runs tests from the template root, where plugin.json is
        handler = new PluginServiceImpl(new PluginConfig(), PluginManifest.load(Path.of("plugin.json")));
    }

    @Test
    void initializeReturnsManifestInfo() {
        Capture<InitializeResponse> response = new Capture<>();
        handler.initialize(InitializeRequest.newBuilder().setPluginName("java-plugin-template").build(), response);
        assertTrue(response.completed);
        assertTrue(response.value.getSuccess());
        assertEquals("java-plugin-template", response.value.getPluginInfo().getName());
        assertEquals("hello", response.value.getPluginInfo().getCommands(0).getName());
    }

    @Test
    void healthIsHealthy() {
        Capture<HealthResponse> response = new Capture<>();
        handler.health(HealthRequest.getDefaultInstance(), response);
        assertTrue(response.value.getHealthy());
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

    @Test
    void listCommandsAndShutdown() {
        Capture<ListCommandsResponse> list = new Capture<>();
        handler.listCommands(ListCommandsRequest.getDefaultInstance(), list);
        assertFalse(list.value.getCommandsList().isEmpty());

        Capture<ShutdownResponse> shutdown = new Capture<>();
        handler.shutdown(ShutdownRequest.getDefaultInstance(), shutdown);
        assertTrue(shutdown.value.getSuccess());
    }
}
