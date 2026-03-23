resource "dfcloud_network" "network" {
  name = "my-network"

  location = {
    region   = "eu-west-1"
    provider = "aws"
  }

  cidr_block = "192.168.0.0/16"
}
