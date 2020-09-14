package ncloud

import (
	"fmt"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"time"

	"log"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceNcloudBlockStorage() *schema.Resource {
	return &schema.Resource{
		Create: resourceNcloudBlockStorageCreate,
		Read:   resourceNcloudBlockStorageRead,
		Delete: resourceNcloudBlockStorageDelete,
		Update: resourceNcloudBlockStorageUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(DefaultCreateTimeout),
			Delete: schema.DefaultTimeout(DefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"server_instance_no": {
				Type:     schema.TypeString,
				Required: true,
			},
			"size": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(10, 1000),
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"disk_detail_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"snapshot_no": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"block_storage_no": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"device_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"product_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"operation": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_status": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "Use `status` instead",
			},
			"instance_operation": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "Use `operation` instead",
			},
			"instance_status_name": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "This field no longer support",
			},
			"instance_no": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "Use `block_storage_no` instead",
			},
		},
	}
}

func resourceNcloudBlockStorageCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)

	id, err := createBlockStorage(d, config)
	if err != nil {
		return err
	}

	d.SetId(ncloud.StringValue(id))
	log.Printf("[INFO] Block Storage ID: %s", d.Id())

	return resourceNcloudBlockStorageRead(d, meta)
}

func resourceNcloudBlockStorageRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)

	r, err := getBlockStorage(config, d.Id())
	if err != nil {
		return err
	}

	if r == nil {
		d.SetId("")
	}

	instance := ConvertToMap(r)

	SetSingularResourceDataFromMapSchema(resourceNcloudBlockStorage(), d, instance)

	if !config.SupportVPC {
		// Set deprecated field on classic
		d.Set("instance_operation", r.Operation)
		d.Set("instance_status", r.Status)
		d.Set("instance_no", r.BlockStorageInstanceNo)
	}

	return nil
}

func resourceNcloudBlockStorageDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)

	if err := deleteBlockStorage(d, config, d.Id()); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func resourceNcloudBlockStorageUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceNcloudBlockStorageRead(d, meta)
}

func createBlockStorage(d *schema.ResourceData, config *ProviderConfig) (*string, error) {
	var id *string
	var err error

	if config.SupportVPC {
		id, err = createVpcBlockStorage(d, config)
	} else {
		id, err = createClassicBlockStorage(d, config)
	}

	if err != nil {
		return nil, err
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{"INIT", "CREAT"},
		Target:  []string{"ATTAC"},
		Refresh: func() (interface{}, string, error) {
			instance, err := getBlockStorage(config, ncloud.StringValue(id))
			if err != nil {
				return 0, "", err
			}
			return instance, ncloud.StringValue(instance.Status), nil
		},
		Timeout:    DefaultCreateTimeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return nil, fmt.Errorf("error waiting for BlockStorageInstance state to be \"ATTAC\": %s", err)
	}

	return id, nil
}

func createClassicBlockStorage(d *schema.ResourceData, config *ProviderConfig) (*string, error) {
	reqParams := &server.CreateBlockStorageInstanceRequest{
		ServerInstanceNo:        ncloud.String(d.Get("server_instance_no").(string)),
		BlockStorageSize:        ncloud.Int64(int64(d.Get("size").(int))),
		BlockStorageName:        StringPtrOrNil(d.GetOk("name")),
		BlockStorageDescription: StringPtrOrNil(d.GetOk("description")),
		DiskDetailTypeCode:      StringPtrOrNil(d.GetOk("disk_detail_type")),
	}

	logCommonRequest("createClassicBlockStorage", reqParams)

	resp, err := config.Client.server.V2Api.CreateBlockStorageInstance(reqParams)
	if err != nil {
		logErrorResponse("createClassicBlockStorage", err, reqParams)
		return nil, err
	}
	logCommonResponse("createClassicBlockStorage", GetCommonResponse(resp))

	instance := resp.BlockStorageInstanceList[0]

	return instance.BlockStorageInstanceNo, nil
}

