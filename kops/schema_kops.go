package kops

import "github.com/hashicorp/terraform/helper/schema"

func kopsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"admin_access": {
			Type:        schema.TypeList,
			Description: "Admin Access",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"api_load_balancer_type": {
			Type:        schema.TypeString,
			Description: "Sets the API loadbalancer type to either 'public' or 'internal'",
			Optional:    true,
		},
		"associate_public_ip": {
			Type:        schema.TypeBool,
			Description: "Specify associate_public_ip=[true|false] to enable/disable association of public IP for master ASG",
			Optional:    true,
			Default:     "",
		},
		"authorization": {
			Type:        schema.TypeString,
			Description: "Authorization, RBAC or AlwaysAllow",
			ForceNew:    true,
			Optional:    true,
			Default:     "AlwaysAllow",
		},
		"bastion": {
			Type:        schema.TypeBool,
			Description: "Set to enable a bastion instance group. Only applies to private topology",
			Optional:    true,
			Default:     false,
		},
		"cloud": {
			Type:        schema.TypeString,
			Description: "Name of Cloud Provider",
			Optional:    true,
			ForceNew:    true,
			Default:     "aws",
		},
		"cloud_labels": {
			Type:        schema.TypeString,
			Description: "A list of KV pairs used to tag all instance groups in AWS (eg Owner=John Doe,Team=Some Team)",
			Optional:    true,
		},
		"config": {
			Type:        schema.TypeString,
			Description: "yaml config file(default is $HOME/.kops.yaml)",
			Optional:    true,
		},
		"dns": {
			Type:        schema.TypeString,
			Description: "DNS hosted zone to use: public|private. (default Public)",
			Optional:    true,
		},
		"dry_run": {
			Type:        schema.TypeBool,
			Description: "If true, only print the object that would be sent, without sending it. This can be used to create a cluster YAML or JSON manifest",
			Optional:    true,
			Default:     false,
		},
		"encrypt_etcd_storage": {
			Type:        schema.TypeBool,
			Description: "Generate key in aws kms and use it for encrypt etcd volumee",
			Optional:    true,
			Default:     true,
		},
		"etcd_version": {
			Type:        schema.TypeString,
			Description: "etcd version",
			Optional:    true,
			ForceNew:    true,
			Default:     "3.2.24",
		},
		"image": {
			Type:        schema.TypeString,
			Description: "AMI Image for all volumes",
			Optional:    true,
			Default:     "ami-03b850a018c8cd25e",
		},
		"k8s_version": {
			Type:        schema.TypeString,
			Description: "k8s version",
			Optional:    true,
			ForceNew:    true,
			Default:     "v1.11.5",
		},
		"master_per_zone": {
			Type:        schema.TypeInt,
			Description: "Masters Per Zone",
			ForceNew:    true,
			Optional:    true,
			Default:     1,
		},
		"master_security_groups": {
			Type:        schema.TypeList,
			Description: "Add precreated additional security groups to masters",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"master_size": {
			Type:        schema.TypeString,
			Description: "Master Nodes Instances Size e.g. t2.medium",
			Required:    true,
		},
		"master_volume_size": {
			Type:        schema.TypeInt,
			Description: "Master Root Volume Size",
			ForceNew:    true,
			Required:    true,
		},
		"master_zones": {
			Type:        schema.TypeList,
			Description: "Zones in which to run masters (must be an odd number)",
			Required:    true,
			ForceNew:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"model": {
			Type:        schema.TypeString,
			Description: "Models to apply(separate multiple models with commas) (default proto,cloudup)",
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "Name of cluster",
			Required:    true,
			ForceNew:    true,
		},
		"network_cidr": {
			Type:        schema.TypeString,
			Description: "Set to override the default network CIDR",
			Required:    true,
			ForceNew:    true,
		},
		"networking": {
			Type:        schema.TypeString,
			Description: "Networking mode to use.  kubenet (default), classic, external, kopeio-vxlan (or kopeio), weave, flannel-vxlan (or flannel), flannel-udp, calico, canal, kube-router, romana, amazon-vpc-routed-eni, cilium, cni. (default kubenet)",
			Optional:    true,
			ForceNew:    true,
			Default:     "kubenet",
		},
		"node_max_size": {
			Type:        schema.TypeInt,
			Description: "Node Max Size",
			ForceNew:    true,
			Required:    true,
		},
		"node_min_size": {
			Type:        schema.TypeInt,
			Description: "Node Min Size",
			ForceNew:    true,
			Required:    true,
		},
		"node_size": {
			Type:        schema.TypeString,
			Description: "Worker Nodes Instances Size e.g. t2.medium",
			Required:    true,
		},
		"node_volume_size": {
			Type:        schema.TypeInt,
			Description: "Node Root Volume Size",
			ForceNew:    true,
			Required:    true,
		},
		"node_zones": {
			Type:        schema.TypeList,
			Description: "The list of node zones",
			Required:    true,
			ForceNew:    true,
			Elem: &schema.Schema{
				Type:     schema.TypeString,
				MinItems: 1},
		},
		"node_security_groups": {
			Type:        schema.TypeList,
			Description: "Add precreated additional security groups to nodes",
			Required:    true,
			ForceNew:    true,
			Elem: &schema.Schema{
				Type:     schema.TypeString,
				MinItems: 1},
		},
		"non_masquerade_cidr": {
			Type:        schema.TypeString,
			Description: "non masquerade cidr",
			Optional:    true,
			Default:     "100.64.0.1/10",
		},
		"out": {
			Type:        schema.TypeString,
			Description: "Path to write any local output",
			Optional:    true,
		},
		"output": {
			Type:        schema.TypeString,
			Description: "Output format.One of json | yaml.Used with the dry-run",
			Optional:    true,
		},
		"ssh_access": {
			Type:        schema.TypeList,
			Description: "Restrict SSH access to this CIDR.  If not set, access will not be restricted by IP. (default [0.0.0.0/0])",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ssh_public_key": {
			Type:        schema.TypeString,
			Description: "ssh public key path",
			Required:    true,
			ForceNew:    true,
		},
		"state_store": {
			Type:        schema.TypeString,
			Description: "State Store",
			Required:    true,
			ForceNew:    true,
		},
		"subnets": {
			Type:        schema.TypeList,
			Description: "Set to use shared subnets",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"target": {
			Type:        schema.TypeString,
			Description: "Valid targets : direct, terraform, cloudformation",
			Optional:    true,
			Default:     "direct",
		},
		"topology": {
			Type:        schema.TypeString,
			Description: "Topology",
			Optional:    true,
			Default:     "public",
		},
		"utility_subnets": {
			Type:        schema.TypeList,
			Description: "utility_subnets Set to use shared utility subnets",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"vpc_id": {
			Type:        schema.TypeString,
			Description: "VPC ID, Set to use a shared VPC",
			Optional:    true,
		},

		"zones": {
			Type:        schema.TypeList,
			Description: "Zones in which to run the cluster",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}
