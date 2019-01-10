provider "aws" {
  region  = "us-east-1"
  version = "~> 1.45"
}

// Im going to integrate this provider into the kops provider. Do you use centos without yum ¯\_(ツ)_/¯
# provider "helm" {
#   debug           = "true"
#   enable_tls      = "false"
#   install_tiller  = "false"
#   namespace       = "kube-system"
#   service_account = "tiller"

#   kubernetes {}
# }
#################################################################################################
# KOPS CLUSTER
##################################################################################################
data "aws_acm_certificate" "domain_cert" {
  domain      = "api.k8s.urbanradikal.com"
  most_recent = true
  types       = ["AMAZON_ISSUED"]
}
resource "kops_cluster" "aux_cluster" {

  admin_access           = ["0.0.0.0/0"] // optional,
  api_ssl_certificate    = "${data.aws_acm_certificate.domain_cert.arn}"
  api_load_balancer_type = ""     // optional
  associate_public_ip    = "true" // optional
  authorization          = "AlwaysAllow"
  bastion                = "false"
  cloud                  = "aws" // Only AWS for now
  cloud_labels           = "Owner=Kalada Opuiyo,env=test"
  dns                    = "public"
  dry_run                = "false" // not implemented
  etcd_version           = "3.2.24"
  encrypt_etcd_storage   = "true"
  image                  = "ami-03b850a018c8cd25e"
  k8s_version            = "v1.11.6"
  kube_dns               = "CoreDNS"
  master_per_zone        = 1  // optional, default is 1 per zone odd numbers only
  master_security_groups = [] // optional, not implemented
  master_size            = "t2.micro"
  master_volume_size     = 20
  master_zones           = ["us-east-1f"] // odd numbers only
  model                  = ""             // optional, not implemented
  name                   = "k8s.urbanradikal.com"
  network_cidr           = "10.0.0.0/16"
  networking             = "calico"
  node_max_size          = 5
  node_min_size          = 2
  node_security_groups   = [] // optional, not implemented
  node_size              = "t2.micro"
  node_volume_size       = 20
  node_zones             = ["us-east-1a", "us-east-1c"]
  out                    = ""            // optional, not implemented terraform or yaml
  output                 = ""            // optional, not implemented directory to output files output_dir
  ssh_access             = ["0.0.0.0/0"] // optional
  ssh_public_key         = "~/.ssh/kalada-admin.pub"
  state_store            = "s3://${aws_s3_bucket.kops_state.id}"
  subnets                = []       // optional, not implemented
  target                 = ""       // optional, not implemented
  topology               = "public" // public, private
  utility_subnets        = []       // optional, not implemented
  network_id             = ""       // optional, not tested shared vpc id

  kubelet {
    anonymous_auth               = "false"
    authentication_token_webhook = "true"
    authorization_mode           = "Webhook"
  }


  depends_on = ["aws_iam_user.kops"]
}
data "kops_cloud_resources" "cluster_cloud_resources" {
  cluster_name = "${kops_cluster.aux_cluster.id}"
  state_store  = "${kops_cluster.aux_cluster.state_store}"
}
output "cloud_resources" {
  value = {
    # load_balancer_id                = "${data.kops_cloud_resources.cluster_cloud_resources.load_balancer_id}"
    autoscaling_config_masters_id   = "${data.kops_cloud_resources.cluster_cloud_resources.autoscaling_config_masters_id}"
    autoscaling_config_nodes_id     = "${data.kops_cloud_resources.cluster_cloud_resources.autoscaling_config_nodes_id}"
    autoscaling_group_masters_id    = "${data.kops_cloud_resources.cluster_cloud_resources.autoscaling_group_masters_id}"
    autoscaling_group_nodes_id      = "${data.kops_cloud_resources.cluster_cloud_resources.autoscaling_group_nodes_id}"
    dhcp_options_id                 = "${data.kops_cloud_resources.cluster_cloud_resources.dhcp_options_id}"
    etcd_volumes_id                 = "${data.kops_cloud_resources.cluster_cloud_resources.etcd_volumes_id}"
    iam_instance_profile_masters_id = "${data.kops_cloud_resources.cluster_cloud_resources.iam_instance_profile_masters_id}"
    iam_instance_profile_nodes_id   = "${data.kops_cloud_resources.cluster_cloud_resources.iam_instance_profile_nodes_id}"
    iam_role_masters_id             = "${data.kops_cloud_resources.cluster_cloud_resources.iam_role_masters_id}"
    iam_role_nodes_id               = "${data.kops_cloud_resources.cluster_cloud_resources.iam_role_nodes_id}"
    instance_bastion_id             = "${data.kops_cloud_resources.cluster_cloud_resources.instance_bastion_id}"
    instance_masters_id             = "${data.kops_cloud_resources.cluster_cloud_resources.instance_masters_id}"
    instance_nodes_id               = "${data.kops_cloud_resources.cluster_cloud_resources.instance_nodes_id}"
    internet_gateway_id             = "${data.kops_cloud_resources.cluster_cloud_resources.internet_gateway_id}"
    keypair_id                      = "${data.kops_cloud_resources.cluster_cloud_resources.keypair_id}"
    route_table_id                  = "${data.kops_cloud_resources.cluster_cloud_resources.route_table_id}"
    route53_records_api             = "${data.kops_cloud_resources.cluster_cloud_resources.route53_records_api}"
    route53_records_etcd_id         = "${data.kops_cloud_resources.cluster_cloud_resources.route53_records_etcd_id}"
    security_group_elbs_id          = "${data.kops_cloud_resources.cluster_cloud_resources.security_group_elbs_id}"
    security_group_masters_id       = "${data.kops_cloud_resources.cluster_cloud_resources.security_group_masters_id}"
    security_group_nodes_id         = "${data.kops_cloud_resources.cluster_cloud_resources.security_group_nodes_id}"
    subnets_id                      = "${data.kops_cloud_resources.cluster_cloud_resources.subnets_id}"
    vpc_id                          = "${data.kops_cloud_resources.cluster_cloud_resources.vpc_id}"
  }

}
##################################################################################################
# S3
##################################################################################################
resource "aws_s3_bucket" "kops_state" {
  bucket        = "k8s.urbanradikal.com"
  acl           = "private"
  region        = "us-east-1"
  force_destroy = "true"

  versioning {
    enabled = true
  }

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}

