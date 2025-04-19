# Project State: Minecraft Network

## Overview

This project is a Minecraft network setup using Kubernetes and Helm. It includes a proxy server and a lobby server, both of which are containerized and managed using Kubernetes. The project also integrates with NATS for messaging and uses JetStream for data streaming.

## Directory Structure

```
minecraft-network/
├── charts/
│   └── network/
│       ├── templates/
│       │   ├── _helpers.tpl
│       │   ├── lobby-server.yaml
│       │   ├── proxy-configmap.yaml
│       │   ├── proxy-secret.yaml
│       │   └── proxy.yaml
│       ├── Chart.lock
│       ├── Chart.yaml
│       ├── values.yaml
│       └── .helmignore
├── scripts/
│   ├── chart-upgrade.sh
│   ├── kind-config.yaml.in
│   ├── kind-setup.sh
│   └── test.sh
├── servers/
│   ├── lobby_minestom/
│   │   ├── build.gradle.kts
│   │   ├── build-and-load.sh
│   │   ├── compose.yaml
│   │   ├── Dockerfile
│   │   ├── gradle/
│   │   │   └── wrapper/
│   │   │       └── gradle-wrapper.properties
│   │   ├── gradlew
│   │   ├── gradlew.bat
│   │   ├── settings.gradle.kts
│   │   ├── src/
│   │   │   └── main/
│   │   │       └── java/
│   │   │           └── fr/
│   │   │               └── bafbi/
│   │   │                   └── lobby/
│   │   │                       └── Main.java
│   │   ├── .dockerignore
│   │   ├── .gitignore
│   │   └── README.md
│   └── proxy_gate/
│       ├── .github/
│       │   └── workflows/
│       │       └── workflow.yml
│       ├── plugins/
│       │   ├── bossbar/
│       │   │   └── bossbar.go
│       │   ├── globalchat/
│       │   │   └── globalchat.go
│       │   ├── network/
│       │   │   └── network.go
│       │   ├── ping/
│       │   │   └── ping.go
│       │   ├── tablist/
│       │   │   └── tablist.go
│       │   └── titlecmd/
│       │       └── titlecmd.go
│       ├── util/
│       │   ├── mini/
│       │   │   └── mini.go
│       │   └── util.go
│       ├── build-and-load.sh
│       ├── config.yml
│       ├── Dockerfile
│       ├── go.mod
│       ├── go.sum
│       ├── LICENSE
│       ├── Makefile
│       ├── README.md
│       └── renovate.json
├── values/
│   └── dev-values.yaml
└── README.md
```

## Components

### Helm Chart

- **Chart.yaml**: Defines the Helm chart for the network.
- **values.yaml**: Default values for the Helm chart.
- **templates/**: Contains Kubernetes resource templates for the lobby and proxy servers.
  - **_helpers.tpl**: Helper template for common functions.
  - **lobby-server.yaml**: StatefulSet and Service for the lobby server.
  - **proxy.yaml**: Deployment and Service for the proxy server.
  - **proxy-configmap.yaml**: ConfigMap for the proxy server configuration.
  - **proxy-secret.yaml**: Secret for the proxy server.

### Scripts

- **chart-upgrade.sh**: Script to upgrade the Helm chart.
- **kind-config.yaml.in**: Template for the kind cluster configuration.
- **kind-setup.sh**: Script to set up a kind cluster and deploy the Helm chart.
- **test.sh**: Script to run tests.

### Servers

#### Lobby Server (lobby_minestom)

- **Dockerfile**: Dockerfile to build the lobby server.
- **build.gradle.kts**: Gradle build script.
- **build-and-load.sh**: Script to build and load the Docker image into the kind cluster.
- **compose.yaml**: Docker Compose file for local development.
- **src/main/java/fr/bafbi/lobby/Main.java**: Main class for the lobby server.
- **gradle/wrapper/**: Gradle wrapper files.

#### Proxy Server (proxy_gate)

- **Dockerfile**: Dockerfile to build the proxy server.
- **build-and-load.sh**: Script to build and load the Docker image into the kind cluster.
- **config.yml**: Configuration file for the proxy server.
- **go.mod**: Go module file.
- **go.sum**: Go dependencies file.
- **Makefile**: Makefile for building and testing the proxy server.
- **README.md**: Documentation for the proxy server.
- **renovate.json**: Configuration for Renovate dependency updates.
- **plugins/**: Directory containing various plugins for the proxy server.
  - **bossbar**: Plugin to display a boss bar to players.
  - **globalchat**: Plugin to broadcast chat messages to all players.
  - **network**: Plugin to handle dynamic lobby registration.
  - **ping**: Plugin to handle ping events.
  - **tablist**: Plugin to set a custom header and footer in the tab list.
  - **titlecmd**: Plugin to handle title commands.
- **util/**: Utility functions for the proxy server.
  - **mini**: Utilities for parsing and manipulating Minecraft text colors and styles.

### Values

- **dev-values.yaml**: Development-specific values for the Helm chart.

## Dependencies

- **NATS**: Used for messaging between the proxy and lobby servers.
- **JetStream**: Used for data streaming.
- **Minestom**: Used for the lobby server.
- **Minekube Gate**: Used for the proxy server.

## Setup and Deployment

1. **Set up the kind cluster**:
   ```sh
   ./scripts/kind-setup.sh
   ```

2. **Build and load Docker images**:
   - For the lobby server:
     ```sh
     ./servers/lobby_minestom/build-and-load.sh
     ```
   - For the proxy server:
     ```sh
     ./servers/proxy_gate/build-and-load.sh
     ```

3. **Deploy the Helm chart**:
   ```sh
   ./scripts/chart-upgrade.sh
   ```

## Testing

Run the test script:
```sh
./scripts/test.sh
```

## Future Work

- Implement more advanced features and plugins for the proxy server.
- Improve the CI/CD pipeline.
- Add more comprehensive tests.
- Enhance the documentation.
