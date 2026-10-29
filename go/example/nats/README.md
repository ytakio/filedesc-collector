# Go NATS filedesc-collector example

This example demonstrates distributing protobuf schemas via NATS KV and
writing received raw protobuf messages to an MCAP file.

## Architecture

- `docker-compose.yml` starts a NATS server with JetStream enabled.
- `sender` publishes:
  - individual `FileDescriptorProto` messages to the `SCHEMAS` KV bucket,
    keyed by proto file path (e.g., `foxglove/SceneUpdate.proto`).
  - a JSON value to the `SUBJECTS` KV bucket keyed by NATS subject name
    (e.g., `sensor.dummy`). The JSON has the form:

    ```json
    {
      "schema": {
        "path": "foxglove/SceneUpdate.proto",
        "full_name": "foxglove.SceneUpdate"
      }
    }
    ```

  - serialized protobuf messages to NATS subjects (e.g., `sensor.dummy`).
- `receiver` starts first and uses NATS KV watchers to wait for each
  subject's info entry in `SUBJECTS`. For every subject it resolves the
  `FileDescriptorProto` chain from `SCHEMAS` (including transitive
  dependencies), registers the corresponding MCAP schema and channel, and
  subscribes to the subject. Received raw protobuf messages are written to
  `output.mcap`.

## Run

From this directory (`go/example/nats/`):

```bash
make example
```

Or step by step:

```bash
make up        # start NATS
make receiver  # run receiver in one terminal (waits for schemas via KV watch)
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
