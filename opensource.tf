## Terraform script for provisioning open-source alternatives to AWS Services

# Provider configuration
provider "aws" {
  region = "us-east-1"
}

# 1. Networking (Calico for VPC-like networking)
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
  enable_dns_support = true
  enable_dns_hostnames = true
  tags = {
    Name = "Main-VPC"
  }
}

# Subnets
resource "aws_subnet" "subnet_1" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
  availability_zone = "us-east-1a"
  tags = {
    Name = "Main-Subnet-1"
  }
}

resource "aws_subnet" "subnet_2" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.2.0/24"
  availability_zone = "us-east-1b"
  tags = {
    Name = "Main-Subnet-2"
  }
}

# 2. Internet Gateway
resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.main.id
  tags = {
    Name = "Main-InternetGateway"
  }
}

# 3. Route Table
resource "aws_route_table" "main_rt" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Name = "Main-Route-Table"
  }
}

# Route Table Association
resource "aws_route_table_association" "subnet_1_assoc" {
  subnet_id      = aws_subnet.subnet_1.id
  route_table_id = aws_route_table.main_rt.id
}

resource "aws_route_table_association" "subnet_2_assoc" {
  subnet_id      = aws_subnet.subnet_2.id
  route_table_id = aws_route_table.main_rt.id
}


# 4. Nginx Load Balancer Setup
resource "aws_instance" "nginx_lb" {
  ami           = "ami-0c55b159cbfafe1f0" # Amazon Linux 2 AMI
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.subnet_1.id

  user_data = <<-EOF
              #!/bin/bash
              yum update -y
              amazon-linux-extras install nginx1 -y
              systemctl enable nginx
              systemctl start nginx
            EOF

  tags = {
    Name = "Nginx-LoadBalancer"
  }
}

# 5. Kubernetes Cluster (K3s)
resource "null_resource" "k3s_master" {
  provisioner "local-exec" {
    command = <<-EOT
      ssh -o StrictHostKeyChecking=no ec2-user@${aws_instance.nginx_lb.public_ip} "curl -sfL https://get.k3s.io | sh -"
    EOT
  }
  depends_on = [aws_instance.nginx_lb]
}

# 6. MinIO Object Storage
resource "aws_instance" "minio" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.subnet_1.id

  user_data = <<-EOF
              #!/bin/bash
              yum update -y
              curl -O https://dl.min.io/server/minio/release/linux-amd64/minio
              chmod +x minio
              mv minio /usr/local/bin/
              mkdir -p /data/minio
              export MINIO_ACCESS_KEY="minioadmin"
              export MINIO_SECRET_KEY="minioadmin"
              nohup minio server /data/minio --console-address ":9001" &
            EOF

  tags = {
    Name = "MinIO-ObjectStorage"
  }
}


# 7. PostgreSQL Database
resource "aws_instance" "postgres" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.subnet_1.id

  user_data = <<-EOF
              #!/bin/bash
              yum update -y
              yum install postgresql-server -y
              postgresql-setup initdb
              systemctl enable postgresql
              systemctl start postgresql
            EOF

  tags = {
    Name = "PostgreSQL-Database"
  }
}


# Next Steps:
# - ScyllaDB Setup (DynamoDB alternative)
# - Monitoring (Prometheus & Grafana)
# - Queue and Messaging (RabbitMQ)
# - Caching (Redis)
# - DNS and API Gateway

# These components will be added incrementally in the next updates.
