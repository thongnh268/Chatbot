// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.11.2
// source: user.proto

package user

import (
	proto "github.com/golang/protobuf/proto"
	common "github.com/subiz/header/common"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type SCondition_EventTime int32

const (
	SCondition_none    SCondition_EventTime = 0 // not an event
	SCondition_current SCondition_EventTime = 1
	SCondition_latest  SCondition_EventTime = 2
	SCondition_past    SCondition_EventTime = 3 // exists
)

// Enum value maps for SCondition_EventTime.
var (
	SCondition_EventTime_name = map[int32]string{
		0: "none",
		1: "current",
		2: "latest",
		3: "past",
	}
	SCondition_EventTime_value = map[string]int32{
		"none":    0,
		"current": 1,
		"latest":  2,
		"past":    3,
	}
)

func (x SCondition_EventTime) Enum() *SCondition_EventTime {
	p := new(SCondition_EventTime)
	*p = x
	return p
}

func (x SCondition_EventTime) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SCondition_EventTime) Descriptor() protoreflect.EnumDescriptor {
	return file_user_proto_enumTypes[0].Descriptor()
}

func (SCondition_EventTime) Type() protoreflect.EnumType {
	return &file_user_proto_enumTypes[0]
}

func (x SCondition_EventTime) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *SCondition_EventTime) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = SCondition_EventTime(num)
	return nil
}

// Deprecated: Use SCondition_EventTime.Descriptor instead.
func (SCondition_EventTime) EnumDescriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{0, 0}
}

type SCondition_JoinOperator int32

const (
	SCondition_and SCondition_JoinOperator = 0
	SCondition_or  SCondition_JoinOperator = 1
)

// Enum value maps for SCondition_JoinOperator.
var (
	SCondition_JoinOperator_name = map[int32]string{
		0: "and",
		1: "or",
	}
	SCondition_JoinOperator_value = map[string]int32{
		"and": 0,
		"or":  1,
	}
)

func (x SCondition_JoinOperator) Enum() *SCondition_JoinOperator {
	p := new(SCondition_JoinOperator)
	*p = x
	return p
}

func (x SCondition_JoinOperator) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SCondition_JoinOperator) Descriptor() protoreflect.EnumDescriptor {
	return file_user_proto_enumTypes[1].Descriptor()
}

func (SCondition_JoinOperator) Type() protoreflect.EnumType {
	return &file_user_proto_enumTypes[1]
}

func (x SCondition_JoinOperator) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *SCondition_JoinOperator) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = SCondition_JoinOperator(num)
	return nil
}

// Deprecated: Use SCondition_JoinOperator.Descriptor instead.
func (SCondition_JoinOperator) EnumDescriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{0, 1}
}

type SCondition_Function int32

const (
	SCondition_minute_of_day SCondition_Function = 0
	SCondition_hour_of_day   SCondition_Function = 1
	SCondition_day_of_week   SCondition_Function = 2
	SCondition_day_ago       SCondition_Function = 3
)

// Enum value maps for SCondition_Function.
var (
	SCondition_Function_name = map[int32]string{
		0: "minute_of_day",
		1: "hour_of_day",
		2: "day_of_week",
		3: "day_ago",
	}
	SCondition_Function_value = map[string]int32{
		"minute_of_day": 0,
		"hour_of_day":   1,
		"day_of_week":   2,
		"day_ago":       3,
	}
)

func (x SCondition_Function) Enum() *SCondition_Function {
	p := new(SCondition_Function)
	*p = x
	return p
}

func (x SCondition_Function) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SCondition_Function) Descriptor() protoreflect.EnumDescriptor {
	return file_user_proto_enumTypes[2].Descriptor()
}

func (SCondition_Function) Type() protoreflect.EnumType {
	return &file_user_proto_enumTypes[2]
}

func (x SCondition_Function) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *SCondition_Function) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = SCondition_Function(num)
	return nil
}

// Deprecated: Use SCondition_Function.Descriptor instead.
func (SCondition_Function) EnumDescriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{0, 2}
}

