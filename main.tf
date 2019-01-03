provider "aws" {
  region  = "us-east-1"
  version = "~> 1.45"
}
provider "helm" {
  debug           = "true"
  enable_tls      = "false"
  install_tiller  = "false"
  namespace       = "kube-system"
  service_account = "tiller"

  kubernetes {}
}

resource "kops_cluster" "aux_cluster" {

  etcd_version       = "3.2.24"
  image              = "ami-03b850a018c8cd25e"
  k8s_version        = "v1.11.5"
  master_count       = 1
  master_size        = "t2.medium"
  master_volume_size = 20
  master_zones       = ["us-east-1a"]
  name               = "green.k8s.urbanradikal.com"
  network_cidr       = "10.0.0.0/16"
  node_max_size      = 5
  node_min_size      = 3
  node_size          = "t2.medium"
  node_volume_size   = 20
  node_zones         = ["us-east-1a", "us-east-1b", "us-east-1c"]
  ssh_public_key     = "~/.ssh/kalada-admin.pub"
  state_store        = "s3://k8s.kaladaopuiyo.com"

  depends_on = ["aws_s3_bucket.kops_state"]
}
##########################################################################
# S3
##########################################################################
resource "aws_s3_bucket" "kops_state" {
  bucket        = "k8s.kaladaopuiyo.com"
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
