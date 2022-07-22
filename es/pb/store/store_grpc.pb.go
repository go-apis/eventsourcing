// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.0
// source: store.proto

package store

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// StoreClient is the client API for Store service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StoreClient interface {
	NewTx(ctx context.Context, in *NewTxRequest, opts ...grpc.CallOption) (*NewTxResponse, error)
	Commit(ctx context.Context, in *Tx, opts ...grpc.CallOption) (*CommitResponse, error)
	Rollback(ctx context.Context, in *Tx, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Events(ctx context.Context, in *EventsRequest, opts ...grpc.CallOption) (*EventsResponse, error)
	SaveEvents(ctx context.Context, in *SaveEventsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	EventStream(ctx context.Context, in *EventStreamRequest, opts ...grpc.CallOption) (Store_EventStreamClient, error)
}

type storeClient struct {
	cc grpc.ClientConnInterface
}

func NewStoreClient(cc grpc.ClientConnInterface) StoreClient {
	return &storeClient{cc}
}

func (c *storeClient) NewTx(ctx context.Context, in *NewTxRequest, opts ...grpc.CallOption) (*NewTxResponse, error) {
	out := new(NewTxResponse)
	err := c.cc.Invoke(ctx, "/Store/NewTx", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storeClient) Commit(ctx context.Context, in *Tx, opts ...grpc.CallOption) (*CommitResponse, error) {
	out := new(CommitResponse)
	err := c.cc.Invoke(ctx, "/Store/Commit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storeClient) Rollback(ctx context.Context, in *Tx, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/Store/Rollback", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storeClient) Events(ctx context.Context, in *EventsRequest, opts ...grpc.CallOption) (*EventsResponse, error) {
	out := new(EventsResponse)
	err := c.cc.Invoke(ctx, "/Store/Events", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storeClient) SaveEvents(ctx context.Context, in *SaveEventsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/Store/SaveEvents", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storeClient) EventStream(ctx context.Context, in *EventStreamRequest, opts ...grpc.CallOption) (Store_EventStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Store_ServiceDesc.Streams[0], "/Store/EventStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &storeEventStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Store_EventStreamClient interface {
	Recv() (*EventStreamResponse, error)
	grpc.ClientStream
}

type storeEventStreamClient struct {
	grpc.ClientStream
}

func (x *storeEventStreamClient) Recv() (*EventStreamResponse, error) {
	m := new(EventStreamResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// StoreServer is the server API for Store service.
// All implementations must embed UnimplementedStoreServer
// for forward compatibility
type StoreServer interface {
	NewTx(context.Context, *NewTxRequest) (*NewTxResponse, error)
	Commit(context.Context, *Tx) (*CommitResponse, error)
	Rollback(context.Context, *Tx) (*emptypb.Empty, error)
	Events(context.Context, *EventsRequest) (*EventsResponse, error)
	SaveEvents(context.Context, *SaveEventsRequest) (*emptypb.Empty, error)
	EventStream(*EventStreamRequest, Store_EventStreamServer) error
	mustEmbedUnimplementedStoreServer()
}

// UnimplementedStoreServer must be embedded to have forward compatible implementations.
type UnimplementedStoreServer struct {
}

func (UnimplementedStoreServer) NewTx(context.Context, *NewTxRequest) (*NewTxResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewTx not implemented")
}
func (UnimplementedStoreServer) Commit(context.Context, *Tx) (*CommitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Commit not implemented")
}
func (UnimplementedStoreServer) Rollback(context.Context, *Tx) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Rollback not implemented")
}
func (UnimplementedStoreServer) Events(context.Context, *EventsRequest) (*EventsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Events not implemented")
}
func (UnimplementedStoreServer) SaveEvents(context.Context, *SaveEventsRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SaveEvents not implemented")
}
func (UnimplementedStoreServer) EventStream(*EventStreamRequest, Store_EventStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method EventStream not implemented")
}
func (UnimplementedStoreServer) mustEmbedUnimplementedStoreServer() {}

// UnsafeStoreServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StoreServer will
// result in compilation errors.
type UnsafeStoreServer interface {
	mustEmbedUnimplementedStoreServer()
}

func RegisterStoreServer(s grpc.ServiceRegistrar, srv StoreServer) {
	s.RegisterService(&Store_ServiceDesc, srv)
}

func _Store_NewTx_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NewTxRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StoreServer).NewTx(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Store/NewTx",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StoreServer).NewTx(ctx, req.(*NewTxRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Store_Commit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Tx)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StoreServer).Commit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Store/Commit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StoreServer).Commit(ctx, req.(*Tx))
	}
	return interceptor(ctx, in, info, handler)
}

func _Store_Rollback_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Tx)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StoreServer).Rollback(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Store/Rollback",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StoreServer).Rollback(ctx, req.(*Tx))
	}
	return interceptor(ctx, in, info, handler)
}

func _Store_Events_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EventsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StoreServer).Events(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Store/Events",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StoreServer).Events(ctx, req.(*EventsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Store_SaveEvents_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SaveEventsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StoreServer).SaveEvents(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Store/SaveEvents",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StoreServer).SaveEvents(ctx, req.(*SaveEventsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Store_EventStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(EventStreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StoreServer).EventStream(m, &storeEventStreamServer{stream})
}

type Store_EventStreamServer interface {
	Send(*EventStreamResponse) error
	grpc.ServerStream
}

type storeEventStreamServer struct {
	grpc.ServerStream
}

func (x *storeEventStreamServer) Send(m *EventStreamResponse) error {
	return x.ServerStream.SendMsg(m)
}

// Store_ServiceDesc is the grpc.ServiceDesc for Store service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Store_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Store",
	HandlerType: (*StoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NewTx",
			Handler:    _Store_NewTx_Handler,
		},
		{
			MethodName: "Commit",
			Handler:    _Store_Commit_Handler,
		},
		{
			MethodName: "Rollback",
			Handler:    _Store_Rollback_Handler,
		},
		{
			MethodName: "Events",
			Handler:    _Store_Events_Handler,
		},
		{
			MethodName: "SaveEvents",
			Handler:    _Store_SaveEvents_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "EventStream",
			Handler:       _Store_EventStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "store.proto",
}
