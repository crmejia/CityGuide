# City Guide
Is a webapp to help you discover and share cool spots in cities.
These guides are useful for backpackers, digital nomads, and 
people new in town that are keen to discover cool things.

## Installation on an Ubuntu Server
### Prerequisites
Login as root on your server and create a user for the app.
```bash
# adduser cityguide
# usermod -aG sudo cityguide
# su cityguide
# sudo apt update #verify the user was added to sudoers
```
Install nginx, if not installed.

### Deploying
1. `ssh` into your server as the `cityguide` user.
2. Stop the service(if running) `service cityguide stop`
3. Copy(using `scp` or copy/paste) the systemd service unit.
```bash
$ scp cityguide/deploy/cityguide.service root@<YOUR_SERVER>:/lib/systemd/system/
```
4. Download the binary, move it to `/usr/local/bin`, and make it executable
```bash
$ curl -L -o cityguide https://github.com/crmejia/CityGuide/releases/download/<RELEASE_VERSION>/cityguide
$ mv cityguide /usr/local/bin
$ chmod 0100 /usr/local/bin/cityguide
```
5. Start service  `sudo service cityguide start`.
6. Check the status `service cityguide status`.

### Setting up reverse proxy with nginx
1. Copy(using `scp` or copy/paste) the nginx site configuration.
```bash
$ scp cityguide/deploy/cityguide.nginx-reverse-proxy root@<YOUR_SERVER>:/etc/nginx/sites-available/cityguide
```
2. Create a symlink `sudo ln -s /etc/nginx/sites-available/cityguide /etc/nginx/sites-enabled/cityguide`
3. Reload the nginx conf `sudo nginx -s reload`
4. Open your browser and visit the site.