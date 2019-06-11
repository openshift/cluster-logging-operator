// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v1/services/feed_item_service.proto

package services // import "google.golang.org/genproto/googleapis/ads/googleads/v1/services"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/golang/protobuf/ptypes/wrappers"
import resources "google.golang.org/genproto/googleapis/ads/googleads/v1/resources"
import _ "google.golang.org/genproto/googleapis/api/annotations"
import status "google.golang.org/genproto/googleapis/rpc/status"
import field_mask "google.golang.org/genproto/protobuf/field_mask"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Request message for [FeedItemService.GetFeedItem][google.ads.googleads.v1.services.FeedItemService.GetFeedItem].
type GetFeedItemRequest struct {
	// The resource name of the feed item to fetch.
	ResourceName         string   `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetFeedItemRequest) Reset()         { *m = GetFeedItemRequest{} }
func (m *GetFeedItemRequest) String() string { return proto.CompactTextString(m) }
func (*GetFeedItemRequest) ProtoMessage()    {}
func (*GetFeedItemRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_feed_item_service_07bfb4ad3b8f0fac, []int{0}
}
func (m *GetFeedItemRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetFeedItemRequest.Unmarshal(m, b)
}
func (m *GetFeedItemRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetFeedItemRequest.Marshal(b, m, deterministic)
}
func (dst *GetFeedItemRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetFeedItemRequest.Merge(dst, src)
}
func (m *GetFeedItemRequest) XXX_Size() int {
	return xxx_messageInfo_GetFeedItemRequest.Size(m)
}
func (m *GetFeedItemRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetFeedItemRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetFeedItemRequest proto.InternalMessageInfo

func (m *GetFeedItemRequest) GetResourceName() string {
	if m != nil {
		return m.ResourceName
	}
	return ""
}

// Request message for [FeedItemService.MutateFeedItems][google.ads.googleads.v1.services.FeedItemService.MutateFeedItems].
type MutateFeedItemsRequest struct {
	// The ID of the customer whose feed items are being modified.
	CustomerId string `protobuf:"bytes,1,opt,name=customer_id,json=customerId,proto3" json:"customer_id,omitempty"`
	// The list of operations to perform on individual feed items.
	Operations []*FeedItemOperation `protobuf:"bytes,2,rep,name=operations,proto3" json:"operations,omitempty"`
	// If true, successful operations will be carried out and invalid
	// operations will return errors. If false, all operations will be carried
	// out in one transaction if and only if they are all valid.
	// Default is false.
	PartialFailure bool `protobuf:"varint,3,opt,name=partial_failure,json=partialFailure,proto3" json:"partial_failure,omitempty"`
	// If true, the request is validated but not executed. Only errors are
	// returned, not results.
	ValidateOnly         bool     `protobuf:"varint,4,opt,name=validate_only,json=validateOnly,proto3" json:"validate_only,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MutateFeedItemsRequest) Reset()         { *m = MutateFeedItemsRequest{} }
func (m *MutateFeedItemsRequest) String() string { return proto.CompactTextString(m) }
func (*MutateFeedItemsRequest) ProtoMessage()    {}
func (*MutateFeedItemsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_feed_item_service_07bfb4ad3b8f0fac, []int{1}
}
func (m *MutateFeedItemsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MutateFeedItemsRequest.Unmarshal(m, b)
}
func (m *MutateFeedItemsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MutateFeedItemsRequest.Marshal(b, m, deterministic)
}
func (dst *MutateFeedItemsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MutateFeedItemsRequest.Merge(dst, src)
}
func (m *MutateFeedItemsRequest) XXX_Size() int {
	return xxx_messageInfo_MutateFeedItemsRequest.Size(m)
}
func (m *MutateFeedItemsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MutateFeedItemsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MutateFeedItemsRequest proto.InternalMessageInfo

func (m *MutateFeedItemsRequest) GetCustomerId() string {
	if m != nil {
		return m.CustomerId
	}
	return ""
}

func (m *MutateFeedItemsRequest) GetOperations() []*FeedItemOperation {
	if m != nil {
		return m.Operations
	}
	return nil
}

func (m *MutateFeedItemsRequest) GetPartialFailure() bool {
	if m != nil {
		return m.PartialFailure
	}
	return false
}

func (m *MutateFeedItemsRequest) GetValidateOnly() bool {
	if m != nil {
		return m.ValidateOnly
	}
	return false
}

// A single operation (create, update, remove) on an feed item.
type FeedItemOperation struct {
	// FieldMask that determines which resource fields are modified in an update.
	UpdateMask *field_mask.FieldMask `protobuf:"bytes,4,opt,name=update_mask,json=updateMask,proto3" json:"update_mask,omitempty"`
	// The mutate operation.
	//
	// Types that are valid to be assigned to Operation:
	//	*FeedItemOperation_Create
	//	*FeedItemOperation_Update
	//	*FeedItemOperation_Remove
	Operation            isFeedItemOperation_Operation `protobuf_oneof:"operation"`
	XXX_NoUnkeyedLiteral struct{}                      `json:"-"`
	XXX_unrecognized     []byte                        `json:"-"`
	XXX_sizecache        int32                         `json:"-"`
}

func (m *FeedItemOperation) Reset()         { *m = FeedItemOperation{} }
func (m *FeedItemOperation) String() string { return proto.CompactTextString(m) }
func (*FeedItemOperation) ProtoMessage()    {}
func (*FeedItemOperation) Descriptor() ([]byte, []int) {
	return fileDescriptor_feed_item_service_07bfb4ad3b8f0fac, []int{2}
}
func (m *FeedItemOperation) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FeedItemOperation.Unmarshal(m, b)
}
func (m *FeedItemOperation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FeedItemOperation.Marshal(b, m, deterministic)
}
func (dst *FeedItemOperation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FeedItemOperation.Merge(dst, src)
}
func (m *FeedItemOperation) XXX_Size() int {
	return xxx_messageInfo_FeedItemOperation.Size(m)
}
func (m *FeedItemOperation) XXX_DiscardUnknown() {
	xxx_messageInfo_FeedItemOperation.DiscardUnknown(m)
}

var xxx_messageInfo_FeedItemOperation proto.InternalMessageInfo

func (m *FeedItemOperation) GetUpdateMask() *field_mask.FieldMask {
	if m != nil {
		return m.UpdateMask
	}
	return nil
}

type isFeedItemOperation_Operation interface {
	isFeedItemOperation_Operation()
}

type FeedItemOperation_Create struct {
	Create *resources.FeedItem `protobuf:"bytes,1,opt,name=create,proto3,oneof"`
}

type FeedItemOperation_Update struct {
	Update *resources.FeedItem `protobuf:"bytes,2,opt,name=update,proto3,oneof"`
}

type FeedItemOperation_Remove struct {
	Remove string `protobuf:"bytes,3,opt,name=remove,proto3,oneof"`
}

func (*FeedItemOperation_Create) isFeedItemOperation_Operation() {}

func (*FeedItemOperation_Update) isFeedItemOperation_Operation() {}

func (*FeedItemOperation_Remove) isFeedItemOperation_Operation() {}

func (m *FeedItemOperation) GetOperation() isFeedItemOperation_Operation {
	if m != nil {
		return m.Operation
	}
	return nil
}

func (m *FeedItemOperation) GetCreate() *resources.FeedItem {
	if x, ok := m.GetOperation().(*FeedItemOperation_Create); ok {
		return x.Create
	}
	return nil
}

func (m *FeedItemOperation) GetUpdate() *resources.FeedItem {
	if x, ok := m.GetOperation().(*FeedItemOperation_Update); ok {
		return x.Update
	}
	return nil
}

func (m *FeedItemOperation) GetRemove() string {
	if x, ok := m.GetOperation().(*FeedItemOperation_Remove); ok {
		return x.Remove
	}
	return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*FeedItemOperation) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _FeedItemOperation_OneofMarshaler, _FeedItemOperation_OneofUnmarshaler, _FeedItemOperation_OneofSizer, []interface{}{
		(*FeedItemOperation_Create)(nil),
		(*FeedItemOperation_Update)(nil),
		(*FeedItemOperation_Remove)(nil),
	}
}

