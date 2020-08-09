# Container image for mutual TLS access to to Docker Daemon
A secure, easier-to-use alternative to docker daemon's built-in mutual TLS settings. Mutual TLS from a client is terminated at the container then proxied to a docker daemon via a bind-mounted docker socket. 

## Requirements
* Docker daemon with accessible IP
* Hostname for your docker daemon IP (either DNS name or /etc/host entry)
  * Note: `host.docker.internal` may already be present on [Docker Desktop](https://docs.docker.com/docker-for-mac/networking/#use-cases-and-workarounds) 

## Usage
* Generate `ca`, `server`, and `client` certs (official [instructions](https://docs.docker.com/engine/security/https/#create-a-ca-server-and-client-keys-with-openssl))
    * Self-signed example
    ```bash
    HOST=host.docker.internal
  
    mkdir -p certs
    pushd certs
    
    # CA
    openssl req -new -x509 -days 365 -sha256 -subj "/C=ZZ/ST=ZZ/L=ZZ/O=ZZ/CN=$HOST" -out ca.pem -keyout ca-key.pem -newkey rsa:4096 -nodes
    
    # Server
    openssl req -subj "/CN=$HOST" -sha256 -new -out server.csr -keyout server-key.pem -newkey rsa:4096 -nodes
    echo subjectAltName = DNS:$HOST > extfile.cnf
    echo extendedKeyUsage = serverAuth >> extfile.cnf
    openssl x509 -req -days 365 -sha256 -in server.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile extfile.cnf
    
    # Client
    openssl req -subj '/CN=client' -new -out client.csr -keyout key.pem -newkey rsa:4096 -nodes
    echo extendedKeyUsage = clientAuth > extfile-client.cnf
    openssl x509 -req -days 365 -sha256 -in client.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out cert.pem -extfile extfile-client.cnf
    
    rm -v client.csr server.csr extfile.cnf extfile-client.cnf
    
    popd
    ```

* Build the image on a local Docker daemon
    ```bash
    # Linux
    docker build --tag tlssocketproxy -f Dockerfile.linux .

    # Windows
    docker build --tag tlssocketproxy -f Dockerfile.windows .
    ```

* Run the image on the to-be-secured Docker daemon
    ```bash
    # Linux
    docker create --name tlssocketproxy-ctr --restart unless-stopped --volume '/var/run/docker.sock:/var/run/docker.sock' --user root -p 23760:2376 tlssocketproxy
    docker cp certs tlssocketproxy-ctr:/certs
    docker start tlssocketproxy-ctr

    # Windows
    docker create --name tlssocketproxy-ctr --restart unless-stopped --volume '\\.\pipe\docker_engine:\\.\pipe\docker_engine' --user ContainerAdministrator -p 23760:2376 tlssocketproxy
    docker cp certs tlssocketproxy-ctr:/certs
    docker start tlssocketproxy-ctr
    ```

```
docker --host tcp://host.docker.internal:23760 --tlsverify --tlscacert certs/ca.pem --tlscert certs/cert.pem --tlskey certs/key.pem ps
```
