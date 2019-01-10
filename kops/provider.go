package kops

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider Func
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"kops_cloud_resources": dataSourceKopsCloudResources(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"kops_cluster": resourceKopsCluster(),
		},
	}
}
