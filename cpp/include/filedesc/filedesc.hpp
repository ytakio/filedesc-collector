#pragma once

#include <google/protobuf/descriptor.h>
#include <google/protobuf/descriptor.pb.h>

#include <optional>
#include <string_view>
#include <unordered_set>
#include <vector>

namespace filedesc {

// Collects FileDescriptorProto messages including transitive dependencies.
// This class is header-only so it can be used without compiling a separate
// library.
//
// A collector can be reused for multiple independent collections; its internal
// state is reset at the beginning of each collection. Passing a null descriptor
// returns std::nullopt.
class FileDescriptorSetCollector final {
 public:
  using Proto = google::protobuf::FileDescriptorProto;
  using Protos = std::vector<Proto>;
  using OptionalProtos = std::optional<Protos>;

  FileDescriptorSetCollector() = default;

  // The collector is stateless; default copy/move are safe and convenient.
  FileDescriptorSetCollector(const FileDescriptorSetCollector&) = default;
  FileDescriptorSetCollector& operator=(const FileDescriptorSetCollector&) =
      default;
  FileDescriptorSetCollector(FileDescriptorSetCollector&&) = default;
  FileDescriptorSetCollector& operator=(FileDescriptorSetCollector&&) = default;

  // Return FileDescriptorProtos for the file containing ``descriptor`` and its
  // deps. Returns std::nullopt if ``descriptor`` is null.
  [[nodiscard]]
  auto CollectForMessage(const google::protobuf::Descriptor* descriptor) const
      -> OptionalProtos {
    if (!descriptor) {
      return std::nullopt;
    }

    return CollectForDescriptor(descriptor->file());
  }

  // Return FileDescriptorProtos for ``file_descriptor`` and its deps.
  // Returns std::nullopt if ``file_descriptor`` is null.
  [[nodiscard]]
  auto CollectForDescriptor(
      const google::protobuf::FileDescriptor* file_descriptor) const
      -> OptionalProtos {
    if (!file_descriptor) {
      return std::nullopt;
    }

    // FileDescriptor names are owned by the protobuf descriptor pool, which
    // outlives this function call (and typically the program), so string_view
    // keys are safe here.
    std::unordered_set<std::string_view> seen;

    auto collectFile =
        [&seen](const auto& self,
                const google::protobuf::FileDescriptor* fd) -> OptionalProtos {
      if (!fd) {
        return std::nullopt;
      }

      // Detect cycles and insert the file name in a single hash lookup.
      // emplace() constructs the string_view in place, avoiding a temporary.
      const bool inserted = seen.emplace(fd->name()).second;
      if (!inserted) {
        return Protos{};
      }

      // Recursively collect dependencies first so the result is topologically
      // ordered.
      Protos result;
      for (int i = 0; i < fd->dependency_count(); ++i) {
        auto deps = self(self, fd->dependency(i));
        if (!deps.has_value()) {
          return std::nullopt;
        }
        result.insert(result.end(), deps->begin(), deps->end());
      }

      // Append this file's descriptor proto.
      Proto proto;
      fd->CopyTo(&proto);
      result.push_back(std::move(proto));

      return result;
    };

    return collectFile(collectFile, file_descriptor);
  }
};

}  // namespace filedesc
