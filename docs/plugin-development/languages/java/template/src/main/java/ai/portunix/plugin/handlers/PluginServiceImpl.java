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
