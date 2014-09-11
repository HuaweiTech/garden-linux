Garden in Go, on linux

* [![Coverage Status](https://coveralls.io/repos/cloudfoundry-incubator/garden-linux/badge.png)](https://coveralls.io/r/cloudfoundry-incubator/garden-linux)
* [Tracker](https://www.pivotaltracker.com/s/projects/962374)

# Running

For development, you can just spin up the Vagrant VM and run the server
locally, pointing at its host:

```bash
# if you need it:
vagrant plugin install vagrant-omnibus

# then:
librarian-chef install
vagrant up
ssh-copy-id vagrant@192.168.50.5
ssh vagrant@192.168.50.5 sudo cp -r .ssh/ /root/.ssh/
./scripts/add-route
./scripts/run-garden-remote-linux

# or run from inside the vm:
vagrant ssh
sudo su -
goto garden-linux
./scripts/run-garden-linux
```

This runs the server locally and configures the Linux backend to do everything
over SSH to the Vagrant box.

# Testing

## Pre-requisites

* [Docker](https://www.docker.io/) v0.9.0 or later (for creating a root filesystem)
* [git](http://git-scm.com/) (for garden and its dependencies on github)
* [mercurial](http://mercurial.selenic.com/) (for some dependencies not on github)

Run **all** the following commands **as root**.

Make a directory to contain go code:
```
# mkdir ~/go
```

From now on, we assume this directory is in `/root/go`.

Install Go 1.2.1 or later. For example, install [gvm](https://github.com/moovweb/gvm) and issue:
```
# gvm install go1.2.1
# gvm use go1.2.1
```

Extend `$GOPATH` and `$PATH`:
```
# export GOPATH=/root/go:$GOPATH
# export PATH=$PATH:/root/go/bin
```

Install [godep](https://github.com/kr/godep) (used to manage garden's dependencies):
```
# go get github.com/kr/godep
```

Get garden and its dependencies:
```
# go get github.com/cloudfoundry-incubator/garden-linux
# cd /root/go/src/github.com/cloudfoundry-incubator/garden-linux
# godep restore
```

Make the C code:
```
# make
```

Create a root filesystem, extract it (still as root), and point to it:
```
# make garden-test-rootfs.tar
# gzip garden-test-rootfs.tar
# mkdir -p /var/garden/rootfs
# tar xzf garden-test-rootfs.tar.gz -C /var/garden/rootfs
# export GARDEN_TEST_ROOTFS=/var/garden/rootfs
```
(You may wish to save the root filesystem tar.gz file for future use.)

Install ginkgo (used to test garden):
```
# go install github.com/onsi/ginkgo/ginkgo
```

Run the tests (skipping performance measurements):
```
# ginkgo -r -skipMeasurements
```
