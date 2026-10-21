#include <gtest/gtest.h>

#include <google/protobuf/descriptor.pb.h>
#include <google/protobuf/duration.pb.h>
#include <google/protobuf/timestamp.pb.h>

#include "filedesc/filedesc.hpp"

using google::protobuf::Descriptor;
using google::protobuf::Duration;
using google::protobuf::FileDescriptor;
using google::protobuf::FileDescriptorProto;
using google::protobuf::FileDescriptorSet;
using google::protobuf::Timestamp;

namespace filedesc {

static std::vector<std::string>
FileNames(const std::vector<FileDescriptorProto> &protos) {
  std::vector<std::string> names;
  for (const auto &p : protos) {
    names.push_back(p.name());
  }
  return names;
}

TEST(FileDescriptorSetCollector, CollectForMessage) {
  auto protos = CollectForMessage(Timestamp::descriptor());
  EXPECT_FALSE(protos.empty());
  auto names = FileNames(protos);
  EXPECT_EQ(names.back(), "google/protobuf/timestamp.proto");
}

TEST(FileDescriptorSetCollector, CollectForMessageWithDependency) {
  auto protos = CollectForMessage(Duration::descriptor());
  auto names = FileNames(protos);
  EXPECT_TRUE(std::find(names.begin(), names.end(),
                        "google/protobuf/duration.proto") != names.end());
}

TEST(FileDescriptorSetCollector, CollectForDescriptor) {
  auto protos = CollectForDescriptor(FileDescriptorSet::descriptor()->file());
  EXPECT_FALSE(protos.empty());
}

TEST(FileDescriptorSetCollector, DoesNotDuplicate) {
  FileDescriptorSetCollector collector;
  auto first = collector.CollectForMessage(Timestamp::descriptor());
  ASSERT_EQ(first.size(), 1);
  EXPECT_EQ(first[0].name(), "google/protobuf/timestamp.proto");

  auto second = collector.CollectForMessage(Duration::descriptor());
  ASSERT_EQ(second.size(), 1);
  EXPECT_EQ(second[0].name(), "google/protobuf/duration.proto");
}

TEST(FileDescriptorSetCollector, RejectsNullMessage) {
  EXPECT_THROW(CollectForMessage(static_cast<const Descriptor *>(nullptr)),
               std::invalid_argument);
}

TEST(FileDescriptorSetCollector, RejectsNullDescriptor) {
  EXPECT_THROW(
      CollectForDescriptor(static_cast<const FileDescriptor *>(nullptr)),
      std::invalid_argument);
}

} // namespace filedesc
