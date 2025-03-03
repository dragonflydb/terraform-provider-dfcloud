terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.50.0"
    }

    dfcloud = {
      source = "registry.terraform.io/dragonflydb/dfcloud"
    }
  }
}

provider "aws" {
  # Configuration options
}

provider "dfcloud" {
  # Configuration options
}

data "aws_caller_identity" "current" {}

# client VPC
resource "aws_vpc" "client" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "client"
  }
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

# private connection
resource "dfcloud_connection" "connection" {
  depends_on = [aws_vpc.client, dfcloud_network.network]

  name = "connection"
  peer = {
    account_id = data.aws_caller_identity.current.account_id
    region     = "us-east-1"
    vpc_id     = aws_vpc.client.id
  }

  network_id = dfcloud_network.network.id
}

resource "aws_vpc_peering_connection_accepter" "accepter" {
  depends_on = [dfcloud_connection.connection]

  vpc_peering_connection_id = dfcloud_connection.connection.peer_connection_id
  auto_accept               = true
}
