package ncloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	RegisterDataSource("ncloud_sourcebuild_project_docker", dataSourceNcloudSourceBuildDocker())
}

func dataSourceNcloudSourceBuildDocker() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNcloudSourceBuildDockerRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"docker": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNcloudSourceBuildDockerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*ProviderConfig)

	logCommonRequest("GetDockerEnv", "")
	resp, err := config.Client.sourcebuild.V1Api.GetDockerEnv(context.Background())
	if err != nil {
		logErrorResponse("GetDockerEnv", err, "")
		return diag.FromErr(err)
	}
	logResponse("GetDockerEnv", resp)

	resources := []map[string]interface{}{}

	for _, r := range resp.Docker {
		docker := map[string]interface{}{
			"id":   *r.Id,
			"name": *r.Name,
		}

		resources = append(resources, docker)
	}

	if f, ok := d.GetOk("filter"); ok {
		resources = ApplyFilters(f.(*schema.Set), resources, dataSourceNcloudSourceBuildDocker().Schema)
	}

	d.SetId(config.RegionCode)
	d.Set("docker", resources)

	return nil
}
