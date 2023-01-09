# City Guide
Is a webapp to help you discover and share cool spots in cities. These guides are useful for backpackers, digital 
nomads, and people new in town that are keen to discover cool things.
## Development Setup
## Deployment 
These instructions are for a Digital Ocean Ubuntu droplet though they are interchangeable for other GNU Linux systems.  

### Provision a Digital Ocean droplet
_The following instructions are meant to provision a Digital Ocean droplet. Feel free to use your favorite cloud or 
bare metal and skip to the next section._

Install [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli) and create a ssh key pair to access 
the droplet.

Edit the file `deploy/infrastructure/terraform.tfvars` and change `public_key` to the public key you've created. Once
again, feel free to change the defaults. Then proceed to create the infra.

```bash
cd deploy/infrastructure
terraform apply
# verify changes and if correct, type `yes`
```
The output of the command is the IP of the droplet. Save it, we'll point our domain to it later.
### Installation on an Ubuntu Server
Check out the [hardening guide](docs/hardening.md) for instructions on how to harden sshd, setup a firewall with ufw, and setup fail2ban.

### Prerequisites
_Skip this section if you are using the provided droplet this configuration is already taken care of._

Login as root on your server and create a user for the app.
```bash
# adduser cityguide
# usermod -aG sudo cityguide
# su cityguide
# sudo apt update #verify the user was added to sudoers
```
Install:
* [Docker Engine](https://docs.docker.com/engine/install/ubuntu) 
* [Docker Compose](https://docs.docker.com/compose/install/linux/)
* [Caddy](https://caddyserver.com/docs/install#debian-ubuntu-raspbian)


### Deploying 
1. Modify `cityguide/deploy/Caddyfile` to point to your domain and also to your email for HTTPS cert validation.
2. Copy(using `scp` or copy/paste) the `docker-compose.yml` and `Caddyfile`.
```bash
$ scp cityguide/deploy/docker-compose.yml cityguide@<YOUR_SERVER>:/home/cityguide
$ scp cityguide/deploy/Caddyfile cityguide@<YOUR_SERVER>:/home/cityguide
```
3. `ssh` into your server as the `cityguide` user and navigate home `cd ~`.
4. run `docker compose up -d`
5. Add an A record to your domain that points to the IP address of the droplet that we saved earlier. 
6. Open your browser and visit the site.