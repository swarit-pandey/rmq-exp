// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v3.19.6
// source: examples/book/book.proto

package book

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

type EnumSample int32

const (
	EnumSample_UNKNOWN EnumSample = 0
	EnumSample_STARTED EnumSample = 1
	EnumSample_RUNNING EnumSample = 1
)

// Enum value maps for EnumSample.
var (
	EnumSample_name = map[int32]string{
		0: "UNKNOWN",
		1: "STARTED",
		// Duplicate value: 1: "RUNNING",
	}
	EnumSample_value = map[string]int32{
		"UNKNOWN": 0,
		"STARTED": 1,
		"RUNNING": 1,
	}
)

func (x EnumSample) Enum() *EnumSample {
	p := new(EnumSample)
	*p = x
	return p
}

func (x EnumSample) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (EnumSample) Descriptor() protoreflect.EnumDescriptor {
	return file_examples_book_book_proto_enumTypes[0].Descriptor()
}

func (EnumSample) Type() protoreflect.EnumType {
	return &file_examples_book_book_proto_enumTypes[0]
}

func (x EnumSample) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use EnumSample.Descriptor instead.
func (EnumSample) EnumDescriptor() ([]byte, []int) {
	return file_examples_book_book_proto_rawDescGZIP(), []int{0}
}

type Book struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Isbn   int64  `protobuf:"varint,1,opt,name=isbn,proto3" json:"isbn,omitempty"`
	Title  string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Author string `protobuf:"bytes,3,opt,name=author,proto3" json:"author,omitempty"`
}

func (x *Book) Reset() {
	*x = Book{}
	mi := &file_examples_book_book_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Book) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Book) ProtoMessage() {}

func (x *Book) ProtoReflect() protoreflect.Message {
	mi := &file_examples_book_book_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Book.ProtoReflect.Descriptor instead.
func (*Book) Descriptor() ([]byte, []int) {
	return file_examples_book_book_proto_rawDescGZIP(), []int{0}
}

func (x *Book) GetIsbn() int64 {
	if x != nil {
		return x.Isbn
	}
	return 0
}

func (x *Book) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Book) GetAuthor() string {
	if x != nil {
		return x.Author
	}
	return ""
}

type GetBookRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Isbn int64 `protobuf:"varint,1,opt,name=isbn,proto3" json:"isbn,omitempty"`
}

func (x *GetBookRequest) Reset() {
	*x = GetBookRequest{}
	mi := &file_examples_book_book_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetBookRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBookRequest) ProtoMessage() {}

func (x *GetBookRequest) ProtoReflect() protoreflect.Message {
	mi := &file_examples_book_book_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBookRequest.ProtoReflect.Descriptor instead.
func (*GetBookRequest) Descriptor() ([]byte, []int) {
	return file_examples_book_book_proto_rawDescGZIP(), []int{1}
}

func (x *GetBookRequest) GetIsbn() int64 {
	if x != nil {
		return x.Isbn
	}
	return 0
}

type GetBookViaAuthor struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Author string `protobuf:"bytes,1,opt,name=author,proto3" json:"author,omitempty"`
}

func (x *GetBookViaAuthor) Reset() {
	*x = GetBookViaAuthor{}
	mi := &file_examples_book_book_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetBookViaAuthor) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBookViaAuthor) ProtoMessage() {}

func (x *GetBookViaAuthor) ProtoReflect() protoreflect.Message {
	mi := &file_examples_book_book_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBookViaAuthor.ProtoReflect.Descriptor instead.
func (*GetBookViaAuthor) Descriptor() ([]byte, []int) {
	return file_examples_book_book_proto_rawDescGZIP(), []int{2}
}

func (x *GetBookViaAuthor) GetAuthor() string {
	if x != nil {
		return x.Author
	}
	return ""
}

type BookStore struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name  string           `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Books map[int64]string `protobuf:"bytes,2,rep,name=books,proto3" json:"books,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *BookStore) Reset() {
	*x = BookStore{}
	mi := &file_examples_book_book_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BookStore) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BookStore) ProtoMessage() {}

func (x *BookStore) ProtoReflect() protoreflect.Message {
	mi := &file_examples_book_book_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BookStore.ProtoReflect.Descriptor instead.
func (*BookStore) Descriptor() ([]byte, []int) {
	return file_examples_book_book_proto_rawDescGZIP(), []int{3}
}

func (x *BookStore) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *BookStore) GetBooks() map[int64]string {
	if x != nil {
		return x.Books
	}
	return nil
}

var File_examples_book_book_proto protoreflect.FileDescriptor

