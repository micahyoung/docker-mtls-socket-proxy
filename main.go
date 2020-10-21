package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

func main() {
	log.SetFlags(log.Lshortfile)

	clientCertPath := flag.String("clientCertPath", filepath.Join(".", "certs", "cert.pem"), "server cert")
	clientKeyPath := flag.String("clientKeyPath", filepath.Join(".", "certs", "key.pem"), "server key")
	serverCertPath := flag.String("serverCertPath", filepath.Join(".", "certs", "server-cert.pem"), "server cert")
	serverKeyPath := flag.String("serverKeyPath", filepath.Join(".", "certs", "server-key.pem"), "server key")
	caPath := flag.String("caPath", filepath.Join(".", "certs", "ca.pem"), "ca cert")
	caKeyPath := flag.String("caKeyPath", filepath.Join(".", "certs", "ca-key.pem"), "ca key")
	hostname := flag.String("hostname", "localhost", "hostname for generated cert")
	listenAddr := flag.String("listenAddr", ":2376", "server key")
	flag.Parse()

	if err := run(*clientCertPath, *clientKeyPath, *serverCertPath, *serverKeyPath, *caPath, *caKeyPath, *hostname, *listenAddr); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run(clientCertPath, clientKeyPath, serverCertPath, serverKeyPath, caPath, caKeyPath, hostname, listenAddr string) error {
	if _, err := os.Stat(filepath.Dir(serverCertPath)); os.IsNotExist(err) {
		return fmt.Errorf("refusing to run without dir %s", filepath.Dir(serverCertPath))
	}

	if err := generateCerts(clientCertPath, clientKeyPath, serverCertPath, serverKeyPath, caPath, caKeyPath, hostname); err != nil {
		return errors.Wrap(err, "generating certs")
	}

	instructions, err := clientCertCommand(hostname, clientCertPath, clientKeyPath, caPath)
	fmt.Println(instructions)

	serverCert, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
	if err != nil {
		return errors.Wrap(err, "loading server cert")
	}

	// Create a CA certificate pool and add cert.pem to it
	caCert, err := ioutil.ReadFile(caPath)
	if err != nil {
		return errors.Wrap(err, "reading ca cert")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create the TLS Config with the CA pool and enable Client certificate validation
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	// Create TCP listener with TLS config
	listener, err := tls.Listen("tcp", listenAddr, tlsConfig)
	if err != nil {
		return errors.Wrap(err, "listening")
	}
	defer listener.Close()

	log.Printf("Listening: %v\n\n", listenAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err) //non-fatal error
			continue
		}
		go proxyConn(conn)
	}
}

// ref: https://medium.com/buildpacks/pack-with-a-remote-docker-daemon-41aab804b839
func generateCerts(clientCertPath, clientKeyPath, serverCertPath, serverKeyPath, caPath, caKeyPath, hostname string) error {
	// Certificate Authority cert
	if _, err := os.Stat(caPath); os.IsNotExist(err) {
		log.Printf("generating ca cert %s %s\n", caPath, caKeyPath)
		cmd := exec.Command("openssl", "req", "-new", "-x509", "-days", "365", "-sha256", "-newkey", "rsa:4096", "-nodes", "-subj", fmt.Sprintf("/C=ZZ/ST=ZZ/L=ZZ/O=ZZ/CN=%s", hostname), "-out", caPath, "-keyout", caKeyPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)

			return err
		}
		log.Printf("%s\n", out)
	}

	if _, err := os.Stat(serverCertPath); os.IsNotExist(err) {
		log.Printf("generating server cert %s %s\n", serverCertPath, serverKeyPath)
		cmd := exec.Command("openssl", "req", "-new", "-newkey", "rsa:4096", "-nodes", "-subj", fmt.Sprintf("/CN=%s", hostname), "-out", "server.csr", "-keyout", serverKeyPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)

			return err
		}
		log.Printf("%s\n", out)

		if err := ioutil.WriteFile("extfile.cnf", []byte(fmt.Sprintf("subjectAltName = DNS:%s,IP:%s\nextendedKeyUsage = serverAuth\n", hostname, hostname)), 0666); err != nil {
			return err
		}

		cmd = exec.Command("openssl", "x509", "-req", "-days", "365", "-sha256", "-extfile", "extfile.cnf", "-in", "server.csr", "-CA", caPath, "-CAkey", caKeyPath, "-CAcreateserial", "-out", serverCertPath)
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)

			return err
		}
		log.Printf("%s\n", out)
	}

	if _, err := os.Stat(clientCertPath); os.IsNotExist(err) {
		log.Printf("generating client cert %s %s\n", clientCertPath, clientKeyPath)
		cmd := exec.Command("openssl", "req", "-subj", "/CN=client", "-new", "-newkey", "rsa:4096", "-nodes", "-out", "client.csr", "-keyout", clientKeyPath)

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)

			return err
		}
		log.Printf("%s\n", out)

		if err := ioutil.WriteFile("extfile-client.cnf", []byte("extendedKeyUsage = clientAuth\n"), 0666); err != nil {
			return err
		}

		cmd = exec.Command("openssl", "x509", "-req", "-days", "365", "-sha256", "-extfile", "extfile-client.cnf", "-in", "client.csr", "-CA", caPath, "-CAkey", caKeyPath, "-CAcreateserial", "-out", clientCertPath)
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)

			return err
		}
		log.Printf("%s\n", out)
	}

	os.Remove("client.csr")
	os.Remove("server.csr")
	os.Remove("extfile.cnf")
	os.Remove("extfile-client.cnf")

	return nil
}

