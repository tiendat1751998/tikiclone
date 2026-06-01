terraform {
  backend "s3" {
    bucket = "tiki-clone-terraform-state"
    key    = "production/terraform.tfstate"
    region = "ap-southeast-1"
    encrypt = true
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
  }
}

provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
    }
  }
}

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "5.5.0"

  name = "tiki-${var.environment}"
  cidr = var.vpc_cidr

  azs             = var.availability_zones
  private_subnets = var.private_subnet_cidrs
  public_subnets  = var.public_subnet_cidrs

  enable_nat_gateway   = true
  single_nat_gateway   = false
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Environment = var.environment
    Project     = "tiki-clone"
  }
}

module "eks" {
  source = "terraform-aws-modules/eks/aws"
  version = "19.16.0"

  cluster_name    = "tiki-${var.environment}"
  cluster_version = "1.28"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  cluster_endpoint_public_access = var.environment == "production" ? false : true

  eks_managed_node_groups = {
    general = {
      desired_size = var.general_node_desired_size
      min_size     = var.general_node_min_size
      max_size     = var.general_node_max_size

      instance_types = var.general_node_instance_types

      tags = {
        Environment = var.environment
        NodeGroup   = "general"
      }
    }

    memory = {
      desired_size = var.memory_node_desired_size
      min_size     = var.memory_node_min_size
      max_size     = var.memory_node_max_size

      instance_types = var.memory_node_instance_types

      tags = {
        Environment = var.environment
        NodeGroup   = "memory"
      }
    }

    burst = {
      desired_size = 0
      min_size     = 0
      max_size     = var.burst_node_max_size

      instance_types = var.burst_node_instance_types

      tags = {
        Environment = var.environment
        NodeGroup   = "burst"
      }
    }
  }

  tags = {
    Environment = var.environment
    Project     = "tiki-clone"
  }
}
