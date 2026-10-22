// mcapwriter is a small example that writes an MCAP file using the
// filedesc-collector to supply protobuf FileDescriptorProtos and the Foxglove
// SDK C++ library to write the file.

#include <foxglove/channel.hpp>
#include <foxglove/error.hpp>
#include <foxglove/foxglove.hpp>
#include <foxglove/mcap.hpp>
#include <foxglove/messages.hpp>
#include <foxglove/schema.hpp>
#include <google/protobuf/descriptor.pb.h>
#include <google/protobuf/timestamp.pb.h>

#include <chrono>
#include <cstdint>
#include <iostream>
#include <string>

#include "filedesc/filedesc.hpp"

namespace gp = google::protobuf;

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
    foxglove::McapWriterOptions mcap_options = {};
    mcap_options.path = "output.mcap";
    auto writer_result = foxglove::McapWriter::create(mcap_options);
    if (!writer_result.has_value()) {
      std::cerr << "Failed to create writer: "
                << foxglove::strerror(writer_result.error()) << '\n';
      return 1;
    }
    auto writer = std::move(writer_result.value());

    // Create a generic protobuf channel for google.protobuf.Timestamp using a
    // FileDescriptorSet collected by filedesc-collector as its schema.
    const gp::Descriptor *timestamp_descriptor = gp::Timestamp::descriptor();
    std::string timestamp_fdset =
        BuildFileDescriptorSet(timestamp_descriptor);

    foxglove::Schema timestamp_schema;
    timestamp_schema.encoding = "protobuf";
    timestamp_schema.name = timestamp_descriptor->full_name();
    timestamp_schema.data =
        reinterpret_cast<const std::byte *>(timestamp_fdset.data());
    timestamp_schema.data_len = timestamp_fdset.size();

    auto timestamp_channel_result =
        foxglove::RawChannel::create("/timestamp", "protobuf",
                                     std::move(timestamp_schema));
    if (!timestamp_channel_result.has_value()) {
      std::cerr << "Failed to create timestamp channel: "
                << foxglove::strerror(timestamp_channel_result.error())
                << '\n';
      return 1;
    }
    auto timestamp_channel = std::move(timestamp_channel_result.value());

    // Create a SceneUpdate channel using the Foxglove SDK's built-in message
    // type. The SDK handles protobuf serialization and schema registration.
    auto scene_channel_result =
        foxglove::messages::SceneUpdateChannel::create("/sensor/dummy");
    if (!scene_channel_result.has_value()) {
      std::cerr << "Failed to create scene channel: "
                << foxglove::strerror(scene_channel_result.error()) << '\n';
      return 1;
    }
    auto scene_channel = std::move(scene_channel_result.value());

    // Use a fixed epoch: 2026-10-22 09:00:00 UTC.
    const auto start_time = std::chrono::system_clock::time_point(
        std::chrono::seconds(1'796'625'600));

    for (int i = 0; i < 100; ++i) {
      const auto t = start_time + std::chrono::seconds(i);
      const auto since_epoch = t.time_since_epoch();
      const auto seconds =
          std::chrono::duration_cast<std::chrono::seconds>(since_epoch);

      // Log a google.protobuf.Timestamp sample.
      gp::Timestamp timestamp;
      timestamp.set_seconds(seconds.count());
      timestamp.set_nanos(static_cast<int32_t>(
          std::chrono::duration_cast<std::chrono::nanoseconds>(
              since_epoch - seconds)
              .count()));
      std::string timestamp_data = timestamp.SerializeAsString();
      const uint64_t log_time_ns = static_cast<uint64_t>(
          std::chrono::duration_cast<std::chrono::nanoseconds>(since_epoch)
              .count());

      timestamp_channel.log(
          reinterpret_cast<const std::byte *>(timestamp_data.data()),
          timestamp_data.size(), log_time_ns);

      // Log a foxglove.SceneUpdate sample.
      foxglove::messages::SceneUpdate scene_update;
      auto &entity = scene_update.entities.emplace_back();
      entity.id = "entity-" + std::to_string(i);
      entity.frame_id = "base_link";

      auto &cube = entity.cubes.emplace_back();
      cube.pose.emplace();
      cube.pose->position = foxglove::messages::Vector3{
          static_cast<double>(i), 0.0, 0.0};
      cube.pose->orientation = foxglove::messages::Quaternion{0.0, 0.0, 0.0, 1.0};
      cube.size = foxglove::messages::Vector3{1.0, 1.0, 1.0};
      cube.color = foxglove::messages::Color{1.0, 0.5, 0.0, 1.0};

      scene_channel.log(scene_update, log_time_ns);
    }

    std::cout << "Wrote output.mcap\n";
    return 0;
  } catch (const std::exception &e) {
    std::cerr << "error: " << e.what() << "\n";
    return 1;
  }
}

