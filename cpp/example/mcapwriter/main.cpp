// mcapwriter is a small example that writes an MCAP file using the
// filedesc-collector to supply protobuf FileDescriptorProtos and the Foxglove
// MCAP C++ library (https://github.com/foxglove/mcap) to write the file.

#define MCAP_IMPLEMENTATION
#include <google/protobuf/descriptor.pb.h>
#include <google/protobuf/message.h>
#include <google/protobuf/timestamp.pb.h>

#include <chrono>
#include <cstdint>
#include <cstdio>
#include <functional>
#include <iostream>
#include <mcap/writer.hpp>
#include <memory>
#include <optional>
#include <string>
#include <vector>

#include "filedesc/filedesc.hpp"
#include "foxglove/SceneUpdate.pb.h"

namespace gp = google::protobuf;

namespace {

// channelConfig describes one channel to write: its topic, the protobuf
// descriptor used for its schema, and a populate function that builds each
// sample.
struct ChannelConfig {
  std::string topic;
  const gp::Descriptor* descriptor;
  std::function<std::unique_ptr<gp::Message>(
      const std::chrono::system_clock::time_point& t, int i)>
      populate;
};

// channelConfigs is the table of channels produced by this example.
// Adding a new channel only requires appending a new entry here.
const std::vector<ChannelConfig> channel_configs = {
    {
        "/timestamp",
        gp::Timestamp::descriptor(),
        [](const std::chrono::system_clock::time_point& t, int /*i*/) {
          auto timestamp = std::make_unique<gp::Timestamp>();
          const auto since_epoch = t.time_since_epoch();
          const auto seconds =
              std::chrono::duration_cast<std::chrono::seconds>(since_epoch);
          timestamp->set_seconds(seconds.count());
          timestamp->set_nanos(static_cast<int32_t>(
              std::chrono::duration_cast<std::chrono::nanoseconds>(since_epoch -
                                                                   seconds)
                  .count()));
          return timestamp;
        },
    },
    {
        "/sensor/dummy",
        foxglove::SceneUpdate::descriptor(),
        [](const std::chrono::system_clock::time_point& /*t*/, int i) {
          auto scene_update = std::make_unique<foxglove::SceneUpdate>();
          auto* entity = scene_update->add_entities();

          char entity_id[32];
          std::snprintf(entity_id, sizeof(entity_id), "entity-%02d", i);
          entity->set_id(entity_id);

          entity->set_frame_id("base_link");

          auto* cube = entity->add_cubes();
          auto* pose = cube->mutable_pose();

          auto* position = pose->mutable_position();
          position->set_x(static_cast<double>(i));
          position->set_y(0.0);
          position->set_z(0.0);

          auto* orientation = pose->mutable_orientation();
          orientation->set_x(0.0);
          orientation->set_y(0.0);
          orientation->set_z(0.0);
          orientation->set_w(1.0);

          auto* size = cube->mutable_size();
          size->set_x(1.0);
          size->set_y(1.0);
          size->set_z(1.0);

          auto* color = cube->mutable_color();
          color->set_r(1.0);
          color->set_g(0.5);
          color->set_b(0.0);
          color->set_a(1.0);

          return scene_update;
        },
    },
};

// Create an MCAP schema from a protobuf descriptor, assigning the given ID.
// Returns std::nullopt on failure.
std::optional<mcap::Schema> CreateSchema(uint16_t id,
                                         const gp::Descriptor* descriptor) {
  filedesc::FileDescriptorSetCollector collector;
  auto protos = collector.CollectForMessage(descriptor);
  if (!protos.has_value()) {
    return std::nullopt;
  }

  gp::FileDescriptorSet fd_set{};
  for (auto& proto : protos.value()) {
    *fd_set.add_file() = std::move(proto);
  }

  std::string data{};
  if (!fd_set.SerializeToString(&data)) {
    return std::nullopt;
  }

  mcap::Schema schema(descriptor->full_name(), "protobuf", std::move(data));
  schema.id = id;
  return schema;
}

// MessageWithData keeps the serialized protobuf bytes alive while the
// contained mcap::Message borrows them.
struct MessageWithData {
  mcap::Message message;
  std::string data;
};

// Create an MCAP message on ``channel_id`` with the given sequence and log
// time. The serialized bytes are owned by the returned MessageWithData so the
// caller does not need to manage a separate buffer.
// Returns std::nullopt on failure.
std::optional<MessageWithData> CreateMessage(uint16_t channel_id,
                                             uint32_t sequence,
                                             uint64_t log_time_ns,
                                             const gp::Message& msg) {
  MessageWithData result{};
  if (!msg.SerializeToString(&result.data)) {
    return std::nullopt;
  }

  result.message.channelId = channel_id;
  result.message.sequence = sequence;
  result.message.logTime = log_time_ns;
  result.message.publishTime = log_time_ns;
  result.message.data = reinterpret_cast<const std::byte*>(result.data.data());
  result.message.dataSize = result.data.size();
  return result;
}

// Run writes the MCAP file. Returns std::nullopt on success or an error
// message on failure.
std::optional<std::string> Run() {
  // Open an MCAP file for writing.
  mcap::McapWriter writer;
  mcap::McapWriterOptions options("protobuf");
  options.library = "filedesc-collector mcapwriter example";
  options.chunkSize = mcap::DefaultChunkSize;
  auto open_status = writer.open("output.mcap", options);
  if (!open_status.ok()) {
    return "create mcap writer: " + std::string(open_status.message);
  }

  // Assign IDs dynamically instead of hardcoding them.
  uint16_t next_schema_id = 1;
  uint16_t next_channel_id = 0;

  // Register schemas and channels from the channel_configs table.
  std::vector<uint16_t> channel_ids;
  channel_ids.reserve(channel_configs.size());
  for (const auto& cfg : channel_configs) {
    auto schema = CreateSchema(next_schema_id, cfg.descriptor);
    if (!schema.has_value()) {
      return "create schema for " + cfg.topic;
    }
    writer.addSchema(schema.value());
    next_schema_id++;

    mcap::Channel channel(cfg.topic, "protobuf", schema.value().id);
    channel.id = next_channel_id;
    channel.metadata["source"] = "mcapwriter example";
    writer.addChannel(channel);
    channel_ids.push_back(channel.id);
    next_channel_id++;
  }

  // Use the current time as the start epoch.
  const auto start_time = std::chrono::system_clock::now();

  // Write 100 samples at 1 Hz.
  for (int i = 0; i < 100; ++i) {
    const auto t = start_time + std::chrono::seconds(i);
    const auto since_epoch = t.time_since_epoch();
    const uint64_t log_time_ns = static_cast<uint64_t>(
        std::chrono::duration_cast<std::chrono::nanoseconds>(since_epoch)
            .count());

    for (size_t j = 0; j < channel_configs.size(); ++j) {
      const auto& cfg = channel_configs[j];
      auto msg = cfg.populate(t, i);
      auto m = CreateMessage(channel_ids[j], static_cast<uint32_t>(i + 1),
                             log_time_ns, *msg);
      if (!m.has_value()) {
        return "create message " + std::to_string(i) + " on " + cfg.topic;
      }
      auto status = writer.write(m.value().message);
      if (!status.ok()) {
        return "write message " + std::to_string(i) + " on " + cfg.topic +
               ": " + std::string(status.message);
      }
    }
  }

  // The MCAP writer's close() returns void; errors during close are
  // reported by the destructor.
  writer.close();
  return std::nullopt;
}

}  // namespace

int main() {
  if (auto err = Run(); err.has_value()) {
    std::cerr << "error: " << err.value() << "\n";
    return 1;
  }
  std::cout << "Wrote output.mcap\n";
  return 0;
}
