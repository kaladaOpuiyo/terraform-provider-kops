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
		Read:   dataSourceKopsCloudResourcesRead,
		Schema: kopsCloudResourcesSchema(),
	}
}
func dataSourceKopsCloudResourcesRead(d *schema.ResourceData, meta interface{}) error {

	var cloud fi.Cloud

	name := d.Get("cluster_name").(string)
	d.SetId(name)

	autoScalingConfigMasters := []string{}
	autoScalingConfigNodes := []string{}
	autoScalingGroupMasters := []string{}
	autoScalingGroupNodes := []string{}
	iamInstanceProfileMasters := []string{}
	iamInstanceProfileNodes := []string{}
	iamRoleMasters := []string{}
	iamRoleNodes := []string{}
	instanceMasters := []string{}
	instanceNodes := []string{}
	instanceBastion := []string{}
	route53RecordsAPI := []string{}
	route53RecordsEtcd := []string{}
	securityGroupMasters := []string{}
	securityGroupNodes := []string{}
	securityGroupELBs := []string{}
	subnets := []string{}
	etcdVolumes := []string{}

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

		for _, v := range clusterResources {

			switch v.Type {

			case "vpc":
				d.Set("vpc_id", v.ID)
			case "dhcp-options":
				d.Set("dhcp_options_id", v.ID)

			case "internet-gateway":
				d.Set("internet_gateway_id", v.ID)

			case "route-table":
				d.Set("route_table_id", v.ID)

			case "load-balancer":
				d.Set("load_balancer_id", v.ID)

			case "keypair":
				d.Set("keypair_id", v.ID)

			case "autoscaling-config":
				if strings.Contains(v.Name, "master") {
					autoScalingConfigMasters = append(autoScalingConfigMasters, v.ID)
				} else {
					autoScalingConfigNodes = append(autoScalingConfigNodes, v.ID)
				}

			case "autoscaling-group":
				if strings.Contains(v.Name, "master") {
					autoScalingGroupMasters = append(autoScalingGroupMasters, v.ID)
				} else {
					autoScalingGroupNodes = append(autoScalingGroupNodes, v.ID)
				}

			case "iam-instance-profile":
				if strings.Contains(v.Name, "master") {
					iamInstanceProfileMasters = append(iamInstanceProfileMasters, v.ID)
				} else {
					iamInstanceProfileNodes = append(iamInstanceProfileNodes, v.ID)
				}

			case "iam-role":
				if strings.Contains(v.Name, "master") {
					iamRoleMasters = append(iamRoleMasters, v.ID)
				} else {
					iamRoleNodes = append(iamRoleNodes, v.ID)
				}

			case "instance":
				if strings.Contains(v.Name, "master") {
					instanceMasters = append(instanceMasters, v.ID)
				} else if strings.Contains(v.Name, "nodes") {
					instanceNodes = append(instanceNodes, v.ID)
				} else {
					instanceBastion = append(instanceBastion, v.ID)
				}

			case "route53-record":
				if strings.Contains(v.Name, "api") {
					route53RecordsAPI = append(route53RecordsAPI, v.ID)

				} else if strings.Contains(v.Name, "etcd") {
					route53RecordsEtcd = append(route53RecordsEtcd, v.ID)
				}

			case "security-group":
				if strings.Contains(v.Name, "master") {
					securityGroupMasters = append(securityGroupMasters, v.ID)
				} else if strings.Contains(v.Name, "nodes") {
					securityGroupNodes = append(securityGroupNodes, v.ID)
				} else if strings.Contains(v.Name, "elb") {
					securityGroupELBs = append(securityGroupELBs, v.ID)
				}

			case "subnet":
				subnets = append(subnets, v.ID)
			case "volume":
				etcdVolumes = append(etcdVolumes, v.ID)

			default:
				fmt.Printf("The resource type %s, is not implemented.\nid: %s", v.Name, v.ID)
			}
		}

	}

	d.Set("autoscaling_config_masters_id", autoScalingConfigMasters)
	d.Set("autoscaling_config_nodes_id", autoScalingConfigNodes)
	d.Set("autoscaling_group_masters_id", autoScalingGroupMasters)
	d.Set("autoscaling_group_nodes_id", autoScalingGroupNodes)
	d.Set("iam_instance_profile_masters_id", iamInstanceProfileMasters)
	d.Set("iam_instance_profile_nodes_id", iamInstanceProfileNodes)
	d.Set("iam_role_masters_id", iamRoleMasters)
	d.Set("iam_role_nodes_id", iamRoleNodes)
	d.Set("instance_masters_id", instanceMasters)
	d.Set("instance_nodes_id", instanceNodes)
	d.Set("instance_bastion_id", instanceBastion)
	d.Set("route53_records_api", route53RecordsAPI)
	d.Set("route53_records_etcd_id", route53RecordsEtcd)
	d.Set("security_group_masters_id", securityGroupMasters)
	d.Set("security_group_nodes_id", securityGroupNodes)
	d.Set("security_group_elbs_id", securityGroupELBs)
	d.Set("subnets_id", subnets)
	d.Set("etcd_volumes_id", etcdVolumes)

	return nil
}
