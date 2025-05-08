provider "docker" {
  host = "unix:///var/run/docker.sock"
}

resource "docker_container" "nginx" {
  name  = "nginx_lb"
  image = "nginx:latest"
  ports {
    internal = 80
    external = 80
  }
  restart = "always"
}

resource "docker_container" "k3s_master" {
  name  = "k3s_master"
  image = "rancher/k3s:v1.22.4-k3s1"
  ports {
    internal = 6443
    external = 6443
  }
  environment = {
    K3S_KUBECONFIG_MODE = "0644"
  }
  command = "--no-deploy traefik"
  restart = "always"
}

resource "docker_container" "minio" {
  name  = "minio"
  image = "minio/minio:latest"
  ports {
    internal = 9000
    external = 9000
  }
  environment = {
    MINIO_ACCESS_KEY = "minioadmin"
    MINIO_SECRET_KEY = "minioadmin"
  }
  command = "server /data"
  restart = "always"
}

resource "docker_container" "postgres" {
  name  = "postgres_db"
  image = "postgres:latest"
  ports {
    internal = 5432
    external = 5432
  }
  environment = {
    POSTGRES_USER     = "postgres"
    POSTGRES_PASSWORD = "password"
  }
  volumes = [
    "/var/lib/postgresql/data"
  ]
  restart = "always"
}

resource "docker_network" "my_network" {
  name = "my_network"
}

resource "docker_container" "nginx" {
  name    = "nginx_lb"
  image   = "nginx"
  network = docker_network.my_network.name
  ports {
    internal = 80
    external = 80
  }
  depends_on = [
    docker_container.k3s_master,
    docker_container.minio,
    docker_container.postgres
  ]
}

output "nginx_lb_ip" {
  value = docker_container.nginx_lb.ip_address
}

output "minio_ip" {
  value = docker_container.minio.ip_address
}

output "postgres_ip" {
  value = docker_container.postgres.ip_address
}