// can be
// {id, join, event_time, conditions} (grouped event condition)
// {id, join, conditions} (grouped condition)
// {id, key, operator, value} (user condition || simple event condition)
type SCondition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id            *string       `protobuf:"bytes,3,opt,name=id" json:"id,omitempty"`
	Key           *string       `protobuf:"bytes,4,opt,name=key" json:"key,omitempty"`           // unique
	Operator      *string       `protobuf:"bytes,5,opt,name=operator" json:"operator,omitempty"` // = # regex
	Value         *string       `protobuf:"bytes,6,opt,name=value" json:"value,omitempty"`       // JSON
	Join          *string       `protobuf:"bytes,7,opt,name=join" json:"join,omitempty"`
	EventTypeTime *string       `protobuf:"bytes,8,opt,name=event_type_time,json=eventTypeTime" json:"event_type_time,omitempty"` //
	EventType     *string       `protobuf:"bytes,10,opt,name=event_type,json=eventType" json:"event_type,omitempty"`
	Conditions    []*SCondition `protobuf:"bytes,9,rep,name=conditions" json:"conditions,omitempty"`
	Priority      *int32        `protobuf:"varint,12,opt,name=priority" json:"priority,omitempty"`
	EventExisted  *bool         `protobuf:"varint,13,opt,name=event_existed,json=eventExisted" json:"event_existed,omitempty"`
	Function      *string       `protobuf:"bytes,14,opt,name=function" json:"function,omitempty"` // used to transform value of left side before evaluate expression
}

func (x *SCondition) Reset() {
	*x = SCondition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SCondition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SCondition) ProtoMessage() {}

func (x *SCondition) ProtoReflect() protoreflect.Message {
	mi := &file_user_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SCondition.ProtoReflect.Descriptor instead.
func (*SCondition) Descriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{0}
}

func (x *SCondition) GetId() string {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return ""
}

func (x *SCondition) GetKey() string {
	if x != nil && x.Key != nil {
		return *x.Key
	}
	return ""
}

func (x *SCondition) GetOperator() string {
	if x != nil && x.Operator != nil {
		return *x.Operator
	}
	return ""
}

func (x *SCondition) GetValue() string {
	if x != nil && x.Value != nil {
		return *x.Value
	}
	return ""
}

func (x *SCondition) GetJoin() string {
	if x != nil && x.Join != nil {
		return *x.Join
	}
	return ""
}

func (x *SCondition) GetEventTypeTime() string {
	if x != nil && x.EventTypeTime != nil {
		return *x.EventTypeTime
	}
	return ""
}

func (x *SCondition) GetEventType() string {
	if x != nil && x.EventType != nil {
		return *x.EventType
	}
	return ""
}

func (x *SCondition) GetConditions() []*SCondition {
	if x != nil {
		return x.Conditions
	}
	return nil
}

func (x *SCondition) GetPriority() int32 {
	if x != nil && x.Priority != nil {
		return *x.Priority
	}
	return 0
}

func (x *SCondition) GetEventExisted() bool {
	if x != nil && x.EventExisted != nil {
		return *x.EventExisted
	}
	return false
}

func (x *SCondition) GetFunction() string {
	if x != nil && x.Function != nil {
		return *x.Function
	}
	return ""
}

type UserSearchRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ctx            *common.Context `protobuf:"bytes,1,opt,name=ctx" json:"ctx,omitempty"`
	AccountId      *string         `protobuf:"bytes,2,opt,name=account_id,json=accountId" json:"account_id,omitempty"`
	SegmentationId *string         `protobuf:"bytes,3,opt,name=segmentation_id,json=segmentationId" json:"segmentation_id,omitempty"`
	Query          *string         `protobuf:"bytes,4,opt,name=query" json:"query,omitempty"`
	Anchor         *string         `protobuf:"bytes,5,opt,name=anchor" json:"anchor,omitempty"`
	Limit          *int32          `protobuf:"varint,6,opt,name=limit" json:"limit,omitempty"`
	AgentId        *string         `protobuf:"bytes,8,opt,name=agent_id,json=agentId" json:"agent_id,omitempty"` // search my user of agent
	Unread         *bool           `protobuf:"varint,9,opt,name=unread" json:"unread,omitempty"`                 // search my user of agent
	Condition      *SCondition     `protobuf:"bytes,10,opt,name=condition" json:"condition,omitempty"`
}