func clientCertCommand(hostname, clientCertPath, clientKeyPath, caPath string) (string, error) {
	var args []interface{}
	args = append(args, hostname)

	dataFiles, err := fileContents(clientCertPath, clientKeyPath, caPath)
	if err != nil {
		return "", err
	}

	data := struct{ Hostname, CertData, KeyData, CAData interface{} }{
		hostname, dataFiles[0], dataFiles[1], dataFiles[2]}

	t := template.Must(template.New("tmpl").Parse(`
##### BEGIN COPY #####
set -o errexit

mkdir -p ~/.docker/{{.Hostname}}

cat > ~/.docker/{{.Hostname}}/cert.pem <<EOF
{{.CertData}}
EOF

cat > ~/.docker/{{.Hostname}}/key.pem <<EOF
{{.KeyData}}
EOF

cat > ~/.docker/{{.Hostname}}/ca.pem <<EOF
{{.CAData}}
EOF

cat > ~/.docker/{{.Hostname}}/env.sh <<EOF
export DOCKER_HOST=tcp://{{.Hostname}}:2376
export DOCKER_CERT_PATH=~/.docker/{{.Hostname}}
export DOCKER_TLS_VERIFY=1
EOF
##### END COPY #####
`))

	envContent := &bytes.Buffer{}
	if err := t.Execute(envContent, data); err != nil {
		return "", nil
	}

	return envContent.String(), nil
}

func fileContents(paths ...string) ([]interface{}, error) {
	var contents []interface{}

	for _, path := range paths {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		contents = append(contents, string(content))
	}

	return contents, nil
}

func proxyConn(conn net.Conn) {
	defer conn.Close()

	rConn, err := dialDockerSocket()
	if err != nil {
		log.Print(err)
	}
	defer rConn.Close()

	Pipe(conn, rConn)

	log.Printf("handleConnection end: %s\n", conn.RemoteAddr())
}

func chanFromConn(conn net.Conn) chan []byte {
	c := make(chan []byte)

	go func() {
		b := make([]byte, 1024)

		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
				// Copy the buffer so it doesn't get changed while read by the recipient.
				copy(res, b[:n])
				c <- res
			}
			if err != nil {
				c <- nil
				break
			}
		}
	}()

	return c
}

func Pipe(conn1 net.Conn, conn2 net.Conn) {
	chan1 := chanFromConn(conn1)
	chan2 := chanFromConn(conn2)

	for {
		select {
		case b1 := <-chan1:
			if b1 == nil {
				return
			} else {
				conn2.Write(b1)
			}
		case b2 := <-chan2:
			if b2 == nil {
				return
			} else {
				conn1.Write(b2)
			}
		}
	}
}
