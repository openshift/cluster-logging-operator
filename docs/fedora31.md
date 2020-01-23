## Installing docker on Fedora 31 (and maybe later)

In Fedora 31, default kernel settings changed and are incompatible with docker.

Here's how to fix that (steps taken from [here](https://linuxconfig.org/how-to-install-docker-on-fedora-31))

You should `dnf erase` any existing docker, docker-ce or moby packages first.

```
sudo dnf install -y grubby
sudo grubby --update-kernel=ALL --args="systemd.unified_cgroup_hierarchy=0"
sudo reboot
# After reboot
sudo dnf config-manager --add-repo=https://download.docker.com/linux/fedora/docker-ce.repo
sudo dnf install -y docker-ce
sudo systemctl enable --now docker
sudo groupadd docker
sudo usermod -aG docker $USER
```