func (x *UserSearchRequest) Reset() {
	*x = UserSearchRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserSearchRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserSearchRequest) ProtoMessage() {}

func (x *UserSearchRequest) ProtoReflect() protoreflect.Message {
	mi := &file_user_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserSearchRequest.ProtoReflect.Descriptor instead.
func (*UserSearchRequest) Descriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{1}
}

func (x *UserSearchRequest) GetCtx() *common.Context {
	if x != nil {
		return x.Ctx
	}
	return nil
}

func (x *UserSearchRequest) GetAccountId() string {
	if x != nil && x.AccountId != nil {
		return *x.AccountId
	}
	return ""
}

func (x *UserSearchRequest) GetSegmentationId() string {
	if x != nil && x.SegmentationId != nil {
		return *x.SegmentationId
	}
	return ""
}

func (x *UserSearchRequest) GetQuery() string {
	if x != nil && x.Query != nil {
		return *x.Query
	}
	return ""
}

func (x *UserSearchRequest) GetAnchor() string {
	if x != nil && x.Anchor != nil {
		return *x.Anchor
	}
	return ""
}

func (x *UserSearchRequest) GetLimit() int32 {
	if x != nil && x.Limit != nil {
		return *x.Limit
	}
	return 0
}

func (x *UserSearchRequest) GetAgentId() string {
	if x != nil && x.AgentId != nil {
		return *x.AgentId
	}
	return ""
}

func (x *UserSearchRequest) GetUnread() bool {
	if x != nil && x.Unread != nil {
		return *x.Unread
	}
	return false
}

func (x *UserSearchRequest) GetCondition() *SCondition {
	if x != nil {
		return x.Condition
	}
	return nil
}

type ESNote struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ctx        *common.Context `protobuf:"bytes,1,opt,name=ctx" json:"ctx,omitempty"`
	AccountId  *string         `protobuf:"bytes,2,opt,name=account_id,json=accountId" json:"account_id,omitempty"`
	Id         *string         `protobuf:"bytes,3,opt,name=id" json:"id,omitempty"`
	CreatorId  *string         `protobuf:"bytes,5,opt,name=creator_id,json=creatorId" json:"creator_id,omitempty"`
	Text       *string         `protobuf:"bytes,6,opt,name=text" json:"text,omitempty"`
	Created    *int64          `protobuf:"varint,7,opt,name=created" json:"created,omitempty"`
	TargetId   *string         `protobuf:"bytes,9,opt,name=target_id,json=targetId" json:"target_id,omitempty"`
	TargetType *string         `protobuf:"bytes,10,opt,name=target_type,json=targetType" json:"target_type,omitempty"`
	Format     *string         `protobuf:"bytes,13,opt,name=format" json:"format,omitempty"`
}

func (x *ESNote) Reset() {
	*x = ESNote{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ESNote) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ESNote) ProtoMessage() {}

func (x *ESNote) ProtoReflect() protoreflect.Message {
	mi := &file_user_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ESNote.ProtoReflect.Descriptor instead.
func (*ESNote) Descriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{2}
}

func (x *ESNote) GetCtx() *common.Context {
	if x != nil {
		return x.Ctx
	}
	return nil
}

func (x *ESNote) GetAccountId() string {
	if x != nil && x.AccountId != nil {
		return *x.AccountId
	}
	return ""
}

func (x *ESNote) GetId() string {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return ""
}

func (x *ESNote) GetCreatorId() string {
	if x != nil && x.CreatorId != nil {
		return *x.CreatorId
	}
	return ""
}

