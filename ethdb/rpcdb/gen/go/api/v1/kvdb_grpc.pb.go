// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: kvdb.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	KV_Get_FullMethodName             = "/ethereum.rpcdb.api.v1.KV/Get"
	KV_Batch_FullMethodName           = "/ethereum.rpcdb.api.v1.KV/Batch"
	KV_NewIterator_FullMethodName     = "/ethereum.rpcdb.api.v1.KV/NewIterator"
	KV_IteratorError_FullMethodName   = "/ethereum.rpcdb.api.v1.KV/IteratorError"
	KV_IteratorKey_FullMethodName     = "/ethereum.rpcdb.api.v1.KV/IteratorKey"
	KV_IteratorNext_FullMethodName    = "/ethereum.rpcdb.api.v1.KV/IteratorNext"
	KV_IteratorRelease_FullMethodName = "/ethereum.rpcdb.api.v1.KV/IteratorRelease"
	KV_IteratorValue_FullMethodName   = "/ethereum.rpcdb.api.v1.KV/IteratorValue"
)

// KVClient is the client API for KV service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KVClient interface {
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	Batch(ctx context.Context, in *BatchRequest, opts ...grpc.CallOption) (*BatchResponse, error)
	NewIterator(ctx context.Context, in *NewIteratorRequest, opts ...grpc.CallOption) (*NewIteratorResponse, error)
	IteratorError(ctx context.Context, in *IteratorErrorRequest, opts ...grpc.CallOption) (*IteratorErrorResponse, error)
	IteratorKey(ctx context.Context, in *IteratorKeyRequest, opts ...grpc.CallOption) (*IteratorKeyResponse, error)
	IteratorNext(ctx context.Context, in *IteratorNextRequest, opts ...grpc.CallOption) (*IteratorNextResponse, error)
	IteratorRelease(ctx context.Context, in *IteratorReleaseRequest, opts ...grpc.CallOption) (*IteratorReleaseResponse, error)
	IteratorValue(ctx context.Context, in *IteratorValueRequest, opts ...grpc.CallOption) (*IteratorValueResponse, error)
}

type kVClient struct {
	cc grpc.ClientConnInterface
}

func NewKVClient(cc grpc.ClientConnInterface) KVClient {
	return &kVClient{cc}
}