##################################################################################################
# IAM
##################################################################################################
locals {
  create_kops_user = "true"
}

variable "kops_attach_policy" {
  default = [
    "AmazonEC2FullAccess",
    "AmazonRoute53FullAccess",
    "AmazonS3FullAccess",
    "IAMFullAccess",
    "AmazonVPCFullAccess",
  ]
}


resource "aws_iam_user" "kops" {
  count = "${local.create_kops_user ? 1 : 0}"
  name  = "kops"
}

resource "aws_iam_group" "kops" {
  count = "${local.create_kops_user ? 1 : 0}"
  name  = "kops"
}

resource "aws_iam_group_policy_attachment" "kops_attach" {
  count      = "${local.create_kops_user ? length(var.kops_attach_policy) : 0}"
  group      = "${aws_iam_group.kops.name}"
  policy_arn = "arn:aws:iam::aws:policy/${element(var.kops_attach_policy, count.index)}"
}

resource "aws_iam_user_group_membership" "kops_membership" {
  count = "${local.create_kops_user ? 1 : 0}"
  user  = "${aws_iam_user.kops.name}"

  groups = [
    "${aws_iam_group.kops.name}",
  ]
}

resource "aws_iam_access_key" "kops" {
  count = "${local.create_kops_user ? 1 : 0}"
  user  = "${aws_iam_user.kops.name}"
}
