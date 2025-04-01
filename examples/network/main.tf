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


# private network
resource "dfcloud_network" "network" {
  name = "network"
  location = {
    region   = "us-east-1"
    provider = "aws"
  }
  cidr_block = "192.168.0.0/16"
}

resource "dfcloud_datastore" "datastore" {
  name = "tf-test-no-pass"

  tier = {
    max_memory_bytes = 3000000000
    performance_tier = "dev"
    replicas         = 1
  }

  location = {
    region   = "us-east-1"
    provider = "aws"
  }

  disable_pass_key = true
  network_id = dfcloud_network.network.id
}
