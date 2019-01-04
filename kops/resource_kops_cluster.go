package kops

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	commands "k8s.io/kops/pkg/commands"
	dnsGossip "k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/pkg/resources"
	resourceops "k8s.io/kops/pkg/resources/ops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/vfs"
)

func resourceKopsCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceKopsCreate,
		Read:   resourceKopsRead,
		Update: resourceKopsUpdate,
		Delete: resourceKopsDelete,

		Schema: map[string]*schema.Schema{
			"admin_access": {
				Type:        schema.TypeList,
				Description: "Admin Access",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"api_load_balancer_type": {
				Type:        schema.TypeString,
				Description: "Api load balance type, internal or public",
				Optional:    true,
				Default:     "public",
			},
			"authorization": {
				Type:        schema.TypeString,
				Description: "Authorization, RBAC or AlwaysAllow",
				Optional:    true,
				Default:     "AlwaysAllow",
			},
			"associate_public_ip": {
				Type:        schema.TypeBool,
				Description: "associate public ip",
				Optional:    true,
				Default:     true,
			},
			"bastion": {
				Type:        schema.TypeBool,
				Description: "create a bastion host",
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
				Description: "Cloud Labels",
				Optional:    true,
			},
			"image": {
				Type:        schema.TypeString,
				Description: "AMI Image",
				Optional:    true,
				Default:     "ami-03b850a018c8cd25e",
			},
			"master_size": {
				Type:        schema.TypeString,
				Description: "Master Nodes Instances Size e.g. t2.medium",
				Required:    true,
			},
			"node_size": {
				Type:        schema.TypeString,
				Description: "Worker Nodes Instances Size e.g. t2.medium",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of Cluster",
				Required:    true,
				ForceNew:    true,
			},
			"dns": {
				Type:        schema.TypeString,
				Description: "dns",
				Optional:    true,
			},

			"topology": {
				Type:        schema.TypeString,
				Description: "Topology",
				Optional:    true,
				Default:     "public",
			},
			"non_masquerade_cidr": {
				Type:        schema.TypeString,
				Description: "non masquerade cidr",
				Optional:    true,
				Default:     "100.64.0.1/10",
			},

			"state_store": {
				Type:        schema.TypeString,
				Description: "State Store",
				Required:    true,
				ForceNew:    true,
			},
			"dry_run": {
				Type:        schema.TypeBool,
				Description: "dry run",
				Optional:    true,
				Default:     true,
			},
			"encrypt_etcd_storage": {
				Type:        schema.TypeBool,
				Description: "encrypt etcd storage",
				Optional:    true,
				Default:     true,
			},
			"master_count": {
				Type:        schema.TypeInt,
				Description: "Master Count",
				ForceNew:    true,
				Required:    true,
			},
			"master_volume_size": {
				Type:        schema.TypeInt,
				Description: "Master Volume Size",
				ForceNew:    true,
				Required:    true,
			},
			"node_volume_size": {
				Type:        schema.TypeInt,
				Description: "Node Volume Size",
				ForceNew:    true,
				Required:    true,
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
			"etcd_version": {
				Type:        schema.TypeString,
				Description: "etcd version",
				Optional:    true,
				ForceNew:    true,
				Default:     "3.2.24",
			},
			"k8s_version": {
				Type:        schema.TypeString,
				Description: "k8s version",
				Optional:    true,
				ForceNew:    true,
				Default:     "v1.11.5",
			},
			"ssh_public_key": {
				Type:        schema.TypeString,
				Description: "ssh public key path",
				Required:    true,
				ForceNew:    true,
			},
			"network_cidr": {
				Type:        schema.TypeString,
				Description: "network cidr block",
				Required:    true,
				ForceNew:    true,
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

			"master_zones": {
				Type:        schema.TypeList,
				Description: "The list of master zones",
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Description: "vpc id",
				Optional:    true,
			},
		},
	}
}

//Sourced:k8s.io/kops/
func resourceKopsCreate(d *schema.ResourceData, meta interface{}) error {

	var err error

	adminAccess := make([]string, len(d.Get("admin_access").([]interface{})))
	if len(adminAccess) == 0 {
		return fmt.Errorf("Must provide node zones")
	}
	for i, v := range d.Get("admin_access").([]interface{}) {
		adminAccess[i] = fmt.Sprint(v)
	}
	allowList := true
	apiLoadBalancerType := fmt.Sprint(d.Get("api_load_balancer_type"))
	authorization := fmt.Sprint(d.Get("authorization"))
	associatePublicIP := d.Get("associate_public_ip").(bool)
	bastion := d.Get("bastion").(bool)
	cloudLabels, err := parseCloudLabels(d.Get("cloud_labels").(string))
	if err != nil {
		return fmt.Errorf("error parsing global cloud labels: %v", err)
	}

	registryBase, err := vfs.Context.BuildVfsPath(d.Get("state_store").(string))

	if err != nil {
		return fmt.Errorf("error parsing registry path %q: %v", d.Get("state_store").(string), err)
	}
	clientset := vfsclientset.NewVFSClientset(registryBase, allowList)
	cloud := fmt.Sprint(d.Get("cloud"))
	cluster := &api.Cluster{}
	clusterName := fmt.Sprint(d.Get("name"))
	dns := fmt.Sprint(d.Get("dns"))
	encryptEtcdStorage := d.Get("encrypt_etcd_storage").(bool)
	etcdVersion := fmt.Sprint(d.Get("etcd_version"))
	image := fmt.Sprint(d.Get("image"))
	instanceGroups := []*api.InstanceGroup{}
	k8sVersion := fmt.Sprint(d.Get("k8s_version"))
	masterCount := d.Get("master_count").(int)
	masters := []*api.InstanceGroup{}
	masterSize := fmt.Sprint(d.Get("master_size"))
	masterVolumeSize := fi.Int32(int32(d.Get("master_volume_size").(int)))
	masterZones := make([]string, len(d.Get("master_zones").([]interface{})))
	if len(masterZones) == 0 {
		return fmt.Errorf("Must provide node zones")
	}
	for i, v := range d.Get("master_zones").([]interface{}) {
		masterZones[i] = fmt.Sprint(v)
	}
	networkCidr := fmt.Sprint(d.Get("network_cidr"))
	nodeMaxSize := fi.Int32(int32(d.Get("node_max_size").(int)))
	nodeMinSize := fi.Int32(int32(d.Get("node_min_size").(int)))
	nodes := &api.InstanceGroup{}
	nodeSize := fmt.Sprint(d.Get("node_size"))
	nodeVolumeSize := fi.Int32(int32(d.Get("node_volume_size").(int)))
	nodeZones := make([]string, len(d.Get("node_zones").([]interface{})))
	if len(nodeZones) == 0 {
		return fmt.Errorf("Must provide node zones")
	}
	for i, v := range d.Get("node_zones").([]interface{}) {
		nodeZones[i] = fmt.Sprint(v)
	}
	nonMasqueradeCIDR := fmt.Sprint(d.Get("non_masquerade_cidr"))
	topology := fmt.Sprint(d.Get("topology"))
	vpcID := fmt.Sprint(d.Get("vpc_id"))

	cluster.ObjectMeta.Name = clusterName
	cluster.Spec = api.ClusterSpec{
		Channel:             "stable",
		CloudProvider:       cloud,
		ConfigBase:          registryBase.Join(cluster.ObjectMeta.Name).Path(),
		KubernetesAPIAccess: adminAccess,
		KubernetesVersion:   k8sVersion,
		SSHAccess:           adminAccess,
		Topology:            &api.TopologySpec{},
		NetworkCIDR:         networkCidr,
	}

	if vpcID != "" {
		cluster.Spec.NetworkID = vpcID
	}

	cluster.Spec.IAM = &api.IAMSpec{
		AllowContainerRegistry: true,
		Legacy:                 false,
	}

	// WIP
	if bastion && topology != "public" {
		bastionGroup := &api.InstanceGroup{}
		bastionGroup.Spec.Role = api.InstanceGroupRoleBastion
		bastionGroup.ObjectMeta.Name = "bastions"
		bastionGroup.Spec.Image = image
		instanceGroups = append(instanceGroups, bastionGroup)

		cluster.Spec.Topology.Bastion = &api.BastionSpec{
			BastionPublicName: "bastion." + clusterName,
		}

	}

	if len(cloudLabels) != 0 {
		cluster.Spec.CloudLabels = cloudLabels
	}

	cluster.Spec.API = &api.AccessSpec{}
	cluster.Spec.API.DNS = &api.DNSAccessSpec{}
	cluster.Spec.Authorization = &api.AuthorizationSpec{}
	if strings.EqualFold(authorization, "AlwaysAllow") {
		cluster.Spec.Authorization.AlwaysAllow = &api.AlwaysAllowAuthorizationSpec{}
	} else if strings.EqualFold(authorization, "RBAC") {
		cluster.Spec.Authorization.RBAC = &api.RBACAuthorizationSpec{}
	} else {
		return fmt.Errorf("unknown authorization mode %q", authorization)
	}
	// Will make networking selectable... eventually :)
	cluster.Spec.Networking = &api.NetworkingSpec{}
	cluster.Spec.Networking.Calico = &api.CalicoNetworkingSpec{}
	cluster.Spec.Networking.Calico.MajorVersion = "v3"
	cluster.Spec.Topology.DNS = &api.DNSSpec{}
	if dns == "private" {
		cluster.Spec.Topology.DNS.Type = api.DNSTypePrivate
	} else {
		cluster.Spec.Topology.DNS.Type = api.DNSTypePublic
	}

	cluster.Spec.Topology.Masters = api.TopologyPublic
	cluster.Spec.Topology.Nodes = api.TopologyPublic
	cluster.Spec.NonMasqueradeCIDR = nonMasqueradeCIDR

	cluster.Spec.Kubelet = &api.KubeletConfigSpec{
		AnonymousAuth: fi.Bool(false),

		// Hard Coded for now testing dont forget to add RBAC when createing these rules
		AuthenticationTokenWebhook: fi.Bool(true),
		AuthorizationMode:          "Webhook",
	}

	if cluster.Spec.API == nil {
		cluster.Spec.API = &api.AccessSpec{}
	}
	if cluster.Spec.API.IsEmpty() {
		if apiLoadBalancerType != "" {
			cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}
		} else {
			switch cluster.Spec.Topology.Masters {
			case api.TopologyPublic:
				if dnsGossip.IsGossipHostname(cluster.Name) {
					// gossip DNS names don't work outside the cluster, so we use a LoadBalancer instead
					cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}
				} else {
					cluster.Spec.API.DNS = &api.DNSAccessSpec{}
				}

			case api.TopologyPrivate:
				cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}

			default:
				return fmt.Errorf("unknown master topology type: %q", cluster.Spec.Topology.Masters)
			}
		}
	}
	if cluster.Spec.API.LoadBalancer != nil && cluster.Spec.API.LoadBalancer.Type == "" {
		switch apiLoadBalancerType {
		case "", "public":
			cluster.Spec.API.LoadBalancer.Type = api.LoadBalancerTypePublic
		case "internal":
			cluster.Spec.API.LoadBalancer.Type = api.LoadBalancerTypeInternal
		default:
			return fmt.Errorf("unknown api-loadbalancer-type: %q", apiLoadBalancerType)
		}
	}

	// if cluster.Spec.API.LoadBalancer != nil && apiSSLCertificate != "" {
	// 	cluster.Spec.API.LoadBalancer.SSLCertificate = apiSSLCertificate
	// }

	keys := make(map[string]bool)
	subnetZones := append(nodeZones, masterZones...)

	for _, subnetZone := range subnetZones {
		if _, value := keys[subnetZone]; !value {
			keys[subnetZone] = true
			cluster.Spec.Subnets = append(cluster.Spec.Subnets, api.ClusterSubnetSpec{
				Name: subnetZone,
				Zone: subnetZone,
				Type: api.SubnetTypePublic,
			})
		}

	}

	// Create master ig,Testing only
	for i := 0; i < masterCount; i++ {

		zone := masterZones[i%len(masterZones)]
		name := zone
		if int(masterCount) > len(masterZones) {
			name += "-" + strconv.Itoa(1+(i/len(masterZones)))
		}

		master := &api.InstanceGroup{}
		master.ObjectMeta.Name = "master-" + name
		master.Spec = api.InstanceGroupSpec{
			AssociatePublicIP: fi.Bool(associatePublicIP),
			Image:             image,
			MachineType:       masterSize,
			Role:              api.InstanceGroupRoleMaster,
			RootVolumeSize:    masterVolumeSize,
			Subnets:           []string{masterZones[i%len(masterZones)]},
		}

		masters = append(masters, master)

		_, err = clientset.InstanceGroupsFor(cluster).Create(master)
		if err != nil {
			return err
		}

	}

	for _, etcdClusterName := range cloudup.EtcdClusters {
		etcdCluster := &api.EtcdClusterSpec{
			Name:    etcdClusterName,
			Version: etcdVersion,
		}

		for _, master := range masters {
			etcdMember := &api.EtcdMemberSpec{}
			if encryptEtcdStorage {
				etcdMember.EncryptedVolume = fi.Bool(encryptEtcdStorage)
			}
			etcdMember.Name = master.ObjectMeta.Name
			etcdMember.InstanceGroup = fi.String(master.ObjectMeta.Name)
			etcdCluster.Members = append(etcdCluster.Members, etcdMember)
		}

		cluster.Spec.EtcdClusters = append(cluster.Spec.EtcdClusters, etcdCluster)
	}

	nodes.ObjectMeta.Name = "nodes"
	nodes.Spec = api.InstanceGroupSpec{
		AssociatePublicIP: fi.Bool(associatePublicIP),
		Image:             image,
		MachineType:       nodeSize,
		MaxSize:           nodeMaxSize,
		MinSize:           nodeMinSize,
		Role:              api.InstanceGroupRoleNode,
		RootVolumeSize:    nodeVolumeSize,
		Subnets:           nodeZones,
	}

	_, err = clientset.InstanceGroupsFor(cluster).Create(nodes)
	if err != nil {
		return err
	}

	sshCredentialStore, err := clientset.SSHCredentialStore(cluster)
	if err != nil {
		return err
	}

	f := utils.ExpandPath(d.Get("ssh_public_key").(string))
	pubKey, err := ioutil.ReadFile(f)
	if err != nil {
		return fmt.Errorf("error reading SSH key file %q: %v", f, err)
	}
	err = sshCredentialStore.AddSSHPublicKey(fi.SecretNameSSHPrimary, pubKey)
	if err != nil {
		return fmt.Errorf("error adding SSH public key: %v", err)
	}

	if err := cloudup.PerformAssignments(cluster); err != nil {
		return err
	}

	_, err = clientset.CreateCluster(cluster)
	if err != nil {
		return err
	}

	apply := &cloudup.ApplyClusterCmd{
		Cluster:    cluster,
		Clientset:  clientset,
		TargetName: cloudup.TargetDirect,
	}

	err = apply.Run()
	if err != nil {
		return err
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	secretStore, err := clientset.SecretStore(cluster)
	if err != nil {
		return err
	}

	conf, err := kubeconfig.BuildKubecfg(cluster, keyStore, secretStore, &commands.CloudDiscoveryStatusStore{})

	if err != nil {
		return err
	}

	conf.WriteKubecfg()

	d.SetId(clusterName)

	return resourceKopsRead(d, meta)
}

