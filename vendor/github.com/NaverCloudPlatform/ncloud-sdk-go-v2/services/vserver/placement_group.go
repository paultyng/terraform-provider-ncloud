/*
 * vserver
 *
 * VPC Compute 관련 API<br/>https://ncloud.apigw.ntruss.com/vserver/v2
 *
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package vserver

type PlacementGroup struct {

	// 물리배치그룹번호
PlacementGroupNo *string `json:"placementGroupNo,omitempty"`

	// 물리배치그룹이름
PlacementGroupName *string `json:"placementGroupName,omitempty"`

	// 물리배치그룹유형
PlacementGroupType *CommonCode `json:"placementGroupType,omitempty"`
}
