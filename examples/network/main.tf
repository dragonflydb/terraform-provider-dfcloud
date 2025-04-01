terraform {
  required_providers {
    dfcloud = {
      source = "registry.terraform.io/dragonflydb/dfcloud"
    }
  }
}

provider "dfcloud" {
  # Configuration options
  api_host = "api.dev.dragonflydb.cloud"
}


# private network
resource "dfcloud_network" "network" {
  name = "network"
  location = {
    region   = "eu-west-1"
    provider = "aws"
  }
  cidr_block = "192.168.0.0/16"
}


resource "dfcloud_datastore" "test" {
  name       = "my-cache-datastore-test"
  network_id = dfcloud_network.network.id

  location = {
    region             = "eu-west-1"
    availability_zones = ["euw1-az2"]
    provider           = "aws"
  }

  disable_pass_key = true

  tier = {
    max_memory_bytes = 3000000000
    performance_tier = "dev"
    replicas         = 0
  }

  dragonfly = {
    cache_mode = false
  }
}
