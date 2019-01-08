provider "aws" {
  region  = "us-east-1"
  version = "~> 1.45"
}
// Added for testing. Would be nice if helm could wait
// for the kops cluster to validate before trying to install....¯\_(ツ)_/¯
provider "helm" {
  debug           = "true"
  enable_tls      = "false"
  install_tiller  = "false"
  namespace       = "kube-system"
  service_account = "tiller"

  kubernetes {}
}
#################################################################################################
# KOPS CLUSTER
##################################################################################################

resource "kops_cluster" "aux_cluster" {

  admin_access           = ["0.0.0.0/0"] // optional,
  api_load_balancer_type = ""            // optional, not implemented, Testing
  associate_public_ip    = "true"        // optional
  authorization          = "AlwaysAllow"
  bastion                = "false" // working out the bugs leave as for testing
  cloud                  = "aws"   // Only AWS for now
  cloud_labels           = "Owner=Kalada Opuiyo,env=test"
  dns                    = "public" // working out bugs leave as for testing
  dry_run                = "false"
  etcd_version           = "3.2.24"
  encrypt_etcd_storage   = "true"
  image                  = "ami-03b850a018c8cd25e"
  k8s_version            = "v1.11.5"
  master_per_zone        = 1  // optional, default is 1 per zone odd numbers only
  master_security_groups = [] // optional, not implemented
  master_size            = "t2.medium"
  master_volume_size     = 20
  master_zones           = ["us-east-1f"]
  model                  = "" // optional, not implemented
  name                   = "k8s.urbanradikal.com"
  network_cidr           = "10.0.0.0/16"
  networking             = "calico"
  node_max_size          = 5
  node_min_size          = 2
  node_security_groups   = [] // optional, not implemented
  node_size              = "t2.medium"
  node_volume_size       = 20
  node_zones             = ["us-east-1a", "us-east-1c"]
  out                    = ""            // optional, not implemented
  output                 = ""            // optional, not implemented
  ssh_access             = ["0.0.0.0/0"] // optional
  ssh_public_key         = "~/.ssh/kalada-admin.pub"
  state_store            = "s3://${aws_s3_bucket.kops_state.id}"
  subnets                = []       // optional, not implemented
  target                 = ""       // optional, not implemented
  topology               = "public" // working out bugs leave as for testing
  utility_subnets        = []       // optional, not implemented
  validate_on_creation   = "false"  // optional works but still testing logic leave false
  vpc_id                 = ""       // optional, not tested
  zones                  = []       // optional, not implemented

  kubelet {
    anonymous_auth               = "false"
    authentication_token_webhook = "true"
    authorization_mode           = "webhook"

  }
  depends_on = ["aws_iam_user.kops"]
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
