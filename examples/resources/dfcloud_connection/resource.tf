# Create a network first
resource "dfcloud_network" "network" {
  name = "my-network"

  location = {
    region   = "us-east-1"
    provider = "aws"
  }

  cidr_block = "192.168.0.0/16"
}

# AWS VPC peering connection
resource "dfcloud_connection" "aws_peering" {
  name       = "my-aws-connection"
  network_id = dfcloud_network.network.id

  peer = {
    account_id = "123456789012"
    region     = "us-east-1"
    vpc_id     = "vpc-0123456789abcdef0"
  }
}