func (c *kVClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, KV_Get_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVClient) Batch(ctx context.Context, in *BatchRequest, opts ...grpc.CallOption) (*BatchResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BatchResponse)
	err := c.cc.Invoke(ctx, KV_Batch_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVClient) NewIterator(ctx context.Context, in *NewIteratorRequest, opts ...grpc.CallOption) (*NewIteratorResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(NewIteratorResponse)
	err := c.cc.Invoke(ctx, KV_NewIterator_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVClient) IteratorError(ctx context.Context, in *IteratorErrorRequest, opts ...grpc.CallOption) (*IteratorErrorResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(IteratorErrorResponse)
	err := c.cc.Invoke(ctx, KV_IteratorError_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVClient) IteratorKey(ctx context.Context, in *IteratorKeyRequest, opts ...grpc.CallOption) (*IteratorKeyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(IteratorKeyResponse)
	err := c.cc.Invoke(ctx, KV_IteratorKey_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVClient) IteratorNext(ctx context.Context, in *IteratorNextRequest, opts ...grpc.CallOption) (*IteratorNextResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(IteratorNextResponse)
	err := c.cc.Invoke(ctx, KV_IteratorNext_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVClient) IteratorRelease(ctx context.Context, in *IteratorReleaseRequest, opts ...grpc.CallOption) (*IteratorReleaseResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(IteratorReleaseResponse)
	err := c.cc.Invoke(ctx, KV_IteratorRelease_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVClient) IteratorValue(ctx context.Context, in *IteratorValueRequest, opts ...grpc.CallOption) (*IteratorValueResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(IteratorValueResponse)
	err := c.cc.Invoke(ctx, KV_IteratorValue_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KVServer is the server API for KV service.
// All implementations must embed UnimplementedKVServer
// for forward compatibility.
type KVServer interface {
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Batch(context.Context, *BatchRequest) (*BatchResponse, error)
	NewIterator(context.Context, *NewIteratorRequest) (*NewIteratorResponse, error)
	IteratorError(context.Context, *IteratorErrorRequest) (*IteratorErrorResponse, error)
	IteratorKey(context.Context, *IteratorKeyRequest) (*IteratorKeyResponse, error)
	IteratorNext(context.Context, *IteratorNextRequest) (*IteratorNextResponse, error)
	IteratorRelease(context.Context, *IteratorReleaseRequest) (*IteratorReleaseResponse, error)
	IteratorValue(context.Context, *IteratorValueRequest) (*IteratorValueResponse, error)
	mustEmbedUnimplementedKVServer()
}

// UnimplementedKVServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedKVServer struct{}

func (UnimplementedKVServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedKVServer) Batch(context.Context, *BatchRequest) (*BatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Batch not implemented")
}
func (UnimplementedKVServer) NewIterator(context.Context, *NewIteratorRequest) (*NewIteratorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewIterator not implemented")
}
func (UnimplementedKVServer) IteratorError(context.Context, *IteratorErrorRequest) (*IteratorErrorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IteratorError not implemented")
}
func (UnimplementedKVServer) IteratorKey(context.Context, *IteratorKeyRequest) (*IteratorKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IteratorKey not implemented")
}
func (UnimplementedKVServer) IteratorNext(context.Context, *IteratorNextRequest) (*IteratorNextResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IteratorNext not implemented")
}
func (UnimplementedKVServer) IteratorRelease(context.Context, *IteratorReleaseRequest) (*IteratorReleaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IteratorRelease not implemented")
}
func (UnimplementedKVServer) IteratorValue(context.Context, *IteratorValueRequest) (*IteratorValueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IteratorValue not implemented")
}
func (UnimplementedKVServer) mustEmbedUnimplementedKVServer() {}
func (UnimplementedKVServer) testEmbeddedByValue()            {}

// UnsafeKVServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KVServer will
// result in compilation errors.
type UnsafeKVServer interface {
	mustEmbedUnimplementedKVServer()
}

func RegisterKVServer(s grpc.ServiceRegistrar, srv KVServer) {
	// If the following call pancis, it indicates UnimplementedKVServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&KV_ServiceDesc, srv)
}

func _KV_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KV_Get_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KV_Batch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServer).Batch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KV_Batch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServer).Batch(ctx, req.(*BatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KV_NewIterator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NewIteratorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServer).NewIterator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KV_NewIterator_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServer).NewIterator(ctx, req.(*NewIteratorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KV_IteratorError_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IteratorErrorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServer).IteratorError(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KV_IteratorError_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServer).IteratorError(ctx, req.(*IteratorErrorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KV_IteratorKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IteratorKeyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServer).IteratorKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KV_IteratorKey_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServer).IteratorKey(ctx, req.(*IteratorKeyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KV_IteratorNext_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IteratorNextRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServer).IteratorNext(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KV_IteratorNext_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServer).IteratorNext(ctx, req.(*IteratorNextRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KV_IteratorRelease_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IteratorReleaseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServer).IteratorRelease(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KV_IteratorRelease_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServer).IteratorRelease(ctx, req.(*IteratorReleaseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KV_IteratorValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IteratorValueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServer).IteratorValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KV_IteratorValue_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServer).IteratorValue(ctx, req.(*IteratorValueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// KV_ServiceDesc is the grpc.ServiceDesc for KV service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var KV_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ethereum.rpcdb.api.v1.KV",
	HandlerType: (*KVServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _KV_Get_Handler,
		},
		{
			MethodName: "Batch",
			Handler:    _KV_Batch_Handler,
		},
		{
			MethodName: "NewIterator",
			Handler:    _KV_NewIterator_Handler,
		},
		{
			MethodName: "IteratorError",
			Handler:    _KV_IteratorError_Handler,
		},
		{
			MethodName: "IteratorKey",
			Handler:    _KV_IteratorKey_Handler,
		},
		{
			MethodName: "IteratorNext",
			Handler:    _KV_IteratorNext_Handler,
		},
		{
			MethodName: "IteratorRelease",
			Handler:    _KV_IteratorRelease_Handler,
		},
		{
			MethodName: "IteratorValue",
			Handler:    _KV_IteratorValue_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kvdb.proto",
}
