provider "aws" {
  region  = "us-east-1"
  version = "~> 1.45"
}


resource "kops_cluster" "aux_cluster" {


  master_zones   = ["us-east-1a"]
  name           = "green.k8s.urbanradikal.com"
  node_zones     = ["us-east-1a"]
  ssh_public_key = "~/.ssh/kalada-admin.pub"
  state_store    = "s3://k8s.kaladaopuiyo.com"

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
