package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		// 1. VPC Setup
		vpc, err := ec2.NewVpc(ctx, "mainVpc", &ec2.VpcArgs{
			CidrBlock:          pulumi.String("10.0.0.0/16"),
			EnableDnsSupport:   pulumi.Bool(true),
			EnableDnsHostnames: pulumi.Bool(true),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("Main-VPC"),
			},
		})
		if err != nil {
			return err
		}

		// 2. Subnets
		subnet1, err := ec2.NewSubnet(ctx, "subnet1", &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String("10.0.1.0/24"),
			AvailabilityZone: pulumi.String("us-east-1a"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("Main-Subnet-1"),
			},
		})
		if err != nil {
			return err
		}

		subnet2, err := ec2.NewSubnet(ctx, "subnet2", &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String("10.0.2.0/24"),
			AvailabilityZone: pulumi.String("us-east-1b"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("Main-Subnet-2"),
			},
		})
		if err != nil {
			return err
		}

		// 3. Internet Gateway
		igw, err := ec2.NewInternetGateway(ctx, "mainIgw", &ec2.InternetGatewayArgs{
			VpcId: vpc.ID(),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("Main-InternetGateway"),
			},
		})
		if err != nil {
			return err
		}

		// 4. Route Table
		rt, err := ec2.NewRouteTable(ctx, "mainRt", &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
			Routes: ec2.RouteTableRouteArray{
				ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String("0.0.0.0/0"),
					GatewayId: igw.ID(),
				},
			},
			Tags: pulumi.StringMap{
				"Name": pulumi.String("Main-Route-Table"),
			},
		})
		if err != nil {
			return err
		}

		// Associations
		_, err = ec2.NewRouteTableAssociation(ctx, "subnet1Assoc", &ec2.RouteTableAssociationArgs{
			SubnetId:     subnet1.ID(),
			RouteTableId: rt.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2.NewRouteTableAssociation(ctx, "subnet2Assoc", &ec2.RouteTableAssociationArgs{
			SubnetId:     subnet2.ID(),
			RouteTableId: rt.ID(),
		})
		if err != nil {
			return err
		}

		// Nginx Load Balancer
		lb, err := ec2.NewInstance(ctx, "nginxLb", &ec2.InstanceArgs{
			Ami:           pulumi.String("ami-0c55b159cbfafe1f0"),
			InstanceType:  pulumi.String("t2.micro"),
			SubnetId:      subnet1.ID(),
			UserData:      pulumi.String(`#!/bin/bash
								yum update -y
								amazon-linux-extras install nginx1 -y
								systemctl enable nginx
								systemctl start nginx`),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("Nginx-Load-Balancer"),
			},
		})
		if err != nil {
			return err
		}

		// K3s Kubernetes Cluster
		k3sMaster, err := ec2.NewInstance(ctx, "k3sMaster", &ec2.InstanceArgs{
			Ami:           pulumi.String("ami-0c55b159cbfafe1f0"),
			InstanceType:  pulumi.String("t2.micro"),
			SubnetId:      subnet2.ID(),
			UserData:      pulumi.String(`#!/bin/bash
								curl -sfL https://get.k3s.io | sh -`),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("K3s-Master"),
			},
		})
		if err != nil {
			return err
		}

		// MinIO Object Storage
		minioInstance, err := ec2.NewInstance(ctx, "minio", &ec2.InstanceArgs{
			Ami:           pulumi.String("ami-0c55b159cbfafe1f0"),
			InstanceType:  pulumi.String("t2.micro"),
			SubnetId:      subnet1.ID(),
			UserData: pulumi.String(`#!/bin/bash
								yum update -y
								curl -O https://dl.min.io/server/minio/release/linux-amd64/minio
								chmod +x minio
								mv minio /usr/local/bin/
								mkdir -p /data/minio
								export MINIO_ACCESS_KEY=minioadmin
								export MINIO_SECRET_KEY=minioadmin
								nohup minio server /data/minio --console-address ':9001' &`),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("MinIO-Object-Storage"),
			},
		})
		if err != nil {
			return err
		}

		// PostgreSQL Database
		postgresInstance, err := ec2.NewInstance(ctx, "postgresDb", &ec2.InstanceArgs{
			Ami:           pulumi.String("ami-0c55b159cbfafe1f0"),
			InstanceType:  pulumi.String("t2.micro"),
			SubnetId:      subnet1.ID(),
			UserData: pulumi.String(`#!/bin/bash
								yum update -y
								yum install postgresql-server -y
								postgresql-setup initdb
								systemctl enable postgresql
								systemctl start postgresql`),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("PostgreSQL-Database"),
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
