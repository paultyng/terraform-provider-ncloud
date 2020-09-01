package ncloud

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
)

func validElem(i interface{}) bool {
	return reflect.ValueOf(i).Elem().IsValid()
}

func validField(f reflect.Value) bool {
	return (!f.CanAddr() || f.CanAddr() && !f.IsNil()) && f.IsValid()
}

func StringField(f reflect.Value) *string {
	if f.Kind() == reflect.Ptr && f.Type().String() == "*string" {
		return f.Interface().(*string)
	} else if f.Kind() == reflect.Slice && f.Type().String() == "string" {
		return ncloud.String(f.Interface().(string))
	}
	return nil
}

func GetCommonResponse(i interface{}) *CommonResponse {
	if i == nil || !validElem(i) {
		return &CommonResponse{}
	}
	var requestId *string
	var returnCode *string
	var returnMessage *string

	if f := reflect.ValueOf(i).Elem().FieldByName("RequestId"); validField(f) {
		requestId = StringField(f)
	}
	if f := reflect.ValueOf(i).Elem().FieldByName("ReturnCode"); validField(f) {
		returnCode = StringField(f)
	}
	if f := reflect.ValueOf(i).Elem().FieldByName("ReturnMessage"); validField(f) {
		returnMessage = StringField(f)
	}
	return &CommonResponse{
		RequestId:     requestId,
		ReturnCode:    returnCode,
		ReturnMessage: returnMessage,
	}
}

//GetCommonErrorBody parse common error message
func GetCommonErrorBody(err error) (*CommonError, error) {
	sa := strings.Split(err.Error(), "Body: ")
	var errMsg string

	if len(sa) != 2 {
		return nil, fmt.Errorf("error body is incorrect: %s", err)
	}

	errMsg = sa[1]

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(errMsg), &m); err != nil {
		return nil, err
	}

	e := m["responseError"].(map[string]interface{})

	return &CommonError{
		ReturnCode:    e["returnCode"].(string),
		ReturnMessage: e["returnMessage"].(string),
	}, nil
}

func GetRegion(i interface{}) *Region {
	if i == nil || !reflect.ValueOf(i).Elem().IsValid() {
		return &Region{}
	}
	var regionNo *string
	var regionCode *string
	var regionName *string
	if f := reflect.ValueOf(i).Elem().FieldByName("RegionNo"); validField(f) {
		regionNo = StringField(f)
	}
	if f := reflect.ValueOf(i).Elem().FieldByName("RegionCode"); validField(f) {
		regionCode = StringField(f)
	}
	if f := reflect.ValueOf(i).Elem().FieldByName("RegionName"); validField(f) {
		regionName = StringField(f)
	}

	return &Region{
		RegionNo:   regionNo,
		RegionCode: regionCode,
		RegionName: regionName,
	}
}

func GetZone(i interface{}) *Zone {
	if i == nil || !reflect.ValueOf(i).Elem().IsValid() {
		return &Zone{}
	}
	var zoneNo *string
	var zoneDescription *string
	var zoneName *string
	var zoneCode *string
	var regionNo *string
	var regionCode *string

	if f := reflect.ValueOf(i).Elem().FieldByName("ZoneNo"); validField(f) {
		zoneNo = StringField(f)
	}
	if f := reflect.ValueOf(i).Elem().FieldByName("ZoneName"); validField(f) {
		zoneName = StringField(f)
	}
	if f := reflect.ValueOf(i).Elem().FieldByName("ZoneCode"); validField(f) {
		zoneCode = StringField(f)
	}
	if f := reflect.ValueOf(i).Elem().FieldByName("ZoneDescription"); validField(f) {
		zoneDescription = StringField(f)
	}
	if f := reflect.ValueOf(i).Elem().FieldByName("RegionNo"); validField(f) {
		regionNo = StringField(f)
	}
	if f := reflect.ValueOf(i).Elem().FieldByName("RegionCode"); validField(f) {
		regionCode = StringField(f)
	}

	return &Zone{
		ZoneNo:          zoneNo,
		ZoneName:        zoneName,
		ZoneCode:        zoneCode,
		ZoneDescription: zoneDescription,
		RegionNo:        regionNo,
		RegionCode:      regionCode,
	}
}

//StringPtrOrNil return *string from interface{}
func StringPtrOrNil(v interface{}, ok bool) *string {
	if !ok {
		return nil
	}
	return ncloud.String(v.(string))
}

//StringOrEmpty Get string from *pointer
func StringOrEmpty(v *string) string {
	if v != nil {
		return *v
	}

	return ""
}

//StringPtrArrToStringArr Convert []*string to []string
func StringPtrArrToStringArr(ptrArray []*string) []string {
	var arr []string
	for _, v := range ptrArray {
		arr = append(arr, *v)
	}

	return arr
}
