// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: fuse_pod_manager/v1alpha1/fuse_pod_manager.proto

package fuse_pod_managerv1alpha1

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

type ListFusePodsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListFusePodsRequest) Reset() {
	*x = ListFusePodsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListFusePodsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFusePodsRequest) ProtoMessage() {}

func (x *ListFusePodsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFusePodsRequest.ProtoReflect.Descriptor instead.
func (*ListFusePodsRequest) Descriptor() ([]byte, []int) {
	return file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescGZIP(), []int{0}
}

type ListFusePodsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pods []*FusePod `protobuf:"bytes,1,rep,name=pods,proto3" json:"pods,omitempty"`
}

func (x *ListFusePodsResponse) Reset() {
	*x = ListFusePodsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListFusePodsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFusePodsResponse) ProtoMessage() {}

func (x *ListFusePodsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFusePodsResponse.ProtoReflect.Descriptor instead.
func (*ListFusePodsResponse) Descriptor() ([]byte, []int) {
	return file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescGZIP(), []int{1}
}

func (x *ListFusePodsResponse) GetPods() []*FusePod {
	if x != nil {
		return x.Pods
	}
	return nil
}

type CreateFusePodRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Bucket    string `protobuf:"bytes,1,opt,name=bucket,proto3" json:"bucket,omitempty"`
	MountPath string `protobuf:"bytes,2,opt,name=mount_path,json=mountPath,proto3" json:"mount_path,omitempty"`
	VolumeId  string `protobuf:"bytes,3,opt,name=volume_id,json=volumeId,proto3" json:"volume_id,omitempty"`
}

func (x *CreateFusePodRequest) Reset() {
	*x = CreateFusePodRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateFusePodRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateFusePodRequest) ProtoMessage() {}

func (x *CreateFusePodRequest) ProtoReflect() protoreflect.Message {
	mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateFusePodRequest.ProtoReflect.Descriptor instead.
func (*CreateFusePodRequest) Descriptor() ([]byte, []int) {
	return file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescGZIP(), []int{2}
}

func (x *CreateFusePodRequest) GetBucket() string {
	if x != nil {
		return x.Bucket
	}
	return ""
}

func (x *CreateFusePodRequest) GetMountPath() string {
	if x != nil {
		return x.MountPath
	}
	return ""
}

func (x *CreateFusePodRequest) GetVolumeId() string {
	if x != nil {
		return x.VolumeId
	}
	return ""
}

type CreateFusePodResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *CreateFusePodResponse) Reset() {
	*x = CreateFusePodResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateFusePodResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateFusePodResponse) ProtoMessage() {}

