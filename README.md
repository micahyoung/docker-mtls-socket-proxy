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

### Start your container
This will start the process and generate self-signed certificates (or re-use previously generated ones). 

#### Linux/MacOS
```bash
hostname=<"my-docker-host" or "10.1.2.3">

docker run --detach \
    --name tlsproxy \
    --publish 2376:2376 \
    --volume tlsproxy-certs:/certs \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    --restart=always \
    docker-mtls-socket-proxy \
        -hostname $hostname
```

# Windows
```powershell
$hostname=<"my-docker-host" or "10.1.2.3">

docker run --detach `
    --name tlsproxy `
    --publish 2376:2376 `
    --volume tlsproxy-certs:c:/certs `
    --volume \\.\pipe\docker_engine:\\.\pipe\docker_engine `
    --restart=always `
    docker-mtls-socket-proxy `
        -hostname $hostname
```

### Copy client credentials to your client

* On Docker host, print the logs from the container
```
docker logs tlsproxy
```

* On Docker host, copy/paste all output between `BEGIN COPY` and `END COPY`

```
##### BEGIN COPY #####
<highlight and copy these 127 lines>
##### END COPY #####
```

* On client, execute clipboard

Syntax check and execute
```bash
# MacOS 
pbpaste | wc -l  # should be ~130 depending on random key lengths
pbpaste | bash

# Linux
xclip -o -selection clipboard | wc -l
xclip -o -selection clipboard | bash
```

  * Note: You can also copy paste data just each certs/key from the logs if preferred
    * ~/.docker/[hostname]/cert.pem
    * ~/.docker/[hostname]/key.pem
    * ~/.docker/[hostname]/ca.pem

    
