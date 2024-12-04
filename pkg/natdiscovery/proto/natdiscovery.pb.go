//
//SPDX-License-Identifier: Apache-2.0
//
//Copyright Contributors to the Submariner project.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v3.19.6
// source: pkg/natdiscovery/proto/natdiscovery.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ResponseType int32

const (
	ResponseType_OK                   ResponseType = 0
	ResponseType_NAT_DETECTED         ResponseType = 1
	ResponseType_UNKNOWN_DST_CLUSTER  ResponseType = 2
	ResponseType_UNKNOWN_DST_ENDPOINT ResponseType = 3
	ResponseType_MALFORMED            ResponseType = 4
)

// Enum value maps for ResponseType.
var (
	ResponseType_name = map[int32]string{
		0: "OK",
		1: "NAT_DETECTED",
		2: "UNKNOWN_DST_CLUSTER",
		3: "UNKNOWN_DST_ENDPOINT",
		4: "MALFORMED",
	}
	ResponseType_value = map[string]int32{
		"OK":                   0,
		"NAT_DETECTED":         1,
		"UNKNOWN_DST_CLUSTER":  2,
		"UNKNOWN_DST_ENDPOINT": 3,
		"MALFORMED":            4,
	}
)

func (x ResponseType) Enum() *ResponseType {
	p := new(ResponseType)
	*p = x
	return p
}

func (x ResponseType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ResponseType) Descriptor() protoreflect.EnumDescriptor {
	return file_pkg_natdiscovery_proto_natdiscovery_proto_enumTypes[0].Descriptor()
}

func (ResponseType) Type() protoreflect.EnumType {
	return &file_pkg_natdiscovery_proto_natdiscovery_proto_enumTypes[0]
}

func (x ResponseType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ResponseType.Descriptor instead.
func (ResponseType) EnumDescriptor() ([]byte, []int) {
	return file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescGZIP(), []int{0}
}

type SubmarinerNATDiscoveryMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version int32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	// Types that are assignable to Message:
	//
	//	*SubmarinerNATDiscoveryMessage_Request
	//	*SubmarinerNATDiscoveryMessage_Response
	Message isSubmarinerNATDiscoveryMessage_Message `protobuf_oneof:"message"`
}

func (x *SubmarinerNATDiscoveryMessage) Reset() {
	*x = SubmarinerNATDiscoveryMessage{}
	mi := &file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmarinerNATDiscoveryMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmarinerNATDiscoveryMessage) ProtoMessage() {}

func (x *SubmarinerNATDiscoveryMessage) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubmarinerNATDiscoveryMessage.ProtoReflect.Descriptor instead.
func (*SubmarinerNATDiscoveryMessage) Descriptor() ([]byte, []int) {
	return file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescGZIP(), []int{0}
}

func (x *SubmarinerNATDiscoveryMessage) GetVersion() int32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (m *SubmarinerNATDiscoveryMessage) GetMessage() isSubmarinerNATDiscoveryMessage_Message {
	if m != nil {
		return m.Message
	}
	return nil
}

func (x *SubmarinerNATDiscoveryMessage) GetRequest() *SubmarinerNATDiscoveryRequest {
	if x, ok := x.GetMessage().(*SubmarinerNATDiscoveryMessage_Request); ok {
		return x.Request
	}
	return nil
}

func (x *SubmarinerNATDiscoveryMessage) GetResponse() *SubmarinerNATDiscoveryResponse {
	if x, ok := x.GetMessage().(*SubmarinerNATDiscoveryMessage_Response); ok {
		return x.Response
	}
	return nil
}

type isSubmarinerNATDiscoveryMessage_Message interface {
	isSubmarinerNATDiscoveryMessage_Message()
}

type SubmarinerNATDiscoveryMessage_Request struct {
	Request *SubmarinerNATDiscoveryRequest `protobuf:"bytes,2,opt,name=request,proto3,oneof"`
}

type SubmarinerNATDiscoveryMessage_Response struct {
	Response *SubmarinerNATDiscoveryResponse `protobuf:"bytes,3,opt,name=response,proto3,oneof"`
}

func (*SubmarinerNATDiscoveryMessage_Request) isSubmarinerNATDiscoveryMessage_Message() {}

