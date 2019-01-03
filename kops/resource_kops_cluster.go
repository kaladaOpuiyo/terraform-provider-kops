package kops

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	commands "k8s.io/kops/pkg/commands"
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
			"cloud": {
				Type:        schema.TypeString,
				Description: "Name of Cloud Provider",
				Optional:    true,
				ForceNew:    true,
				Default:     "aws",
			},
			"image": {
				Type:        schema.TypeString,
				Description: "AMI Image",
				Optional:    true,
				ForceNew:    true,
				Default:     "ami-03b850a018c8cd25e",
			},
			"master_size": {
				Type:        schema.TypeString,
				Description: "Master Nodes Instances Size e.g. t2.medium",
				Required:    true,
				ForceNew:    true,
			},
			"node_size": {
				Type:        schema.TypeString,
				Description: "Worker Nodes Instances Size e.g. t2.medium",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of Cluster",
				Required:    true,
				ForceNew:    true,
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
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"master_zones": {
				Type:        schema.TypeList,
				Description: "The list of master zones",
				Required:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceKopsCreate(d *schema.ResourceData, meta interface{}) error {

	var err error

	allowList := true
	registryBase, err := vfs.Context.BuildVfsPath(d.Get("state_store").(string))

	if err != nil {
		return fmt.Errorf("error parsing registry path %q: %v", d.Get("state_store").(string), err)
	}

	adminAccess := []string{"0.0.0.0/0"}
	clientset := vfsclientset.NewVFSClientset(registryBase, allowList)
	cloud := fmt.Sprint(d.Get("cloud"))
	clusterName := fmt.Sprint(d.Get("name"))
	etcdVersion := fmt.Sprint(d.Get("etcd_version"))
	image := fmt.Sprint(d.Get("image"))
	k8sVersion := fmt.Sprint(d.Get("k8s_version"))
	masterCount := d.Get("master_count").(int)
	masterSize := fmt.Sprint(d.Get("master_size"))
	masterVolumeSize := fi.Int32(int32(d.Get("master_volume_size").(int)))
	networkCidr := fmt.Sprint(d.Get("network_cidr"))
	nodeMaxSize := fi.Int32(int32(d.Get("node_max_size").(int)))
	nodeMinSize := fi.Int32(int32(d.Get("node_min_size").(int)))
	nodeSize := fmt.Sprint(d.Get("node_size"))
	nodeVolumeSize := fi.Int32(int32(d.Get("node_volume_size").(int)))

	nodeZones := make([]string, len(d.Get("node_zones").([]interface{})))

	if len(nodeZones) == 0 {
		return fmt.Errorf("Must provide node zones")
	}

	for i, v := range d.Get("node_zones").([]interface{}) {
		nodeZones[i] = fmt.Sprint(v)
	}

	masterZones := make([]string, len(d.Get("master_zones").([]interface{})))

	if len(masterZones) == 0 {
		return fmt.Errorf("Must provide node zones")
	}

	for i, v := range d.Get("master_zones").([]interface{}) {
		masterZones[i] = fmt.Sprint(v)
	}

	cluster := &api.Cluster{}

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
	cluster.Spec.IAM = &api.IAMSpec{
		AllowContainerRegistry: true,
		Legacy:                 false,
	}
	//**********************************************************
	// These will be added to resource schema, here for testing

	cluster.Spec.API = &api.AccessSpec{}
	cluster.Spec.API.DNS = &api.DNSAccessSpec{}
	cluster.Spec.Authorization = &api.AuthorizationSpec{}
	cluster.Spec.Authorization.RBAC = &api.RBACAuthorizationSpec{}
	cluster.Spec.Networking = &api.NetworkingSpec{}
	// Will make networking selectable... eventually :)
	cluster.Spec.Networking.Calico = &api.CalicoNetworkingSpec{}
	cluster.Spec.Networking.Calico.MajorVersion = "v3"
	cluster.Spec.Topology.DNS = &api.DNSSpec{}
	cluster.Spec.Topology.DNS.Type = api.DNSTypePublic
	cluster.Spec.Topology.Masters = api.TopologyPublic
	cluster.Spec.Topology.Nodes = api.TopologyPublic

	cluster.Spec.Kubelet = &api.KubeletConfigSpec{
		AnonymousAuth: fi.Bool(false),

		// Hard Coded for now testing dont forget to add RBAC
		AuthenticationTokenWebhook: fi.Bool(true),
		AuthorizationMode:          "Webhook",
	}
	//**********************************************************

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

	for _, etcdClusterName := range cloudup.EtcdClusters {
		etcdCluster := &api.EtcdClusterSpec{
			Name:    etcdClusterName,
			Version: etcdVersion,
		}
		for _, masterZone := range masterZones {
			etcdMember := &api.EtcdMemberSpec{
				Name:          masterZone,
				InstanceGroup: fi.String("master-" + masterZone),
			}
			etcdCluster.Members = append(etcdCluster.Members, etcdMember)

		}
		cluster.Spec.EtcdClusters = append(cluster.Spec.EtcdClusters, etcdCluster)
	}

	{
		if err := cloudup.PerformAssignments(cluster); err != nil {
			return err
		}

		_, err := clientset.CreateCluster(cluster)
		if err != nil {
			return err
		}

	}
	{
		// Create master ig, Test only

		masters := []*api.InstanceGroup{}

		for i := 0; i < masterCount; i++ {
			master := &api.InstanceGroup{}
			master.ObjectMeta.Name = "master-" + masterZones[i%len(masterZones)]
			if int(masterCount) > len(masterZones) {
				master.ObjectMeta.Name += "-" + strconv.Itoa(1+(i/len(masterZones)))
			}
			master.Spec = api.InstanceGroupSpec{
				Role:           api.InstanceGroupRoleMaster,
				Subnets:        masterZones,
				Image:          image,
				MachineType:    masterSize,
				RootVolumeSize: masterVolumeSize,
			}

			_, err := clientset.InstanceGroupsFor(cluster).Create(master)
			if err != nil {
				return err
			}

			masters = append(masters, master)
		}

	}
	{
		// Create node ig,Testing only
		nodes := &api.InstanceGroup{}

		nodes.ObjectMeta.Name = "nodes"
		nodes.Spec = api.InstanceGroupSpec{
			Image:          image,
			MachineType:    nodeSize,
			MaxSize:        nodeMaxSize,
			MinSize:        nodeMinSize,
			Role:           api.InstanceGroupRoleNode,
			RootVolumeSize: nodeVolumeSize,
			Subnets:        nodeZones,
		}

		_, err := clientset.InstanceGroupsFor(cluster).Create(nodes)
		if err != nil {
			return err
		}

	}

	{
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
