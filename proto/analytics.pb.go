// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.6.1
// source: analytics.proto

package proto

import (
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type GetQueriesListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PeriodStartFrom *timestamp.Timestamp `protobuf:"bytes,1,opt,name=period_start_from,json=periodStartFrom,proto3" json:"period_start_from,omitempty"`
	PeriodStartTo   *timestamp.Timestamp `protobuf:"bytes,2,opt,name=period_start_to,json=periodStartTo,proto3" json:"period_start_to,omitempty"`
	ClusterName     string               `protobuf:"bytes,3,opt,name=cluster_name,json=clusterName,proto3" json:"cluster_name,omitempty"`
	Order           string               `protobuf:"bytes,4,opt,name=order,proto3" json:"order,omitempty"`
	Limit           int64                `protobuf:"varint,5,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (x *GetQueriesListRequest) Reset() {
	*x = GetQueriesListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_analytics_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetQueriesListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetQueriesListRequest) ProtoMessage() {}

func (x *GetQueriesListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_analytics_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetQueriesListRequest.ProtoReflect.Descriptor instead.
func (*GetQueriesListRequest) Descriptor() ([]byte, []int) {
	return file_analytics_proto_rawDescGZIP(), []int{0}
}

func (x *GetQueriesListRequest) GetPeriodStartFrom() *timestamp.Timestamp {
	if x != nil {
		return x.PeriodStartFrom
	}
	return nil
}

func (x *GetQueriesListRequest) GetPeriodStartTo() *timestamp.Timestamp {
	if x != nil {
		return x.PeriodStartTo
	}
	return nil
}

func (x *GetQueriesListRequest) GetClusterName() string {
	if x != nil {
		return x.ClusterName
	}
	return ""
}

func (x *GetQueriesListRequest) GetOrder() string {
	if x != nil {
		return x.Order
	}
	return ""
}

func (x *GetQueriesListRequest) GetLimit() int64 {
	if x != nil {
		return x.Limit
	}
	return 0
}

type GetQueriesListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Queries  []*Query      `protobuf:"bytes,1,rep,name=queries,proto3" json:"queries,omitempty"`
	Mappings []*MetricInfo `protobuf:"bytes,5,rep,name=mappings,proto3" json:"mappings,omitempty"`
}

func (x *GetQueriesListResponse) Reset() {
	*x = GetQueriesListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_analytics_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetQueriesListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetQueriesListResponse) ProtoMessage() {}

func (x *GetQueriesListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_analytics_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetQueriesListResponse.ProtoReflect.Descriptor instead.
func (*GetQueriesListResponse) Descriptor() ([]byte, []int) {
	return file_analytics_proto_rawDescGZIP(), []int{1}
}

func (x *GetQueriesListResponse) GetQueries() []*Query {
	if x != nil {
		return x.Queries
	}
	return nil
}

func (x *GetQueriesListResponse) GetMappings() []*MetricInfo {
	if x != nil {
		return x.Mappings
	}
	return nil
}

type Query struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string                   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Fingerprint string                   `protobuf:"bytes,2,opt,name=fingerprint,proto3" json:"fingerprint,omitempty"`
	Text        string                   `protobuf:"bytes,3,opt,name=text,proto3" json:"text,omitempty"`
	Parameters  []string                 `protobuf:"bytes,4,rep,name=parameters,proto3" json:"parameters,omitempty"`
	PlanIds     []string                 `protobuf:"bytes,6,rep,name=plan_ids,json=planIds,proto3" json:"plan_ids,omitempty"`
	Metrics     map[string]*MetricValues `protobuf:"bytes,5,rep,name=metrics,proto3" json:"metrics,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Query) Reset() {
	*x = Query{}
	if protoimpl.UnsafeEnabled {
		mi := &file_analytics_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Query) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Query) ProtoMessage() {}

func (x *Query) ProtoReflect() protoreflect.Message {
	mi := &file_analytics_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Query.ProtoReflect.Descriptor instead.
func (*Query) Descriptor() ([]byte, []int) {
	return file_analytics_proto_rawDescGZIP(), []int{2}
}

func (x *Query) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Query) GetFingerprint() string {
	if x != nil {
		return x.Fingerprint
	}
	return ""
}

func (x *Query) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *Query) GetParameters() []string {
	if x != nil {
		return x.Parameters
	}
	return nil
}