func (x *ESNote) GetText() string {
	if x != nil && x.Text != nil {
		return *x.Text
	}
	return ""
}

func (x *ESNote) GetCreated() int64 {
	if x != nil && x.Created != nil {
		return *x.Created
	}
	return 0
}

func (x *ESNote) GetTargetId() string {
	if x != nil && x.TargetId != nil {
		return *x.TargetId
	}
	return ""
}

func (x *ESNote) GetTargetType() string {
	if x != nil && x.TargetType != nil {
		return *x.TargetType
	}
	return ""
}

func (x *ESNote) GetFormat() string {
	if x != nil && x.Format != nil {
		return *x.Format
	}
	return ""
}

type SearchNoteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ctx       *common.Context `protobuf:"bytes,1,opt,name=ctx" json:"ctx,omitempty"`
	AccountId *string         `protobuf:"bytes,2,opt,name=account_id,json=accountId" json:"account_id,omitempty"`
	AgentId   *string         `protobuf:"bytes,3,opt,name=agent_id,json=agentId" json:"agent_id,omitempty"`
	Keyword   *string         `protobuf:"bytes,4,opt,name=keyword" json:"keyword,omitempty"`
	FromSec   *int64          `protobuf:"varint,5,opt,name=from_sec,json=fromSec" json:"from_sec,omitempty"` // unix seconds
	ToSec     *int64          `protobuf:"varint,6,opt,name=to_sec,json=toSec" json:"to_sec,omitempty"`       // unix seconds
	Limit     *int64          `protobuf:"varint,7,opt,name=limit" json:"limit,omitempty"`
	Anchor    *string         `protobuf:"bytes,8,opt,name=anchor" json:"anchor,omitempty"`
}

func (x *SearchNoteRequest) Reset() {
	*x = SearchNoteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchNoteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchNoteRequest) ProtoMessage() {}

func (x *SearchNoteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_user_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchNoteRequest.ProtoReflect.Descriptor instead.
func (*SearchNoteRequest) Descriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{3}
}

func (x *SearchNoteRequest) GetCtx() *common.Context {
	if x != nil {
		return x.Ctx
	}
	return nil
}

func (x *SearchNoteRequest) GetAccountId() string {
	if x != nil && x.AccountId != nil {
		return *x.AccountId
	}
	return ""
}

func (x *SearchNoteRequest) GetAgentId() string {
	if x != nil && x.AgentId != nil {
		return *x.AgentId
	}
	return ""
}

func (x *SearchNoteRequest) GetKeyword() string {
	if x != nil && x.Keyword != nil {
		return *x.Keyword
	}
	return ""
}

func (x *SearchNoteRequest) GetFromSec() int64 {
	if x != nil && x.FromSec != nil {
		return *x.FromSec
	}
	return 0
}

func (x *SearchNoteRequest) GetToSec() int64 {
	if x != nil && x.ToSec != nil {
		return *x.ToSec
	}
	return 0
}

func (x *SearchNoteRequest) GetLimit() int64 {
	if x != nil && x.Limit != nil {
		return *x.Limit
	}
	return 0
}

func (x *SearchNoteRequest) GetAnchor() string {
	if x != nil && x.Anchor != nil {
		return *x.Anchor
	}
	return ""
}

type SearchNoteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Total  *int64    `protobuf:"varint,2,opt,name=total" json:"total,omitempty"`
	Result []*ESNote `protobuf:"bytes,3,rep,name=result" json:"result,omitempty"`
	Anchor *string   `protobuf:"bytes,4,opt,name=anchor" json:"anchor,omitempty"`
}

func (x *SearchNoteResponse) Reset() {
	*x = SearchNoteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchNoteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchNoteResponse) ProtoMessage() {}

func (x *SearchNoteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_user_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchNoteResponse.ProtoReflect.Descriptor instead.
func (*SearchNoteResponse) Descriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{4}
}

func (x *SearchNoteResponse) GetTotal() int64 {
	if x != nil && x.Total != nil {
		return *x.Total
	}
	return 0
}