func _FeedItemOperation_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*FeedItemOperation)
	// operation
	switch x := m.Operation.(type) {
	case *FeedItemOperation_Create:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Create); err != nil {
			return err
		}
	case *FeedItemOperation_Update:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Update); err != nil {
			return err
		}
	case *FeedItemOperation_Remove:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		b.EncodeStringBytes(x.Remove)
	case nil:
	default:
		return fmt.Errorf("FeedItemOperation.Operation has unexpected type %T", x)
	}
	return nil
}

func _FeedItemOperation_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*FeedItemOperation)
	switch tag {
	case 1: // operation.create
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(resources.FeedItem)
		err := b.DecodeMessage(msg)
		m.Operation = &FeedItemOperation_Create{msg}
		return true, err
	case 2: // operation.update
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(resources.FeedItem)
		err := b.DecodeMessage(msg)
		m.Operation = &FeedItemOperation_Update{msg}
		return true, err
	case 3: // operation.remove
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.Operation = &FeedItemOperation_Remove{x}
		return true, err
	default:
		return false, nil
	}
}

func _FeedItemOperation_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*FeedItemOperation)
	// operation
	switch x := m.Operation.(type) {
	case *FeedItemOperation_Create:
		s := proto.Size(x.Create)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *FeedItemOperation_Update:
		s := proto.Size(x.Update)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *FeedItemOperation_Remove:
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(len(x.Remove)))
		n += len(x.Remove)
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Response message for an feed item mutate.
type MutateFeedItemsResponse struct {
	// Errors that pertain to operation failures in the partial failure mode.
	// Returned only when partial_failure = true and all errors occur inside the
	// operations. If any errors occur outside the operations (e.g. auth errors),
	// we return an RPC level error.
	PartialFailureError *status.Status `protobuf:"bytes,3,opt,name=partial_failure_error,json=partialFailureError,proto3" json:"partial_failure_error,omitempty"`
	// All results for the mutate.
	Results              []*MutateFeedItemResult `protobuf:"bytes,2,rep,name=results,proto3" json:"results,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *MutateFeedItemsResponse) Reset()         { *m = MutateFeedItemsResponse{} }
func (m *MutateFeedItemsResponse) String() string { return proto.CompactTextString(m) }
func (*MutateFeedItemsResponse) ProtoMessage()    {}
func (*MutateFeedItemsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_feed_item_service_07bfb4ad3b8f0fac, []int{3}
}
func (m *MutateFeedItemsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MutateFeedItemsResponse.Unmarshal(m, b)
}
func (m *MutateFeedItemsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MutateFeedItemsResponse.Marshal(b, m, deterministic)
}
func (dst *MutateFeedItemsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MutateFeedItemsResponse.Merge(dst, src)
}
func (m *MutateFeedItemsResponse) XXX_Size() int {
	return xxx_messageInfo_MutateFeedItemsResponse.Size(m)
}
func (m *MutateFeedItemsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MutateFeedItemsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MutateFeedItemsResponse proto.InternalMessageInfo

func (m *MutateFeedItemsResponse) GetPartialFailureError() *status.Status {
	if m != nil {
		return m.PartialFailureError
	}
	return nil
}

func (m *MutateFeedItemsResponse) GetResults() []*MutateFeedItemResult {
	if m != nil {
		return m.Results
	}
	return nil
}

// The result for the feed item mutate.
type MutateFeedItemResult struct {
	// Returned for successful operations.
	ResourceName         string   `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MutateFeedItemResult) Reset()         { *m = MutateFeedItemResult{} }
