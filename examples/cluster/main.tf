terraform {
  required_providers {
    dfcloud = {
      source = "registry.terraform.io/dragonflydb/dfcloud"
    }
  }
}

provider "dfcloud" {
}

resource "dfcloud_datastore" "cache-cluster" {
  name = "frontend-cache-cluster"

  location = {
    region   = "us-central1"
    provider = "gcp"
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