func (x *SearchNoteResponse) GetResult() []*ESNote {
	if x != nil {
		return x.Result
	}
	return nil
}

func (x *SearchNoteResponse) GetAnchor() string {
	if x != nil && x.Anchor != nil {
		return *x.Anchor
	}
	return ""
}

var File_user_proto protoreflect.FileDescriptor

var file_user_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x75, 0x73,
	0x65, 0x72, 0x1a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xf3, 0x03, 0x0a, 0x0a, 0x53, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x1a, 0x0a, 0x08, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x14, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6a, 0x6f, 0x69, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6a, 0x6f, 0x69, 0x6e, 0x12, 0x26, 0x0a, 0x0f, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12,
	0x1d, 0x0a, 0x0a, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x0a, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x30,
	0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x09, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x10, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x53, 0x43, 0x6f, 0x6e, 0x64, 0x69,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0a, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x18, 0x0c, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x12, 0x23, 0x0a, 0x0d,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x65, 0x78, 0x69, 0x73, 0x74, 0x65, 0x64, 0x18, 0x0d, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x0c, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x45, 0x78, 0x69, 0x73, 0x74, 0x65,
	0x64, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0e, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x38, 0x0a,
	0x09, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x6e, 0x6f,
	0x6e, 0x65, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x10,
	0x01, 0x12, 0x0a, 0x0a, 0x06, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x10, 0x02, 0x12, 0x08, 0x0a,
	0x04, 0x70, 0x61, 0x73, 0x74, 0x10, 0x03, 0x22, 0x1f, 0x0a, 0x0c, 0x4a, 0x6f, 0x69, 0x6e, 0x4f,
	0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x07, 0x0a, 0x03, 0x61, 0x6e, 0x64, 0x10, 0x00,
	0x12, 0x06, 0x0a, 0x02, 0x6f, 0x72, 0x10, 0x01, 0x22, 0x4c, 0x0a, 0x08, 0x46, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x11, 0x0a, 0x0d, 0x6d, 0x69, 0x6e, 0x75, 0x74, 0x65, 0x5f, 0x6f,
	0x66, 0x5f, 0x64, 0x61, 0x79, 0x10, 0x00, 0x12, 0x0f, 0x0a, 0x0b, 0x68, 0x6f, 0x75, 0x72, 0x5f,
	0x6f, 0x66, 0x5f, 0x64, 0x61, 0x79, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x64, 0x61, 0x79, 0x5f,
	0x6f, 0x66, 0x5f, 0x77, 0x65, 0x65, 0x6b, 0x10, 0x02, 0x12, 0x0b, 0x0a, 0x07, 0x64, 0x61, 0x79,
	0x5f, 0x61, 0x67, 0x6f, 0x10, 0x03, 0x22, 0xa5, 0x02, 0x0a, 0x11, 0x55, 0x73, 0x65, 0x72, 0x53,
	0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x03,
	0x63, 0x74, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x03, 0x63, 0x74, 0x78, 0x12,
	0x1d, 0x0a, 0x0a, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x27,
	0x0a, 0x0f, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69,
	0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x12, 0x16, 0x0a,
	0x06, 0x61, 0x6e, 0x63, 0x68, 0x6f, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61,
	0x6e, 0x63, 0x68, 0x6f, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x61,
	0x67, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61,
	0x67, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x75, 0x6e, 0x72, 0x65, 0x61, 0x64,
	0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x75, 0x6e, 0x72, 0x65, 0x61, 0x64, 0x12, 0x2e,
	0x0a, 0x09, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0a, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x10, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x53, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x09, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xfd,
	0x01, 0x0a, 0x06, 0x45, 0x53, 0x4e, 0x6f, 0x74, 0x65, 0x12, 0x21, 0x0a, 0x03, 0x63, 0x74, 0x78,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x03, 0x63, 0x74, 0x78, 0x12, 0x1d, 0x0a, 0x0a,
	0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x6f, 0x72, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65,
	0x78, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x12, 0x18,
	0x0a, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74,
	0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x22, 0xea,
	0x01, 0x0a, 0x11, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x4e, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x03, 0x63, 0x74, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0f, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65,
	0x78, 0x74, 0x52, 0x03, 0x63, 0x74, 0x78, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x5f,
	0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x49,
	0x64, 0x12, 0x18, 0x0a, 0x07, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x66,
	0x72, 0x6f, 0x6d, 0x5f, 0x73, 0x65, 0x63, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x66,
	0x72, 0x6f, 0x6d, 0x53, 0x65, 0x63, 0x12, 0x15, 0x0a, 0x06, 0x74, 0x6f, 0x5f, 0x73, 0x65, 0x63,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x74, 0x6f, 0x53, 0x65, 0x63, 0x12, 0x14, 0x0a,
	0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x6c, 0x69,
	0x6d, 0x69, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6e, 0x63, 0x68, 0x6f, 0x72, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x6e, 0x63, 0x68, 0x6f, 0x72, 0x22, 0x68, 0x0a, 0x12, 0x53,
	0x65, 0x61, 0x72, 0x63, 0x68, 0x4e, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x12, 0x24, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x45,
	0x53, 0x4e, 0x6f, 0x74, 0x65, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x16, 0x0a,
	0x06, 0x61, 0x6e, 0x63, 0x68, 0x6f, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61,
	0x6e, 0x63, 0x68, 0x6f, 0x72, 0x42, 0x1e, 0x5a, 0x1c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x75, 0x62, 0x69, 0x7a, 0x2f, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72,
	0x2f, 0x75, 0x73, 0x65, 0x72,
}