func createVpcBlockStorage(d *schema.ResourceData, config *ProviderConfig) (*string, error) {
	reqParams := &vserver.CreateBlockStorageInstanceRequest{
		RegionCode:                     &config.RegionCode,
		BlockStorageSize:               ncloud.Int32(int32(d.Get("size").(int))),
		ServerInstanceNo:               ncloud.String(d.Get("server_instance_no").(string)),
		BlockStorageName:               StringPtrOrNil(d.GetOk("name")),
		BlockStorageDescription:        StringPtrOrNil(d.GetOk("description")),
		BlockStorageDiskDetailTypeCode: StringPtrOrNil(d.GetOk("disk_detail_type")),
		BlockStorageSnapshotInstanceNo: StringPtrOrNil(d.GetOk("snapshot_no")),
		ZoneCode:                       StringPtrOrNil(d.GetOk("zone")),
	}

	logCommonRequest("createVpcBlockStorage", reqParams)

	resp, err := config.Client.vserver.V2Api.CreateBlockStorageInstance(reqParams)
	if err != nil {
		logErrorResponse("createVpcBlockStorage", err, reqParams)
		return nil, err
	}
	logCommonResponse("createVpcBlockStorage", GetCommonResponse(resp))

	instance := resp.BlockStorageInstanceList[0]

	return instance.BlockStorageInstanceNo, nil
}

func getBlockStorage(config *ProviderConfig, id string) (*BlockStorage, error) {
	if config.SupportVPC {
		return getVpcBlockStorage(config, id)
	}

	return getClassicBlockStorage(config, id)
}

func getClassicBlockStorage(config *ProviderConfig, id string) (*BlockStorage, error) {
	reqParams := &server.GetBlockStorageInstanceListRequest{
		BlockStorageInstanceNoList: ncloud.StringList([]string{id}),
	}

	logCommonRequest("getClassicBlockStorage", reqParams)

	resp, err := config.Client.server.V2Api.GetBlockStorageInstanceList(reqParams)
	if err != nil {
		logErrorResponse("getClassicBlockStorage", err, reqParams)
		return nil, err
	}
	logResponse("getClassicBlockStorage", resp)

	if len(resp.BlockStorageInstanceList) > 0 {
		inst := resp.BlockStorageInstanceList[0]

		return &BlockStorage{
			BlockStorageInstanceNo:  inst.BlockStorageInstanceNo,
			ServerInstanceNo:        inst.ServerInstanceNo,
			ServerName:              inst.ServerName,
			BlockStorageType:        inst.BlockStorageType.Code,
			BlockStorageName:        inst.BlockStorageName,
			BlockStorageSize:        ncloud.Int64(*inst.BlockStorageSize / GIGABYTE),
			DeviceName:              inst.DeviceName,
			BlockStorageProductCode: inst.BlockStorageProductCode,
			Status:                  inst.BlockStorageInstanceStatus.Code,
			Operation:               inst.BlockStorageInstanceOperation.Code,
			StatusName:              inst.BlockStorageInstanceStatusName,
			Description:             inst.BlockStorageInstanceDescription,
			DiskType:                inst.DiskType.Code,
			DiskDetailType:          inst.DiskDetailType.Code,
		}, nil
	}

	return nil, nil
}

func getVpcBlockStorage(config *ProviderConfig, id string) (*BlockStorage, error) {
	reqParams := &vserver.GetBlockStorageInstanceDetailRequest{
		RegionCode:             &config.RegionCode,
		BlockStorageInstanceNo: ncloud.String(id),
	}

	logCommonRequest("getVpcBlockStorage", reqParams)

	resp, err := config.Client.vserver.V2Api.GetBlockStorageInstanceDetail(reqParams)
	if err != nil {
		logErrorResponse("getVpcBlockStorage", err, reqParams)
		return nil, err
	}
	logResponse("getVpcBlockStorage", resp)

	if len(resp.BlockStorageInstanceList) > 0 {
		inst := resp.BlockStorageInstanceList[0]

		return &BlockStorage{
			BlockStorageInstanceNo:  inst.BlockStorageInstanceNo,
			ServerInstanceNo:        inst.ServerInstanceNo,
			BlockStorageType:        inst.BlockStorageType.Code,
			BlockStorageName:        inst.BlockStorageName,
			BlockStorageSize:        ncloud.Int64(*inst.BlockStorageSize / GIGABYTE),
			DeviceName:              inst.DeviceName,
			BlockStorageProductCode: inst.BlockStorageProductCode,
			Status:                  inst.BlockStorageInstanceStatus.Code,
			Operation:               inst.BlockStorageInstanceOperation.Code,
			StatusName:              inst.BlockStorageInstanceStatusName,
			Description:             inst.BlockStorageDescription,
			DiskType:                inst.BlockStorageDiskType.Code,
			DiskDetailType:          inst.BlockStorageDiskDetailType.Code,
			ZoneCode:                inst.ZoneCode,
		}, nil
	}

	return nil, nil
}

