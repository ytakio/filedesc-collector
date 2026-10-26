// Tests for filedesc::FileDescriptorSetCollector.
//
// These tests verify that the collector returns the target file descriptor and
// its transitive dependencies in topological order, deduplicates files within a
// single collection, and rejects null input.
#include <google/protobuf/descriptor.pb.h>
#include <google/protobuf/duration.pb.h>
#include <google/protobuf/timestamp.pb.h>
#include <gtest/gtest.h>

#include "filedesc/filedesc.hpp"

using google::protobuf::Descriptor;
using google::protobuf::Duration;
using google::protobuf::FileDescriptor;
using google::protobuf::FileDescriptorProto;
using google::protobuf::FileDescriptorSet;
using google::protobuf::Timestamp;

namespace filedesc {

static auto FileNames(const std::vector<FileDescriptorProto>& protos)
    -> std::vector<std::string> {
  auto names = std::vector<std::string>{};
  names.reserve(protos.size());
  for (const auto& p : protos) {
    names.push_back(p.name());
  }
  return names;
}

TEST(FileDescriptorSetCollector, CollectForMessage) {
  auto collector = FileDescriptorSetCollector{};
  auto protos = collector.CollectForMessage(Timestamp::descriptor());
  ASSERT_TRUE(protos.has_value());
  EXPECT_FALSE(protos->empty());
  EXPECT_EQ(FileNames(*protos).back(), "google/protobuf/timestamp.proto");
}

TEST(FileDescriptorSetCollector, CollectForMessageWithDependency) {
  auto collector = FileDescriptorSetCollector{};
  auto protos = collector.CollectForMessage(Duration::descriptor());
  ASSERT_TRUE(protos.has_value());
  auto names = FileNames(*protos);
  EXPECT_TRUE(std::find(names.begin(), names.end(),
                        "google/protobuf/duration.proto") != names.end());
}

TEST(FileDescriptorSetCollector, CollectForDescriptor) {
  auto collector = FileDescriptorSetCollector{};
  auto protos =
      collector.CollectForDescriptor(FileDescriptorSet::descriptor()->file());
  ASSERT_TRUE(protos.has_value());
  EXPECT_FALSE(protos->empty());
}

TEST(FileDescriptorSetCollector, DoesNotDuplicate) {
  auto collector = FileDescriptorSetCollector{};
  auto first = collector.CollectForMessage(Timestamp::descriptor());
  ASSERT_TRUE(first.has_value());
  ASSERT_EQ(first->size(), 1);
  EXPECT_EQ(first->front().name(), "google/protobuf/timestamp.proto");

  auto second = collector.CollectForMessage(Duration::descriptor());
  ASSERT_TRUE(second.has_value());
  ASSERT_EQ(second->size(), 1);
  EXPECT_EQ(second->front().name(), "google/protobuf/duration.proto");
}

TEST(FileDescriptorSetCollector, RejectsNullMessage) {
  auto collector = FileDescriptorSetCollector{};
  auto protos =
      collector.CollectForMessage(static_cast<const Descriptor*>(nullptr));
  EXPECT_FALSE(protos.has_value());
}

TEST(FileDescriptorSetCollector, RejectsNullDescriptor) {
  auto collector = FileDescriptorSetCollector{};
  auto protos = collector.CollectForDescriptor(
      static_cast<const FileDescriptor*>(nullptr));
  EXPECT_FALSE(protos.has_value());
}

}  // namespace filedesc