func (x *CreateFusePodResponse) ProtoReflect() protoreflect.Message {
	mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateFusePodResponse.ProtoReflect.Descriptor instead.
func (*CreateFusePodResponse) Descriptor() ([]byte, []int) {
	return file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescGZIP(), []int{3}
}

func (x *CreateFusePodResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type DeleteFusePodRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *DeleteFusePodRequest) Reset() {
	*x = DeleteFusePodRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteFusePodRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteFusePodRequest) ProtoMessage() {}

func (x *DeleteFusePodRequest) ProtoReflect() protoreflect.Message {
	mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteFusePodRequest.ProtoReflect.Descriptor instead.
func (*DeleteFusePodRequest) Descriptor() ([]byte, []int) {
	return file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescGZIP(), []int{4}
}

func (x *DeleteFusePodRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type DeleteFusePodResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteFusePodResponse) Reset() {
	*x = DeleteFusePodResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteFusePodResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteFusePodResponse) ProtoMessage() {}

func (x *DeleteFusePodResponse) ProtoReflect() protoreflect.Message {
	mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteFusePodResponse.ProtoReflect.Descriptor instead.
func (*DeleteFusePodResponse) Descriptor() ([]byte, []int) {
	return file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescGZIP(), []int{5}
}

type FusePod struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Bucket    string `protobuf:"bytes,2,opt,name=bucket,proto3" json:"bucket,omitempty"`
	MountPath string `protobuf:"bytes,3,opt,name=mount_path,json=mountPath,proto3" json:"mount_path,omitempty"`
	VolumeId  string `protobuf:"bytes,4,opt,name=volume_id,json=volumeId,proto3" json:"volume_id,omitempty"`
}

func (x *FusePod) Reset() {
	*x = FusePod{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FusePod) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FusePod) ProtoMessage() {}

func (x *FusePod) ProtoReflect() protoreflect.Message {
	mi := &file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FusePod.ProtoReflect.Descriptor instead.
func (*FusePod) Descriptor() ([]byte, []int) {
	return file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescGZIP(), []int{6}
}

func (x *FusePod) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FusePod) GetBucket() string {
	if x != nil {
		return x.Bucket
	}
	return ""
}

func (x *FusePod) GetMountPath() string {
	if x != nil {
		return x.MountPath
	}
	return ""
}

func (x *FusePod) GetVolumeId() string {
	if x != nil {
		return x.VolumeId
	}
	return ""
}

var File_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto protoreflect.FileDescriptor

var file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDesc = []byte{
	0x0a, 0x30, 0x66, 0x75, 0x73, 0x65, 0x5f, 0x70, 0x6f, 0x64, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67,
	0x65, 0x72, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x66, 0x75, 0x73, 0x65,
	0x5f, 0x70, 0x6f, 0x64, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x19, 0x66, 0x75, 0x73, 0x65, 0x5f, 0x70, 0x6f, 0x64, 0x5f, 0x6d, 0x61, 0x6e,
	0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x22, 0x15, 0x0a,
	0x13, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x22, 0x4e, 0x0a, 0x14, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x75, 0x73, 0x65,
	0x50, 0x6f, 0x64, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x36, 0x0a, 0x04,
	0x70, 0x6f, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x66, 0x75, 0x73,
	0x65, 0x5f, 0x70, 0x6f, 0x64, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x52, 0x04,
	0x70, 0x6f, 0x64, 0x73, 0x22, 0x6a, 0x0a, 0x14, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x46, 0x75,
	0x73, 0x65, 0x50, 0x6f, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x75,
	0x63, 0x6b, 0x65, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x70, 0x61,
	0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x50,
	0x61, 0x74, 0x68, 0x12, 0x1b, 0x0a, 0x09, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x5f, 0x69, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x49, 0x64,
	0x22, 0x2b, 0x0a, 0x15, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f,
	0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x2a, 0x0a,
	0x14, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x17, 0x0a, 0x15, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x71, 0x0a, 0x07, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x6d, 0x6f, 0x75,
	0x6e, 0x74, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6d,
	0x6f, 0x75, 0x6e, 0x74, 0x50, 0x61, 0x74, 0x68, 0x12, 0x1b, 0x0a, 0x09, 0x76, 0x6f, 0x6c, 0x75,
	0x6d, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x76, 0x6f, 0x6c,
	0x75, 0x6d, 0x65, 0x49, 0x64, 0x32, 0xf0, 0x02, 0x0a, 0x15, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f,
	0x64, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x6f, 0x0a, 0x0c, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x73, 0x12,
	0x2e, 0x2e, 0x66, 0x75, 0x73, 0x65, 0x5f, 0x70, 0x6f, 0x64, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74,
	0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x2f, 0x2e, 0x66, 0x75, 0x73, 0x65, 0x5f, 0x70, 0x6f, 0x64, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74,
	0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x72, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f,
	0x64, 0x12, 0x2f, 0x2e, 0x66, 0x75, 0x73, 0x65, 0x5f, 0x70, 0x6f, 0x64, 0x5f, 0x6d, 0x61, 0x6e,
	0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x30, 0x2e, 0x66, 0x75, 0x73, 0x65, 0x5f, 0x70, 0x6f, 0x64, 0x5f, 0x6d, 0x61,
	0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x72, 0x0a, 0x0d, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x46, 0x75,
	0x73, 0x65, 0x50, 0x6f, 0x64, 0x12, 0x2f, 0x2e, 0x66, 0x75, 0x73, 0x65, 0x5f, 0x70, 0x6f, 0x64,
	0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x30, 0x2e, 0x66, 0x75, 0x73, 0x65, 0x5f, 0x70, 0x6f,
	0x64, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0xfe, 0x01, 0x0a, 0x1d, 0x63, 0x6f, 0x6d,
	0x2e, 0x66, 0x75, 0x73, 0x65, 0x5f, 0x70, 0x6f, 0x64, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x42, 0x13, 0x46, 0x75, 0x73, 0x65,
	0x50, 0x6f, 0x64, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50,
	0x01, 0x5a, 0x4b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x68, 0x6f,
	0x77, 0x6e, 0x33, 0x64, 0x2f, 0x73, 0x33, 0x2d, 0x63, 0x73, 0x69, 0x2f, 0x66, 0x75, 0x73, 0x65,
	0x5f, 0x70, 0x6f, 0x64, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x3b, 0x66, 0x75, 0x73, 0x65, 0x5f, 0x70, 0x6f, 0x64, 0x5f, 0x6d,
	0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xa2, 0x02,
	0x03, 0x46, 0x58, 0x58, 0xaa, 0x02, 0x17, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x4d, 0x61,
	0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xca, 0x02,
	0x17, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x5c,
	0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xe2, 0x02, 0x23, 0x46, 0x75, 0x73, 0x65, 0x50,
	0x6f, 0x64, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x5c, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02,
	0x18, 0x46, 0x75, 0x73, 0x65, 0x50, 0x6f, 0x64, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x3a,
	0x3a, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescOnce sync.Once
	file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescData = file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDesc
)

func file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescGZIP() []byte {
	file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescOnce.Do(func() {
		file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescData = protoimpl.X.CompressGZIP(file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescData)
	})
	return file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDescData
}

var file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_goTypes = []interface{}{
	(*ListFusePodsRequest)(nil),   // 0: fuse_pod_manager.v1alpha1.ListFusePodsRequest
	(*ListFusePodsResponse)(nil),  // 1: fuse_pod_manager.v1alpha1.ListFusePodsResponse
	(*CreateFusePodRequest)(nil),  // 2: fuse_pod_manager.v1alpha1.CreateFusePodRequest
	(*CreateFusePodResponse)(nil), // 3: fuse_pod_manager.v1alpha1.CreateFusePodResponse
	(*DeleteFusePodRequest)(nil),  // 4: fuse_pod_manager.v1alpha1.DeleteFusePodRequest
	(*DeleteFusePodResponse)(nil), // 5: fuse_pod_manager.v1alpha1.DeleteFusePodResponse
	(*FusePod)(nil),               // 6: fuse_pod_manager.v1alpha1.FusePod
}
var file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_depIdxs = []int32{
	6, // 0: fuse_pod_manager.v1alpha1.ListFusePodsResponse.pods:type_name -> fuse_pod_manager.v1alpha1.FusePod
	0, // 1: fuse_pod_manager.v1alpha1.FusePodManagerService.ListFusePods:input_type -> fuse_pod_manager.v1alpha1.ListFusePodsRequest
	2, // 2: fuse_pod_manager.v1alpha1.FusePodManagerService.CreateFusePod:input_type -> fuse_pod_manager.v1alpha1.CreateFusePodRequest
	4, // 3: fuse_pod_manager.v1alpha1.FusePodManagerService.DeleteFusePod:input_type -> fuse_pod_manager.v1alpha1.DeleteFusePodRequest
	1, // 4: fuse_pod_manager.v1alpha1.FusePodManagerService.ListFusePods:output_type -> fuse_pod_manager.v1alpha1.ListFusePodsResponse
	3, // 5: fuse_pod_manager.v1alpha1.FusePodManagerService.CreateFusePod:output_type -> fuse_pod_manager.v1alpha1.CreateFusePodResponse
	5, // 6: fuse_pod_manager.v1alpha1.FusePodManagerService.DeleteFusePod:output_type -> fuse_pod_manager.v1alpha1.DeleteFusePodResponse
	4, // [4:7] is the sub-list for method output_type
	1, // [1:4] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_init() }
func file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_init() {
	if File_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListFusePodsRequest); i {
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
		file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListFusePodsResponse); i {
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
		file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateFusePodRequest); i {
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
		file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateFusePodResponse); i {
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
		file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteFusePodRequest); i {
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
		file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteFusePodResponse); i {
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
		file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FusePod); i {
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
			RawDescriptor: file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_goTypes,
		DependencyIndexes: file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_depIdxs,
		MessageInfos:      file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_msgTypes,
	}.Build()
	File_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto = out.File
	file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_rawDesc = nil
	file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_goTypes = nil
	file_fuse_pod_manager_v1alpha1_fuse_pod_manager_proto_depIdxs = nil
}