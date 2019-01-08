package kops

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	commands "k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/pkg/resources"
	ops "k8s.io/kops/pkg/resources/ops"
	"k8s.io/kops/pkg/validation"
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
		Schema: kopsSchema(),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
		},
	}

}

//Sourced:k8s.io/kops/
func resourceKopsCreate(d *schema.ResourceData, meta interface{}) error {

	var err error

	adminAccess := make([]string, len(d.Get("admin_access").([]interface{})))
	if len(adminAccess) == 0 {
		adminAccess = []string{"0.0.0.0/0"}
	} else {
		for i, v := range d.Get("admin_access").([]interface{}) {
			adminAccess[i] = fmt.Sprint(v)
		}
	}

	allowList := true
	apiLoadBalancerType := fmt.Sprint(d.Get("api_load_balancer_type"))
	associatePublicIP := d.Get("associate_public_ip").(bool)
	authorization := fmt.Sprint(d.Get("authorization"))
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
	networking := fmt.Sprint(d.Get("networking"))
	cluster := &api.Cluster{}
	clusterName := fmt.Sprint(d.Get("name"))
	dns := fmt.Sprint(d.Get("dns"))
	encryptEtcdStorage := d.Get("encrypt_etcd_storage").(bool)
	etcdVersion := fmt.Sprint(d.Get("etcd_version"))
	image := fmt.Sprint(d.Get("image"))
	instanceGroups := []*api.InstanceGroup{}
	k8sVersion := fmt.Sprint(d.Get("k8s_version"))
	masterPerZone := fi.Int32(int32(d.Get("master_per_zone").(int)))
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
	sshAccess := make([]string, len(d.Get("ssh_access").([]interface{})))
	if len(sshAccess) == 0 {
		sshAccess = []string{"0.0.0.0/0"}
	} else {
		for i, v := range d.Get("ssh_access").([]interface{}) {
			sshAccess[i] = fmt.Sprint(v)
		}
	}
	topology := fmt.Sprint(d.Get("topology"))
	validateOnCreation := d.Get("validate_on_creation").(bool)
	vpcID := fmt.Sprint(d.Get("vpc_id"))

	cluster.ObjectMeta.Name = clusterName
	cluster.Spec = api.ClusterSpec{
		Channel:             "stable",
		CloudProvider:       cloud,
		ConfigBase:          registryBase.Join(cluster.ObjectMeta.Name).Path(),
		KubernetesAPIAccess: adminAccess,
		KubernetesVersion:   k8sVersion,
		SSHAccess:           sshAccess,
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

	cluster.Spec.Networking = &api.NetworkingSpec{}
	switch networking {
	case "classic":
		cluster.Spec.Networking.Classic = &api.ClassicNetworkingSpec{}
	case "kubenet":
		cluster.Spec.Networking.Kubenet = &api.KubenetNetworkingSpec{}
	case "external":
		cluster.Spec.Networking.External = &api.ExternalNetworkingSpec{}
	case "cni":
		cluster.Spec.Networking.CNI = &api.CNINetworkingSpec{}
	case "kopeio-vxlan", "kopeio":
		cluster.Spec.Networking.Kopeio = &api.KopeioNetworkingSpec{}
	case "weave":
		cluster.Spec.Networking.Weave = &api.WeaveNetworkingSpec{}

		if cluster.Spec.CloudProvider == "aws" {
			// AWS supports "jumbo frames" of 9001 bytes and weave adds up to 87 bytes overhead
			// sets the default to the largest number that leaves enough overhead and is divisible by 4
			jumboFrameMTUSize := int32(8912)
			cluster.Spec.Networking.Weave.MTU = &jumboFrameMTUSize
		}
	case "flannel", "flannel-vxlan":
		cluster.Spec.Networking.Flannel = &api.FlannelNetworkingSpec{
			Backend: "vxlan",
		}
	case "flannel-udp":
		cluster.Spec.Networking.Flannel = &api.FlannelNetworkingSpec{
			Backend: "udp",
		}
	case "calico":
		cluster.Spec.Networking.Calico = &api.CalicoNetworkingSpec{
			MajorVersion: "v3",
		}

	case "canal":
		cluster.Spec.Networking.Canal = &api.CanalNetworkingSpec{}
	case "kube-router":
		cluster.Spec.Networking.Kuberouter = &api.KuberouterNetworkingSpec{}
	case "romana":
		cluster.Spec.Networking.Romana = &api.RomanaNetworkingSpec{}
	case "amazonvpc", "amazon-vpc-routed-eni":
		cluster.Spec.Networking.AmazonVPC = &api.AmazonVPCNetworkingSpec{}
	case "cilium":
		cluster.Spec.Networking.Cilium = &api.CiliumNetworkingSpec{}
	case "lyftvpc":
		cluster.Spec.Networking.LyftVPC = &api.LyftVPCNetworkingSpec{}
	default:
		return fmt.Errorf("unknown networking mode %q", networking)
	}

	cluster.Spec.Topology.DNS = &api.DNSSpec{}
	if dns == "private" {
		cluster.Spec.Topology.DNS.Type = api.DNSTypePrivate
	} else {
		cluster.Spec.Topology.DNS.Type = api.DNSTypePublic
	}
	// Need to handle Private networks as well for now just public ¯\_(ツ)_/¯
	switch topology {
	case api.TopologyPublic:
		cluster.Spec.Topology.Masters = api.TopologyPublic
		cluster.Spec.Topology.Nodes = api.TopologyPublic
		if bastion {
			return fmt.Errorf("bastion supports topology='private' only")
		}

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
	case api.TopologyPrivate:
		return fmt.Errorf("Need to handle Private networks as well for now just public")

	default:
		return fmt.Errorf("invalid topology %s", topology)
	}
	cluster.Spec.NonMasqueradeCIDR = nonMasqueradeCIDR

	cluster.Spec.Kubelet = &api.KubeletConfigSpec{
		AnonymousAuth: fi.Bool(false),

		// Hard Coded for now testing dont forget to add RBAC when creating these rules
		AuthenticationTokenWebhook: fi.Bool(true),
		AuthorizationMode:          "Webhook",

		// Optional Parameters to be implemented
		// ClientCAFile:               "",
		// TLSCertFile:                "",
		// TLSPrivateKeyFile:          "",
		// LoadBalancerType:           "",
		// IdleTimeoutSeconds:         "",
		// SecurityGroupOverride:      "",
		// AdditionalSecurityGroups:   " ",
		// UseForInternalApi:          "",
		// SSLCertificate:             "",
	}

	// Super scaled down from the cmd implementation mainly here for testing
	// if cluster.Spec.API.IsEmpty() {
	if apiLoadBalancerType != "" {
		cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}
	}
	// else {
	// 	switch cluster.Spec.Topology.Masters {
	// 	case api.TopologyPublic:
	// 		if dnsGossip.IsGossipHostname(cluster.Spec) {
	// 			// gossip DNS names don't work outside the cluster, so we use a LoadBalancer instead
	// 			cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}
	// 		} else {
	// 			cluster.Spec.API.DNS = &api.DNSAccessSpec{}
	// 		}

	// 	case api.TopologyPrivate:
	// 		cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}

	// 	default:
	// 		return fmt.Errorf("unknown master topology type: %q", cluster.Spec.Topology.Masters)
	// 	}
	// }
	// }
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

	// Create master ig(s)
	for i := 0; i < len(masterZones); i++ {

		zone := masterZones[i%len(masterZones)]
		name := zone

		master := &api.InstanceGroup{}
		master.ObjectMeta.Name = "master-" + name
		master.Spec = api.InstanceGroupSpec{
			AssociatePublicIP: fi.Bool(associatePublicIP),
			Image:             image,
			MachineType:       masterSize,
			Role:              api.InstanceGroupRoleMaster,
			RootVolumeSize:    masterVolumeSize,
			MaxSize:           masterPerZone,
			MinSize:           masterPerZone,
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

	// Create nodes ig
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
	// privateDNS := cluster.Spec.Topology != nil && cluster.Spec.Topology.DNS.Type == kops.DNSTypePrivate
	// Probably need to look into doing this another way ¯\_(ツ)_/¯
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

	if validateOnCreation {

		list, err := clientset.InstanceGroupsFor(cluster).List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("cannot get InstanceGroups")
		}

		config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{CurrentContext: clusterName}).ClientConfig()
		if err != nil {
			return fmt.Errorf("Cannot load kubecfg settings for %q: %v", clusterName, err)
		}

		k8sClient, err := kubernetes.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("Cannot build kubernetes api client for %q: %v", clusterName, err)
		}
		validateClusterState := &resource.StateChangeConf{
			Pending: []string{"Validating"},
			Target:  []string{"Ready"},
			Refresh: func() (interface{}, string, error) {

				result, err := validation.ValidateCluster(cluster, list, k8sClient)

				var clusterStatus string

				if err != nil {
					clusterStatus = "Validating"
				}
				if len(result.Failures) != 0 {
					clusterStatus = "Validating"

				} else {
					clusterStatus = "Ready"
				}

				return result, clusterStatus, nil
			},
			Timeout:                   9 * time.Minute,
			Delay:                     20 * time.Second,
			MinTimeout:                20 * time.Second,
			ContinuousTargetOccurence: 5,
		}
		_, err = validateClusterState.WaitForState()
		if err != nil {
			return fmt.Errorf("Error Validating cluster: %s", err)
		}

	}

	d.SetId(clusterName)
	return resourceKopsRead(d, meta)

}

