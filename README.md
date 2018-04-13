## MulanFaaS - Serverless Functions Made Simple for Software Development

## Get started with MulanFaaS

### Pre-requisites:

#### Docker

For Mac

* [Docker CE for Mac Edge Edition](https://store.docker.com/editions/community/docker-ce-desktop-mac)

For Windows 

* Use Windows 10 Pro or Enterprise only
* Install [Docker CE for Windows](https://store.docker.com/editions/community/docker-ce-desktop-windows)
* Install [Git Bash](https://git-scm.com/downloads)

> Note: please use Git Bash for all steps: do not attempt to use *WSL* or *Bash for Windows*.

Linux - Ubuntu or Debian

* Docker CE for Linux

> You can install Docker CE from the [Docker Store](https://store.docker.com).

#### Clone MulanFaaS to your workspace

```
$ git clone https://github.com/mulansoft/mulanfaas.git
```

#### Prepare the environment


```
$ cd mulanfaas && make run-prepare
```

If you get an error after `make run-prepare`:
```
Error response from daemon: This node is already part of a swarm. Use "docker swarm leave" to leave this swarm and join another one.
```

You can fix it by run this command:
```
$ docker swarm leave --force
```

#### Install OpenFaaS CLI
```bash
$ make install-faas-cli
```
#### Run MulanFaaS

```
$ make run
```

#### Use the UI Portal

You can now test out the OpenFaaS UI by going to http://127.0.0.1/ui/ , username is `user`, password is `mulanfaas` - if you're deploying to a Linux VM then replace 127.0.0.1 with the IP address from the output you see on the `ifconfig` command.

> Note that we are using `127.0.0.1` instead of `localhost`, which may hang on some Linux distributions due to conflicts between IPv4/IPv6 networking.

#### Test via curl

```bash
$ curl -u user:mulanfaas -X POST http://localhost/function/func_echoit -d "hello MulanFaaS"
hello OpenFaaS
$ curl -X POST http://localhost/function/func_echoit -d "hello MulanFaaS"
401 Unauthorized
```

#### Monitoring dashboard
OpenFaaS tracks metrics on your functions automatically using Prometheus. The metrics can be turned into a useful dashboard with free and Open Source tools like [Grafana](https://grafana.com).


Open Grafana in your browser, login with username `admin` password `admin` and navigate to the pre-made OpenFaaS dashboard at:

[http://127.0.0.1:3000/dashboard/db/openfaas](http://127.0.0.1:3000/dashboard/db/openfaas)