func resourceKopsRead(d *schema.ResourceData, meta interface{}) error {

	name := d.Id()

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

	log.Printf("[INFO] Received Kops Cluster: %#v", name)
	err = d.Set("name", cluster.Name)
	if err != nil {
		return err
	}

	return nil
}

func resourceKopsUpdate(d *schema.ResourceData, meta interface{}) error {

	return resourceKopsRead(d, meta)
}

func resourceKopsDelete(d *schema.ResourceData, meta interface{}) error {

	var err error

	name := d.Id()

	registryBase, err := vfs.Context.BuildVfsPath(d.Get("state_store").(string))

	if err != nil {
		return fmt.Errorf("error parsing registry path %q: %v", d.Get("state_store").(string), err)
	}
	allowList := true

	clientset := vfsclientset.NewVFSClientset(registryBase, allowList)

	log.Printf("[INFO] Reading Kops Cluster %s", name)

	cluster, err := clientset.GetCluster(name)
	if err != nil {
		return err
	}

	cloud, err := cloudup.BuildCloud(cluster)
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

	err = resourceops.DeleteResources(cloud, clusterResources)
	if err != nil {
		return err
	}

	{
		err := clientset.DeleteCluster(cluster)
		if err != nil {
			log.Printf("[DEBUG] Received error: %#v", err)
			return err
		}

	}

	conf := kubeconfig.NewKubeconfigBuilder()
	conf.Context = name

	if err = conf.DeleteKubeConfig(); err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
	}

	d.SetId("")

	return nil
}