var file_examples_book_book_proto_rawDesc = []byte{
	0x0a, 0x18, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x2f, 0x62, 0x6f, 0x6f, 0x6b, 0x2f,
	0x62, 0x6f, 0x6f, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x62, 0x6f, 0x6f, 0x6b,
	0x22, 0x48, 0x0a, 0x04, 0x42, 0x6f, 0x6f, 0x6b, 0x12, 0x12, 0x0a, 0x04, 0x69, 0x73, 0x62, 0x6e,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x69, 0x73, 0x62, 0x6e, 0x12, 0x14, 0x0a, 0x05,
	0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74,
	0x6c, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x22, 0x24, 0x0a, 0x0e, 0x47, 0x65,
	0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x69, 0x73, 0x62, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x69, 0x73, 0x62, 0x6e,
	0x22, 0x2a, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x56, 0x69, 0x61, 0x41, 0x75,
	0x74, 0x68, 0x6f, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x22, 0x8b, 0x01, 0x0a,
	0x09, 0x42, 0x6f, 0x6f, 0x6b, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x30,
	0x0a, 0x05, 0x62, 0x6f, 0x6f, 0x6b, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x62, 0x6f, 0x6f, 0x6b, 0x2e, 0x42, 0x6f, 0x6f, 0x6b, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x42,
	0x6f, 0x6f, 0x6b, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x62, 0x6f, 0x6f, 0x6b, 0x73,
	0x1a, 0x38, 0x0a, 0x0a, 0x42, 0x6f, 0x6f, 0x6b, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x2a, 0x37, 0x0a, 0x0a, 0x45, 0x6e,
	0x75, 0x6d, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e,
	0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x54, 0x41, 0x52, 0x54, 0x45, 0x44,
	0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x52, 0x55, 0x4e, 0x4e, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x1a,
	0x02, 0x10, 0x01, 0x32, 0xe6, 0x01, 0x0a, 0x0b, 0x42, 0x6f, 0x6f, 0x6b, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x2d, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x12, 0x14,
	0x2e, 0x62, 0x6f, 0x6f, 0x6b, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x0a, 0x2e, 0x62, 0x6f, 0x6f, 0x6b, 0x2e, 0x42, 0x6f, 0x6f, 0x6b,
	0x22, 0x00, 0x12, 0x3b, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x73, 0x56, 0x69,
	0x61, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x12, 0x16, 0x2e, 0x62, 0x6f, 0x6f, 0x6b, 0x2e, 0x47,
	0x65, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x56, 0x69, 0x61, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x1a,
	0x0a, 0x2e, 0x62, 0x6f, 0x6f, 0x6b, 0x2e, 0x42, 0x6f, 0x6f, 0x6b, 0x22, 0x00, 0x30, 0x01, 0x12,
	0x37, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x47, 0x72, 0x65, 0x61, 0x74, 0x65, 0x73, 0x74, 0x42, 0x6f,
	0x6f, 0x6b, 0x12, 0x14, 0x2e, 0x62, 0x6f, 0x6f, 0x6b, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x6f,
	0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0a, 0x2e, 0x62, 0x6f, 0x6f, 0x6b, 0x2e,
	0x42, 0x6f, 0x6f, 0x6b, 0x22, 0x00, 0x28, 0x01, 0x12, 0x32, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x42,
	0x6f, 0x6f, 0x6b, 0x73, 0x12, 0x14, 0x2e, 0x62, 0x6f, 0x6f, 0x6b, 0x2e, 0x47, 0x65, 0x74, 0x42,
	0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0a, 0x2e, 0x62, 0x6f, 0x6f,
	0x6b, 0x2e, 0x42, 0x6f, 0x6f, 0x6b, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x42, 0x30, 0x5a, 0x2e,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x77, 0x61, 0x72, 0x69,
	0x74, 0x2d, 0x70, 0x61, 0x6e, 0x64, 0x65, 0x79, 0x2f, 0x72, 0x6d, 0x71, 0x2d, 0x65, 0x78, 0x70,
	0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x2f, 0x62, 0x6f, 0x6f, 0x6b, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_examples_book_book_proto_rawDescOnce sync.Once
	file_examples_book_book_proto_rawDescData = file_examples_book_book_proto_rawDesc
)

func file_examples_book_book_proto_rawDescGZIP() []byte {
	file_examples_book_book_proto_rawDescOnce.Do(func() {
		file_examples_book_book_proto_rawDescData = protoimpl.X.CompressGZIP(file_examples_book_book_proto_rawDescData)
	})
	return file_examples_book_book_proto_rawDescData
}

var file_examples_book_book_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_examples_book_book_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_examples_book_book_proto_goTypes = []any{
	(EnumSample)(0),          // 0: book.EnumSample
	(*Book)(nil),             // 1: book.Book
	(*GetBookRequest)(nil),   // 2: book.GetBookRequest
	(*GetBookViaAuthor)(nil), // 3: book.GetBookViaAuthor
	(*BookStore)(nil),        // 4: book.BookStore
	nil,                      // 5: book.BookStore.BooksEntry
}
var file_examples_book_book_proto_depIdxs = []int32{
	5, // 0: book.BookStore.books:type_name -> book.BookStore.BooksEntry
	2, // 1: book.BookService.GetBook:input_type -> book.GetBookRequest
	3, // 2: book.BookService.GetBooksViaAuthor:input_type -> book.GetBookViaAuthor
	2, // 3: book.BookService.GetGreatestBook:input_type -> book.GetBookRequest
	2, // 4: book.BookService.GetBooks:input_type -> book.GetBookRequest
	1, // 5: book.BookService.GetBook:output_type -> book.Book
	1, // 6: book.BookService.GetBooksViaAuthor:output_type -> book.Book
	1, // 7: book.BookService.GetGreatestBook:output_type -> book.Book
	1, // 8: book.BookService.GetBooks:output_type -> book.Book
	5, // [5:9] is the sub-list for method output_type
	1, // [1:5] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_examples_book_book_proto_init() }
func file_examples_book_book_proto_init() {
	if File_examples_book_book_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_examples_book_book_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_examples_book_book_proto_goTypes,
		DependencyIndexes: file_examples_book_book_proto_depIdxs,
		EnumInfos:         file_examples_book_book_proto_enumTypes,
		MessageInfos:      file_examples_book_book_proto_msgTypes,
	}.Build()
	File_examples_book_book_proto = out.File
	file_examples_book_book_proto_rawDesc = nil
	file_examples_book_book_proto_goTypes = nil
	file_examples_book_book_proto_depIdxs = nil
}