func (x *Query) GetPlanIds() []string {
	if x != nil {
		return x.PlanIds
	}
	return nil
}

func (x *Query) GetMetrics() map[string]*MetricValues {
	if x != nil {
		return x.Metrics
	}
	return nil
}

type MetricInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Type  string `protobuf:"bytes,2,opt,name=Type,proto3" json:"Type,omitempty"`
	Kind  string `protobuf:"bytes,3,opt,name=Kind,proto3" json:"Kind,omitempty"`
	Title string `protobuf:"bytes,5,opt,name=Title,proto3" json:"Title,omitempty"`
}

func (x *MetricInfo) Reset() {
	*x = MetricInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_analytics_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricInfo) ProtoMessage() {}

func (x *MetricInfo) ProtoReflect() protoreflect.Message {
	mi := &file_analytics_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricInfo.ProtoReflect.Descriptor instead.
func (*MetricInfo) Descriptor() ([]byte, []int) {
	return file_analytics_proto_rawDescGZIP(), []int{3}
}

func (x *MetricInfo) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *MetricInfo) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *MetricInfo) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *MetricInfo) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

var File_analytics_proto protoreflect.FileDescriptor

var file_analytics_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x10, 0x62, 0x6f, 0x72, 0x65, 0x61, 0x6c, 0x69, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x0c, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xf2, 0x01, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x51, 0x75, 0x65, 0x72, 0x69, 0x65, 0x73, 0x4c,
	0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x46, 0x0a, 0x11, 0x70, 0x65,
	0x72, 0x69, 0x6f, 0x64, 0x5f, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x66, 0x72, 0x6f, 0x6d, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x52, 0x0f, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x53, 0x74, 0x61, 0x72, 0x74, 0x46, 0x72,
	0x6f, 0x6d, 0x12, 0x42, 0x0a, 0x0f, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x5f, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x5f, 0x74, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0d, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x53,
	0x74, 0x61, 0x72, 0x74, 0x54, 0x6f, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6c,
	0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6f, 0x72, 0x64,
	0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x12,
	0x14, 0x0a, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05,
	0x6c, 0x69, 0x6d, 0x69, 0x74, 0x22, 0x85, 0x01, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x51, 0x75, 0x65,
	0x72, 0x69, 0x65, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x31, 0x0a, 0x07, 0x71, 0x75, 0x65, 0x72, 0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x17, 0x2e, 0x62, 0x6f, 0x72, 0x65, 0x61, 0x6c, 0x69, 0x73, 0x2e, 0x76, 0x31, 0x62,
	0x65, 0x74, 0x61, 0x31, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x07, 0x71, 0x75, 0x65, 0x72,
	0x69, 0x65, 0x73, 0x12, 0x38, 0x0a, 0x08, 0x6d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x73, 0x18,
	0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x62, 0x6f, 0x72, 0x65, 0x61, 0x6c, 0x69, 0x73,
	0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x08, 0x6d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x73, 0x22, 0xa4, 0x02,
	0x0a, 0x05, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x20, 0x0a, 0x0b, 0x66, 0x69, 0x6e, 0x67, 0x65,
	0x72, 0x70, 0x72, 0x69, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x66, 0x69,
	0x6e, 0x67, 0x65, 0x72, 0x70, 0x72, 0x69, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78,
	0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x12, 0x1e, 0x0a,
	0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x12, 0x19, 0x0a,
	0x08, 0x70, 0x6c, 0x61, 0x6e, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x07, 0x70, 0x6c, 0x61, 0x6e, 0x49, 0x64, 0x73, 0x12, 0x3e, 0x0a, 0x07, 0x6d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x62, 0x6f, 0x72, 0x65,
	0x61, 0x6c, 0x69, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x51, 0x75, 0x65,
	0x72, 0x79, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x07, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x1a, 0x5a, 0x0a, 0x0c, 0x4d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x34, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x62, 0x6f, 0x72, 0x65,
	0x61, 0x6c, 0x69, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x22, 0x5c, 0x0a, 0x0a, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x49, 0x6e,
	0x66, 0x6f, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x4b, 0x69, 0x6e, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x4b, 0x69, 0x6e, 0x64, 0x12, 0x14, 0x0a, 0x05,
	0x54, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x54, 0x69, 0x74,
	0x6c, 0x65, 0x32, 0xa2, 0x01, 0x0a, 0x0e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x41, 0x6e, 0x61, 0x6c,
	0x79, 0x74, 0x69, 0x63, 0x73, 0x12, 0x8f, 0x01, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x51, 0x75, 0x65,
	0x72, 0x69, 0x65, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x27, 0x2e, 0x62, 0x6f, 0x72, 0x65, 0x61,
	0x6c, 0x69, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x51,
	0x75, 0x65, 0x72, 0x69, 0x65, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x28, 0x2e, 0x62, 0x6f, 0x72, 0x65, 0x61, 0x6c, 0x69, 0x73, 0x2e, 0x76, 0x31, 0x62,
	0x65, 0x74, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x51, 0x75, 0x65, 0x72, 0x69, 0x65, 0x73, 0x4c,
	0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2a, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x24, 0x22, 0x1f, 0x2f, 0x76, 0x30, 0x2f, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69,
	0x63, 0x73, 0x2f, 0x47, 0x65, 0x74, 0x51, 0x75, 0x65, 0x72, 0x69, 0x65, 0x73, 0x4d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x73, 0x3a, 0x01, 0x2a, 0x42, 0x08, 0x5a, 0x06, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_analytics_proto_rawDescOnce sync.Once
	file_analytics_proto_rawDescData = file_analytics_proto_rawDesc
)

