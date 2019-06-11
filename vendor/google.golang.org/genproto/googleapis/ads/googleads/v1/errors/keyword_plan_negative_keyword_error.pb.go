// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v1/errors/keyword_plan_negative_keyword_error.proto

package errors // import "google.golang.org/genproto/googleapis/ads/googleads/v1/errors"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/annotations"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Enum describing possible errors from applying a keyword plan negative
// keyword.
type KeywordPlanNegativeKeywordErrorEnum_KeywordPlanNegativeKeywordError int32

const (
	// Enum unspecified.
	KeywordPlanNegativeKeywordErrorEnum_UNSPECIFIED KeywordPlanNegativeKeywordErrorEnum_KeywordPlanNegativeKeywordError = 0
	// The received error code is not known in this version.
	KeywordPlanNegativeKeywordErrorEnum_UNKNOWN KeywordPlanNegativeKeywordErrorEnum_KeywordPlanNegativeKeywordError = 1
)

var KeywordPlanNegativeKeywordErrorEnum_KeywordPlanNegativeKeywordError_name = map[int32]string{
	0: "UNSPECIFIED",
	1: "UNKNOWN",
}
var KeywordPlanNegativeKeywordErrorEnum_KeywordPlanNegativeKeywordError_value = map[string]int32{
	"UNSPECIFIED": 0,
	"UNKNOWN":     1,
}

func (x KeywordPlanNegativeKeywordErrorEnum_KeywordPlanNegativeKeywordError) String() string {
	return proto.EnumName(KeywordPlanNegativeKeywordErrorEnum_KeywordPlanNegativeKeywordError_name, int32(x))
}
func (KeywordPlanNegativeKeywordErrorEnum_KeywordPlanNegativeKeywordError) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_keyword_plan_negative_keyword_error_64e283750637729c, []int{0, 0}
}

// Container for enum describing possible errors from applying a keyword plan
// negative keyword.
type KeywordPlanNegativeKeywordErrorEnum struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KeywordPlanNegativeKeywordErrorEnum) Reset()         { *m = KeywordPlanNegativeKeywordErrorEnum{} }
func (m *KeywordPlanNegativeKeywordErrorEnum) String() string { return proto.CompactTextString(m) }
func (*KeywordPlanNegativeKeywordErrorEnum) ProtoMessage()    {}
func (*KeywordPlanNegativeKeywordErrorEnum) Descriptor() ([]byte, []int) {
	return fileDescriptor_keyword_plan_negative_keyword_error_64e283750637729c, []int{0}
}
func (m *KeywordPlanNegativeKeywordErrorEnum) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeywordPlanNegativeKeywordErrorEnum.Unmarshal(m, b)
}
func (m *KeywordPlanNegativeKeywordErrorEnum) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeywordPlanNegativeKeywordErrorEnum.Marshal(b, m, deterministic)
}
func (dst *KeywordPlanNegativeKeywordErrorEnum) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeywordPlanNegativeKeywordErrorEnum.Merge(dst, src)
}
func (m *KeywordPlanNegativeKeywordErrorEnum) XXX_Size() int {
	return xxx_messageInfo_KeywordPlanNegativeKeywordErrorEnum.Size(m)
}
func (m *KeywordPlanNegativeKeywordErrorEnum) XXX_DiscardUnknown() {
	xxx_messageInfo_KeywordPlanNegativeKeywordErrorEnum.DiscardUnknown(m)
}

var xxx_messageInfo_KeywordPlanNegativeKeywordErrorEnum proto.InternalMessageInfo

func init() {
	proto.RegisterType((*KeywordPlanNegativeKeywordErrorEnum)(nil), "google.ads.googleads.v1.errors.KeywordPlanNegativeKeywordErrorEnum")
	proto.RegisterEnum("google.ads.googleads.v1.errors.KeywordPlanNegativeKeywordErrorEnum_KeywordPlanNegativeKeywordError", KeywordPlanNegativeKeywordErrorEnum_KeywordPlanNegativeKeywordError_name, KeywordPlanNegativeKeywordErrorEnum_KeywordPlanNegativeKeywordError_value)
}

func init() {
	proto.RegisterFile("google/ads/googleads/v1/errors/keyword_plan_negative_keyword_error.proto", fileDescriptor_keyword_plan_negative_keyword_error_64e283750637729c)
}

