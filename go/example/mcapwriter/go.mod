module github.com/ytakio/filedesc-collector/go/example/mcapwriter

go 1.27

require (
	github.com/foxglove/mcap/go/mcap v1.5.0
	github.com/ytakio/filedesc-collector v0.0.0
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.16.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/ytakio/filedesc-collector => ../../..

tool (
	google.golang.org/protobuf/cmd/protoc-gen-go
)