// parseCloudLabels takes a CSV list of key=value records and parses them into a map. Nested '='s are supported via
// quoted strings (eg `foo="bar=baz"` parses to map[string]string{"foo":"bar=baz"}. Nested commas are not supported.
func parseCloudLabels(s string) (map[string]string, error) {

	// Replace commas with newlines to allow a single pass with csv.Reader.
	// We can't use csv.Reader for the initial split because it would see each key=value record as a single field
	// and significantly complicates using quoted fields as keys or values.
	records := strings.Replace(s, ",", "\n", -1)

	// Let the CSV library do the heavy-lifting in handling nested ='s
	r := csv.NewReader(strings.NewReader(records))
	r.Comma = '='
	r.FieldsPerRecord = 2
	r.LazyQuotes = false
	r.TrimLeadingSpace = true
	kvPairs, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("One or more key=value pairs are malformed:\n%s\n:%v", records, err)
	}

	m := make(map[string]string, len(kvPairs))
	for _, pair := range kvPairs {
		m[pair[0]] = pair[1]
	}
	return m, nil
}
func trimCommonPrefix(names []string) []string {
	// Trim shared prefix to keep the lengths sane
	// (this only applies to new clusters...)
	for len(names) != 0 && len(names[0]) > 1 {
		prefix := names[0][:1]
		allMatch := true
		for _, name := range names {
			if !strings.HasPrefix(name, prefix) {
				allMatch = false
			}
		}

		if !allMatch {
			break
		}

		for i := range names {
			names[i] = strings.TrimPrefix(names[i], prefix)
		}
	}

	return names
}
