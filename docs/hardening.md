
## Creation
I added the key manually as the tf module was not reading the fingerprint.

## Hardening
### configure sshd 
File `/etc/ssh/sshd_config`
Change port to something else in the range 1025 - 65500.
Disable password authentication
```
Port <SSH-PORT>
PasswordAuthentication no
```
Create an alias for the ssh connection:
alias <random-name>="ssh -p <SSH-PORT> <USERNAME>@<LOCAL-IP>"

### firewall with ufw
```bash
ufw default deny incoming
 ufw default allow outgoing
ufw allow <PORT>/tcp comment 'SSH'
ufw allow http
ufw allow https
ufw enable
ufw status #verify
```

### Install and setup fail2ban
```bash
apt install fail2ban
cp /etc/fail2ban/fail2ban.conf /etc/fail2ban/fail2ban.local
```
Add an SSH jail to the end of the file
```
/etc/fail2ban/fail2ban.local
[sshd]
enabled = true
port = 549
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = -1
```

### Users
* root
* cityguide