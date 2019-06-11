// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v1/services/ad_group_extension_setting_service.proto

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

// Request message for
// [AdGroupExtensionSettingService.GetAdGroupExtensionSetting][google.ads.googleads.v1.services.AdGroupExtensionSettingService.GetAdGroupExtensionSetting].
type GetAdGroupExtensionSettingRequest struct {
	// The resource name of the ad group extension setting to fetch.
	ResourceName         string   `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetAdGroupExtensionSettingRequest) Reset()         { *m = GetAdGroupExtensionSettingRequest{} }
func (m *GetAdGroupExtensionSettingRequest) String() string { return proto.CompactTextString(m) }
func (*GetAdGroupExtensionSettingRequest) ProtoMessage()    {}
func (*GetAdGroupExtensionSettingRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad_group_extension_setting_service_fd5e4f4a8f156262, []int{0}
}
func (m *GetAdGroupExtensionSettingRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetAdGroupExtensionSettingRequest.Unmarshal(m, b)
}
func (m *GetAdGroupExtensionSettingRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetAdGroupExtensionSettingRequest.Marshal(b, m, deterministic)
}
func (dst *GetAdGroupExtensionSettingRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetAdGroupExtensionSettingRequest.Merge(dst, src)
}
func (m *GetAdGroupExtensionSettingRequest) XXX_Size() int {
	return xxx_messageInfo_GetAdGroupExtensionSettingRequest.Size(m)
}
func (m *GetAdGroupExtensionSettingRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetAdGroupExtensionSettingRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetAdGroupExtensionSettingRequest proto.InternalMessageInfo

func (m *GetAdGroupExtensionSettingRequest) GetResourceName() string {
	if m != nil {
		return m.ResourceName
	}
	return ""
}

// Request message for
// [AdGroupExtensionSettingService.MutateAdGroupExtensionSettings][google.ads.googleads.v1.services.AdGroupExtensionSettingService.MutateAdGroupExtensionSettings].
type MutateAdGroupExtensionSettingsRequest struct {
	// The ID of the customer whose ad group extension settings are being
	// modified.
	CustomerId string `protobuf:"bytes,1,opt,name=customer_id,json=customerId,proto3" json:"customer_id,omitempty"`
	// The list of operations to perform on individual ad group extension
	// settings.
	Operations []*AdGroupExtensionSettingOperation `protobuf:"bytes,2,rep,name=operations,proto3" json:"operations,omitempty"`
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

func (m *MutateAdGroupExtensionSettingsRequest) Reset()         { *m = MutateAdGroupExtensionSettingsRequest{} }
func (m *MutateAdGroupExtensionSettingsRequest) String() string { return proto.CompactTextString(m) }
func (*MutateAdGroupExtensionSettingsRequest) ProtoMessage()    {}
func (*MutateAdGroupExtensionSettingsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad_group_extension_setting_service_fd5e4f4a8f156262, []int{1}
}
func (m *MutateAdGroupExtensionSettingsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MutateAdGroupExtensionSettingsRequest.Unmarshal(m, b)
}
func (m *MutateAdGroupExtensionSettingsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MutateAdGroupExtensionSettingsRequest.Marshal(b, m, deterministic)
}
func (dst *MutateAdGroupExtensionSettingsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MutateAdGroupExtensionSettingsRequest.Merge(dst, src)
}
func (m *MutateAdGroupExtensionSettingsRequest) XXX_Size() int {
	return xxx_messageInfo_MutateAdGroupExtensionSettingsRequest.Size(m)
}
func (m *MutateAdGroupExtensionSettingsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MutateAdGroupExtensionSettingsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MutateAdGroupExtensionSettingsRequest proto.InternalMessageInfo

func (m *MutateAdGroupExtensionSettingsRequest) GetCustomerId() string {
	if m != nil {
		return m.CustomerId
	}
	return ""
}

func (m *MutateAdGroupExtensionSettingsRequest) GetOperations() []*AdGroupExtensionSettingOperation {
	if m != nil {
		return m.Operations
	}
	return nil
}

func (m *MutateAdGroupExtensionSettingsRequest) GetPartialFailure() bool {
	if m != nil {
		return m.PartialFailure
	}
	return false
}

func (m *MutateAdGroupExtensionSettingsRequest) GetValidateOnly() bool {
	if m != nil {
		return m.ValidateOnly
	}
	return false
}

// A single operation (create, update, remove) on an ad group extension setting.
type AdGroupExtensionSettingOperation struct {
	// FieldMask that determines which resource fields are modified in an update.
	UpdateMask *field_mask.FieldMask `protobuf:"bytes,4,opt,name=update_mask,json=updateMask,proto3" json:"update_mask,omitempty"`
	// The mutate operation.
	//
	// Types that are valid to be assigned to Operation:
	//	*AdGroupExtensionSettingOperation_Create
	//	*AdGroupExtensionSettingOperation_Update
	//	*AdGroupExtensionSettingOperation_Remove
	Operation            isAdGroupExtensionSettingOperation_Operation `protobuf_oneof:"operation"`
	XXX_NoUnkeyedLiteral struct{}                                     `json:"-"`
	XXX_unrecognized     []byte                                       `json:"-"`
	XXX_sizecache        int32                                        `json:"-"`
}

func (m *AdGroupExtensionSettingOperation) Reset()         { *m = AdGroupExtensionSettingOperation{} }
func (m *AdGroupExtensionSettingOperation) String() string { return proto.CompactTextString(m) }
func (*AdGroupExtensionSettingOperation) ProtoMessage()    {}
func (*AdGroupExtensionSettingOperation) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad_group_extension_setting_service_fd5e4f4a8f156262, []int{2}
}
func (m *AdGroupExtensionSettingOperation) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AdGroupExtensionSettingOperation.Unmarshal(m, b)
}
func (m *AdGroupExtensionSettingOperation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AdGroupExtensionSettingOperation.Marshal(b, m, deterministic)
}
func (dst *AdGroupExtensionSettingOperation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AdGroupExtensionSettingOperation.Merge(dst, src)
}
func (m *AdGroupExtensionSettingOperation) XXX_Size() int {
	return xxx_messageInfo_AdGroupExtensionSettingOperation.Size(m)
}
func (m *AdGroupExtensionSettingOperation) XXX_DiscardUnknown() {
	xxx_messageInfo_AdGroupExtensionSettingOperation.DiscardUnknown(m)
}

var xxx_messageInfo_AdGroupExtensionSettingOperation proto.InternalMessageInfo

func (m *AdGroupExtensionSettingOperation) GetUpdateMask() *field_mask.FieldMask {
	if m != nil {
		return m.UpdateMask
	}
	return nil
}

type isAdGroupExtensionSettingOperation_Operation interface {
	isAdGroupExtensionSettingOperation_Operation()
}

type AdGroupExtensionSettingOperation_Create struct {
	Create *resources.AdGroupExtensionSetting `protobuf:"bytes,1,opt,name=create,proto3,oneof"`
}

type AdGroupExtensionSettingOperation_Update struct {
	Update *resources.AdGroupExtensionSetting `protobuf:"bytes,2,opt,name=update,proto3,oneof"`
}

type AdGroupExtensionSettingOperation_Remove struct {
	Remove string `protobuf:"bytes,3,opt,name=remove,proto3,oneof"`
}

func (*AdGroupExtensionSettingOperation_Create) isAdGroupExtensionSettingOperation_Operation() {}

func (*AdGroupExtensionSettingOperation_Update) isAdGroupExtensionSettingOperation_Operation() {}

func (*AdGroupExtensionSettingOperation_Remove) isAdGroupExtensionSettingOperation_Operation() {}

func (m *AdGroupExtensionSettingOperation) GetOperation() isAdGroupExtensionSettingOperation_Operation {
	if m != nil {
		return m.Operation
	}
	return nil
}

func (m *AdGroupExtensionSettingOperation) GetCreate() *resources.AdGroupExtensionSetting {
	if x, ok := m.GetOperation().(*AdGroupExtensionSettingOperation_Create); ok {
		return x.Create
	}
	return nil
}

func (m *AdGroupExtensionSettingOperation) GetUpdate() *resources.AdGroupExtensionSetting {
	if x, ok := m.GetOperation().(*AdGroupExtensionSettingOperation_Update); ok {
		return x.Update
	}
	return nil
}

func (m *AdGroupExtensionSettingOperation) GetRemove() string {
	if x, ok := m.GetOperation().(*AdGroupExtensionSettingOperation_Remove); ok {
		return x.Remove
	}
	return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*AdGroupExtensionSettingOperation) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _AdGroupExtensionSettingOperation_OneofMarshaler, _AdGroupExtensionSettingOperation_OneofUnmarshaler, _AdGroupExtensionSettingOperation_OneofSizer, []interface{}{
		(*AdGroupExtensionSettingOperation_Create)(nil),
		(*AdGroupExtensionSettingOperation_Update)(nil),
		(*AdGroupExtensionSettingOperation_Remove)(nil),
	}
}

func _AdGroupExtensionSettingOperation_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*AdGroupExtensionSettingOperation)
	// operation
	switch x := m.Operation.(type) {
	case *AdGroupExtensionSettingOperation_Create:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Create); err != nil {
			return err
		}
	case *AdGroupExtensionSettingOperation_Update:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Update); err != nil {
			return err
		}
	case *AdGroupExtensionSettingOperation_Remove:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		b.EncodeStringBytes(x.Remove)
	case nil:
	default:
		return fmt.Errorf("AdGroupExtensionSettingOperation.Operation has unexpected type %T", x)
	}
	return nil
}

func _AdGroupExtensionSettingOperation_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*AdGroupExtensionSettingOperation)
	switch tag {
	case 1: // operation.create
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(resources.AdGroupExtensionSetting)
		err := b.DecodeMessage(msg)
		m.Operation = &AdGroupExtensionSettingOperation_Create{msg}
		return true, err
	case 2: // operation.update
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(resources.AdGroupExtensionSetting)
		err := b.DecodeMessage(msg)
		m.Operation = &AdGroupExtensionSettingOperation_Update{msg}
		return true, err
	case 3: // operation.remove
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.Operation = &AdGroupExtensionSettingOperation_Remove{x}
		return true, err
	default:
		return false, nil
	}
}

func _AdGroupExtensionSettingOperation_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*AdGroupExtensionSettingOperation)
	// operation
	switch x := m.Operation.(type) {
	case *AdGroupExtensionSettingOperation_Create:
		s := proto.Size(x.Create)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *AdGroupExtensionSettingOperation_Update:
		s := proto.Size(x.Update)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *AdGroupExtensionSettingOperation_Remove:
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(len(x.Remove)))
		n += len(x.Remove)
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Response message for an ad group extension setting mutate.
type MutateAdGroupExtensionSettingsResponse struct {
	// Errors that pertain to operation failures in the partial failure mode.
	// Returned only when partial_failure = true and all errors occur inside the
	// operations. If any errors occur outside the operations (e.g. auth errors),
	// we return an RPC level error.
	PartialFailureError *status.Status `protobuf:"bytes,3,opt,name=partial_failure_error,json=partialFailureError,proto3" json:"partial_failure_error,omitempty"`
	// All results for the mutate.
	Results              []*MutateAdGroupExtensionSettingResult `protobuf:"bytes,2,rep,name=results,proto3" json:"results,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                               `json:"-"`
	XXX_unrecognized     []byte                                 `json:"-"`
	XXX_sizecache        int32                                  `json:"-"`
}

