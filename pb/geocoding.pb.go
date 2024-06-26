// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.26.1
// source: geocoding.proto

package pb

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

type GeocodingApi struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GeocodingApi) Reset() {
	*x = GeocodingApi{}
	if protoimpl.UnsafeEnabled {
		mi := &file_geocoding_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GeocodingApi) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeocodingApi) ProtoMessage() {}

func (x *GeocodingApi) ProtoReflect() protoreflect.Message {
	mi := &file_geocoding_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeocodingApi.ProtoReflect.Descriptor instead.
func (*GeocodingApi) Descriptor() ([]byte, []int) {
	return file_geocoding_proto_rawDescGZIP(), []int{0}
}

type GeocodingApi_SearchResults struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Results          []*GeocodingApi_Geoname `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
	GenerationtimeMs float32                 `protobuf:"fixed32,2,opt,name=generationtime_ms,json=generationtimeMs,proto3" json:"generationtime_ms,omitempty"`
}

func (x *GeocodingApi_SearchResults) Reset() {
	*x = GeocodingApi_SearchResults{}
	if protoimpl.UnsafeEnabled {
		mi := &file_geocoding_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GeocodingApi_SearchResults) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeocodingApi_SearchResults) ProtoMessage() {}

func (x *GeocodingApi_SearchResults) ProtoReflect() protoreflect.Message {
	mi := &file_geocoding_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeocodingApi_SearchResults.ProtoReflect.Descriptor instead.
func (*GeocodingApi_SearchResults) Descriptor() ([]byte, []int) {
	return file_geocoding_proto_rawDescGZIP(), []int{0, 0}
}

func (x *GeocodingApi_SearchResults) GetResults() []*GeocodingApi_Geoname {
	if x != nil {
		return x.Results
	}
	return nil
}

func (x *GeocodingApi_SearchResults) GetGenerationtimeMs() float32 {
	if x != nil {
		return x.GenerationtimeMs
	}
	return 0
}

type GeocodingApi_Geoname struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          int32   `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string  `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Latitude    float32 `protobuf:"fixed32,4,opt,name=latitude,proto3" json:"latitude,omitempty"`
	Longitude   float32 `protobuf:"fixed32,5,opt,name=longitude,proto3" json:"longitude,omitempty"`
	Ranking     float32 `protobuf:"fixed32,6,opt,name=ranking,proto3" json:"ranking,omitempty"`
	Elevation   float32 `protobuf:"fixed32,7,opt,name=elevation,proto3" json:"elevation,omitempty"`
	FeatureCode string  `protobuf:"bytes,8,opt,name=feature_code,json=featureCode,proto3" json:"feature_code,omitempty"`
	CountryCode string  `protobuf:"bytes,9,opt,name=country_code,json=countryCode,proto3" json:"country_code,omitempty"`
	CountryId   int32   `protobuf:"varint,18,opt,name=country_id,json=countryId,proto3" json:"country_id,omitempty"`
	Country     string  `protobuf:"bytes,19,opt,name=country,proto3" json:"country,omitempty"`
	Admin1Id    int32   `protobuf:"varint,10,opt,name=admin1_id,json=admin1Id,proto3" json:"admin1_id,omitempty"`
	Admin2Id    int32   `protobuf:"varint,11,opt,name=admin2_id,json=admin2Id,proto3" json:"admin2_id,omitempty"`
	Admin3Id    int32   `protobuf:"varint,12,opt,name=admin3_id,json=admin3Id,proto3" json:"admin3_id,omitempty"`
	Admin4Id    int32   `protobuf:"varint,13,opt,name=admin4_id,json=admin4Id,proto3" json:"admin4_id,omitempty"`
	Admin1      string  `protobuf:"bytes,20,opt,name=admin1,proto3" json:"admin1,omitempty"`
	Admin2      string  `protobuf:"bytes,21,opt,name=admin2,proto3" json:"admin2,omitempty"`
	Admin3      string  `protobuf:"bytes,22,opt,name=admin3,proto3" json:"admin3,omitempty"`
	Admin4      string  `protobuf:"bytes,23,opt,name=admin4,proto3" json:"admin4,omitempty"`
	Timezone    string  `protobuf:"bytes,14,opt,name=timezone,proto3" json:"timezone,omitempty"`
	Population  uint32  `protobuf:"varint,15,opt,name=population,proto3" json:"population,omitempty"`
	// map<int32, string> alternativeNames = 16;
	Postcodes []string `protobuf:"bytes,17,rep,name=postcodes,proto3" json:"postcodes,omitempty"`
}

func (x *GeocodingApi_Geoname) Reset() {
	*x = GeocodingApi_Geoname{}
	if protoimpl.UnsafeEnabled {
		mi := &file_geocoding_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GeocodingApi_Geoname) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeocodingApi_Geoname) ProtoMessage() {}

func (x *GeocodingApi_Geoname) ProtoReflect() protoreflect.Message {
	mi := &file_geocoding_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeocodingApi_Geoname.ProtoReflect.Descriptor instead.
func (*GeocodingApi_Geoname) Descriptor() ([]byte, []int) {
	return file_geocoding_proto_rawDescGZIP(), []int{0, 1}
}

func (x *GeocodingApi_Geoname) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *GeocodingApi_Geoname) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GeocodingApi_Geoname) GetLatitude() float32 {
	if x != nil {
		return x.Latitude
	}
	return 0
}

func (x *GeocodingApi_Geoname) GetLongitude() float32 {
	if x != nil {
		return x.Longitude
	}
	return 0
}

func (x *GeocodingApi_Geoname) GetRanking() float32 {
	if x != nil {
		return x.Ranking
	}
	return 0
}

func (x *GeocodingApi_Geoname) GetElevation() float32 {
	if x != nil {
		return x.Elevation
	}
	return 0
}

func (x *GeocodingApi_Geoname) GetFeatureCode() string {
	if x != nil {
		return x.FeatureCode
	}
	return ""
}

func (x *GeocodingApi_Geoname) GetCountryCode() string {
	if x != nil {
		return x.CountryCode
	}
	return ""
}

func (x *GeocodingApi_Geoname) GetCountryId() int32 {
	if x != nil {
		return x.CountryId
	}
	return 0
}

func (x *GeocodingApi_Geoname) GetCountry() string {
	if x != nil {
		return x.Country
	}
	return ""
}

func (x *GeocodingApi_Geoname) GetAdmin1Id() int32 {
	if x != nil {
		return x.Admin1Id
	}
	return 0
}

func (x *GeocodingApi_Geoname) GetAdmin2Id() int32 {
	if x != nil {
		return x.Admin2Id
	}
	return 0
}

func (x *GeocodingApi_Geoname) GetAdmin3Id() int32 {
	if x != nil {
		return x.Admin3Id
	}
	return 0
}

func (x *GeocodingApi_Geoname) GetAdmin4Id() int32 {
	if x != nil {
		return x.Admin4Id
	}
	return 0
}

func (x *GeocodingApi_Geoname) GetAdmin1() string {
	if x != nil {
		return x.Admin1
	}
	return ""
}

func (x *GeocodingApi_Geoname) GetAdmin2() string {
	if x != nil {
		return x.Admin2
	}
	return ""
}

func (x *GeocodingApi_Geoname) GetAdmin3() string {
	if x != nil {
		return x.Admin3
	}
	return ""
}

func (x *GeocodingApi_Geoname) GetAdmin4() string {
	if x != nil {
		return x.Admin4
	}
	return ""
}

func (x *GeocodingApi_Geoname) GetTimezone() string {
	if x != nil {
		return x.Timezone
	}
	return ""
}

func (x *GeocodingApi_Geoname) GetPopulation() uint32 {
	if x != nil {
		return x.Population
	}
	return 0
}

func (x *GeocodingApi_Geoname) GetPostcodes() []string {
	if x != nil {
		return x.Postcodes
	}
	return nil
}

var File_geocoding_proto protoreflect.FileDescriptor

var file_geocoding_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x67, 0x65, 0x6f, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xcc, 0x05, 0x0a, 0x0c, 0x47, 0x65, 0x6f, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x41,
	0x70, 0x69, 0x1a, 0x6d, 0x0a, 0x0d, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x73, 0x12, 0x2f, 0x0a, 0x07, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x47, 0x65, 0x6f, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67,
	0x41, 0x70, 0x69, 0x2e, 0x47, 0x65, 0x6f, 0x6e, 0x61, 0x6d, 0x65, 0x52, 0x07, 0x72, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x73, 0x12, 0x2b, 0x0a, 0x11, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x02, 0x52,
	0x10, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x4d,
	0x73, 0x1a, 0xcc, 0x04, 0x0a, 0x07, 0x47, 0x65, 0x6f, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x02, 0x52, 0x08, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x1c, 0x0a,
	0x09, 0x6c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x02,
	0x52, 0x09, 0x6c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x72,
	0x61, 0x6e, 0x6b, 0x69, 0x6e, 0x67, 0x18, 0x06, 0x20, 0x01, 0x28, 0x02, 0x52, 0x07, 0x72, 0x61,
	0x6e, 0x6b, 0x69, 0x6e, 0x67, 0x12, 0x1c, 0x0a, 0x09, 0x65, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x02, 0x52, 0x09, 0x65, 0x6c, 0x65, 0x76, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x21, 0x0a, 0x0c, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x63,
	0x6f, 0x64, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x66, 0x65, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72,
	0x79, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x72, 0x79, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x72, 0x79, 0x5f, 0x69, 0x64, 0x18, 0x12, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x72, 0x79, 0x18, 0x13, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x1b, 0x0a, 0x09, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x31, 0x5f, 0x69, 0x64, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x31, 0x49, 0x64, 0x12,
	0x1b, 0x0a, 0x09, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x32, 0x5f, 0x69, 0x64, 0x18, 0x0b, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x08, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x32, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x33, 0x5f, 0x69, 0x64, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x08, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x33, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x34, 0x5f, 0x69, 0x64, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x34, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x31,
	0x18, 0x14, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x31, 0x12, 0x16,
	0x0a, 0x06, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x32, 0x18, 0x15, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x32, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x33,
	0x18, 0x16, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x33, 0x12, 0x16,
	0x0a, 0x06, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x34, 0x18, 0x17, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x34, 0x12, 0x1a, 0x0a, 0x08, 0x74, 0x69, 0x6d, 0x65, 0x7a, 0x6f,
	0x6e, 0x65, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x69, 0x6d, 0x65, 0x7a, 0x6f,
	0x6e, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x6f, 0x70, 0x75, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x0f, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x70, 0x6f, 0x70, 0x75, 0x6c, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x6f, 0x73, 0x74, 0x63, 0x6f, 0x64, 0x65, 0x73, 0x18,
	0x11, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x70, 0x6f, 0x73, 0x74, 0x63, 0x6f, 0x64, 0x65, 0x73,
	0x42, 0x05, 0x5a, 0x03, 0x70, 0x62, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_geocoding_proto_rawDescOnce sync.Once
	file_geocoding_proto_rawDescData = file_geocoding_proto_rawDesc
)

func file_geocoding_proto_rawDescGZIP() []byte {
	file_geocoding_proto_rawDescOnce.Do(func() {
		file_geocoding_proto_rawDescData = protoimpl.X.CompressGZIP(file_geocoding_proto_rawDescData)
	})
	return file_geocoding_proto_rawDescData
}

var file_geocoding_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_geocoding_proto_goTypes = []interface{}{
	(*GeocodingApi)(nil),               // 0: GeocodingApi
	(*GeocodingApi_SearchResults)(nil), // 1: GeocodingApi.SearchResults
	(*GeocodingApi_Geoname)(nil),       // 2: GeocodingApi.Geoname
}
var file_geocoding_proto_depIdxs = []int32{
	2, // 0: GeocodingApi.SearchResults.results:type_name -> GeocodingApi.Geoname
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_geocoding_proto_init() }
func file_geocoding_proto_init() {
	if File_geocoding_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_geocoding_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GeocodingApi); i {
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
		file_geocoding_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GeocodingApi_SearchResults); i {
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
		file_geocoding_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GeocodingApi_Geoname); i {
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
			RawDescriptor: file_geocoding_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_geocoding_proto_goTypes,
		DependencyIndexes: file_geocoding_proto_depIdxs,
		MessageInfos:      file_geocoding_proto_msgTypes,
	}.Build()
	File_geocoding_proto = out.File
	file_geocoding_proto_rawDesc = nil
	file_geocoding_proto_goTypes = nil
	file_geocoding_proto_depIdxs = nil
}