func resourceKopsRead(d *schema.ResourceData, meta interface{}) error {

	name := d.Id()

	//check if diff in state_store
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

	// d.Set("api_load_balancer_type", "") //not implemented
	// d.Set("associate_public_ip", cluster.Spec)
	// d.Set("bastion", false) // need to determine like ^
	// d.Set("dry_run", false) // need to determine like ^
	// d.Set("encrypt_etcd_storage", fi.BoolValue(etcdMember.EncryptedVolume))
	// d.Set("etcd_version", etcdCluster.Version)
	// d.Set("subnets", cluster.Spec.Subnets) // subnets slice
	// d.Set("target", cluster.Spec.Target) Force new
	// d.Set("utility_subnets", cluster.Spec.Subnets) // need to find if exist
	// etcdCluster := api.EtcdClusterSpec{}
	// etcdMember := api.EtcdMemberSpec{}
	d.Set("admin_access", cluster.Spec.KubernetesAPIAccess)
	d.Set("authorization", cluster.Spec.Authorization.AlwaysAllow) // set for now need to pull not set
	d.Set("cloud_labels", cluster.Spec.CloudLabels)
	d.Set("config", cluster.Spec.ConfigBase) // computed
	d.Set("dns", strings.ToLower(string(cluster.Spec.Topology.DNS.Type)))
	d.Set("k8s_version", cluster.Spec.KubernetesVersion)
	d.Set("name", cluster.Name)
	d.Set("network_cidr", cluster.Spec.NetworkCIDR)
	d.Set("networking", cluster.Spec.Networking)
	d.Set("non_masquerade_cidr", cluster.Spec.NonMasqueradeCIDR)
	d.Set("ssh_access", cluster.Spec.SSHAccess)
	d.Set("state_store", strings.Split(cluster.Spec.ConfigBase, "/")) // Force new
	d.Set("topology", cluster.Spec.Topology.Masters)
	d.Set("vpc_id", cluster.Spec.NetworkID)

	list, err := clientset.InstanceGroupsFor(cluster).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("cannot get InstanceGroups for %q: %v", cluster.ObjectMeta.Name, err)
	}

	for _, ig := range list.Items {

		// Will need to deal with mult-zoned masters
		if strings.Contains(ig.Name, "master") {
			d.Set("image", ig.Spec.Image)
			d.Set("master_per_zone", ig.Spec.MaxSize)
			d.Set("master_security_groups", ig.Spec.SecurityGroupOverride)
			d.Set("master_size", ig.Spec.MachineType)
			d.Set("master_volume_size", ig.Spec.RootVolumeSize)
			d.Set("master_zones", ig.Spec.Subnets) // Need to iterate each master
		}
		if strings.Contains(ig.Name, "node") {
			d.Set("node_max_size", ig.Spec.MaxSize)
			d.Set("node_min_size", ig.Spec.MinSize)
			d.Set("node_security_groups", ig.Spec.SecurityGroupOverride)
			d.Set("node_size", ig.Spec.MachineType)
			d.Set("node_volume_size", ig.Spec.RootVolumeSize)
			d.Set("node_zones", ig.Spec.Subnets)
		}

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

	allResources, err := ops.ListResources(cloud, name, "")
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

	err = ops.DeleteResources(cloud, clusterResources)
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
