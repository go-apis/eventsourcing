// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.20.0
// source: store.proto

package store

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	structpb "google.golang.org/protobuf/types/known/structpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type NewTxRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *NewTxRequest) Reset() {
	*x = NewTxRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewTxRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewTxRequest) ProtoMessage() {}

func (x *NewTxRequest) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewTxRequest.ProtoReflect.Descriptor instead.
func (*NewTxRequest) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{0}
}

type NewTxResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionId string `protobuf:"bytes,1,opt,name=transactionId,proto3" json:"transactionId,omitempty"`
}

func (x *NewTxResponse) Reset() {
	*x = NewTxResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewTxResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewTxResponse) ProtoMessage() {}

func (x *NewTxResponse) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewTxResponse.ProtoReflect.Descriptor instead.
func (*NewTxResponse) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{1}
}

func (x *NewTxResponse) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

type Tx struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionId string `protobuf:"bytes,1,opt,name=transactionId,proto3" json:"transactionId,omitempty"`
}

func (x *Tx) Reset() {
	*x = Tx{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tx) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tx) ProtoMessage() {}

func (x *Tx) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tx.ProtoReflect.Descriptor instead.
func (*Tx) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{2}
}

func (x *Tx) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

type CommitResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UpdatedRows int32 `protobuf:"varint,1,opt,name=updatedRows,proto3" json:"updatedRows,omitempty"`
}

func (x *CommitResponse) Reset() {
	*x = CommitResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommitResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommitResponse) ProtoMessage() {}

func (x *CommitResponse) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommitResponse.ProtoReflect.Descriptor instead.
func (*CommitResponse) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{3}
}

func (x *CommitResponse) GetUpdatedRows() int32 {
	if x != nil {
		return x.UpdatedRows
	}
	return 0
}

type SaveEventsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionId string   `protobuf:"bytes,1,opt,name=transactionId,proto3" json:"transactionId,omitempty"`
	Events        []*Event `protobuf:"bytes,2,rep,name=events,proto3" json:"events,omitempty"`
}

func (x *SaveEventsRequest) Reset() {
	*x = SaveEventsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SaveEventsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SaveEventsRequest) ProtoMessage() {}

func (x *SaveEventsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SaveEventsRequest.ProtoReflect.Descriptor instead.
func (*SaveEventsRequest) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{4}
}

func (x *SaveEventsRequest) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

func (x *SaveEventsRequest) GetEvents() []*Event {
	if x != nil {
		return x.Events
	}
	return nil
}

type EventsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionId *string `protobuf:"bytes,1,opt,name=transactionId,proto3,oneof" json:"transactionId,omitempty"`
	ServiceName   *string `protobuf:"bytes,2,opt,name=serviceName,proto3,oneof" json:"serviceName,omitempty"`
	Namespace     *string `protobuf:"bytes,3,opt,name=namespace,proto3,oneof" json:"namespace,omitempty"`
	AggregateType *string `protobuf:"bytes,4,opt,name=aggregateType,proto3,oneof" json:"aggregateType,omitempty"`
	AggregateId   *string `protobuf:"bytes,5,opt,name=aggregateId,proto3,oneof" json:"aggregateId,omitempty"`
	FromVersion   *int64  `protobuf:"varint,6,opt,name=fromVersion,proto3,oneof" json:"fromVersion,omitempty"`
}

func (x *EventsRequest) Reset() {
	*x = EventsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventsRequest) ProtoMessage() {}

func (x *EventsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventsRequest.ProtoReflect.Descriptor instead.
func (*EventsRequest) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{5}
}

func (x *EventsRequest) GetTransactionId() string {
	if x != nil && x.TransactionId != nil {
		return *x.TransactionId
	}
	return ""
}

func (x *EventsRequest) GetServiceName() string {
	if x != nil && x.ServiceName != nil {
		return *x.ServiceName
	}
	return ""
}

func (x *EventsRequest) GetNamespace() string {
	if x != nil && x.Namespace != nil {
		return *x.Namespace
	}
	return ""
}

func (x *EventsRequest) GetAggregateType() string {
	if x != nil && x.AggregateType != nil {
		return *x.AggregateType
	}
	return ""
}

func (x *EventsRequest) GetAggregateId() string {
	if x != nil && x.AggregateId != nil {
		return *x.AggregateId
	}
	return ""
}

func (x *EventsRequest) GetFromVersion() int64 {
	if x != nil && x.FromVersion != nil {
		return *x.FromVersion
	}
	return 0
}

type EventsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Events []*Event `protobuf:"bytes,1,rep,name=events,proto3" json:"events,omitempty"`
}

func (x *EventsResponse) Reset() {
	*x = EventsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventsResponse) ProtoMessage() {}

func (x *EventsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventsResponse.ProtoReflect.Descriptor instead.
func (*EventsResponse) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{6}
}

func (x *EventsResponse) GetEvents() []*Event {
	if x != nil {
		return x.Events
	}
	return nil
}

type EventStreamRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceName string   `protobuf:"bytes,1,opt,name=serviceName,proto3" json:"serviceName,omitempty"`
	EventTypes  []string `protobuf:"bytes,2,rep,name=eventTypes,proto3" json:"eventTypes,omitempty"`
}

func (x *EventStreamRequest) Reset() {
	*x = EventStreamRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventStreamRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventStreamRequest) ProtoMessage() {}

func (x *EventStreamRequest) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventStreamRequest.ProtoReflect.Descriptor instead.
func (*EventStreamRequest) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{7}
}

func (x *EventStreamRequest) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *EventStreamRequest) GetEventTypes() []string {
	if x != nil {
		return x.EventTypes
	}
	return nil
}

type EventStreamResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Events []*Event `protobuf:"bytes,1,rep,name=events,proto3" json:"events,omitempty"`
}

func (x *EventStreamResponse) Reset() {
	*x = EventStreamResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventStreamResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventStreamResponse) ProtoMessage() {}

func (x *EventStreamResponse) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventStreamResponse.ProtoReflect.Descriptor instead.
func (*EventStreamResponse) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{8}
}

