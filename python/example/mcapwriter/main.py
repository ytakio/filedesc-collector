#!/usr/bin/env python3
"""mcapwriter is a small example that writes an MCAP file using the Foxglove
SDK and filedesc-collector.

This example mirrors the Go mcapwriter example: it writes ``/timestamp``
(``google.protobuf.Timestamp``) and ``/sensor/dummy`` (``foxglove.SceneUpdate``)
channels.

For ``/timestamp`` we build a protobuf schema from a collected
``FileDescriptorSet`` and log serialized ``Timestamp`` messages through a
generic ``foxglove.Channel``.

For ``/sensor/dummy`` we use the Foxglove SDK's built-in
``SceneUpdateChannel`` and ``SceneUpdate`` message class, so no manual schema
or serialization is required.
"""

from __future__ import annotations

from datetime import datetime, timedelta, timezone

import foxglove
from foxglove import Channel, Schema
from foxglove.channels import SceneUpdateChannel
from foxglove.messages import (
    Color,
    CubePrimitive,
    Pose,
    Quaternion,
    SceneEntity,
    SceneUpdate,
    Vector3,
)
from google.protobuf.descriptor_pb2 import FileDescriptorSet
from google.protobuf.timestamp_pb2 import Timestamp

from filedesc import FileDescriptorSetCollector


def build_timestamp_schema() -> Schema:
    """Build a Foxglove protobuf schema for ``google.protobuf.Timestamp``.

    The schema data is a serialized ``FileDescriptorSet`` collected by
    filedesc-collector from the protobuf descriptor.
    """
    collector = FileDescriptorSetCollector()
    protos = collector.collect_for_message(Timestamp.DESCRIPTOR)
    fd_set = FileDescriptorSet(file=protos)
    return Schema(
        name=Timestamp.DESCRIPTOR.full_name,
        encoding="protobuf",
        data=fd_set.SerializeToString(),
    )


def write_mcap(output_path: str) -> None:
    """Write a small MCAP file with sample messages."""
    timestamp_channel = Channel(
        topic="/timestamp",
        message_encoding="protobuf",
        schema=build_timestamp_schema(),
    )
    scene_channel = SceneUpdateChannel("/sensor/dummy")

    with foxglove.open_mcap(output_path, allow_overwrite=True):
        start = datetime(2026, 10, 22, 9, 0, 0, tzinfo=timezone.utc)
        for i in range(100):
            t = start + timedelta(seconds=i)
            log_time_ns = int(t.timestamp() * 1e9)

            timestamp = Timestamp(
                seconds=int(t.timestamp()),
                nanos=t.microsecond * 1000,
            )
            timestamp_channel.log(timestamp.SerializeToString(), log_time=log_time_ns)

            scene_channel.log(
                SceneUpdate(
                    entities=[
                        SceneEntity(
                            id=f"entity-{i:02d}",
                            frame_id="base_link",
                            cubes=[
                                CubePrimitive(
                                    pose=Pose(
                                        position=Vector3(x=float(i), y=0.0, z=0.0),
                                        orientation=Quaternion(
                                            x=0.0, y=0.0, z=0.0, w=1.0
                                        ),
                                    ),
                                    size=Vector3(x=1.0, y=1.0, z=1.0),
                                    color=Color(r=1.0, g=0.5, b=0.0, a=1.0),
                                )
                            ],
                        )
                    ]
                ),
                log_time=log_time_ns,
            )


def main() -> int:
    output_path = "output.mcap"
    write_mcap(output_path)
    print(f"Wrote {output_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