func (*SubmarinerNATDiscoveryMessage_Response) isSubmarinerNATDiscoveryMessage_Message() {}

type SubmarinerNATDiscoveryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestNumber uint64           `protobuf:"varint,1,opt,name=request_number,json=requestNumber,proto3" json:"request_number,omitempty"`
	Sender        *EndpointDetails `protobuf:"bytes,2,opt,name=sender,proto3" json:"sender,omitempty"`
	Receiver      *EndpointDetails `protobuf:"bytes,3,opt,name=receiver,proto3" json:"receiver,omitempty"`
	// The following information would allow the receiver to identify
	// and log if any form of NAT traversal is happening on the path
	UsingSrc *IPPortPair `protobuf:"bytes,4,opt,name=using_src,json=usingSrc,proto3" json:"using_src,omitempty"`
	UsingDst *IPPortPair `protobuf:"bytes,5,opt,name=using_dst,json=usingDst,proto3" json:"using_dst,omitempty"`
}

func (x *SubmarinerNATDiscoveryRequest) Reset() {
	*x = SubmarinerNATDiscoveryRequest{}
	mi := &file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmarinerNATDiscoveryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmarinerNATDiscoveryRequest) ProtoMessage() {}

func (x *SubmarinerNATDiscoveryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubmarinerNATDiscoveryRequest.ProtoReflect.Descriptor instead.
func (*SubmarinerNATDiscoveryRequest) Descriptor() ([]byte, []int) {
	return file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescGZIP(), []int{1}
}

func (x *SubmarinerNATDiscoveryRequest) GetRequestNumber() uint64 {
	if x != nil {
		return x.RequestNumber
	}
	return 0
}

func (x *SubmarinerNATDiscoveryRequest) GetSender() *EndpointDetails {
	if x != nil {
		return x.Sender
	}
	return nil
}

func (x *SubmarinerNATDiscoveryRequest) GetReceiver() *EndpointDetails {
	if x != nil {
		return x.Receiver
	}
	return nil
}

func (x *SubmarinerNATDiscoveryRequest) GetUsingSrc() *IPPortPair {
	if x != nil {
		return x.UsingSrc
	}
	return nil
}

func (x *SubmarinerNATDiscoveryRequest) GetUsingDst() *IPPortPair {
	if x != nil {
		return x.UsingDst
	}
	return nil
}

type SubmarinerNATDiscoveryResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestNumber      uint64           `protobuf:"varint,1,opt,name=request_number,json=requestNumber,proto3" json:"request_number,omitempty"`
	Response           ResponseType     `protobuf:"varint,2,opt,name=response,proto3,enum=ResponseType" json:"response,omitempty"`
	Sender             *EndpointDetails `protobuf:"bytes,3,opt,name=sender,proto3" json:"sender,omitempty"`
	Receiver           *EndpointDetails `protobuf:"bytes,4,opt,name=receiver,proto3" json:"receiver,omitempty"`
	SrcIpNatDetected   bool             `protobuf:"varint,5,opt,name=src_ip_nat_detected,json=srcIpNatDetected,proto3" json:"src_ip_nat_detected,omitempty"`
	SrcPortNatDetected bool             `protobuf:"varint,6,opt,name=src_port_nat_detected,json=srcPortNatDetected,proto3" json:"src_port_nat_detected,omitempty"`
	DstIpNatDetected   bool             `protobuf:"varint,7,opt,name=dst_ip_nat_detected,json=dstIpNatDetected,proto3" json:"dst_ip_nat_detected,omitempty"`
	// The received SRC IP / SRC port is reported, which will be useful for
	// diagnosing corner cases
	ReceivedSrc *IPPortPair `protobuf:"bytes,8,opt,name=received_src,json=receivedSrc,proto3" json:"received_src,omitempty"`
}

func (x *SubmarinerNATDiscoveryResponse) Reset() {
	*x = SubmarinerNATDiscoveryResponse{}
	mi := &file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmarinerNATDiscoveryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmarinerNATDiscoveryResponse) ProtoMessage() {}

func (x *SubmarinerNATDiscoveryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubmarinerNATDiscoveryResponse.ProtoReflect.Descriptor instead.
func (*SubmarinerNATDiscoveryResponse) Descriptor() ([]byte, []int) {
	return file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescGZIP(), []int{2}
}

func (x *SubmarinerNATDiscoveryResponse) GetRequestNumber() uint64 {
	if x != nil {
		return x.RequestNumber
	}
	return 0
}

func (x *SubmarinerNATDiscoveryResponse) GetResponse() ResponseType {
	if x != nil {
		return x.Response
	}
	return ResponseType_OK
}

func (x *SubmarinerNATDiscoveryResponse) GetSender() *EndpointDetails {
	if x != nil {
		return x.Sender
	}
	return nil
}

func (x *SubmarinerNATDiscoveryResponse) GetReceiver() *EndpointDetails {
	if x != nil {
		return x.Receiver
	}
	return nil
}

func (x *SubmarinerNATDiscoveryResponse) GetSrcIpNatDetected() bool {
	if x != nil {
		return x.SrcIpNatDetected
	}
	return false
}

func (x *SubmarinerNATDiscoveryResponse) GetSrcPortNatDetected() bool {
	if x != nil {
		return x.SrcPortNatDetected
	}
	return false
}

func (x *SubmarinerNATDiscoveryResponse) GetDstIpNatDetected() bool {
	if x != nil {
		return x.DstIpNatDetected
	}
	return false
}

func (x *SubmarinerNATDiscoveryResponse) GetReceivedSrc() *IPPortPair {
	if x != nil {
		return x.ReceivedSrc
	}
	return nil
}

type IPPortPair struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IP   string `protobuf:"bytes,1,opt,name=IP,proto3" json:"IP,omitempty"`
	Port int32  `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
}

func (x *IPPortPair) Reset() {
	*x = IPPortPair{}
	mi := &file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *IPPortPair) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IPPortPair) ProtoMessage() {}

func (x *IPPortPair) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IPPortPair.ProtoReflect.Descriptor instead.
func (*IPPortPair) Descriptor() ([]byte, []int) {
	return file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescGZIP(), []int{3}
}

func (x *IPPortPair) GetIP() string {
	if x != nil {
		return x.IP
	}
	return ""
}

func (x *IPPortPair) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

type EndpointDetails struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// should we hash this for privacy? a hash can be checked against a known list, but can't be decoded
	ClusterId  string `protobuf:"bytes,1,opt,name=cluster_id,json=clusterId,proto3" json:"cluster_id,omitempty"`
	EndpointId string `protobuf:"bytes,2,opt,name=endpoint_id,json=endpointId,proto3" json:"endpoint_id,omitempty"`
}

func (x *EndpointDetails) Reset() {
	*x = EndpointDetails{}
	mi := &file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EndpointDetails) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EndpointDetails) ProtoMessage() {}

func (x *EndpointDetails) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EndpointDetails.ProtoReflect.Descriptor instead.
func (*EndpointDetails) Descriptor() ([]byte, []int) {
	return file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescGZIP(), []int{4}
}

func (x *EndpointDetails) GetClusterId() string {
	if x != nil {
		return x.ClusterId
	}
	return ""
}

func (x *EndpointDetails) GetEndpointId() string {
	if x != nil {
		return x.EndpointId
	}
	return ""
}

var File_pkg_natdiscovery_proto_natdiscovery_proto protoreflect.FileDescriptor

var file_pkg_natdiscovery_proto_natdiscovery_proto_rawDesc = []byte{
	0x0a, 0x29, 0x70, 0x6b, 0x67, 0x2f, 0x6e, 0x61, 0x74, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65,
	0x72, 0x79, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6e, 0x61, 0x74, 0x64, 0x69, 0x73, 0x63,
	0x6f, 0x76, 0x65, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbf, 0x01, 0x0a, 0x1d,
	0x53, 0x75, 0x62, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x4e, 0x41, 0x54, 0x44, 0x69, 0x73,
	0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x18, 0x0a,
	0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x3a, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x53, 0x75, 0x62, 0x6d, 0x61,
	0x72, 0x69, 0x6e, 0x65, 0x72, 0x4e, 0x41, 0x54, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72,
	0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x3d, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x53, 0x75, 0x62, 0x6d, 0x61, 0x72, 0x69, 0x6e,
	0x65, 0x72, 0x4e, 0x41, 0x54, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x09, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0xf2, 0x01,
	0x0a, 0x1d, 0x53, 0x75, 0x62, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x4e, 0x41, 0x54, 0x44,
	0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x25, 0x0a, 0x0e, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0d, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x28, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72,
	0x12, 0x2c, 0x0a, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x10, 0x2e, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x44, 0x65, 0x74,
	0x61, 0x69, 0x6c, 0x73, 0x52, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x12, 0x28,
	0x0a, 0x09, 0x75, 0x73, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x72, 0x63, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0b, 0x2e, 0x49, 0x50, 0x50, 0x6f, 0x72, 0x74, 0x50, 0x61, 0x69, 0x72, 0x52, 0x08,
	0x75, 0x73, 0x69, 0x6e, 0x67, 0x53, 0x72, 0x63, 0x12, 0x28, 0x0a, 0x09, 0x75, 0x73, 0x69, 0x6e,
	0x67, 0x5f, 0x64, 0x73, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x49, 0x50,
	0x50, 0x6f, 0x72, 0x74, 0x50, 0x61, 0x69, 0x72, 0x52, 0x08, 0x75, 0x73, 0x69, 0x6e, 0x67, 0x44,
	0x73, 0x74, 0x22, 0x8b, 0x03, 0x0a, 0x1e, 0x53, 0x75, 0x62, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65,
	0x72, 0x4e, 0x41, 0x54, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x25, 0x0a, 0x0e, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0d, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x29, 0x0a, 0x08,
	0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d,
	0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x08, 0x72,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65,
	0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65,
	0x72, 0x12, 0x2c, 0x0a, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x44, 0x65,
	0x74, 0x61, 0x69, 0x6c, 0x73, 0x52, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x12,
	0x2d, 0x0a, 0x13, 0x73, 0x72, 0x63, 0x5f, 0x69, 0x70, 0x5f, 0x6e, 0x61, 0x74, 0x5f, 0x64, 0x65,
	0x74, 0x65, 0x63, 0x74, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x10, 0x73, 0x72,
	0x63, 0x49, 0x70, 0x4e, 0x61, 0x74, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x65, 0x64, 0x12, 0x31,
	0x0a, 0x15, 0x73, 0x72, 0x63, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x6e, 0x61, 0x74, 0x5f, 0x64,
	0x65, 0x74, 0x65, 0x63, 0x74, 0x65, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x12, 0x73,
	0x72, 0x63, 0x50, 0x6f, 0x72, 0x74, 0x4e, 0x61, 0x74, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x65,
	0x64, 0x12, 0x2d, 0x0a, 0x13, 0x64, 0x73, 0x74, 0x5f, 0x69, 0x70, 0x5f, 0x6e, 0x61, 0x74, 0x5f,
	0x64, 0x65, 0x74, 0x65, 0x63, 0x74, 0x65, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x10,
	0x64, 0x73, 0x74, 0x49, 0x70, 0x4e, 0x61, 0x74, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x65, 0x64,
	0x12, 0x2e, 0x0a, 0x0c, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x64, 0x5f, 0x73, 0x72, 0x63,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x49, 0x50, 0x50, 0x6f, 0x72, 0x74, 0x50,
	0x61, 0x69, 0x72, 0x52, 0x0b, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x64, 0x53, 0x72, 0x63,
	0x22, 0x30, 0x0a, 0x0a, 0x49, 0x50, 0x50, 0x6f, 0x72, 0x74, 0x50, 0x61, 0x69, 0x72, 0x12, 0x0e,
	0x0a, 0x02, 0x49, 0x50, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x49, 0x50, 0x12, 0x12,
	0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x6f,
	0x72, 0x74, 0x22, 0x51, 0x0a, 0x0f, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x44, 0x65,
	0x74, 0x61, 0x69, 0x6c, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65, 0x6e, 0x64, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x49, 0x64, 0x2a, 0x6a, 0x0a, 0x0c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x06, 0x0a, 0x02, 0x4f, 0x4b, 0x10, 0x00, 0x12, 0x10, 0x0a,
	0x0c, 0x4e, 0x41, 0x54, 0x5f, 0x44, 0x45, 0x54, 0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x01, 0x12,
	0x17, 0x0a, 0x13, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x5f, 0x44, 0x53, 0x54, 0x5f, 0x43,
	0x4c, 0x55, 0x53, 0x54, 0x45, 0x52, 0x10, 0x02, 0x12, 0x18, 0x0a, 0x14, 0x55, 0x4e, 0x4b, 0x4e,
	0x4f, 0x57, 0x4e, 0x5f, 0x44, 0x53, 0x54, 0x5f, 0x45, 0x4e, 0x44, 0x50, 0x4f, 0x49, 0x4e, 0x54,
	0x10, 0x03, 0x12, 0x0d, 0x0a, 0x09, 0x4d, 0x41, 0x4c, 0x46, 0x4f, 0x52, 0x4d, 0x45, 0x44, 0x10,
	0x04, 0x42, 0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x73, 0x75, 0x62, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x75,
	0x62, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6e, 0x61, 0x74,
	0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescOnce sync.Once
	file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescData = file_pkg_natdiscovery_proto_natdiscovery_proto_rawDesc
)

func file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescGZIP() []byte {
	file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescOnce.Do(func() {
		file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescData)
	})
	return file_pkg_natdiscovery_proto_natdiscovery_proto_rawDescData
}

var file_pkg_natdiscovery_proto_natdiscovery_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_pkg_natdiscovery_proto_natdiscovery_proto_goTypes = []any{
	(ResponseType)(0),                      // 0: ResponseType
	(*SubmarinerNATDiscoveryMessage)(nil),  // 1: SubmarinerNATDiscoveryMessage
	(*SubmarinerNATDiscoveryRequest)(nil),  // 2: SubmarinerNATDiscoveryRequest
	(*SubmarinerNATDiscoveryResponse)(nil), // 3: SubmarinerNATDiscoveryResponse
	(*IPPortPair)(nil),                     // 4: IPPortPair
	(*EndpointDetails)(nil),                // 5: EndpointDetails
}
var file_pkg_natdiscovery_proto_natdiscovery_proto_depIdxs = []int32{
	2,  // 0: SubmarinerNATDiscoveryMessage.request:type_name -> SubmarinerNATDiscoveryRequest
	3,  // 1: SubmarinerNATDiscoveryMessage.response:type_name -> SubmarinerNATDiscoveryResponse
	5,  // 2: SubmarinerNATDiscoveryRequest.sender:type_name -> EndpointDetails
	5,  // 3: SubmarinerNATDiscoveryRequest.receiver:type_name -> EndpointDetails
	4,  // 4: SubmarinerNATDiscoveryRequest.using_src:type_name -> IPPortPair
	4,  // 5: SubmarinerNATDiscoveryRequest.using_dst:type_name -> IPPortPair
	0,  // 6: SubmarinerNATDiscoveryResponse.response:type_name -> ResponseType
	5,  // 7: SubmarinerNATDiscoveryResponse.sender:type_name -> EndpointDetails
	5,  // 8: SubmarinerNATDiscoveryResponse.receiver:type_name -> EndpointDetails
	4,  // 9: SubmarinerNATDiscoveryResponse.received_src:type_name -> IPPortPair
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_pkg_natdiscovery_proto_natdiscovery_proto_init() }
func file_pkg_natdiscovery_proto_natdiscovery_proto_init() {
	if File_pkg_natdiscovery_proto_natdiscovery_proto != nil {
		return
	}
	file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes[0].OneofWrappers = []any{
		(*SubmarinerNATDiscoveryMessage_Request)(nil),
		(*SubmarinerNATDiscoveryMessage_Response)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_natdiscovery_proto_natdiscovery_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_natdiscovery_proto_natdiscovery_proto_goTypes,
		DependencyIndexes: file_pkg_natdiscovery_proto_natdiscovery_proto_depIdxs,
		EnumInfos:         file_pkg_natdiscovery_proto_natdiscovery_proto_enumTypes,
		MessageInfos:      file_pkg_natdiscovery_proto_natdiscovery_proto_msgTypes,
	}.Build()
	File_pkg_natdiscovery_proto_natdiscovery_proto = out.File
	file_pkg_natdiscovery_proto_natdiscovery_proto_rawDesc = nil
	file_pkg_natdiscovery_proto_natdiscovery_proto_goTypes = nil
	file_pkg_natdiscovery_proto_natdiscovery_proto_depIdxs = nil
}
