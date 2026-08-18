terraform {
  required_providers {
    dfcloud = {
      source = "registry.terraform.io/dragonflydb/dfcloud"
    }
  }
}

provider "dfcloud" {
  # Configuration options
}

resource "dfcloud_datastore" "cache-cluster" {
  name = "frontend-cache-cluster"

  location = {
    provider = "gcp"
    region   = "us-central1"
    availability_zones = ["us-central1-a", "us-central1-a"]
  }

  tier = {
    max_memory_bytes = 6000000000
    performance_tier = "dev"
    replicas         = 1
  }

  cluster = {
    shard_memory = 3000000000
  }

  dragonfly = {
    cache_mode = true
  }
}
