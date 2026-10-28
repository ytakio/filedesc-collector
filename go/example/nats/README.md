# Go NATS filedesc-collector example

This example demonstrates distributing protobuf schemas via NATS KV and
parsing received messages with dynamic descriptors before writing them to an
MCAP file.

## Architecture

- `docker-compose.yml` starts a NATS server with JetStream enabled.
- `sender` publishes:
  - individual `FileDescriptorProto` messages to the `SCHEMAS` KV bucket,
    keyed by proto file path (e.g., `foxglove/SceneUpdate.proto`).
  - a JSON value to the `SUBJECTS` KV bucket keyed by NATS subject name
    (e.g., `sensor.dummy`). The JSON has the form:
    ```json
    {"schema": {"path": "foxglove/SceneUpdate.proto", "full_name": "foxglove.SceneUpdate"}}
    ```
  - serialized protobuf messages to NATS subjects derived from MCAP topics by
    replacing `/` with `.` (e.g., `/sensor/dummy` → `sensor.dummy`).
- `receiver` subscribes to fixed NATS subjects, looks up the proto path and
  message full name from `TOPICS`, fetches the individual
  `FileDescriptorProto` entries from `SCHEMAS`, recursively resolves
  dependencies to reconstruct a `FileDescriptorSet`, builds a dynamic
  descriptor, parses each message, and writes it to `output.mcap`.

## Run

From this directory (`go/example/nats/`):

```bash
make example
```

Or step by step:

```bash
make up        # start NATS
make receiver  # run receiver in one terminal
make sender    # run sender in another terminal
make down      # stop NATS
```

The full example can also be started from the `go/` or `go/example/`
directories:

```bash
cd ../..
make nats-example
```

## Output

The receiver writes `go/example/nats/output.mcap`.

## Clean

```bash
make clean
```

