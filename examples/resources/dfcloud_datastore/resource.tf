# Simple datastore on GCP
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

# Clustered datastore on GCP
resource "dfcloud_datastore" "cache_cluster" {
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

# Datastore in a private network on AWS
resource "dfcloud_network" "network" {
  name = "my-network"

  location = {
    region   = "eu-west-1"
    provider = "aws"
  }

  cidr_block = "192.168.0.0/16"
}

resource "dfcloud_datastore" "private" {
  name       = "my-cache-datastore"
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
