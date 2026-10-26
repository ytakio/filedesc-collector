from collections.abc import Callable
from dataclasses import dataclass
from datetime import datetime, timedelta
from pathlib import Path

from google.protobuf.descriptor import Descriptor
from google.protobuf.descriptor_pb2 import FileDescriptorSet
from google.protobuf.message import Message
from google.protobuf.timestamp_pb2 import Timestamp

from proto.foxglove.Color_pb2 import Color
from proto.foxglove.CubePrimitive_pb2 import CubePrimitive
from proto.foxglove.Pose_pb2 import Pose
from proto.foxglove.Quaternion_pb2 import Quaternion
from proto.foxglove.SceneEntity_pb2 import SceneEntity
from proto.foxglove.SceneUpdate_pb2 import SceneUpdate
from proto.foxglove.Vector3_pb2 import Vector3

from mcap.well_known import MessageEncoding, SchemaEncoding
from mcap.writer import Writer
from filedesc import FileDescriptorSetCollector

LIBRARY_NAME = "github.com/ytakio/filedesc-collector/mcapwriter example"
SOURCE_METADATA = {"source": "mcapwriter example"}
MESSAGE_COUNT = 100
START_TIME = datetime.fromisoformat("2026-10-22T09:00:00+00:00")


@dataclass(frozen=True, slots=True)
class ChannelConfig:
    """Configuration for one MCAP channel."""

    topic: str
    descriptor: Descriptor
    populate: Callable[[datetime, int], Message]


def build_schema_data(descriptor: Descriptor) -> bytes:
    """Return a serialized FileDescriptorSet for *descriptor* and its deps."""
    collector = FileDescriptorSetCollector()
    fd_set = FileDescriptorSet(
        file=collector.collect_for_message(descriptor),
    )

    return fd_set.SerializeToString()


def build_timestamp(t: datetime, _index: int) -> Timestamp:
    """Build a google.protobuf.Timestamp sample."""
    timestamp = Timestamp()
    timestamp.FromDatetime(t)
    return timestamp


def build_scene_update(_t: datetime, index: int) -> SceneUpdate:
    """Build a foxglove.SceneUpdate sample."""
    return SceneUpdate(
        entities=[
            SceneEntity(
                id=f"entity-{index:02d}",
                frame_id="base_link",
                cubes=[
                    CubePrimitive(
                        pose=Pose(
                            position=Vector3(x=float(index), y=0.0, z=0.0),
                            orientation=Quaternion(x=0.0, y=0.0, z=0.0, w=1.0),
                        ),
                        size=Vector3(x=1.0, y=1.0, z=1.0),
                        color=Color(r=1.0, g=0.5, b=0.0, a=1.0),
                    )
                ],
            )
        ]
    )


CHANNEL_CONFIGS = [
    ChannelConfig(
        topic="/timestamp",
        descriptor=Timestamp.DESCRIPTOR,
        populate=build_timestamp,
    ),
    ChannelConfig(
        topic="/sensor/dummy",
        descriptor=SceneUpdate.DESCRIPTOR,
        populate=build_scene_update,
    ),
]


def write_mcap(output_path: Path) -> None:
    """Write a small MCAP file with sample messages."""
    with output_path.open("wb") as stream:
        writer = Writer(stream)
        writer.start(profile="protobuf", library=LIBRARY_NAME)

        channel_ids: list[int] = []
        for cfg in CHANNEL_CONFIGS:
            schema_id = writer.register_schema(
                name=cfg.descriptor.full_name,
                encoding=SchemaEncoding.Protobuf,
                data=build_schema_data(cfg.descriptor),
            )
            channel_id = writer.register_channel(
                topic=cfg.topic,
                message_encoding=MessageEncoding.Protobuf,
                schema_id=schema_id,
                metadata=SOURCE_METADATA,
            )
            channel_ids.append(channel_id)

        for i in range(MESSAGE_COUNT):
            t = START_TIME + timedelta(seconds=i)
            log_time_ns = int(t.timestamp() * 1e9)
            sequence = i + 1

            for channel_id, cfg in zip(channel_ids, CHANNEL_CONFIGS):
                writer.add_message(
                    channel_id=channel_id,
                    log_time=log_time_ns,
                    publish_time=log_time_ns,
                    sequence=sequence,
                    data=cfg.populate(t, i).SerializeToString(),
                )

        writer.finish()


def main() -> int:
    output_path = Path(__file__).with_name("output.mcap")
    write_mcap(output_path)
    print(f"Wrote {output_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
