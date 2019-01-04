provider "aws" {
  region  = "us-east-1"
  version = "~> 1.45"
}
// Added for testing. Would be nice if helm could wait
// for the kops cluster to validate before trying to install....
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

// Commented out Parameter have not been
// implemented by the provider, future work :)
resource "kops_cluster" "aux_cluster" {
  admin_access           = ["0.0.0.0/0"]
  api_load_balancer_type = "public" //Testing
  associate_public_ip    = true     // does nothing for now :p may need to seperate between nodes and masters
  authorization          = "AlwaysAllow"
  bastion                = "false" //Testing
  cloud                  = "aws"
  cloud_labels           = "Owner=Kalada Opuiyo"
  dns                    = "public"
  encrypt_etcd_storage   = true
  etcd_version           = "3.2.24"
  image                  = "ami-03b850a018c8cd25e"
  k8s_version            = "v1.11.5"
  master_count           = 3
  master_size            = "t2.medium"
  master_volume_size     = 20
  master_zones           = ["us-east-1a", "us-east-1b", "us-east-1d"]
  name                   = "k8s.urbanradikal.com"
  network_cidr           = "10.0.0.0/16"
  # networking            = "calico" // This one I consider fun so saving for marriage
  node_max_size    = 5
  node_min_size    = 2
  node_size        = "t2.medium"
  node_volume_size = 20
  node_zones       = ["us-east-1a", "us-east-1b", "us-east-1c"]
  ssh_public_key   = "~/.ssh/kalada-admin.pub"
  state_store      = "s3://k8s.urbanradikal.com"
  topology         = "public"
  vpc_id           = ""
  depends_on       = ["aws_s3_bucket.kops_state"]
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
