## Upgrades
### Server Upgrades
```bash
$ apt-get update
$ apt-get upgrade
$ apt-get dist-upgrade #In order to install packages kept back https://serverfault.com/questions/265410/ubuntu-server-message-says-packages-can-be-updated-but-apt-get-does-not-update
$ apt auto-remove
```

## Operations
### git
Append authorized public keys to .ssh/authorized_keys

You can set up an empty repository for them by running git init with the --bare option, which initializes the repository without a working directory. Note that someone must shell onto the machine and create a bare repository every time you want to add a project:
```bash
$ cd /srv/git
$ mkdir project.git
$ cd project.git
$ git init --bare
Initialized empty Git repository in /srv/git/project.git/
```

On the repo
```
$ cd myproject
$ git init
$ git add .
$ git commit -m 'Initial commit'
$ git remote add origin git@gitserver:/srv/git/project.git
$ git push origin master
```