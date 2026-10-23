// mcapwriter is a small example that writes an MCAP file using the
// filedesc-collector to supply protobuf FileDescriptorProtos and the Foxglove
// MCAP C++ library (https://github.com/foxglove/mcap) to write the file.

#define MCAP_IMPLEMENTATION
#include <mcap/writer.hpp>

#include <google/protobuf/descriptor.pb.h>
#include <google/protobuf/timestamp.pb.h>

#include "filedesc/filedesc.hpp"
#include "foxglove/SceneUpdate.pb.h"

#include <chrono>
#include <cstdint>
#include <iostream>
#include <string>

namespace gp = google::protobuf;

namespace {

// Helpers that build and return small foxglove protobuf messages by value.
// They can be assigned directly into a mutable sub-message slot.
foxglove::Vector3 MakeVector3(double x, double y, double z) {
  foxglove::Vector3 v;
  v.set_x(x);
  v.set_y(y);
  v.set_z(z);
  return v;
}

foxglove::Quaternion MakeQuaternion(double x, double y, double z, double w) {
  foxglove::Quaternion q;
  q.set_x(x);
  q.set_y(y);
  q.set_z(z);
  q.set_w(w);
  return q;
}

foxglove::Color MakeColor(double r, double g, double b, double a) {
  foxglove::Color c;
  c.set_r(r);
  c.set_g(g);
  c.set_b(b);
  c.set_a(a);
  return c;
}

} // namespace

// Build a serialized FileDescriptorSet for ``descriptor`` using
// filedesc-collector.
std::string BuildFileDescriptorSet(const gp::Descriptor *descriptor) {
  filedesc::FileDescriptorSetCollector collector;
  auto protos = collector.CollectForMessage(descriptor);
  if (!protos.has_value()) {
    throw std::runtime_error("failed to collect file descriptors");
  }

  gp::FileDescriptorSet fd_set;
  for (auto &proto : protos.value()) {
    *fd_set.add_file() = std::move(proto);
  }
  return fd_set.SerializeAsString();
}

int main() {
  try {
    // Open an MCAP file for writing.
    mcap::McapWriter writer;
    mcap::McapWriterOptions options("protobuf");
    options.library = "filedesc-collector mcapwriter example";
    auto open_status = writer.open("output.mcap", options);
    if (!open_status.ok()) {
      std::cerr << "Failed to open writer: " << open_status.message << "\n";
      return 1;
    }

    // Register a schema for google.protobuf.Timestamp using a FileDescriptorSet
    // collected by filedesc-collector.
    const gp::Descriptor *timestamp_descriptor = gp::Timestamp::descriptor();
    std::string timestamp_fdset = BuildFileDescriptorSet(timestamp_descriptor);

    mcap::Schema timestamp_schema(timestamp_descriptor->full_name(), "protobuf",
                                  timestamp_fdset);
    writer.addSchema(timestamp_schema);

    mcap::Channel timestamp_channel("/timestamp", "protobuf",
                                    timestamp_schema.id);
    writer.addChannel(timestamp_channel);

    // Register a schema for foxglove.SceneUpdate in the same way.
    const gp::Descriptor *scene_descriptor =
        foxglove::SceneUpdate::descriptor();
    std::string scene_fdset = BuildFileDescriptorSet(scene_descriptor);

    mcap::Schema scene_schema(scene_descriptor->full_name(), "protobuf",
                              scene_fdset);
    writer.addSchema(scene_schema);

    mcap::Channel scene_channel("/sensor/dummy", "protobuf", scene_schema.id);
    writer.addChannel(scene_channel);

    // Use a fixed epoch: 2026-10-22 09:00:00 UTC.
    const auto start_time = std::chrono::system_clock::time_point(
        std::chrono::seconds(1'796'625'600));

    for (int i = 0; i < 100; ++i) {
      const auto t = start_time + std::chrono::seconds(i);
      const auto since_epoch = t.time_since_epoch();
      const auto seconds =
          std::chrono::duration_cast<std::chrono::seconds>(since_epoch);
      const uint64_t log_time_ns = static_cast<uint64_t>(
          std::chrono::duration_cast<std::chrono::nanoseconds>(since_epoch)
              .count());

      // Log a google.protobuf.Timestamp sample.
      gp::Timestamp timestamp;
      timestamp.set_seconds(seconds.count());
      timestamp.set_nanos(static_cast<int32_t>(
          std::chrono::duration_cast<std::chrono::nanoseconds>(since_epoch -
                                                               seconds)
              .count()));
      std::string timestamp_data = timestamp.SerializeAsString();

      mcap::Message timestamp_msg;
      timestamp_msg.channelId = timestamp_channel.id;
      timestamp_msg.sequence = static_cast<uint32_t>(i + 1);
      timestamp_msg.logTime = log_time_ns;
      timestamp_msg.publishTime = log_time_ns;
      timestamp_msg.data =
          reinterpret_cast<const std::byte *>(timestamp_data.data());
      timestamp_msg.dataSize = timestamp_data.size();

      auto timestamp_status = writer.write(timestamp_msg);
      if (!timestamp_status.ok()) {
        std::cerr << "Failed to write timestamp message: "
                  << timestamp_status.message << "\n";
        writer.terminate();
        return 1;
      }

      // Log a foxglove.SceneUpdate sample.
      foxglove::SceneUpdate scene_update;
      auto *entity = scene_update.add_entities();
      entity->set_id("entity-" + std::to_string(i));
      entity->set_frame_id("base_link");

      auto *cube = entity->add_cubes();
      auto *pose = cube->mutable_pose();
      *pose->mutable_position() = MakeVector3(static_cast<double>(i), 0.0, 0.0);
      *pose->mutable_orientation() = MakeQuaternion(0.0, 0.0, 0.0, 1.0);
      *cube->mutable_size() = MakeVector3(1.0, 1.0, 1.0);
      *cube->mutable_color() = MakeColor(1.0, 0.5, 0.0, 1.0);

      std::string scene_data = scene_update.SerializeAsString();

      mcap::Message scene_msg;
      scene_msg.channelId = scene_channel.id;
      scene_msg.sequence = static_cast<uint32_t>(i + 1);
      scene_msg.logTime = log_time_ns;
      scene_msg.publishTime = log_time_ns;
      scene_msg.data = reinterpret_cast<const std::byte *>(scene_data.data());
      scene_msg.dataSize = scene_data.size();

      auto scene_status = writer.write(scene_msg);
      if (!scene_status.ok()) {
        std::cerr << "Failed to write scene message: " << scene_status.message
                  << "\n";
        writer.terminate();
        return 1;
      }
    }

    writer.close();
    std::cout << "Wrote output.mcap\n";
    return 0;
  } catch (const std::exception &e) {
    std::cerr << "error: " << e.what() << "\n";
    return 1;
  }
}