func (m *MutateFeedItemResult) String() string { return proto.CompactTextString(m) }
func (*MutateFeedItemResult) ProtoMessage()    {}
func (*MutateFeedItemResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_feed_item_service_07bfb4ad3b8f0fac, []int{4}
}
func (m *MutateFeedItemResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MutateFeedItemResult.Unmarshal(m, b)
}
func (m *MutateFeedItemResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MutateFeedItemResult.Marshal(b, m, deterministic)
}
func (dst *MutateFeedItemResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MutateFeedItemResult.Merge(dst, src)
}
func (m *MutateFeedItemResult) XXX_Size() int {
	return xxx_messageInfo_MutateFeedItemResult.Size(m)
}
func (m *MutateFeedItemResult) XXX_DiscardUnknown() {
	xxx_messageInfo_MutateFeedItemResult.DiscardUnknown(m)
}

var xxx_messageInfo_MutateFeedItemResult proto.InternalMessageInfo

func (m *MutateFeedItemResult) GetResourceName() string {
	if m != nil {
		return m.ResourceName
	}
	return ""
}

func init() {
	proto.RegisterType((*GetFeedItemRequest)(nil), "google.ads.googleads.v1.services.GetFeedItemRequest")
	proto.RegisterType((*MutateFeedItemsRequest)(nil), "google.ads.googleads.v1.services.MutateFeedItemsRequest")
	proto.RegisterType((*FeedItemOperation)(nil), "google.ads.googleads.v1.services.FeedItemOperation")
	proto.RegisterType((*MutateFeedItemsResponse)(nil), "google.ads.googleads.v1.services.MutateFeedItemsResponse")
	proto.RegisterType((*MutateFeedItemResult)(nil), "google.ads.googleads.v1.services.MutateFeedItemResult")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// FeedItemServiceClient is the client API for FeedItemService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type FeedItemServiceClient interface {
	// Returns the requested feed item in full detail.
	GetFeedItem(ctx context.Context, in *GetFeedItemRequest, opts ...grpc.CallOption) (*resources.FeedItem, error)
	// Creates, updates, or removes feed items. Operation statuses are
	// returned.
	MutateFeedItems(ctx context.Context, in *MutateFeedItemsRequest, opts ...grpc.CallOption) (*MutateFeedItemsResponse, error)
}

