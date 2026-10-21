#pragma once

#include <google/protobuf/descriptor.h>
#include <google/protobuf/descriptor.pb.h>

#include <stdexcept>
#include <string_view>
#include <unordered_set>
#include <vector>

namespace filedesc {

// Collects FileDescriptorProto messages including transitive dependencies.
// This class is header-only so it can be used without compiling a separate
// library.
class FileDescriptorSetCollector {
public:
  FileDescriptorSetCollector() = default;

  // Return FileDescriptorProtos for the file containing ``descriptor`` and its
  // deps.
  [[nodiscard]] std::vector<google::protobuf::FileDescriptorProto>
  CollectForMessage(const google::protobuf::Descriptor *descriptor) {
    if (descriptor == nullptr) {
      throw std::invalid_argument("descriptor must not be null");
    }
    return CollectForDescriptor(descriptor->file());
  }

  // Return FileDescriptorProtos for ``file_descriptor`` and its deps.
  [[nodiscard]] std::vector<google::protobuf::FileDescriptorProto>
  CollectForDescriptor(
      const google::protobuf::FileDescriptor *file_descriptor) {
    if (file_descriptor == nullptr) {
      throw std::invalid_argument("file_descriptor must not be null");
    }
    std::vector<google::protobuf::FileDescriptorProto> result;
    CollectFile(file_descriptor, result);
    return result;
  }

private:
  void CollectFile(const google::protobuf::FileDescriptor *file_descriptor,
                   std::vector<google::protobuf::FileDescriptorProto> &result) {
    if (file_descriptor == nullptr) {
      return;
    }

    const std::string_view name = file_descriptor->name();
    if (!seen_.insert(name).second) {
      return;
    }

    // Recursively collect dependencies first.
    for (int i = 0; i < file_descriptor->dependency_count(); ++i) {
      CollectFile(file_descriptor->dependency(i), result);
    }
    for (int i = 0; i < file_descriptor->public_dependency_count(); ++i) {
      CollectFile(file_descriptor->public_dependency(i), result);
    }
    for (int i = 0; i < file_descriptor->weak_dependency_count(); ++i) {
      CollectFile(file_descriptor->weak_dependency(i), result);
    }

    // Append this file's descriptor proto.
    google::protobuf::FileDescriptorProto proto;
    file_descriptor->CopyTo(&proto);
    result.push_back(std::move(proto));
  }

  // FileDescriptor names are owned by the protobuf descriptor pool, which has
  // program lifetime, so string_view keys are safe.
  std::unordered_set<std::string_view> seen_;
};

// Convenience function to collect FileDescriptorProtos for a message type.
inline std::vector<google::protobuf::FileDescriptorProto>
CollectForMessage(const google::protobuf::Descriptor *descriptor) {
  return FileDescriptorSetCollector().CollectForMessage(descriptor);
}

// Convenience function to collect FileDescriptorProtos for a file descriptor.
inline std::vector<google::protobuf::FileDescriptorProto>
CollectForDescriptor(const google::protobuf::FileDescriptor *file_descriptor) {
  return FileDescriptorSetCollector().CollectForDescriptor(file_descriptor);
}

} // namespace filedesc
