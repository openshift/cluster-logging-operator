// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/cloud/ml/v1/project_service.proto

package ml // import "google.golang.org/genproto/googleapis/cloud/ml/v1"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/annotations"

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

// Requests service account information associated with a project.
type GetConfigRequest struct {
	// Required. The project name.
	//
	// Authorization: requires `Viewer` role on the specified project.
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetConfigRequest) Reset()         { *m = GetConfigRequest{} }
func (m *GetConfigRequest) String() string { return proto.CompactTextString(m) }
func (*GetConfigRequest) ProtoMessage()    {}
func (*GetConfigRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_project_service_81d7c159e503bebf, []int{0}
}
func (m *GetConfigRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetConfigRequest.Unmarshal(m, b)
}
func (m *GetConfigRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetConfigRequest.Marshal(b, m, deterministic)
}
func (dst *GetConfigRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetConfigRequest.Merge(dst, src)
}
func (m *GetConfigRequest) XXX_Size() int {
	return xxx_messageInfo_GetConfigRequest.Size(m)
}
func (m *GetConfigRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetConfigRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetConfigRequest proto.InternalMessageInfo

func (m *GetConfigRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

// Returns service account information associated with a project.
type GetConfigResponse struct {
	// The service account Cloud ML uses to access resources in the project.
	ServiceAccount string `protobuf:"bytes,1,opt,name=service_account,json=serviceAccount,proto3" json:"service_account,omitempty"`
	// The project number for `service_account`.
	ServiceAccountProject int64    `protobuf:"varint,2,opt,name=service_account_project,json=serviceAccountProject,proto3" json:"service_account_project,omitempty"`
	XXX_NoUnkeyedLiteral  struct{} `json:"-"`
	XXX_unrecognized      []byte   `json:"-"`
	XXX_sizecache         int32    `json:"-"`
}

func (m *GetConfigResponse) Reset()         { *m = GetConfigResponse{} }
func (m *GetConfigResponse) String() string { return proto.CompactTextString(m) }
func (*GetConfigResponse) ProtoMessage()    {}
func (*GetConfigResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_project_service_81d7c159e503bebf, []int{1}
}
func (m *GetConfigResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetConfigResponse.Unmarshal(m, b)
}
func (m *GetConfigResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetConfigResponse.Marshal(b, m, deterministic)
}
func (dst *GetConfigResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetConfigResponse.Merge(dst, src)
}
func (m *GetConfigResponse) XXX_Size() int {
	return xxx_messageInfo_GetConfigResponse.Size(m)
}
func (m *GetConfigResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetConfigResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetConfigResponse proto.InternalMessageInfo

func (m *GetConfigResponse) GetServiceAccount() string {
	if m != nil {
		return m.ServiceAccount
	}
	return ""
}

func (m *GetConfigResponse) GetServiceAccountProject() int64 {
	if m != nil {
		return m.ServiceAccountProject
	}
	return 0
}

func init() {
	proto.RegisterType((*GetConfigRequest)(nil), "google.cloud.ml.v1.GetConfigRequest")
	proto.RegisterType((*GetConfigResponse)(nil), "google.cloud.ml.v1.GetConfigResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ProjectManagementServiceClient is the client API for ProjectManagementService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ProjectManagementServiceClient interface {
	// Get the service account information associated with your project. You need
	// this information in order to grant the service account persmissions for
	// the Google Cloud Storage location where you put your model training code
	// for training the model with Google Cloud Machine Learning.
	GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error)
}

type projectManagementServiceClient struct {
	cc *grpc.ClientConn
}

func NewProjectManagementServiceClient(cc *grpc.ClientConn) ProjectManagementServiceClient {
	return &projectManagementServiceClient{cc}
}

func (c *projectManagementServiceClient) GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error) {
	out := new(GetConfigResponse)
	err := c.cc.Invoke(ctx, "/google.cloud.ml.v1.ProjectManagementService/GetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProjectManagementServiceServer is the server API for ProjectManagementService service.
type ProjectManagementServiceServer interface {
	// Get the service account information associated with your project. You need
	// this information in order to grant the service account persmissions for
	// the Google Cloud Storage location where you put your model training code
	// for training the model with Google Cloud Machine Learning.
	GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error)
}

func RegisterProjectManagementServiceServer(s *grpc.Server, srv ProjectManagementServiceServer) {
	s.RegisterService(&_ProjectManagementService_serviceDesc, srv)
}

func _ProjectManagementService_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProjectManagementServiceServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.cloud.ml.v1.ProjectManagementService/GetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProjectManagementServiceServer).GetConfig(ctx, req.(*GetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ProjectManagementService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "google.cloud.ml.v1.ProjectManagementService",
	HandlerType: (*ProjectManagementServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetConfig",
			Handler:    _ProjectManagementService_GetConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "google/cloud/ml/v1/project_service.proto",
}

func init() {
	proto.RegisterFile("google/cloud/ml/v1/project_service.proto", fileDescriptor_project_service_81d7c159e503bebf)
}

var fileDescriptor_project_service_81d7c159e503bebf = []byte{
	// 319 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x91, 0xbf, 0x4a, 0x43, 0x31,
	0x14, 0xc6, 0xb9, 0x55, 0x84, 0x66, 0xf0, 0x4f, 0x44, 0x2c, 0x45, 0xb0, 0x16, 0xb5, 0xc5, 0x21,
	0xa1, 0x2a, 0x0e, 0x8a, 0x83, 0x75, 0x70, 0x12, 0x4a, 0xdd, 0x5c, 0x4a, 0xbc, 0x1e, 0x43, 0x24,
	0xc9, 0x89, 0x37, 0xe9, 0x5d, 0xc4, 0x41, 0x5f, 0xc1, 0xdd, 0x97, 0xf2, 0x15, 0x7c, 0x10, 0xe9,
	0x4d, 0x94, 0xda, 0x0e, 0x6e, 0x87, 0x73, 0x7e, 0x5f, 0xf2, 0x7d, 0xe7, 0x90, 0xae, 0x44, 0x94,
	0x1a, 0x78, 0xae, 0x71, 0x7c, 0xcf, 0x8d, 0xe6, 0x65, 0x8f, 0xbb, 0x02, 0x1f, 0x21, 0x0f, 0x23,
	0x0f, 0x45, 0xa9, 0x72, 0x60, 0xae, 0xc0, 0x80, 0x94, 0x46, 0x92, 0x55, 0x24, 0x33, 0x9a, 0x95,
	0xbd, 0xe6, 0x56, 0x52, 0x0b, 0xa7, 0xb8, 0xb0, 0x16, 0x83, 0x08, 0x0a, 0xad, 0x8f, 0x8a, 0xf6,
	0x3e, 0x59, 0xbd, 0x82, 0x70, 0x89, 0xf6, 0x41, 0xc9, 0x21, 0x3c, 0x8d, 0xc1, 0x07, 0x4a, 0xc9,
	0xa2, 0x15, 0x06, 0x1a, 0x59, 0x2b, 0xeb, 0xd6, 0x87, 0x55, 0xdd, 0x0e, 0x64, 0x6d, 0x8a, 0xf3,
	0x0e, 0xad, 0x07, 0xda, 0x21, 0x2b, 0xe9, 0xff, 0x91, 0xc8, 0x73, 0x1c, 0xdb, 0x90, 0x34, 0xcb,
	0xa9, 0x7d, 0x11, 0xbb, 0xf4, 0x84, 0x6c, 0xce, 0x80, 0xa3, 0x14, 0xa0, 0x51, 0x6b, 0x65, 0xdd,
	0x85, 0xe1, 0xc6, 0x5f, 0xc1, 0x20, 0x0e, 0x0f, 0x3f, 0x32, 0xd2, 0x48, 0xf5, 0xb5, 0xb0, 0x42,
	0x82, 0x01, 0x1b, 0x6e, 0x22, 0x4a, 0x5f, 0x33, 0x52, 0xff, 0xf5, 0x44, 0x77, 0xd9, 0x7c, 0x76,
	0x36, 0x1b, 0xad, 0xb9, 0xf7, 0x0f, 0x15, 0x83, 0xb5, 0x3b, 0x6f, 0x9f, 0x5f, 0xef, 0xb5, 0x1d,
	0xba, 0x3d, 0x59, 0xf5, 0xf3, 0x64, 0x01, 0xe7, 0xc9, 0xaf, 0xe7, 0x07, 0x2f, 0xa7, 0xf2, 0x47,
	0xd0, 0x57, 0xa4, 0x99, 0xa3, 0x99, 0x7b, 0x54, 0x38, 0xc5, 0xca, 0x5e, 0x7f, 0x3d, 0x79, 0x4f,
	0x8e, 0x07, 0x93, 0x8d, 0x0f, 0xb2, 0xdb, 0xe3, 0x84, 0x4b, 0xd4, 0xc2, 0x4a, 0x86, 0x85, 0xe4,
	0x12, 0x6c, 0x75, 0x0f, 0x1e, 0x47, 0xc2, 0x29, 0x3f, 0x7d, 0xee, 0x33, 0xa3, 0xef, 0x96, 0x2a,
	0xe0, 0xe8, 0x3b, 0x00, 0x00, 0xff, 0xff, 0xd0, 0xa5, 0x43, 0x33, 0x0e, 0x02, 0x00, 0x00,
}