func file_analytics_proto_rawDescGZIP() []byte {
	file_analytics_proto_rawDescOnce.Do(func() {
		file_analytics_proto_rawDescData = protoimpl.X.CompressGZIP(file_analytics_proto_rawDescData)
	})
	return file_analytics_proto_rawDescData
}

var file_analytics_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_analytics_proto_goTypes = []interface{}{
	(*GetQueriesListRequest)(nil),  // 0: borealis.v1beta1.GetQueriesListRequest
	(*GetQueriesListResponse)(nil), // 1: borealis.v1beta1.GetQueriesListResponse
	(*Query)(nil),                  // 2: borealis.v1beta1.Query
	(*MetricInfo)(nil),             // 3: borealis.v1beta1.MetricInfo
	nil,                            // 4: borealis.v1beta1.Query.MetricsEntry
	(*timestamp.Timestamp)(nil),    // 5: google.protobuf.Timestamp
	(*MetricValues)(nil),           // 6: borealis.v1beta1.MetricValues
}
var file_analytics_proto_depIdxs = []int32{
	5, // 0: borealis.v1beta1.GetQueriesListRequest.period_start_from:type_name -> google.protobuf.Timestamp
	5, // 1: borealis.v1beta1.GetQueriesListRequest.period_start_to:type_name -> google.protobuf.Timestamp
	2, // 2: borealis.v1beta1.GetQueriesListResponse.queries:type_name -> borealis.v1beta1.Query
	3, // 3: borealis.v1beta1.GetQueriesListResponse.mappings:type_name -> borealis.v1beta1.MetricInfo
	4, // 4: borealis.v1beta1.Query.metrics:type_name -> borealis.v1beta1.Query.MetricsEntry
	6, // 5: borealis.v1beta1.Query.MetricsEntry.value:type_name -> borealis.v1beta1.MetricValues
	0, // 6: borealis.v1beta1.QueryAnalytics.GetQueriesList:input_type -> borealis.v1beta1.GetQueriesListRequest
	1, // 7: borealis.v1beta1.QueryAnalytics.GetQueriesList:output_type -> borealis.v1beta1.GetQueriesListResponse
	7, // [7:8] is the sub-list for method output_type
	6, // [6:7] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_analytics_proto_init() }
func file_analytics_proto_init() {
	if File_analytics_proto != nil {
		return
	}
	file_shared_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_analytics_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetQueriesListRequest); i {
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
		file_analytics_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetQueriesListResponse); i {
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
		file_analytics_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Query); i {
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
		file_analytics_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MetricInfo); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_analytics_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_analytics_proto_goTypes,
		DependencyIndexes: file_analytics_proto_depIdxs,
		MessageInfos:      file_analytics_proto_msgTypes,
	}.Build()
	File_analytics_proto = out.File
	file_analytics_proto_rawDesc = nil
	file_analytics_proto_goTypes = nil
	file_analytics_proto_depIdxs = nil
}
