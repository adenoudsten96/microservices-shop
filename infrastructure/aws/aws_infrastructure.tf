provider "aws" {
  profile = "default"
  region = "us-east-1"
}

resource "aws_vpc" "alex_vpc" {
  cidr_block = "10.200.0.0/16"
}

resource "aws_internet_gateway" "alex_gateway" {
  vpc_id = aws_vpc.alex_vpc.id
  tags = {
    name: "alex_gateway"
  }
}

resource "aws_subnet" "public1" {
  vpc_id = aws_vpc.alex_vpc.id
  cidr_block = "10.200.0.0/24"
  availability_zone = "us-east-1a"
}

resource "aws_route_table" "alex_routes" {
  vpc_id = aws_vpc.alex_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.alex_gateway.id
  }

  tags = {
    name = "alex_routes"
  }
}

resource "aws_route_table_association" "public1" {
  subnet_id = aws_subnet.public1.id
  route_table_id = aws_route_table.alex_routes.id
}

resource "aws_security_group" "kube_node" {
  name        = "kube-node"
  description = "allowed ports for kube-nodes"
  vpc_id      = aws_vpc.alex_vpc.id

  # Allow pinging
  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow SSH
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow HTTP
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow Kubernetes API
  ingress {
    from_port = 6443
    to_port = 6443
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow NodePort traffic
  ingress {
    from_port = 30000
    to_port = 31000
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow access to internet to download container images
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "kube_master" {
  name        = "kube-master"
  description = "allowed ports for kube-master"
  vpc_id      = aws_vpc.alex_vpc.id

  # Allow pinging
  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow SSH
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow HTTP
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow HTTPS
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow Kubernetes API
  ingress {
    from_port = 6443
    to_port = 6443
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow access to internet to download container images
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "kube-master" {
  ami = "ami-07d0cf3af28718ef8"
  instance_type = "t2.medium"
  key_name = "alex"
  vpc_security_group_ids = [aws_security_group.kube_master.id]
  subnet_id = aws_subnet.public1.id
  associate_public_ip_address = true
  tags = {
    name: "kube-master"
  }
}

resource "aws_instance" "kube-node" {
  ami = "ami-07d0cf3af28718ef8"
  instance_type = "t2.medium"
  key_name = "alex"
  vpc_security_group_ids = [aws_security_group.kube_node.id]
  subnet_id = aws_subnet.public1.id
  tags = {
    name: "kube-node"
  }
  associate_public_ip_address = true
}
