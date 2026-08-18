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

resource "dfcloud_datastore" "cache" {
  name = "frontend-cache"

  location = {
    provider = "gcp"
    region   = "us-central1"
    availability_zones = ["us-central1-a", "us-central1-a"]
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
