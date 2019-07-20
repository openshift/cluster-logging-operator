// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v2/enums/product_condition.proto

package enums

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Enum describing the condition of a product offer.
type ProductConditionEnum_ProductCondition int32

const (
	// Not specified.
	ProductConditionEnum_UNSPECIFIED ProductConditionEnum_ProductCondition = 0
	// Used for return value only. Represents value unknown in this version.
	ProductConditionEnum_UNKNOWN ProductConditionEnum_ProductCondition = 1
	// The product condition is new.
	ProductConditionEnum_NEW ProductConditionEnum_ProductCondition = 3
	// The product condition is refurbished.
	ProductConditionEnum_REFURBISHED ProductConditionEnum_ProductCondition = 4
	// The product condition is used.
	ProductConditionEnum_USED ProductConditionEnum_ProductCondition = 5
)

var ProductConditionEnum_ProductCondition_name = map[int32]string{
	0: "UNSPECIFIED",
	1: "UNKNOWN",
	3: "NEW",
	4: "REFURBISHED",
	5: "USED",
}

var ProductConditionEnum_ProductCondition_value = map[string]int32{
	"UNSPECIFIED": 0,
	"UNKNOWN":     1,
	"NEW":         3,
	"REFURBISHED": 4,
	"USED":        5,
}

func (x ProductConditionEnum_ProductCondition) String() string {
	return proto.EnumName(ProductConditionEnum_ProductCondition_name, int32(x))
}

func (ProductConditionEnum_ProductCondition) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_bd718a79cf1e8a2d, []int{0, 0}
}

// Condition of a product offer.
type ProductConditionEnum struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProductConditionEnum) Reset()         { *m = ProductConditionEnum{} }
func (m *ProductConditionEnum) String() string { return proto.CompactTextString(m) }
func (*ProductConditionEnum) ProtoMessage()    {}
func (*ProductConditionEnum) Descriptor() ([]byte, []int) {
	return fileDescriptor_bd718a79cf1e8a2d, []int{0}
}

func (m *ProductConditionEnum) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProductConditionEnum.Unmarshal(m, b)
}
func (m *ProductConditionEnum) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProductConditionEnum.Marshal(b, m, deterministic)
}
func (m *ProductConditionEnum) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProductConditionEnum.Merge(m, src)
}
func (m *ProductConditionEnum) XXX_Size() int {
	return xxx_messageInfo_ProductConditionEnum.Size(m)
}
func (m *ProductConditionEnum) XXX_DiscardUnknown() {
	xxx_messageInfo_ProductConditionEnum.DiscardUnknown(m)
}

var xxx_messageInfo_ProductConditionEnum proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("google.ads.googleads.v2.enums.ProductConditionEnum_ProductCondition", ProductConditionEnum_ProductCondition_name, ProductConditionEnum_ProductCondition_value)
	proto.RegisterType((*ProductConditionEnum)(nil), "google.ads.googleads.v2.enums.ProductConditionEnum")
}

func init() {
	proto.RegisterFile("google/ads/googleads/v2/enums/product_condition.proto", fileDescriptor_bd718a79cf1e8a2d)
}

var fileDescriptor_bd718a79cf1e8a2d = []byte{
	// 308 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x50, 0xcf, 0x4a, 0xc3, 0x30,
	0x18, 0x77, 0x7f, 0x74, 0x92, 0x1d, 0x0c, 0x45, 0x2f, 0xe2, 0x0e, 0xdb, 0x03, 0x24, 0x50, 0xf1,
	0x12, 0x4f, 0xed, 0x96, 0xcd, 0x21, 0xd4, 0xb2, 0xd9, 0x0d, 0xa4, 0x20, 0x75, 0x19, 0xa1, 0xd0,
	0x26, 0xa5, 0x69, 0xf7, 0x40, 0x1e, 0x7d, 0x14, 0x1f, 0x65, 0x4f, 0x21, 0x49, 0x6c, 0x0f, 0x03,
	0xbd, 0x84, 0x1f, 0xdf, 0xef, 0x4f, 0x7e, 0xdf, 0x07, 0x1e, 0xb8, 0x94, 0x3c, 0xdb, 0xe3, 0x84,
	0x29, 0x6c, 0xa1, 0x46, 0x07, 0x17, 0xef, 0x45, 0x9d, 0x2b, 0x5c, 0x94, 0x92, 0xd5, 0xbb, 0xea,
	0x7d, 0x27, 0x05, 0x4b, 0xab, 0x54, 0x0a, 0x54, 0x94, 0xb2, 0x92, 0xce, 0xc8, 0x6a, 0x51, 0xc2,
	0x14, 0x6a, 0x6d, 0xe8, 0xe0, 0x22, 0x63, 0xbb, 0xbd, 0x6b, 0x52, 0x8b, 0x14, 0x27, 0x42, 0xc8,
	0x2a, 0xd1, 0x5e, 0x65, 0xcd, 0x93, 0x0c, 0x5c, 0x87, 0x36, 0x77, 0xda, 0xc4, 0x52, 0x51, 0xe7,
	0x93, 0x57, 0x00, 0x4f, 0xe7, 0xce, 0x15, 0x18, 0x46, 0xc1, 0x3a, 0xa4, 0xd3, 0xe5, 0x7c, 0x49,
	0x67, 0xf0, 0xcc, 0x19, 0x82, 0x41, 0x14, 0x3c, 0x07, 0x2f, 0xdb, 0x00, 0x76, 0x9c, 0x01, 0xe8,
	0x05, 0x74, 0x0b, 0x7b, 0x5a, 0xb6, 0xa2, 0xf3, 0x68, 0xe5, 0x2f, 0xd7, 0x4f, 0x74, 0x06, 0xfb,
	0xce, 0x25, 0xe8, 0x47, 0x6b, 0x3a, 0x83, 0xe7, 0xfe, 0xb1, 0x03, 0xc6, 0x3b, 0x99, 0xa3, 0x7f,
	0x1b, 0xfb, 0x37, 0xa7, 0x3f, 0x87, 0xba, 0x6a, 0xd8, 0x79, 0xf3, 0x7f, 0x7d, 0x5c, 0x66, 0x89,
	0xe0, 0x48, 0x96, 0x1c, 0xf3, 0xbd, 0x30, 0x8b, 0x34, 0x07, 0x2b, 0x52, 0xf5, 0xc7, 0xfd, 0x1e,
	0xcd, 0xfb, 0xd9, 0xed, 0x2d, 0x3c, 0xef, 0xab, 0x3b, 0x5a, 0xd8, 0x28, 0x8f, 0x29, 0x64, 0xa1,
	0x46, 0x1b, 0x17, 0xe9, 0xed, 0xd5, 0x77, 0xc3, 0xc7, 0x1e, 0x53, 0x71, 0xcb, 0xc7, 0x1b, 0x37,
	0x36, 0xfc, 0xb1, 0x3b, 0xb6, 0x43, 0x42, 0x3c, 0xa6, 0x08, 0x69, 0x15, 0x84, 0x6c, 0x5c, 0x42,
	0x8c, 0xe6, 0xe3, 0xc2, 0x14, 0xbb, 0xff, 0x09, 0x00, 0x00, 0xff, 0xff, 0xc9, 0xb3, 0x1a, 0xaf,
	0xd7, 0x01, 0x00, 0x00,
}
