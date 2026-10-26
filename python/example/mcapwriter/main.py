#!/usr/bin/env python3
"""mcapwriter is a small example that writes an MCAP file using the
filedesc-collector to supply protobuf FileDescriptorProtos and the MCAP Python
library to write the file.

This mirrors the Go go/example/mcapwriter example.
"""

from __future__ import annotations

from datetime import datetime, timedelta, timezone

from google.protobuf.descriptor_pb2 import FileDescriptorSet
from google.protobuf.timestamp_pb2 import Timestamp
from mcap.writer import Writer
from mcap.well_known import MessageEncoding, SchemaEncoding

from filedesc import FileDescriptorSetCollector
from filedesc.proto.foxglove.Color_pb2 import Color
from filedesc.proto.foxglove.CubePrimitive_pb2 import CubePrimitive
from filedesc.proto.foxglove.Pose_pb2 import Pose
from filedesc.proto.foxglove.Quaternion_pb2 import Quaternion
from filedesc.proto.foxglove.SceneEntity_pb2 import SceneEntity
from filedesc.proto.foxglove.SceneUpdate_pb2 import SceneUpdate
from filedesc.proto.foxglove.Vector3_pb2 import Vector3


def register_message_schema(writer: Writer, message_descriptor) -> int:
    """Collect a FileDescriptorSet for the message and register it as a schema.

    Returns the schema ID assigned by the MCAP writer.
    """
    collector = FileDescriptorSetCollector()
    fd_protos = collector.collect_for_message(message_descriptor)
    fd_set = FileDescriptorSet(file=fd_protos)
    return writer.register_schema(
        name=message_descriptor.full_name,
        encoding=SchemaEncoding.Protobuf,
        data=fd_set.SerializeToString(),
    )


def write_mcap(output_path: str) -> None:
    """Write a small MCAP file with sample messages."""
    writer = Writer(output_path)
    try:
        writer.start(
            profile="protobuf",
            library="github.com/ytakio/filedesc-collector/mcapwriter example",
        )

        timestamp_schema_id = register_message_schema(writer, Timestamp.DESCRIPTOR)
        timestamp_channel_id = writer.register_channel(
            topic="/timestamp",
            message_encoding=MessageEncoding.Protobuf,
            schema_id=timestamp_schema_id,
            metadata={"source": "mcapwriter example"},
        )

        scene_schema_id = register_message_schema(writer, SceneUpdate.DESCRIPTOR)
        scene_channel_id = writer.register_channel(
            topic="/sensor/dummy",
            message_encoding=MessageEncoding.Protobuf,
            schema_id=scene_schema_id,
            metadata={"source": "mcapwriter example"},
        )

        start = datetime(2026, 10, 22, 9, 0, 0, tzinfo=timezone.utc)
        for i in range(100):
            t = start + timedelta(seconds=i)
            log_time_ns = int(t.timestamp() * 1e9)

            timestamp = Timestamp(
                seconds=int(t.timestamp()),
                nanos=t.microsecond * 1000,
            )
            writer.add_message(
                channel_id=timestamp_channel_id,
                log_time=log_time_ns,
                publish_time=log_time_ns,
                sequence=i + 1,
                data=timestamp.SerializeToString(),
            )

            scene = SceneUpdate(
                entities=[
                    SceneEntity(
                        id=f"entity-{i:02d}",
                        frame_id="base_link",
                        cubes=[
                            CubePrimitive(
                                pose=Pose(
                                    position=Vector3(x=float(i), y=0.0, z=0.0),
                                    orientation=Quaternion(x=0.0, y=0.0, z=0.0, w=1.0),
                                ),
                                size=Vector3(x=1.0, y=1.0, z=1.0),
                                color=Color(r=1.0, g=0.5, b=0.0, a=1.0),
                            )
                        ],
                    )
                ]
            )
            writer.add_message(
                channel_id=scene_channel_id,
                log_time=log_time_ns,
                publish_time=log_time_ns,
                sequence=i + 1,
                data=scene.SerializeToString(),
            )
    finally:
        writer.finish()


def main() -> int:
    output_path = "output.mcap"
    write_mcap(output_path)
    print(f"Wrote {output_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
