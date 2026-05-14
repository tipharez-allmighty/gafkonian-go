# Gafkonian Go - Kafka Clone in Go

## Kafka clone in go
This is a pseudo-replica of Kafka written in Go, inspired by the CodeCrafters challenges. I finished the project in a bit of a rush as I recently started a new job that requires my full attention.

Unfortunately, most of my time was spent parsing binaries rather than implementing advanced Kafka features; however, the project does include some basic core functionality. I built this primarily as a training exercise to practice while following the book Learning Go.

## Features

Supports a subset of the Kafka protocol:
- **Produce API (v0)**: Send records to a specific topic and partition.
- **Fetch API (v16)**: Retrieve records from a specific topic and partition.
- **API Versions (v18)**: Query supported API versions.
- **Describe Topic Partitions (v75)**: Retrieve metadata about topic partitions.

## Prerequisites

- **Go**: version 1.26 or higher.
- **Netcat (nc)**: For testing and interacting with the broker via raw TCP payloads.
- **xxd** and **hexdump**: For converting and inspecting binary data.

## Setup and Installation

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/gafkonian-go/gafkonian-go.git
    cd gafkonian-go
    ```

2.  **Configure environment variables**:
    Copy the example environment file and adjust values as needed:
    ```bash
    cp .env.example .env
    ```

3.  **Install dependencies**:
    ```bash
    go mod download
    ```

## Running the Program

To start the broker, run:

```bash
go run cmd/main.go
```

By default, the server will listen on `0.0.0.0:9092`.

> **Note:** On startup, the broker initializes its partition logs based on the cluster metadata. **Existing records in the partition logs are cleared every time the server starts.**

### Metadata Requirement

The broker requires a valid cluster metadata log to start. This project comes with a sample metadata log located at `tmp/__cluster_metadata/00000000000000000000.log`.

## Interacting with the Broker (Test Payloads)

Use these commands to test the broker's functionality:

**Produce Request (Message Ingestion):**
```bash
echo -n "00000043000000090000000100000000000000000000000000000003666f6f000000000000000100000000000000000000000000000000000000036b65790000000576616c7565" | xxd -r -p | nc localhost 9092
```

**Fetch Request (Message Consumption):**
```bash
echo -n "00000027001000000000000100000000000000000000000000000003666f6f000000000000000000000000" | xxd -r -p | nc localhost 9092
```

**Describe Topic Partitions (Metadata Inquiry):**
```bash
echo -n "00000020004b00000000000700096b61666b612d636c69000204666f6f0000000064ff00" | xxd -r -p | nc localhost 9092 | hexdump -C
```

**API Versions (Protocol Negotiation):**
```bash
echo -n "00000023001200040183eb7d00096b61666b612d636c69000a6b61666b612d636c6904302e3100" | xxd -r -p | nc localhost 9092 | xxd -g 1 -u
```

## Configuration Options

| Variable | Description | Default |
| :--- | :--- | :--- |
| `PORT` | The port the broker listens on. | `9092` |
| `TIMEOUTSECONDS` | TCP connection deadline in seconds. | `10` |
| `METADATA_LOG` | Path to the cluster metadata log file. | `tmp/__cluster_metadata/...` |
| `RECORDS_LOG` | Directory for partition record logs. | `tmp/partitions` |
