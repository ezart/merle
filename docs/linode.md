# Linode server setup

## Create server

On linode.com, create a new Nanode 1G share machine installing Ubuntu 21.10

## Enable SSH key

Copy the id_rsa.pub key to server:

```
$ ssh-copy-id <user>@<server>
```

Verify you can log into server without password:

```
ssh <user>@<server>
```
  
## Update system

```
$ sudo apt update
```

## Install Go language

```
$ sudo apt install golang-go -y
```

```
$ go version
go version go1.17 linux/amd64
```

## Install additional libs for build

```
sudo apt install gcc libpam0g-dev -y
```

## Build Merle

```
$ cd merle
$ go install ./...
```

