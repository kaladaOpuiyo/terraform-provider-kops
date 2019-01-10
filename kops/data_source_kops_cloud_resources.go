package kops

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/pkg/resources"
	resourceops "k8s.io/kops/pkg/resources/ops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/vfs"
)

func dataSourceKopsCloudResources() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceKopsCloudResourcesRead,

		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:        schema.TypeString,
				Description: "Name of cluster",
				Required:    true,
				ForceNew:    true,
			},
			"dhcp_options_id": {
				Type:     schema.TypeString,
				Default:  nil,
				Computed: true,
			},
			"internet_gateway_id": {
				Type:     schema.TypeString,
				Default:  nil,
				Computed: true,
			},
			"keypair_id": {
				Type:     schema.TypeString,
				Default:  nil,
				Computed: true,
			},
			"load_balancer_id": {
				Type:     schema.TypeString,
				Default:  nil,
				Computed: true,
			},
			"route_table_id": {
				Type:     schema.TypeString,
				Default:  nil,
				Computed: true,
			},
			"state_store": {
				Type:        schema.TypeString,
				Description: "State Store",
				Required:    true,
				ForceNew:    true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Default:  nil,
				Computed: true,
			},
		},
	}
}
func dataSourceKopsCloudResourcesRead(d *schema.ResourceData, meta interface{}) error {

	var cloud fi.Cloud

	name := d.Get("cluster_name").(string)
	d.SetId(name)

	registryBase, err := vfs.Context.BuildVfsPath(d.Get("state_store").(string))
	if err != nil {
		return fmt.Errorf("error parsing registry path %q: %v", d.Get("state_store").(string), err)
	}
	allowList := true

	clientset := vfsclientset.NewVFSClientset(registryBase, allowList)

	log.Printf("[INFO] Reading Kops Cluster %s", name)
	cluster, err := clientset.GetCluster(name)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}

	cloud, err = cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	allResources, err := resourceops.ListResources(cloud, name, "")
	if err != nil {
		return err
	}

	clusterResources := make(map[string]*resources.Resource)
	for k, resource := range allResources {
		if resource.Shared {
			continue
		}
		clusterResources[k] = resource
	}
	if len(clusterResources) == 0 {
		fmt.Printf("No cloud resources found")
	} else {

		// to have conditon for id that may not exist e.g. bastion host
		for _, v := range clusterResources {

			if strings.Contains(v.Type, "vpc") {
				d.Set("vpc_id", v.ID)
			} else if strings.Contains(v.Type, "dhcp") {
				d.Set("dhcp_options_id", v.ID)
			} else if strings.Contains(v.Type, "internet-gateway") {
				d.Set("internet_gateway_id", v.ID)
			} else if strings.Contains(v.Type, "route-table") {
				d.Set("route_table_id", v.ID)
			} else if strings.Contains(v.Type, "load-balancer") {
				d.Set("load_balancer_id", v.ID)
			} else if strings.Contains(v.Type, "keypair") {
				d.Set("keypair_id", v.ID)

			}
		}
	}

	return nil
}
