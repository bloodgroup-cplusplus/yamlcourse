package main

import (
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		// Create a custom Docker network
		network, err := docker.NewNetwork(ctx, "my_network", &docker.NetworkArgs{
			Driver: pulumi.String("bridge"),
		})
		if err != nil {
			return err
		}

		// 1. Nginx Load Balancer
		nginx, err := docker.NewContainer(ctx, "nginx_lb", &docker.ContainerArgs{
			Image:   pulumi.String("nginx:latest"),
			Ports:   docker.ContainerPortArray{{Internal: pulumi.Int(80), External: pulumi.Int(80)}},
			Network: network.ID(),
			Restart: pulumi.String("always"),
		})
		if err != nil {
			return err
		}

		// 2. K3s Master
		k3sMaster, err := docker.NewContainer(ctx, "k3s_master", &docker.ContainerArgs{
			Image:   pulumi.String("rancher/k3s:v1.22.4-k3s1"),
			Ports:   docker.ContainerPortArray{{Internal: pulumi.Int(6443), External: pulumi.Int(6443)}},
			Network: network.ID(),
			Command: pulumi.String("--no-deploy traefik"),
			Environment: pulumi.StringMap{
				"K3S_KUBECONFIG_MODE": pulumi.String("0644"),
			},
			Restart: pulumi.String("always"),
		})
		if err != nil {
			return err
		}

		// 3. MinIO Object Storage
		minio, err := docker.NewContainer(ctx, "minio", &docker.ContainerArgs{
			Image:   pulumi.String("minio/minio:latest"),
			Ports:   docker.ContainerPortArray{{Internal: pulumi.Int(9000), External: pulumi.Int(9000)}},
			Network: network.ID(),
			Environment: pulumi.StringMap{
				"MINIO_ACCESS_KEY": pulumi.String("minioadmin"),
				"MINIO_SECRET_KEY": pulumi.String("minioadmin"),
			},
			Command: pulumi.String("server /data"),
			Restart: pulumi.String("always"),
		})
		if err != nil {
			return err
		}

		// 4. PostgreSQL Database
		postgres, err := docker.NewContainer(ctx, "postgres_db", &docker.ContainerArgs{
			Image:   pulumi.String("postgres:latest"),
			Ports:   docker.ContainerPortArray{{Internal: pulumi.Int(5432), External: pulumi.Int(5432)}},
			Network: network.ID(),
			Environment: pulumi.StringMap{
				"POSTGRES_USER":     pulumi.String("postgres"),
				"POSTGRES_PASSWORD": pulumi.String("password"),
			},
			Volumes: pulumi.StringArray{pulumi.String("/var/lib/postgresql/data")},
			Restart: pulumi.String("always"),
		})
		if err != nil {
			return err
		}

		// Outputs
		ctx.Export("nginx_lb_ip", nginx.IPAddress)
		ctx.Export("minio_ip", minio.IPAddress)
		ctx.Export("postgres_ip", postgres.IPAddress)

		return nil
	})
}
