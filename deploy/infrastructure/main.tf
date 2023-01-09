terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

# Configure the DigitalOcean Provider
provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_project" "fleet" {
  name        = var.project_name
  resources   = [digitalocean_droplet.redforce.urn]
}

resource "digitalocean_vpc" "vpc" {
  name = var.vpc_name
  region = var.region
}

data "digitalocean_ssh_key" "ssh_key" {
  name = var.ssh_key_name
}

resource "digitalocean_droplet" "redforce" {
  name   = var.droplet_name
  size   = var.droplet_size
  image  = var.droplet_image
  region = var.region
  ssh_keys = [data.digitalocean_ssh_key.ssh_key.fingerprint]
  vpc_uuid = digitalocean_vpc.vpc.id
  user_data  = file("../scripts/cloud_init.yaml")
}