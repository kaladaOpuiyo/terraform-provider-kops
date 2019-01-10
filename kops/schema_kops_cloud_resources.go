package kops

import "github.com/hashicorp/terraform/helper/schema"

func kopsCloudResourcesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"autoscaling_config_masters_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"autoscaling_config_nodes_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"autoscaling_group_nodes_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"autoscaling_group_masters_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"iam_instance_profile_masters_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"iam_instance_profile_nodes_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		"iam_role_masters_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"iam_role_nodes_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"instance_masters_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"instance_nodes_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"instance_bastion_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"route53_records_api": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"route53_records_etcd_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"security_group_masters_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"security_group_nodes_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"security_group_elbs_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"subnets_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"etcd_volumes_id": {
			Type:     schema.TypeList,
			Default:  nil,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
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
	}

}