func (x *EventStreamResponse) GetEvents() []*Event {
	if x != nil {
		return x.Events
	}
	return nil
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceName   string `protobuf:"bytes,1,opt,name=serviceName,proto3" json:"serviceName,omitempty"`
	Namespace     string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	AggregateType string `protobuf:"bytes,3,opt,name=aggregateType,proto3" json:"aggregateType,omitempty"`
	AggregateId   string `protobuf:"bytes,4,opt,name=aggregateId,proto3" json:"aggregateId,omitempty"`
	Version       int64  `protobuf:"varint,5,opt,name=version,proto3" json:"version,omitempty"`
	Type          string `protobuf:"bytes,6,opt,name=type,proto3" json:"type,omitempty"`
	// @gotags: gorm:"serializer:timestamppb;type:time"
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=timestamp,proto3" json:"timestamp,omitempty" gorm:"serializer:timestamppb;type:time"`
	// @gotags: gorm:"type:jsonb"
	Data []byte `protobuf:"bytes,8,opt,name=data,proto3" json:"data,omitempty" gorm:"type:jsonb"`
	// @gotags: gorm:"serializer:json;type:jsonb"
	Metadata *structpb.Struct `protobuf:"bytes,9,opt,name=Metadata,proto3" json:"Metadata,omitempty" gorm:"serializer:json;type:jsonb"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{9}
}

func (x *Event) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *Event) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *Event) GetAggregateType() string {
	if x != nil {
		return x.AggregateType
	}
	return ""
}

func (x *Event) GetAggregateId() string {
	if x != nil {
		return x.AggregateId
	}
	return ""
}

func (x *Event) GetVersion() int64 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *Event) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Event) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *Event) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Event) GetMetadata() *structpb.Struct {
	if x != nil {
		return x.Metadata
	}
	return nil
}

type ErrorInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code  string `protobuf:"bytes,1,opt,name=code,proto3" json:"code,omitempty"`
	Cause string `protobuf:"bytes,2,opt,name=cause,proto3" json:"cause,omitempty"`
}

func (x *ErrorInfo) Reset() {
	*x = ErrorInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ErrorInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ErrorInfo) ProtoMessage() {}

func (x *ErrorInfo) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ErrorInfo.ProtoReflect.Descriptor instead.
func (*ErrorInfo) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{10}
}

func (x *ErrorInfo) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *ErrorInfo) GetCause() string {
	if x != nil {
		return x.Cause
	}
	return ""
}

type DebugInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Stack string `protobuf:"bytes,1,opt,name=stack,proto3" json:"stack,omitempty"`
}

func (x *DebugInfo) Reset() {
	*x = DebugInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DebugInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DebugInfo) ProtoMessage() {}

func (x *DebugInfo) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DebugInfo.ProtoReflect.Descriptor instead.
func (*DebugInfo) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{11}
}

func (x *DebugInfo) GetStack() string {
	if x != nil {
		return x.Stack
	}
	return ""
}

type RetryInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RetryDelay int32 `protobuf:"varint,1,opt,name=retry_delay,json=retryDelay,proto3" json:"retry_delay,omitempty"`
}

func (x *RetryInfo) Reset() {
	*x = RetryInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_store_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetryInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetryInfo) ProtoMessage() {}

func (x *RetryInfo) ProtoReflect() protoreflect.Message {
	mi := &file_store_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetryInfo.ProtoReflect.Descriptor instead.
func (*RetryInfo) Descriptor() ([]byte, []int) {
	return file_store_proto_rawDescGZIP(), []int{12}
}

func (x *RetryInfo) GetRetryDelay() int32 {
	if x != nil {
		return x.RetryDelay
	}
	return 0
}

var File_store_proto protoreflect.FileDescriptor

var file_store_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65,
	0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x0e, 0x0a, 0x0c, 0x4e, 0x65, 0x77,
	0x54, 0x78, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x35, 0x0a, 0x0d, 0x4e, 0x65, 0x77,
	0x54, 0x78, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0d, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64,
	0x22, 0x2a, 0x0a, 0x02, 0x54, 0x78, 0x12, 0x24, 0x0a, 0x0d, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x22, 0x32, 0x0a, 0x0e,
	0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x20,
	0x0a, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x52, 0x6f, 0x77, 0x73, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x52, 0x6f, 0x77, 0x73,
	0x22, 0x59, 0x0a, 0x11, 0x53, 0x61, 0x76, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x24, 0x0a, 0x0d, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x1e, 0x0a, 0x06, 0x65,
	0x76, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x06, 0x2e, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x22, 0xdf, 0x02, 0x0a, 0x0d,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x29, 0x0a,
	0x0d, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0d, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x49, 0x64, 0x88, 0x01, 0x01, 0x12, 0x25, 0x0a, 0x0b, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52,
	0x0b, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x88, 0x01, 0x01, 0x12,
	0x21, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x02, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x88,
	0x01, 0x01, 0x12, 0x29, 0x0a, 0x0d, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x54,
	0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x48, 0x03, 0x52, 0x0d, 0x61, 0x67, 0x67,
	0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x88, 0x01, 0x01, 0x12, 0x25, 0x0a,
	0x0b, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x49, 0x64, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x04, 0x52, 0x0b, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x49,
	0x64, 0x88, 0x01, 0x01, 0x12, 0x25, 0x0a, 0x0b, 0x66, 0x72, 0x6f, 0x6d, 0x56, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x48, 0x05, 0x52, 0x0b, 0x66, 0x72, 0x6f,
	0x6d, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x42, 0x10, 0x0a, 0x0e, 0x5f,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x42, 0x0e, 0x0a,
	0x0c, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x42, 0x0c, 0x0a,
	0x0a, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x42, 0x10, 0x0a, 0x0e, 0x5f,
	0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x42, 0x0e, 0x0a,
	0x0c, 0x5f, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x49, 0x64, 0x42, 0x0e, 0x0a,
	0x0c, 0x5f, 0x66, 0x72, 0x6f, 0x6d, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x30, 0x0a,
	0x0e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x1e, 0x0a, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x06, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x22,
	0x56, 0x0a, 0x12, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x73, 0x22, 0x35, 0x0a, 0x13, 0x45, 0x76, 0x65, 0x6e, 0x74,
	0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1e,
	0x0a, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x06,
	0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x22, 0xc0,
	0x02, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e,
	0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x61, 0x67, 0x67, 0x72,
	0x65, 0x67, 0x61, 0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x20,
	0x0a, 0x0b, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x49, 0x64, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x49, 0x64,
	0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x38,
	0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x33, 0x0a, 0x08,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x08, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x22, 0x35, 0x0a, 0x09, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x12,
	0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x6f,
	0x64, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x61, 0x75, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x63, 0x61, 0x75, 0x73, 0x65, 0x22, 0x21, 0x0a, 0x09, 0x44, 0x65, 0x62, 0x75,
	0x67, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x22, 0x2c, 0x0a, 0x09, 0x52,
	0x65, 0x74, 0x72, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x74, 0x72,
	0x79, 0x5f, 0x64, 0x65, 0x6c, 0x61, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x72,
	0x65, 0x74, 0x72, 0x79, 0x44, 0x65, 0x6c, 0x61, 0x79, 0x32, 0xa3, 0x02, 0x0a, 0x05, 0x53, 0x74,
	0x6f, 0x72, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x4e, 0x65, 0x77, 0x54, 0x78, 0x12, 0x0d, 0x2e, 0x4e,
	0x65, 0x77, 0x54, 0x78, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x4e, 0x65,
	0x77, 0x54, 0x78, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x20, 0x0a,
	0x06, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x12, 0x03, 0x2e, 0x54, 0x78, 0x1a, 0x0f, 0x2e, 0x43,
	0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x29, 0x0a, 0x08, 0x52, 0x6f, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x12, 0x03, 0x2e, 0x54, 0x78,
	0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x2b, 0x0a, 0x06, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x73, 0x12, 0x0e, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3a, 0x0a, 0x0a, 0x53, 0x61, 0x76, 0x65, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x12, 0x2e, 0x53, 0x61, 0x76, 0x65, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x22, 0x00, 0x12, 0x3a, 0x0a, 0x0b, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x12, 0x13, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x53,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x42,
	0x0a, 0x5a, 0x08, 0x70, 0x62, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_store_proto_rawDescOnce sync.Once
	file_store_proto_rawDescData = file_store_proto_rawDesc
)

func file_store_proto_rawDescGZIP() []byte {
	file_store_proto_rawDescOnce.Do(func() {
		file_store_proto_rawDescData = protoimpl.X.CompressGZIP(file_store_proto_rawDescData)
	})
	return file_store_proto_rawDescData
}

var file_store_proto_msgTypes = make([]protoimpl.MessageInfo, 13)
var file_store_proto_goTypes = []interface{}{
	(*NewTxRequest)(nil),          // 0: NewTxRequest
	(*NewTxResponse)(nil),         // 1: NewTxResponse
	(*Tx)(nil),                    // 2: Tx
	(*CommitResponse)(nil),        // 3: CommitResponse
	(*SaveEventsRequest)(nil),     // 4: SaveEventsRequest
	(*EventsRequest)(nil),         // 5: EventsRequest
	(*EventsResponse)(nil),        // 6: EventsResponse
	(*EventStreamRequest)(nil),    // 7: EventStreamRequest
	(*EventStreamResponse)(nil),   // 8: EventStreamResponse
	(*Event)(nil),                 // 9: Event
	(*ErrorInfo)(nil),             // 10: ErrorInfo
	(*DebugInfo)(nil),             // 11: DebugInfo
	(*RetryInfo)(nil),             // 12: RetryInfo
	(*timestamppb.Timestamp)(nil), // 13: google.protobuf.Timestamp
	(*structpb.Struct)(nil),       // 14: google.protobuf.Struct
	(*emptypb.Empty)(nil),         // 15: google.protobuf.Empty
}
var file_store_proto_depIdxs = []int32{
	9,  // 0: SaveEventsRequest.events:type_name -> Event
	9,  // 1: EventsResponse.events:type_name -> Event
	9,  // 2: EventStreamResponse.events:type_name -> Event
	13, // 3: Event.timestamp:type_name -> google.protobuf.Timestamp
	14, // 4: Event.Metadata:type_name -> google.protobuf.Struct
	0,  // 5: Store.NewTx:input_type -> NewTxRequest
	2,  // 6: Store.Commit:input_type -> Tx
	2,  // 7: Store.Rollback:input_type -> Tx
	5,  // 8: Store.Events:input_type -> EventsRequest
	4,  // 9: Store.SaveEvents:input_type -> SaveEventsRequest
	7,  // 10: Store.EventStream:input_type -> EventStreamRequest
	1,  // 11: Store.NewTx:output_type -> NewTxResponse
	3,  // 12: Store.Commit:output_type -> CommitResponse
	15, // 13: Store.Rollback:output_type -> google.protobuf.Empty
	6,  // 14: Store.Events:output_type -> EventsResponse
	15, // 15: Store.SaveEvents:output_type -> google.protobuf.Empty
	8,  // 16: Store.EventStream:output_type -> EventStreamResponse
	11, // [11:17] is the sub-list for method output_type
	5,  // [5:11] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_store_proto_init() }
func file_store_proto_init() {
	if File_store_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_store_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewTxRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_store_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewTxResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_store_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tx); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_store_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CommitResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_store_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SaveEventsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_store_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_store_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_store_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventStreamRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_store_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventStreamResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_store_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_store_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ErrorInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_store_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DebugInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_store_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RetryInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_store_proto_msgTypes[5].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_store_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   13,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_store_proto_goTypes,
		DependencyIndexes: file_store_proto_depIdxs,
		MessageInfos:      file_store_proto_msgTypes,
	}.Build()
	File_store_proto = out.File
	file_store_proto_rawDesc = nil
	file_store_proto_goTypes = nil
	file_store_proto_depIdxs = nil
}
