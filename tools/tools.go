// +build tools

package tools

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"
	//_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-openapiv2" // todo remove v2
	//_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