type feedItemServiceClient struct {
	cc *grpc.ClientConn
}

func NewFeedItemServiceClient(cc *grpc.ClientConn) FeedItemServiceClient {
	return &feedItemServiceClient{cc}
}

func (c *feedItemServiceClient) GetFeedItem(ctx context.Context, in *GetFeedItemRequest, opts ...grpc.CallOption) (*resources.FeedItem, error) {
	out := new(resources.FeedItem)
	err := c.cc.Invoke(ctx, "/google.ads.googleads.v1.services.FeedItemService/GetFeedItem", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *feedItemServiceClient) MutateFeedItems(ctx context.Context, in *MutateFeedItemsRequest, opts ...grpc.CallOption) (*MutateFeedItemsResponse, error) {
	out := new(MutateFeedItemsResponse)
	err := c.cc.Invoke(ctx, "/google.ads.googleads.v1.services.FeedItemService/MutateFeedItems", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FeedItemServiceServer is the server API for FeedItemService service.
type FeedItemServiceServer interface {
	// Returns the requested feed item in full detail.
	GetFeedItem(context.Context, *GetFeedItemRequest) (*resources.FeedItem, error)
	// Creates, updates, or removes feed items. Operation statuses are
	// returned.
	MutateFeedItems(context.Context, *MutateFeedItemsRequest) (*MutateFeedItemsResponse, error)
}

func RegisterFeedItemServiceServer(s *grpc.Server, srv FeedItemServiceServer) {
	s.RegisterService(&_FeedItemService_serviceDesc, srv)
}

func _FeedItemService_GetFeedItem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFeedItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeedItemServiceServer).GetFeedItem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.ads.googleads.v1.services.FeedItemService/GetFeedItem",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeedItemServiceServer).GetFeedItem(ctx, req.(*GetFeedItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FeedItemService_MutateFeedItems_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MutateFeedItemsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeedItemServiceServer).MutateFeedItems(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.ads.googleads.v1.services.FeedItemService/MutateFeedItems",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeedItemServiceServer).MutateFeedItems(ctx, req.(*MutateFeedItemsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _FeedItemService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "google.ads.googleads.v1.services.FeedItemService",
	HandlerType: (*FeedItemServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetFeedItem",
			Handler:    _FeedItemService_GetFeedItem_Handler,
		},
		{
			MethodName: "MutateFeedItems",
			Handler:    _FeedItemService_MutateFeedItems_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "google/ads/googleads/v1/services/feed_item_service.proto",
}

func init() {
	proto.RegisterFile("google/ads/googleads/v1/services/feed_item_service.proto", fileDescriptor_feed_item_service_07bfb4ad3b8f0fac)
}

var fileDescriptor_feed_item_service_07bfb4ad3b8f0fac = []byte{
	// 708 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x54, 0x4f, 0x6b, 0x13, 0x41,
	0x14, 0x77, 0xb7, 0x52, 0xed, 0x6c, 0xb5, 0x38, 0x56, 0x1b, 0x82, 0x68, 0x58, 0x0b, 0x96, 0x14,
	0x77, 0x49, 0x22, 0xd2, 0x6e, 0xe9, 0x21, 0x85, 0xa6, 0xed, 0xa1, 0xb6, 0x6c, 0xa1, 0x07, 0x09,
	0x2c, 0xd3, 0xec, 0x4b, 0x58, 0xba, 0xbb, 0xb3, 0xce, 0xcc, 0x46, 0x4a, 0xe9, 0x45, 0xf0, 0x13,
	0xf8, 0x0d, 0x04, 0x2f, 0x5e, 0xfd, 0x04, 0x5e, 0xbd, 0x7a, 0xf5, 0xe8, 0xc9, 0xaf, 0x20, 0x82,
	0xec, 0xce, 0x4e, 0xda, 0xa4, 0x86, 0xd8, 0xde, 0xde, 0xbc, 0xf9, 0xfd, 0x7e, 0xef, 0xcd, 0xfb,
	0x33, 0x68, 0xa5, 0x47, 0x69, 0x2f, 0x04, 0x9b, 0xf8, 0xdc, 0x96, 0x66, 0x66, 0xf5, 0x6b, 0x36,
	0x07, 0xd6, 0x0f, 0x3a, 0xc0, 0xed, 0x2e, 0x80, 0xef, 0x05, 0x02, 0x22, 0xaf, 0x70, 0x59, 0x09,
	0xa3, 0x82, 0xe2, 0x8a, 0x84, 0x5b, 0xc4, 0xe7, 0xd6, 0x80, 0x69, 0xf5, 0x6b, 0x96, 0x62, 0x96,
	0x6b, 0xe3, 0xb4, 0x19, 0x70, 0x9a, 0xb2, 0x21, 0x71, 0x29, 0x5a, 0x7e, 0xa4, 0x28, 0x49, 0x60,
	0x93, 0x38, 0xa6, 0x82, 0x88, 0x80, 0xc6, 0xbc, 0xb8, 0x2d, 0x42, 0xda, 0xf9, 0xe9, 0x28, 0xed,
	0xda, 0xdd, 0x00, 0x42, 0xdf, 0x8b, 0x08, 0x3f, 0x2e, 0x10, 0x8f, 0x47, 0x11, 0x6f, 0x19, 0x49,
	0x12, 0x60, 0x4a, 0x61, 0xa1, 0xb8, 0x67, 0x49, 0xc7, 0xe6, 0x82, 0x88, 0xb4, 0xb8, 0x30, 0x57,
	0x11, 0xde, 0x02, 0xd1, 0x02, 0xf0, 0x77, 0x04, 0x44, 0x2e, 0xbc, 0x49, 0x81, 0x0b, 0xfc, 0x14,
	0xdd, 0x51, 0xb9, 0x7a, 0x31, 0x89, 0xa0, 0xa4, 0x55, 0xb4, 0xa5, 0x19, 0x77, 0x56, 0x39, 0x5f,
	0x91, 0x08, 0xcc, 0x1f, 0x1a, 0x7a, 0xb8, 0x9b, 0x0a, 0x22, 0x40, 0xd1, 0xb9, 0xe2, 0x3f, 0x41,
	0x46, 0x27, 0xe5, 0x82, 0x46, 0xc0, 0xbc, 0xc0, 0x2f, 0xd8, 0x48, 0xb9, 0x76, 0x7c, 0x7c, 0x80,
	0x10, 0x4d, 0x80, 0xc9, 0x57, 0x96, 0xf4, 0xca, 0xd4, 0x92, 0x51, 0x6f, 0x58, 0x93, 0x2a, 0x6b,
	0xa9, 0x40, 0x7b, 0x8a, 0xeb, 0x5e, 0x90, 0xc1, 0xcf, 0xd0, 0x5c, 0x42, 0x98, 0x08, 0x48, 0xe8,
	0x75, 0x49, 0x10, 0xa6, 0x0c, 0x4a, 0x53, 0x15, 0x6d, 0xe9, 0xb6, 0x7b, 0xb7, 0x70, 0xb7, 0xa4,
	0x37, 0x7b, 0x5e, 0x9f, 0x84, 0x81, 0x4f, 0x04, 0x78, 0x34, 0x0e, 0x4f, 0x4a, 0x37, 0x73, 0xd8,
	0xac, 0x72, 0xee, 0xc5, 0xe1, 0x89, 0xf9, 0x5e, 0x47, 0xf7, 0x2e, 0xc5, 0xc3, 0x6b, 0xc8, 0x48,
	0x93, 0x9c, 0x98, 0x55, 0x3f, 0x27, 0x1a, 0xf5, 0xb2, 0xca, 0x5c, 0x95, 0xdf, 0x6a, 0x65, 0x0d,
	0xda, 0x25, 0xfc, 0xd8, 0x45, 0x12, 0x9e, 0xd9, 0x78, 0x13, 0x4d, 0x77, 0x18, 0x10, 0x21, 0xeb,
	0x69, 0xd4, 0x97, 0xc7, 0xbe, 0x78, 0x30, 0x29, 0x83, 0x27, 0x6f, 0xdf, 0x70, 0x0b, 0x72, 0x26,
	0x23, 0x45, 0x4b, 0xfa, 0xb5, 0x64, 0x24, 0x19, 0x97, 0xd0, 0x34, 0x83, 0x88, 0xf6, 0x65, 0x95,
	0x66, 0xb2, 0x1b, 0x79, 0xde, 0x30, 0xd0, 0xcc, 0xa0, 0xac, 0xe6, 0x17, 0x0d, 0x2d, 0x5c, 0x6a,
	0x33, 0x4f, 0x68, 0xcc, 0x01, 0xb7, 0xd0, 0x83, 0x91, 0x8a, 0x7b, 0xc0, 0x18, 0x65, 0xb9, 0xa2,
	0x51, 0xc7, 0x2a, 0x31, 0x96, 0x74, 0xac, 0x83, 0x7c, 0xec, 0xdc, 0xfb, 0xc3, 0xbd, 0xd8, 0xcc,
	0xe0, 0x78, 0x1f, 0xdd, 0x62, 0xc0, 0xd3, 0x50, 0xa8, 0x59, 0x78, 0x39, 0x79, 0x16, 0x86, 0x73,
	0x72, 0x73, 0xba, 0xab, 0x64, 0xcc, 0x35, 0x34, 0xff, 0x2f, 0xc0, 0x7f, 0x4d, 0x76, 0xfd, 0x8f,
	0x8e, 0xe6, 0x14, 0xef, 0x40, 0xc6, 0xc3, 0x9f, 0x34, 0x64, 0x5c, 0xd8, 0x14, 0xfc, 0x62, 0x72,
	0x86, 0x97, 0x17, 0xab, 0x7c, 0x95, 0x56, 0x99, 0x8d, 0x77, 0xdf, 0x7f, 0x7e, 0xd0, 0x9f, 0xe3,
	0xe5, 0xec, 0xef, 0x38, 0x1d, 0x4a, 0x7b, 0x5d, 0xed, 0x12, 0xb7, 0xab, 0xf9, 0x67, 0x92, 0xf7,
	0xc5, 0xae, 0x9e, 0xe1, 0xaf, 0x1a, 0x9a, 0x1b, 0x69, 0x17, 0x5e, 0xb9, 0x6a, 0x35, 0xd5, 0x22,
	0x97, 0x57, 0xaf, 0xc1, 0x94, 0xb3, 0x61, 0xae, 0xe6, 0xd9, 0x37, 0x4c, 0x2b, 0xcb, 0xfe, 0x3c,
	0xdd, 0xd3, 0x0b, 0x1f, 0xc3, 0x7a, 0xf5, 0xec, 0x3c, 0x79, 0x27, 0xca, 0x85, 0x1c, 0xad, 0xba,
	0xf1, 0x5b, 0x43, 0x8b, 0x1d, 0x1a, 0x4d, 0x8c, 0xbd, 0x31, 0x3f, 0xd2, 0xa5, 0xfd, 0x6c, 0xff,
	0xf6, 0xb5, 0xd7, 0xdb, 0x05, 0xb3, 0x47, 0x43, 0x12, 0xf7, 0x2c, 0xca, 0x7a, 0x76, 0x0f, 0xe2,
	0x7c, 0x3b, 0xd5, 0x8f, 0x9c, 0x04, 0x7c, 0xfc, 0xe7, 0xbf, 0xa6, 0x8c, 0x8f, 0xfa, 0xd4, 0x56,
	0xb3, 0xf9, 0x59, 0xaf, 0x6c, 0x49, 0xc1, 0xa6, 0xcf, 0x2d, 0x69, 0x66, 0xd6, 0x61, 0xcd, 0x2a,
	0x02, 0xf3, 0x6f, 0x0a, 0xd2, 0x6e, 0xfa, 0xbc, 0x3d, 0x80, 0xb4, 0x0f, 0x6b, 0x6d, 0x05, 0xf9,
	0xa5, 0x2f, 0x4a, 0xbf, 0xe3, 0x34, 0x7d, 0xee, 0x38, 0x03, 0x90, 0xe3, 0x1c, 0xd6, 0x1c, 0x47,
	0xc1, 0x8e, 0xa6, 0xf3, 0x3c, 0x1b, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x02, 0x4f, 0xc6, 0x7f,
	0xa3, 0x06, 0x00, 0x00,
}
