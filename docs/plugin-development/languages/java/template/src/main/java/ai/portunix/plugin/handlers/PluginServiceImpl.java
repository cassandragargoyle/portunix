package ai.portunix.plugin.handlers;

import ai.portunix.plugin.config.PluginConfig;
import ai.portunix.plugin.proto.PluginProto.ExecuteRequest;
import ai.portunix.plugin.proto.PluginProto.ExecuteResponse;
import ai.portunix.plugin.proto.PluginProto.GetInfoRequest;
import ai.portunix.plugin.proto.PluginProto.GetInfoResponse;
import ai.portunix.plugin.proto.PluginProto.HealthCheckRequest;
import ai.portunix.plugin.proto.PluginProto.HealthCheckResponse;
import ai.portunix.plugin.proto.PluginProto.PluginInfo;
import ai.portunix.plugin.proto.PluginProto.ShutdownRequest;
import ai.portunix.plugin.proto.PluginProto.ShutdownResponse;
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
        try {
            String result = service.execute(request.getCommand(), request.getArgsList());
            responseObserver.onNext(ExecuteResponse.newBuilder()
                .setResult(result)
                .setStatus(ExecuteResponse.Status.SUCCESS)
                .build());
        } catch (IllegalArgumentException e) {
            responseObserver.onNext(ExecuteResponse.newBuilder()
                .setStatus(ExecuteResponse.Status.ERROR)
                .setErrorMessage(e.getMessage())
                .build());
        }
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
