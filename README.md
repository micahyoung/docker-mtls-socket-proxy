# Container image for Mutual TLS access to Docker Daemon
A secure, easier-to-use alternative to enabling Docker daemon's [built-in mutual TLS settings](https://docs.docker.com/engine/security/https) that works on default daemon installations.

Mutual TLS connections from docker clients are securely terminated at a container then proxied to a docker daemon via a bind-mounted docker socket. 

<svg width="800" height="600" xmlns="http://www.w3.org/2000/svg" xmlns:svg="http://www.w3.org/2000/svg">
 <!-- Created with Method Draw - http://github.com/duopixel/Method-Draw/ -->
 <title>Docker mTLS Socket Proxy</title>

 <g>
  <title>background</title>
  <rect fill="#fff" id="canvas_background" height="602" width="802" y="-1" x="-1"/>
  <g display="none" id="canvasGrid">
   <rect fill="url(#gridpattern)" stroke-width="0" y="0" x="0" height="100%" width="100%" id="svg_2"/>
  </g>
 </g>
 <g>
  <title>Layer 1</title>
  <rect stroke="#000" id="svg_4" height="478.999993" width="348.99999" y="42" x="368.5" stroke-width="1.5" fill="#fff"/>
  <rect stroke="#000" id="svg_5" height="178.999981" width="252" y="119.000005" x="380.5" stroke-width="1.5" fill="#fff"/>
  <rect stroke="#000" id="svg_3" height="176.000001" width="156.000003" y="139" x="11.499997" stroke-width="1.5" fill="#fff"/>
  <text fill="#000000" stroke-width="0" x="61.177312" y="160.034936" id="svg_7" font-size="24" font-family="Helvetica, Arial, sans-serif" text-anchor="start" xml:space="preserve" transform="matrix(0.8785678744316101,0,0,0.9850172083184621,8.803829103708267,37.49226214475733) " stroke="#000">Client</text>
  <text fill="#000000" stroke-width="0" x="31.170702" y="149.721298" id="svg_9" font-size="24" font-family="Helvetica, Arial, sans-serif" text-anchor="start" xml:space="preserve" transform="matrix(0.8785678744316101,0,0,1.1151290472049373,8.803829103708267,0.7960808425827112) " stroke="#000">Docker API</text>
  <rect stroke="#000" id="svg_1" height="81.999993" width="147" y="395" x="395.5" stroke-width="1.5" fill="#fff"/>
  <text id="svg_6" fill="#000000" stroke-width="0" x="481.544932" y="402.640248" font-size="24" font-family="Helvetica, Arial, sans-serif" text-anchor="start" xml:space="preserve" transform="matrix(0.8785678744316101,0,0,1.1151290472049373,8.803829103708267,0.7960808425827112) " stroke="#000">daemon</text>
  <text id="svg_8" fill="#000000" stroke-width="0" x="486.889883" y="379.862226" font-size="24" font-family="Helvetica, Arial, sans-serif" text-anchor="start" xml:space="preserve" transform="matrix(0.8785678744316101,0,0,1.1151290472049373,8.803829103708267,0.7960808425827112) " stroke="#000">Docker</text>
  <line stroke="#000" id="svg_12" y2="271" x2="377.517797" y1="271" x1="167" stroke-linecap="null" stroke-linejoin="null" stroke-dasharray="null" stroke-width="2" fill="none"/>
  <line stroke="#000" transform="rotate(-90 454.75891113281256,353) " id="svg_13" y2="353" x2="505.517815" y1="353" x1="404" stroke-linecap="null" stroke-linejoin="null" stroke-dasharray="null" stroke-width="2" fill="none"/>
  <text id="svg_17" fill="#000000" stroke-width="0" x="467.524251" y="352.637864" font-size="24" font-family="Helvetica, Arial, sans-serif" text-anchor="start" xml:space="preserve" stroke="#000">Socket/Named Pipe</text>
  <text id="svg_18" fill="#000000" stroke-width="0" x="567.703218" y="310.811921" font-size="24" font-family="Monospace" text-anchor="start" xml:space="preserve" transform="matrix(0.647950416074142,0,0,0.8224160604327283,98.72491743921222,117.31412103977554) " stroke="#000">/var/run/docker.sock</text>
  <text id="svg_19" fill="#000000" stroke-width="0" x="200.059455" y="208.581601" font-size="24" font-family="Helvetica, Arial, sans-serif" text-anchor="start" xml:space="preserve" transform="matrix(0.8785678744316101,0,0,1.1151290472049373,8.803829103708267,0.7960808425827112) " stroke="#000">Mutual TLS</text>
  <text id="svg_20" fill="#000000" stroke-width="0" x="567.703218" y="310.811921" font-size="24" font-family="Monospace" text-anchor="start" xml:space="preserve" transform="matrix(0.647950416074142,0,0,0.8224160604327281,-183.2750825607879,-0.6858789602244697) " stroke="#000">tcp://1.2.3.4:2376</text>
  <text id="svg_21" fill="#000000" stroke-width="0" x="434.531948" y="132.357239" font-size="24" font-family="Helvetica, Arial, sans-serif" text-anchor="start" xml:space="preserve" transform="matrix(0.8785678744316101,0,0,1.1151290472049373,8.803829103708267,0.7960808425827112) " stroke="#000">mTLS Proxy Container</text>
  <text id="svg_23" fill="#000000" stroke-width="0" x="68.076219" y="251.170951" font-size="24" font-family="Helvetica, Arial, sans-serif" text-anchor="start" xml:space="preserve" transform="matrix(0.6382691699593116,0,0,0.8101280641720278,8.914760102167799,68.76722566594626) " stroke="#000">Self-signed</text>
  <text id="svg_24" fill="#000000" stroke-width="0" x="68.082835" y="275.858406" font-size="24" font-family="Helvetica, Arial, sans-serif" text-anchor="start" xml:space="preserve" transform="matrix(0.6382691699593116,0,0,0.8101280641720278,8.914760102167799,68.76722566594626) " stroke="#000">client certs</text>
  <text id="svg_25" fill="#000000" stroke-width="0" x="613.300736" y="228.952241" font-size="24" font-family="Helvetica, Arial, sans-serif" text-anchor="start" xml:space="preserve" transform="matrix(0.6382691699593116,0,0,0.8101280641720278,8.914760102167799,68.76722566594626) " stroke="#000">Self-signed server certs</text>
  <text id="svg_27" fill="#000000" stroke-width="0" x="567.703218" y="319.689602" font-size="24" font-family="Monospace" text-anchor="start" xml:space="preserve" transform="matrix(0.5324783097654483,0,0,0.6758521993650036,103.7441584058656,54.163922754366226) " stroke="#000">SAN IP: 1.2.3.4</text>
  <text id="svg_28" fill="#000000" stroke-width="0" x="418.596924" y="62.410177" font-size="24" font-family="Helvetica, Arial, sans-serif" text-anchor="start" xml:space="preserve" transform="matrix(0.8785678744316101,0,0,1.1151290472049373,8.803829103708267,0.7960808425827112) " stroke="#000">Host Machine</text>
  <text id="svg_29" fill="#000000" stroke-width="0" x="520.752949" y="56.318409" font-size="24" font-family="Monospace" text-anchor="start" xml:space="preserve" transform="matrix(0.5324783097654483,0,0,0.6758521993650036,103.7441584058656,54.163922754366226) " stroke="#000">IP Address: 1.2.3.4</text>
  <text style="cursor: move;" id="svg_30" fill="#000000" stroke-width="0" x="543.289078" y="171.728258" font-size="24" font-family="Monospace" text-anchor="start" xml:space="preserve" transform="matrix(0.5324783097654483,0,0,0.6758521993650036,103.7441584058656,54.163922754366226) " stroke="#000">Published Port: 2376</text>
 </g>
</svg>

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

* On client, Syntax check and execute clipboard contents
    
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

    
