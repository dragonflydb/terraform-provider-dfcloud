terraform {
  required_providers {
    dfcloud = {
      source = "registry.terraform.io/dragonflydb/dfcloud"
    }
  }
}

provider "dfcloud" {
}

resource "dfcloud_datastore" "cache" {
  name = "frontend-cache"

  location = {
    region   = "us-central1"
    provider = "gcp"
  }
 
  tier = {
    max_memory_bytes = 3000000000
    performance_tier = "dev"
    replicas         = 1
  }

  dragonfly = {
    cache_mode = true
  }
}
