package server

import (
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	. "github.com/terraform-providers/terraform-provider-ncloud/internal/common"
	"github.com/terraform-providers/terraform-provider-ncloud/internal/conn"
	"github.com/terraform-providers/terraform-provider-ncloud/internal/verify"
)

func DataSourceNcloudAccessControlGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNcloudAccessControlGroupRead,

		Schema: map[string]*schema.Schema{
			"configuration_no": {
				Type:          schema.TypeString,
				Optional:      true,
				Deprecated:    "use 'id' instead",
				ConflictsWith: []string{"id"},
			},
			"is_default_group": {
				Type:          schema.TypeBool,
				Optional:      true,
				Deprecated:    "use 'is_default' instead",
				ConflictsWith: []string{"is_default"},
			},
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"configuration_no"},
			},
			"is_default": {
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"is_default_group"},
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vpc_no": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"access_control_group_no": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
		},
	}
}

func dataSourceNcloudAccessControlGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*conn.ProviderConfig)
	var resources []map[string]interface{}
	var err error

	if config.SupportVPC {
		resources, err = getVpcAccessControlGroupList(d, config)
	} else {
		resources, err = getClassicAccessControlGroupList(d, config)
	}

	if err != nil {
		return err
	}

	if f, ok := d.GetOk("filter"); ok {
		resources = ApplyFilters(f.(*schema.Set), resources, DataSourceNcloudAccessControlGroup().Schema)
	}

	if err := verify.ValidateOneResult(len(resources)); err != nil {
		return err
	}

	SetSingularResourceDataFromMap(d, resources[0])

	return nil
}

func getVpcAccessControlGroupList(d *schema.ResourceData, config *conn.ProviderConfig) ([]map[string]interface{}, error) {
	reqParams := &vserver.GetAccessControlGroupListRequest{
		RegionCode:             &config.RegionCode,
		AccessControlGroupName: StringPtrOrNil(d.GetOk("name")),
		VpcNo:                  StringPtrOrNil(d.GetOk("vpc_no")),
	}

	if v, ok := d.GetOk("id"); ok {
		reqParams.AccessControlGroupNoList = []*string{ncloud.String(v.(string))}
	}

	LogCommonRequest("getVpcAccessControlGroup", reqParams)

	resp, err := config.Client.Vserver.V2Api.GetAccessControlGroupList(reqParams)
	if err != nil {
		LogErrorResponse("getVpcAccessControlGroup", err, reqParams)
		return nil, err
	}
	LogResponse("getVpcAccessControlGroup", resp)

	var resources []map[string]interface{}
	for _, r := range resp.AccessControlGroupList {
		instance := map[string]interface{}{
			"id":                      *r.AccessControlGroupNo,
			"access_control_group_no": *r.AccessControlGroupNo,
			"name":                    *r.AccessControlGroupName,
			"description":             *r.AccessControlGroupDescription,
			"is_default":              *r.IsDefault,
			"vpc_no":                  *r.VpcNo,
		}

		resources = append(resources, instance)
	}

	return resources, nil
}

func getClassicAccessControlGroupList(d *schema.ResourceData, config *conn.ProviderConfig) ([]map[string]interface{}, error) {
	client := config.Client

	reqParams := server.GetAccessControlGroupListRequest{
		AccessControlGroupName: StringPtrOrNil(d.GetOk("name")),
	}

	if v, ok := d.GetOk("id"); ok {
		reqParams.AccessControlGroupConfigurationNoList = ExpandStringInterfaceList(v.([]interface{}))
	} else if v, ok := d.GetOk("configuration_no"); ok {
		reqParams.AccessControlGroupConfigurationNoList = ExpandStringInterfaceList(v.([]interface{}))
	}

	if v, ok := d.GetOk("is_default"); ok {
		reqParams.IsDefault = ncloud.Bool(v.(bool))
	} else if v, ok := d.GetOk("is_default_group"); ok {
		reqParams.IsDefault = ncloud.Bool(v.(bool))
	}

	reqParams.PageNo = ncloud.Int32(1)

	LogCommonRequest("getClassicAccessControlGroupList", reqParams)

	resp, err := client.Server.V2Api.GetAccessControlGroupList(&reqParams)
	if err != nil {
		LogErrorResponse("getClassicAccessControlGroupList", err, reqParams)
		return nil, err
	}
	LogResponse("getClassicAccessControlGroupList", resp)

	var resources []map[string]interface{}
	for _, r := range resp.AccessControlGroupList {
		instance := map[string]interface{}{
			"id":                      *r.AccessControlGroupConfigurationNo,
			"access_control_group_no": *r.AccessControlGroupConfigurationNo,
			"name":                    *r.AccessControlGroupName,
			"description":             *r.AccessControlGroupDescription,
			"configuration_no":        *r.AccessControlGroupConfigurationNo, // To deprecated
		}

		if r.IsDefaultGroup != nil {
			instance["is_default"] = *r.IsDefaultGroup
			instance["is_default_group"] = *r.IsDefaultGroup // To deprecated
		}

		resources = append(resources, instance)
	}

	return resources, nil
}