var (
	file_user_proto_rawDescOnce sync.Once
	file_user_proto_rawDescData = file_user_proto_rawDesc
)

func file_user_proto_rawDescGZIP() []byte {
	file_user_proto_rawDescOnce.Do(func() {
		file_user_proto_rawDescData = protoimpl.X.CompressGZIP(file_user_proto_rawDescData)
	})
	return file_user_proto_rawDescData
}

var file_user_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_user_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_user_proto_goTypes = []interface{}{
	(SCondition_EventTime)(0),    // 0: user.SCondition.EventTime
	(SCondition_JoinOperator)(0), // 1: user.SCondition.JoinOperator
	(SCondition_Function)(0),     // 2: user.SCondition.Function
	(*SCondition)(nil),           // 3: user.SCondition
	(*UserSearchRequest)(nil),    // 4: user.UserSearchRequest
	(*ESNote)(nil),               // 5: user.ESNote
	(*SearchNoteRequest)(nil),    // 6: user.SearchNoteRequest
	(*SearchNoteResponse)(nil),   // 7: user.SearchNoteResponse
	(*common.Context)(nil),       // 8: common.Context
}
var file_user_proto_depIdxs = []int32{
	3, // 0: user.SCondition.conditions:type_name -> user.SCondition
	8, // 1: user.UserSearchRequest.ctx:type_name -> common.Context
	3, // 2: user.UserSearchRequest.condition:type_name -> user.SCondition
	8, // 3: user.ESNote.ctx:type_name -> common.Context
	8, // 4: user.SearchNoteRequest.ctx:type_name -> common.Context
	5, // 5: user.SearchNoteResponse.result:type_name -> user.ESNote
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_user_proto_init() }
func file_user_proto_init() {
	if File_user_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_user_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SCondition); i {
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
		file_user_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserSearchRequest); i {
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
		file_user_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ESNote); i {
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
		file_user_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchNoteRequest); i {
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
		file_user_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchNoteResponse); i {
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
			RawDescriptor: file_user_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_user_proto_goTypes,
		DependencyIndexes: file_user_proto_depIdxs,
		EnumInfos:         file_user_proto_enumTypes,
		MessageInfos:      file_user_proto_msgTypes,
	}.Build()
	File_user_proto = out.File
	file_user_proto_rawDesc = nil
	file_user_proto_goTypes = nil
	file_user_proto_depIdxs = nil
}
