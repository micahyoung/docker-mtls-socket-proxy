# Container image for Mutual TLS access to Docker Daemon
A secure, easier-to-use alternative to enabling Docker daemon's [built-in mutual TLS settings](https://docs.docker.com/engine/security/https) that works on default daemon installations.

Mutual TLS connections from docker clients are securely terminated at a container then proxied to a docker daemon via a bind-mounted docker socket. 

## Requirements
* Docker daemon with default settings (Linux or Windows)
  * Note: daemon does not need to be listing on network port, only needs default socket listener
* Hostname for your docker daemon host's IP (either DNS name or /etc/host entry)
  * Note: `host.docker.internal` may already be present on [Docker Desktop](https://docs.docker.com/docker-for-mac/networking/#use-cases-and-workarounds) 

## Usage

### Build the image locally on a Docker daemon
#### Linux/MacOS
Note: no certs are included in the image. It will not be runnable without the final steps below.

```bash
docker build --tag docker-mtls-socket-proxy -f Dockerfile.linux .
```
  
#### Windows
```powershell
docker build --tag docker-mtls-socket-proxy -f Dockerfile.windows --build-arg os_tag=1809 .
```

### Generate a cert chain from scratch:
This will write (or re-use) all client, server, ca certs to your local `~/.docker/my-docker-host` dir

#### Linux/MacOS
```bash
docker run --detach \
    --name tlsproxy \
    --publish 2376:2376 \
    --volume $HOME/.docker/my-docker-host:/certs \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    --restart=always \
    docker-mtls-socket-proxy \
        -hostname my-docker-host \
        -ipAddress 10.1.2.3
```

# Windows
```powershell
docker run --detach `
    --name tlsproxy `
    --publish 2376:2376 `
    --volume $env:USERPROFILE\.docker\my-docker-host:c:/certs `
    --volume \\.\pipe\docker_engine:\\.\pipe\docker_engine `
    --restart=always `
    docker-mtls-socket-proxy `
        -hostname my-docker-host `
        -ipAddress 10.1.2.3
```

### Copy client credentials to your remote clients

```bash
mkdir ~/.docker/my-docker-host
ssh <my-docker-host or 10.1.2.3> tar -c -f- -C .docker cert.pem key.pem ca.pem | tar -x -f- -C ~/.docker/my-docker-host

export DOCKER_HOST=tcp://<my-docker-host or 10.1.2.3>:2376
export DOCKER_TLS_VERIFY=1
export DOCKER_CERT_PATH=~/.docker/my-docker-host

docker info
```