func deleteBlockStorage(d *schema.ResourceData, config *ProviderConfig, id string) error {
	var err error
	if config.SupportVPC {
		err = deleteVpcBlockStorage(d, config, id)
	} else {
		err = deleteClassicBlockStorage(d, config, id)
	}

	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{"INIT", "ATTAC"},
		Target:  []string{"TERMINATED"},
		Refresh: func() (interface{}, string, error) {
			instance, err := getBlockStorage(config, id)
			if err != nil {
				return 0, "", err
			}
			if instance == nil { // Instance is terminated.
				return instance, "TERMINATED", nil
			}
			return instance, ncloud.StringValue(instance.Status), nil
		},
		Timeout:    DefaultTimeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for BlockStorageInstance state to be \"TERMINATED\": %s", err)
	}

	return nil
}

func deleteClassicBlockStorage(d *schema.ResourceData, config *ProviderConfig, id string) error {
	reqParams := server.DeleteBlockStorageInstancesRequest{
		BlockStorageInstanceNoList: []*string{ncloud.String(id)},
	}

	var resp *server.DeleteBlockStorageInstancesResponse
	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		var err error
		logCommonRequest("deleteClassicBlockStorage", reqParams)

		resp, err = config.Client.server.V2Api.DeleteBlockStorageInstances(&reqParams)
		if err == nil {
			return resource.NonRetryableError(err)
		}

		errBody, _ := GetCommonErrorBody(err)

		if errBody.ReturnCode == ApiErrorDetachingMountedStorage {
			logErrorResponse("retry deleteClassicBlockStorage", err, reqParams)
			time.Sleep(time.Second * 5)
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})

	if err != nil {
		logErrorResponse("deleteClassicBlockStorage", err, reqParams)
		return err
	}
	logResponse("deleteClassicBlockStorage", resp)

	return nil
}

func deleteVpcBlockStorage(d *schema.ResourceData, config *ProviderConfig, id string) error {
	reqParams := vserver.DeleteBlockStorageInstancesRequest{
		BlockStorageInstanceNoList: []*string{ncloud.String(id)},
	}

	var resp *vserver.DeleteBlockStorageInstancesResponse
	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		var err error
		logCommonRequest("deleteVpcBlockStorage", reqParams)

		resp, err = config.Client.vserver.V2Api.DeleteBlockStorageInstances(&reqParams)
		if err == nil {
			return resource.NonRetryableError(err)
		}

		errBody, _ := GetCommonErrorBody(err)

		if errBody.ReturnCode == ApiErrorDetachingMountedStorage {
			logErrorResponse("retry deleteVpcBlockStorage", err, reqParams)
			time.Sleep(time.Second * 5)
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})

	if err != nil {
		logErrorResponse("deleteVpcBlockStorage", err, reqParams)
		return err
	}
	logResponse("deleteVpcBlockStorage", resp)

	return nil
}

//BlockStorage Dto for block storage
type BlockStorage struct {
	BlockStorageInstanceNo  *string `json:"block_storage_no,omitempty"`
	ServerInstanceNo        *string `json:"server_instance_no,omitempty"`
	ServerName              *string `json:"server_name,omitempty"`
	BlockStorageType        *string `json:"type,omitempty"`
	BlockStorageName        *string `json:"name,omitempty"`
	BlockStorageSize        *int64  `json:"size,omitempty"`
	DeviceName              *string `json:"device_name,omitempty"`
	BlockStorageProductCode *string `json:"product_code,omitempty"`
	Status                  *string `json:"status,omitempty"`
	Operation               *string `json:"operation,omitempty"`
	StatusName              *string `json:"status_name,omitempty"`
	Description             *string `json:"description,omitempty"`
	DiskType                *string `json:"disk_type,omitempty"`
	DiskDetailType          *string `json:"disk_detail_type,omitempty"`
	ZoneCode                *string `json:"zone,omitempty"`
}