func (m *MutateAdGroupExtensionSettingsResponse) Reset() {
	*m = MutateAdGroupExtensionSettingsResponse{}
}
func (m *MutateAdGroupExtensionSettingsResponse) String() string { return proto.CompactTextString(m) }
func (*MutateAdGroupExtensionSettingsResponse) ProtoMessage()    {}
func (*MutateAdGroupExtensionSettingsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad_group_extension_setting_service_fd5e4f4a8f156262, []int{3}
}
func (m *MutateAdGroupExtensionSettingsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MutateAdGroupExtensionSettingsResponse.Unmarshal(m, b)
}
func (m *MutateAdGroupExtensionSettingsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MutateAdGroupExtensionSettingsResponse.Marshal(b, m, deterministic)
}
func (dst *MutateAdGroupExtensionSettingsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MutateAdGroupExtensionSettingsResponse.Merge(dst, src)
}
func (m *MutateAdGroupExtensionSettingsResponse) XXX_Size() int {
	return xxx_messageInfo_MutateAdGroupExtensionSettingsResponse.Size(m)
}
func (m *MutateAdGroupExtensionSettingsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MutateAdGroupExtensionSettingsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MutateAdGroupExtensionSettingsResponse proto.InternalMessageInfo

func (m *MutateAdGroupExtensionSettingsResponse) GetPartialFailureError() *status.Status {
	if m != nil {
		return m.PartialFailureError
	}
	return nil
}

func (m *MutateAdGroupExtensionSettingsResponse) GetResults() []*MutateAdGroupExtensionSettingResult {
	if m != nil {
		return m.Results
	}
	return nil
}

// The result for the ad group extension setting mutate.
type MutateAdGroupExtensionSettingResult struct {
	// Returned for successful operations.
	ResourceName         string   `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MutateAdGroupExtensionSettingResult) Reset()         { *m = MutateAdGroupExtensionSettingResult{} }
func (m *MutateAdGroupExtensionSettingResult) String() string { return proto.CompactTextString(m) }
func (*MutateAdGroupExtensionSettingResult) ProtoMessage()    {}
func (*MutateAdGroupExtensionSettingResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad_group_extension_setting_service_fd5e4f4a8f156262, []int{4}
}
func (m *MutateAdGroupExtensionSettingResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MutateAdGroupExtensionSettingResult.Unmarshal(m, b)
}
func (m *MutateAdGroupExtensionSettingResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MutateAdGroupExtensionSettingResult.Marshal(b, m, deterministic)
}
func (dst *MutateAdGroupExtensionSettingResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MutateAdGroupExtensionSettingResult.Merge(dst, src)
}
func (m *MutateAdGroupExtensionSettingResult) XXX_Size() int {
	return xxx_messageInfo_MutateAdGroupExtensionSettingResult.Size(m)
}
func (m *MutateAdGroupExtensionSettingResult) XXX_DiscardUnknown() {
	xxx_messageInfo_MutateAdGroupExtensionSettingResult.DiscardUnknown(m)
}

var xxx_messageInfo_MutateAdGroupExtensionSettingResult proto.InternalMessageInfo

func (m *MutateAdGroupExtensionSettingResult) GetResourceName() string {
	if m != nil {
		return m.ResourceName
	}
	return ""
}

func init() {
	proto.RegisterType((*GetAdGroupExtensionSettingRequest)(nil), "google.ads.googleads.v1.services.GetAdGroupExtensionSettingRequest")
	proto.RegisterType((*MutateAdGroupExtensionSettingsRequest)(nil), "google.ads.googleads.v1.services.MutateAdGroupExtensionSettingsRequest")
	proto.RegisterType((*AdGroupExtensionSettingOperation)(nil), "google.ads.googleads.v1.services.AdGroupExtensionSettingOperation")
	proto.RegisterType((*MutateAdGroupExtensionSettingsResponse)(nil), "google.ads.googleads.v1.services.MutateAdGroupExtensionSettingsResponse")
	proto.RegisterType((*MutateAdGroupExtensionSettingResult)(nil), "google.ads.googleads.v1.services.MutateAdGroupExtensionSettingResult")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// AdGroupExtensionSettingServiceClient is the client API for AdGroupExtensionSettingService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AdGroupExtensionSettingServiceClient interface {
	// Returns the requested ad group extension setting in full detail.
	GetAdGroupExtensionSetting(ctx context.Context, in *GetAdGroupExtensionSettingRequest, opts ...grpc.CallOption) (*resources.AdGroupExtensionSetting, error)
	// Creates, updates, or removes ad group extension settings. Operation
	// statuses are returned.
	MutateAdGroupExtensionSettings(ctx context.Context, in *MutateAdGroupExtensionSettingsRequest, opts ...grpc.CallOption) (*MutateAdGroupExtensionSettingsResponse, error)
}

type adGroupExtensionSettingServiceClient struct {
	cc *grpc.ClientConn
}

func NewAdGroupExtensionSettingServiceClient(cc *grpc.ClientConn) AdGroupExtensionSettingServiceClient {
	return &adGroupExtensionSettingServiceClient{cc}
}

func (c *adGroupExtensionSettingServiceClient) GetAdGroupExtensionSetting(ctx context.Context, in *GetAdGroupExtensionSettingRequest, opts ...grpc.CallOption) (*resources.AdGroupExtensionSetting, error) {
	out := new(resources.AdGroupExtensionSetting)
	err := c.cc.Invoke(ctx, "/google.ads.googleads.v1.services.AdGroupExtensionSettingService/GetAdGroupExtensionSetting", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adGroupExtensionSettingServiceClient) MutateAdGroupExtensionSettings(ctx context.Context, in *MutateAdGroupExtensionSettingsRequest, opts ...grpc.CallOption) (*MutateAdGroupExtensionSettingsResponse, error) {
	out := new(MutateAdGroupExtensionSettingsResponse)
	err := c.cc.Invoke(ctx, "/google.ads.googleads.v1.services.AdGroupExtensionSettingService/MutateAdGroupExtensionSettings", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AdGroupExtensionSettingServiceServer is the server API for AdGroupExtensionSettingService service.
type AdGroupExtensionSettingServiceServer interface {
	// Returns the requested ad group extension setting in full detail.
	GetAdGroupExtensionSetting(context.Context, *GetAdGroupExtensionSettingRequest) (*resources.AdGroupExtensionSetting, error)
	// Creates, updates, or removes ad group extension settings. Operation
	// statuses are returned.
	MutateAdGroupExtensionSettings(context.Context, *MutateAdGroupExtensionSettingsRequest) (*MutateAdGroupExtensionSettingsResponse, error)
}

func RegisterAdGroupExtensionSettingServiceServer(s *grpc.Server, srv AdGroupExtensionSettingServiceServer) {
	s.RegisterService(&_AdGroupExtensionSettingService_serviceDesc, srv)
}

func _AdGroupExtensionSettingService_GetAdGroupExtensionSetting_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAdGroupExtensionSettingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdGroupExtensionSettingServiceServer).GetAdGroupExtensionSetting(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.ads.googleads.v1.services.AdGroupExtensionSettingService/GetAdGroupExtensionSetting",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdGroupExtensionSettingServiceServer).GetAdGroupExtensionSetting(ctx, req.(*GetAdGroupExtensionSettingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdGroupExtensionSettingService_MutateAdGroupExtensionSettings_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MutateAdGroupExtensionSettingsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdGroupExtensionSettingServiceServer).MutateAdGroupExtensionSettings(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.ads.googleads.v1.services.AdGroupExtensionSettingService/MutateAdGroupExtensionSettings",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdGroupExtensionSettingServiceServer).MutateAdGroupExtensionSettings(ctx, req.(*MutateAdGroupExtensionSettingsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _AdGroupExtensionSettingService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "google.ads.googleads.v1.services.AdGroupExtensionSettingService",
	HandlerType: (*AdGroupExtensionSettingServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAdGroupExtensionSetting",
			Handler:    _AdGroupExtensionSettingService_GetAdGroupExtensionSetting_Handler,
		},
		{
			MethodName: "MutateAdGroupExtensionSettings",
			Handler:    _AdGroupExtensionSettingService_MutateAdGroupExtensionSettings_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "google/ads/googleads/v1/services/ad_group_extension_setting_service.proto",
}

func init() {
	proto.RegisterFile("google/ads/googleads/v1/services/ad_group_extension_setting_service.proto", fileDescriptor_ad_group_extension_setting_service_fd5e4f4a8f156262)
}

var fileDescriptor_ad_group_extension_setting_service_fd5e4f4a8f156262 = []byte{
	// 729 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x95, 0xc1, 0x6b, 0xd4, 0x4e,
	0x14, 0xc7, 0x7f, 0x49, 0x4b, 0x7f, 0x76, 0x52, 0x15, 0x46, 0xc4, 0x65, 0x91, 0xba, 0xa6, 0x55,
	0xcb, 0x1e, 0x12, 0x76, 0xbd, 0xa5, 0xf6, 0xb0, 0x91, 0x76, 0xb7, 0x42, 0x6d, 0x49, 0xa5, 0x07,
	0x59, 0x08, 0xd3, 0xcd, 0x34, 0x84, 0x26, 0x99, 0x38, 0x33, 0x59, 0x2d, 0xa5, 0x17, 0xf1, 0xe4,
	0xd5, 0xff, 0xc0, 0xa3, 0x47, 0xff, 0x0c, 0x6f, 0xe2, 0x5f, 0x20, 0x78, 0xf2, 0x4f, 0x10, 0x04,
	0x99, 0x4c, 0x66, 0x6d, 0x0b, 0xd9, 0x14, 0xd6, 0xdb, 0xdb, 0x37, 0xdf, 0xfd, 0xbc, 0xf7, 0xe6,
	0xbd, 0x79, 0x01, 0xdb, 0x21, 0x21, 0x61, 0x8c, 0x6d, 0x14, 0x30, 0x5b, 0x9a, 0xc2, 0x1a, 0x77,
	0x6c, 0x86, 0xe9, 0x38, 0x1a, 0x61, 0x66, 0xa3, 0xc0, 0x0f, 0x29, 0xc9, 0x33, 0x1f, 0xbf, 0xe1,
	0x38, 0x65, 0x11, 0x49, 0x7d, 0x86, 0x39, 0x8f, 0xd2, 0xd0, 0x2f, 0x35, 0x56, 0x46, 0x09, 0x27,
	0xb0, 0x25, 0xff, 0x6f, 0xa1, 0x80, 0x59, 0x13, 0x94, 0x35, 0xee, 0x58, 0x0a, 0xd5, 0x74, 0xab,
	0x82, 0x51, 0xcc, 0x48, 0x4e, 0xa7, 0x47, 0x93, 0x51, 0x9a, 0x77, 0x15, 0x23, 0x8b, 0x6c, 0x94,
	0xa6, 0x84, 0x23, 0x1e, 0x91, 0x94, 0x95, 0xa7, 0x65, 0x0e, 0x76, 0xf1, 0xeb, 0x30, 0x3f, 0xb2,
	0x8f, 0x22, 0x1c, 0x07, 0x7e, 0x82, 0xd8, 0x71, 0xa9, 0x58, 0xbe, 0xac, 0x78, 0x4d, 0x51, 0x96,
	0x61, 0xaa, 0x08, 0x77, 0xca, 0x73, 0x9a, 0x8d, 0x6c, 0xc6, 0x11, 0xcf, 0xcb, 0x03, 0x73, 0x00,
	0xee, 0xf7, 0x31, 0xef, 0x05, 0x7d, 0x91, 0xde, 0xa6, 0xca, 0x6e, 0x5f, 0x26, 0xe7, 0xe1, 0x57,
	0x39, 0x66, 0x1c, 0xae, 0x80, 0xeb, 0xaa, 0x16, 0x3f, 0x45, 0x09, 0x6e, 0x68, 0x2d, 0x6d, 0x6d,
	0xd1, 0x5b, 0x52, 0xce, 0xe7, 0x28, 0xc1, 0xe6, 0x2f, 0x0d, 0x3c, 0xd8, 0xc9, 0x39, 0xe2, 0xb8,
	0x82, 0xc6, 0x14, 0xee, 0x1e, 0x30, 0x46, 0x39, 0xe3, 0x24, 0xc1, 0xd4, 0x8f, 0x82, 0x12, 0x06,
	0x94, 0x6b, 0x3b, 0x80, 0x87, 0x00, 0x90, 0x0c, 0x53, 0x79, 0x07, 0x0d, 0xbd, 0x35, 0xb7, 0x66,
	0x74, 0x5d, 0xab, 0xae, 0x11, 0x56, 0x45, 0xdc, 0x5d, 0x85, 0xf2, 0xce, 0x51, 0xe1, 0x23, 0x70,
	0x33, 0x43, 0x94, 0x47, 0x28, 0xf6, 0x8f, 0x50, 0x14, 0xe7, 0x14, 0x37, 0xe6, 0x5a, 0xda, 0xda,
	0x35, 0xef, 0x46, 0xe9, 0xde, 0x92, 0x5e, 0x51, 0xfc, 0x18, 0xc5, 0x51, 0x80, 0x38, 0xf6, 0x49,
	0x1a, 0x9f, 0x34, 0xe6, 0x0b, 0xd9, 0x92, 0x72, 0xee, 0xa6, 0xf1, 0x89, 0xf9, 0x59, 0x07, 0xad,
	0xba, 0xf0, 0x70, 0x1d, 0x18, 0x79, 0x56, 0x70, 0x44, 0xe7, 0x0a, 0x8e, 0xd1, 0x6d, 0xaa, 0xba,
	0x54, 0xeb, 0xac, 0x2d, 0xd1, 0xdc, 0x1d, 0xc4, 0x8e, 0x3d, 0x20, 0xe5, 0xc2, 0x86, 0x2f, 0xc0,
	0xc2, 0x88, 0x62, 0xc4, 0xe5, 0xe5, 0x1b, 0x5d, 0xa7, 0xf2, 0x3e, 0x26, 0x63, 0x57, 0x75, 0x21,
	0x83, 0xff, 0xbc, 0x92, 0x25, 0xa8, 0x32, 0x46, 0x43, 0xff, 0x17, 0x54, 0xc9, 0x82, 0x0d, 0xb0,
	0x40, 0x71, 0x42, 0xc6, 0xf2, 0x4a, 0x17, 0xc5, 0x89, 0xfc, 0xed, 0x1a, 0x60, 0x71, 0xd2, 0x03,
	0xf3, 0xab, 0x06, 0x1e, 0xd6, 0x4d, 0x0c, 0xcb, 0x48, 0xca, 0x30, 0xdc, 0x02, 0xb7, 0x2f, 0x75,
	0xcb, 0xc7, 0x94, 0x12, 0x5a, 0x04, 0x30, 0xba, 0x50, 0xa5, 0x4d, 0xb3, 0x91, 0xb5, 0x5f, 0xcc,
	0xb7, 0x77, 0xeb, 0x62, 0x1f, 0x37, 0x85, 0x1c, 0xfa, 0xe0, 0x7f, 0x8a, 0x59, 0x1e, 0x73, 0x35,
	0x56, 0x9b, 0xf5, 0x63, 0x35, 0x35, 0x45, 0xaf, 0xa0, 0x79, 0x8a, 0x6a, 0x3e, 0x03, 0x2b, 0x57,
	0xd0, 0x5f, 0xe9, 0x45, 0x75, 0xdf, 0xcd, 0x83, 0xe5, 0x0a, 0xcc, 0xbe, 0x4c, 0x0e, 0x7e, 0xd7,
	0x40, 0xb3, 0xfa, 0xfd, 0xc2, 0xa7, 0xf5, 0xd5, 0xd5, 0xbe, 0xfe, 0xe6, 0x0c, 0x33, 0x61, 0xba,
	0x6f, 0xbf, 0xfd, 0xf8, 0xa0, 0x3f, 0x81, 0x8e, 0xd8, 0x87, 0xa7, 0x17, 0x4a, 0xde, 0x50, 0x0f,
	0x9e, 0xd9, 0x6d, 0x1b, 0x55, 0x0c, 0x80, 0xdd, 0x3e, 0x83, 0xbf, 0x35, 0xb0, 0x3c, 0x7d, 0x4c,
	0x60, 0x7f, 0xc6, 0x2e, 0xaa, 0xd5, 0xd4, 0x1c, 0xcc, 0x0e, 0x92, 0x13, 0x6b, 0x0e, 0x8a, 0xca,
	0x5d, 0x73, 0x43, 0x54, 0xfe, 0xb7, 0xd4, 0xd3, 0x73, 0x9b, 0x6f, 0xa3, 0x7d, 0x56, 0x59, 0xb8,
	0x93, 0x14, 0x61, 0x1c, 0xad, 0xed, 0xbe, 0xd7, 0xc1, 0xea, 0x88, 0x24, 0xb5, 0x99, 0xb9, 0x2b,
	0xd3, 0x87, 0x65, 0x4f, 0x2c, 0x98, 0x3d, 0xed, 0xe5, 0xa0, 0x04, 0x85, 0x24, 0x46, 0x69, 0x68,
	0x11, 0x1a, 0xda, 0x21, 0x4e, 0x8b, 0xf5, 0xa3, 0xbe, 0x5f, 0x59, 0xc4, 0xaa, 0xbf, 0x9d, 0xeb,
	0xca, 0xf8, 0xa8, 0xcf, 0xf5, 0x7b, 0xbd, 0x4f, 0x7a, 0xab, 0x2f, 0x81, 0xbd, 0x80, 0x59, 0xd2,
	0x14, 0xd6, 0x41, 0xc7, 0x2a, 0x03, 0xb3, 0x2f, 0x4a, 0x32, 0xec, 0x05, 0x6c, 0x38, 0x91, 0x0c,
	0x0f, 0x3a, 0x43, 0x25, 0xf9, 0xa9, 0xaf, 0x4a, 0xbf, 0xe3, 0xf4, 0x02, 0xe6, 0x38, 0x13, 0x91,
	0xe3, 0x1c, 0x74, 0x1c, 0x47, 0xc9, 0x0e, 0x17, 0x8a, 0x3c, 0x1f, 0xff, 0x09, 0x00, 0x00, 0xff,
	0xff, 0xbb, 0x5c, 0x83, 0x6c, 0xe2, 0x07, 0x00, 0x00,
}
