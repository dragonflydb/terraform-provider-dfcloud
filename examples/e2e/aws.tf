# AWS provider

provider "aws" {
  # Configuration options
   region = "us-east-1"
}


data "aws_caller_identity" "current" {}

resource "aws_key_pair" "tf_key" {
  key_name   = "tf_key"
  public_key = tls_private_key.rsa.public_key_openssh
}

resource "tls_private_key" "rsa" {
  algorithm = "RSA"
  rsa_bits  = 4096
}


resource "local_file" "tf_key" {
  content  = tls_private_key.rsa.private_key_pem
  filename = "./tf_key.pem"
}

# client VPC
resource "aws_vpc" "client" {
  cidr_block = "172.16.0.0/16"
 
  tags = {
    Name = "tf-client-vpc"
  }
}

resource "aws_internet_gateway" "test_env_gw" {
vpc_id = aws_vpc.client.id
}


resource "aws_subnet" "my_subnet" {
  vpc_id            = aws_vpc.client.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = "us-east-1a"

tags = {
    Name = "tf-client-subnet"
  }
}

resource "aws_security_group" "security" {
  name = "allow-all"

  vpc_id = aws_vpc.client.id

  ingress {
    cidr_blocks = [
      "0.0.0.0/0"
    ]
    from_port = 22
    to_port   = 22
    protocol  = "tcp"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}                                                                                    


resource "aws_network_interface" "primary" {
  subnet_id   = aws_subnet.my_subnet.id
  private_ips = ["172.16.10.100"]

  tags = {
    Name = "primary_network_interface"
  }
}

# create an instance

resource "aws_instance" "vm" {
  ami           = "ami-011ba4969cf2d6f9b"
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.my_subnet.id

  tags = {
    Name = "tf-client-instance"
  }

    key_name = aws_key_pair.tf_key.key_name

  security_groups = [aws_security_group.security.id, aws_security_group.allow_dfcloud.id]

    associate_public_ip_address = true
}

resource "aws_route_table" "route-public" {
  vpc_id = aws_vpc.client.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test_env_gw.id
  }

  tags = {
    Name = "public-route-table-demo"
  }
}

resource "aws_route_table_association" "public_1" {
  subnet_id      = aws_subnet.my_subnet.id
  route_table_id = aws_route_table.route-public.id
}
