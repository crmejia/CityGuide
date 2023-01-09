// digital ocean access token
variable "do_token" {
  type = string
  sensitive = true
}

variable "project_name"{
    type = string
}

variable "vpc_name"{
    type = string
}

variable "region"{
    type = string
}

variable "droplet_name"{
    type = string
}

variable "droplet_size"{
    type = string
}
variable "droplet_image"{
    type = string
}

variable "ssh_key_name" {
  type = string
}

variable "public_key" {
  type = string
}

output "droplet_ip_address" {
  description = "Public ip address of the droplet"
  value = digitalocean_droplet.redforce.ipv4_address
}