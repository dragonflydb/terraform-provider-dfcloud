terraform {
  required_providers {
    dfcloud = {
      source  = "dragonflydb/dfcloud"
      version = "0.0.5"
    }
  }
}

provider "dfcloud" {
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

resource "dfcloud_datastore" "cache" {
  depends_on = [dfcloud_connection.connection]

  name = "cache"
  location = {
    region   = "us-east-1"
    provider = "aws"
  }
  network_id = dfcloud_network.network.id
  tier = {
    max_memory_bytes = 6000000000
    performance_tier = "dev"
    replicas         = 1
  }
}

# accept the peering connection
resource "aws_vpc_peering_connection_accepter" "accepter" {
  depends_on = [dfcloud_connection.connection]

  vpc_peering_connection_id = dfcloud_connection.connection.peer_connection_id
  auto_accept               = true
}

# add the required route to the client VPC
resource "aws_route" "route" {
  depends_on = [aws_vpc_peering_connection_accepter.accepter]

  route_table_id            = aws_route_table.route-public.id
  destination_cidr_block    = dfcloud_network.network.cidr_block
  vpc_peering_connection_id = dfcloud_connection.connection.peer_connection_id
}


resource "aws_route_table_association" "private_1" {
  subnet_id      = aws_subnet.my_subnet.id
  route_table_id = aws_route_table.route-public.id
}


# now allow in the security group
resource "aws_security_group" "allow_dfcloud" {
  depends_on = [aws_vpc.client]

  vpc_id = aws_vpc.client.id

  egress {
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = [dfcloud_network.network.cidr_block]
  }
  ingress {
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = [dfcloud_network.network.cidr_block]
  }
}

output "redis-endpoint" {
  sensitive = true
  value     = "redis://default:${dfcloud_datastore.cache.password}@${dfcloud_datastore.cache.addr}"
}

output "instance-ip" {
  value = aws_instance.vm.public_ip
}

output "instance-id" {
  value = aws_instance.vm.id
}
