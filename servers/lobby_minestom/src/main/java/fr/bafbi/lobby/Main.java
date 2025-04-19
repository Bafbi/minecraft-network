package fr.bafbi.lobby;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import io.nats.client.*;
import io.nats.client.api.KeyValueConfiguration;
import io.vavr.control.Option;
import io.vavr.control.Try;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.Map;
import net.minestom.server.MinecraftServer;
import net.minestom.server.coordinate.Pos;
import net.minestom.server.entity.Player;
import net.minestom.server.event.GlobalEventHandler;
import net.minestom.server.event.player.AsyncPlayerConfigurationEvent;
import net.minestom.server.extras.bungee.BungeeCordProxy;
import net.minestom.server.extras.velocity.VelocityProxy;
import net.minestom.server.instance.*;
import net.minestom.server.instance.block.Block;
import net.minestom.server.network.ConnectionState;
import net.minestom.server.network.packet.client.handshake.ClientHandshakePacket;
import net.minestom.server.timer.Scheduler;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class Main {

    static Option<Connection> nats;
    static Option<JetStream> js;
    static Option<KeyValueManagement> kvm;
    static Option<KeyValue> kv;
    static final Gson gson = new GsonBuilder().create();

    public static void main(String[] args) {
        Logger logger = LoggerFactory.getLogger(Main.class);

        // init the nats
        nats = Try.of(() -> Nats.connect("nats://network-nats:4222"))
            .onFailure(throwable -> {
                logger.warn("Failed to connect to NATS server", throwable);
            })
            .onSuccess(connection -> {
                logger.info("Connected to NATS server");
            })
            .toOption();

        // init the jetstream
        js = nats.flatMap(connection ->
            Try.of(connection::jetStream)
                .onFailure(throwable -> {
                    logger.warn("Failed to connect to JetStream", throwable);
                })
                .onSuccess(jetStream -> {
                    logger.info("Connected to JetStream");
                })
                .toOption()
        );

        // init the key value management
        kvm = nats.flatMap(nats ->
            Try.of(nats::keyValueManagement)
                .onFailure(throwable -> {
                    logger.warn(
                        "Failed to connect to KeyValueManagement",
                        throwable
                    );
                })
                .onSuccess(keyValueManagement -> {
                    logger.info("Connected to KeyValueManagement");
                })
                .toOption()
        );

        kvm.peek(keyValueManagement -> {
            // create the key value store for servers
            Try.of(keyValueManagement::getBucketNames)
                .flatMap(buckets -> {
                    if (buckets.contains("servers")) {
                        return Try.success(null);
                    } else {
                        return Try.of(() -> {
                            KeyValueConfiguration config =
                                KeyValueConfiguration.builder()
                                    .name("servers")
                                    .build();
                            keyValueManagement.create(config);
                            return null;
                        });
                    }
                })
                .onFailure(throwable -> {
                    logger.warn(
                        "Failed to create servers KeyValue store",
                        throwable
                    );
                })
                .onSuccess(keyValue -> {
                    logger.info("Created servers KeyValue store");
                });
        });

        // init the servers key value
        kv = nats.flatMap(nats ->
            Try.of(() -> nats.keyValue("servers"))
                .onFailure(throwable -> {
                    logger.warn("Failed to connect to KeyValue", throwable);
                })
                .onSuccess(keyValue -> {
                    logger.info("Connected to KeyValue");
                })
                .toOption()
        );

        // Initialization
        MinecraftServer minecraftServer = MinecraftServer.init();

        // use the env variable PROXY_SECRET
        assert System.getenv("PROXY_SECRET") != null;
        VelocityProxy.enable(System.getenv("PROXY_SECRET"));

//        BungeeCordProxy.enable();

        // Create the instance
        InstanceManager instanceManager = MinecraftServer.getInstanceManager();
        InstanceContainer instanceContainer =
            instanceManager.createInstanceContainer();

        // Set the ChunkGenerator
        instanceContainer.setGenerator(unit ->
            unit.modifier().fillHeight(0, 40, Block.GRASS_BLOCK)
        );

        // Add an event callback to specify the spawning instance (and the spawn position)
        GlobalEventHandler globalEventHandler =
            MinecraftServer.getGlobalEventHandler();
        MinecraftServer.getPacketListenerManager().setListener(ConnectionState.HANDSHAKE, ClientHandshakePacket.class,
            (packet, connection) -> {
                // Log the handshake packet
                System.out.println("Received handshake packet: " + packet);
                // log the connection
                System.out.println("Received connection: " + connection);

            }
        );
        globalEventHandler.addListener(
            AsyncPlayerConfigurationEvent.class,
            event -> {
                final Player player = event.getPlayer();
                event.setSpawningInstance(instanceContainer);
                player.setRespawnPoint(new Pos(0, 42, 0));
            }
        );

        Scheduler scheduler = MinecraftServer.getSchedulerManager();
        scheduler.scheduleNextTick(() -> {
            System.out.println("Registering server...");
            // Register the server with the KV store
            kv.peek(keyValue -> {
                String podName = System.getenv("POD_NAME");
                String namespace = System.getenv("NAMESPACE");
                String headlessService = System.getenv("HEADLESS_SERVICE");
                String serverAddress =
                    podName +
                    "." +
                    headlessService +
                    "." +
                    namespace +
                    ".svc.cluster.local:25565";

                // Read labels and annotations from downward API
                Map<String, String> labels = readMetadataFile(
                    "/etc/podinfo/labels"
                );
                Map<String, String> annotations = readMetadataFile(
                    "/etc/podinfo/annotations"
                );

                // Create server info object
                Metadata meta = new Metadata();
                meta.labels = labels;
                meta.annotations = annotations;
                meta.annotations.put("server/address", serverAddress);

                // Convert to JSON
                String json = gson.toJson(meta);

                Try.of(() -> keyValue.put(podName, json))
                    .onFailure(throwable -> {
                        logger.warn(
                            "Failed to publish server info to NATS",
                            throwable
                        );
                    })
                    .onSuccess(natsMessage -> {
                        logger.info("Published server info to NATS: {}", json);
                    });
            });
        });

        // Add a shutdown hook to remove the server from the NATS server
        Runtime.getRuntime()
            .addShutdownHook(
                new Thread(() -> {
                    kv.peek(kv -> {
                        String podName = System.getenv("POD_NAME");

                        Try.run(() -> kv.delete(podName))
                            .onFailure(throwable -> {
                                logger.warn(
                                    "Failed to delete server from NATS",
                                    throwable
                                );
                            })
                            .onSuccess(natsMessage -> {
                                logger.info("Deleted server from NATS");
                            });
                    });
                })
            );

        // Start the server on port 25565
        minecraftServer.start("0.0.0.0", 25565);
    }

    // Server info class that matches our Go ServerInfo struct
    private static class Metadata {
        Map<String, String> labels;
        Map<String, String> annotations;
    }

    // Helper method to read metadata files from downward API
    private static Map<String, String> readMetadataFile(String path) {
        Map<String, String> metadata = new HashMap<>();
        Path filePath = Paths.get(path);

        try {
            if (Files.exists(filePath)) {
                String content = new String(Files.readAllBytes(filePath));
                String[] lines = content.split("\n");
                for (String line : lines) {
                    if (line.contains("=")) {
                        String[] parts = line.split("=", 2);
                        if (parts.length == 2) {
                            // Remove surrounding quotes if present
                            String key =
                                parts[0].trim().replaceAll("^\"|\"$", "");
                            String value =
                                parts[1].trim().replaceAll("^\"|\"$", "");
                            metadata.put(key, value);
                        }
                    }
                }
            }
        } catch (IOException e) {
            LoggerFactory.getLogger(Main.class).error(
                "Failed to read metadata file: {}",
                path,
                e
            );
        }

        return metadata;
    }
}