var fileDescriptor_keyword_plan_negative_keyword_error_64e283750637729c = []byte{
	// 297 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x90, 0xc1, 0x4a, 0xf4, 0x30,
	0x14, 0x85, 0xff, 0xf6, 0x07, 0x85, 0xcc, 0xc2, 0xa1, 0x4b, 0x91, 0x11, 0xaa, 0xeb, 0x84, 0xe2,
	0x2e, 0x2e, 0xa4, 0xe3, 0xd4, 0x71, 0x18, 0xa8, 0x05, 0x99, 0x0a, 0x52, 0x28, 0xd1, 0xc4, 0x50,
	0xec, 0x24, 0x25, 0xa9, 0x15, 0x5f, 0xc7, 0xa5, 0x8f, 0xe2, 0xa3, 0xf8, 0x12, 0x4a, 0x7b, 0xdb,
	0xee, 0x74, 0x56, 0x39, 0x5c, 0xbe, 0x7b, 0xce, 0xc9, 0x45, 0xd7, 0x52, 0x6b, 0x59, 0x0a, 0xc2,
	0xb8, 0x25, 0x20, 0x5b, 0xd5, 0x04, 0x44, 0x18, 0xa3, 0x8d, 0x25, 0xcf, 0xe2, 0xed, 0x55, 0x1b,
	0x9e, 0x57, 0x25, 0x53, 0xb9, 0x12, 0x92, 0xd5, 0x45, 0x23, 0xf2, 0x61, 0xda, 0x41, 0xb8, 0x32,
	0xba, 0xd6, 0xde, 0x0c, 0xd6, 0x31, 0xe3, 0x16, 0x8f, 0x4e, 0xb8, 0x09, 0x30, 0x38, 0x1d, 0x1e,
	0x0d, 0x49, 0x55, 0x41, 0x98, 0x52, 0xba, 0x66, 0x75, 0xa1, 0x95, 0x85, 0x6d, 0xff, 0x09, 0x9d,
	0xac, 0xc1, 0x34, 0x29, 0x99, 0x8a, 0xfb, 0xa0, 0x7e, 0x14, 0xb5, 0x0e, 0x91, 0x7a, 0xd9, 0xfa,
	0x17, 0xe8, 0x78, 0x07, 0xe6, 0x1d, 0xa0, 0xc9, 0x26, 0xbe, 0x4d, 0xa2, 0xcb, 0xd5, 0xd5, 0x2a,
	0x5a, 0x4c, 0xff, 0x79, 0x13, 0xb4, 0xbf, 0x89, 0xd7, 0xf1, 0xcd, 0x5d, 0x3c, 0x75, 0xe6, 0xdf,
	0x0e, 0xf2, 0x1f, 0xf5, 0x16, 0xff, 0x5d, 0x76, 0x7e, 0xba, 0x23, 0x25, 0x69, 0x4b, 0x27, 0xce,
	0xfd, 0xa2, 0xf7, 0x91, 0xba, 0x64, 0x4a, 0x62, 0x6d, 0x24, 0x91, 0x42, 0x75, 0x5f, 0x1a, 0xce,
	0x59, 0x15, 0xf6, 0xb7, 0xeb, 0x9e, 0xc3, 0xf3, 0xee, 0xfe, 0x5f, 0x86, 0xe1, 0x87, 0x3b, 0x5b,
	0x82, 0x59, 0xc8, 0x2d, 0x06, 0xd9, 0xaa, 0x34, 0xc0, 0x5d, 0xa4, 0xfd, 0x1c, 0x80, 0x2c, 0xe4,
	0x36, 0x1b, 0x81, 0x2c, 0x0d, 0x32, 0x00, 0xbe, 0x5c, 0x1f, 0xa6, 0x94, 0x86, 0xdc, 0x52, 0x3a,
	0x22, 0x94, 0xa6, 0x01, 0xa5, 0x00, 0x3d, 0xec, 0x75, 0xed, 0xce, 0x7e, 0x02, 0x00, 0x00, 0xff,
	0xff, 0x6b, 0xa8, 0xb2, 0xa0, 0xfa, 0x01, 0x00, 0x00,
}
