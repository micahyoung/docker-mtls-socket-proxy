package main

import (
	"bytes"
	"crypto/sha1"
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
	"regexp"
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
	serverCSRPath := filepath.Join(filepath.Dir(serverCertPath), "server.csr")
	serverExtPath := filepath.Join(filepath.Dir(serverCertPath), "extfile.cnf")
	clientCSRPath := filepath.Join(filepath.Dir(clientCertPath), "client.csr")
	clientExtPath := filepath.Join(filepath.Dir(clientCertPath), "extfile-client.cnf")

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
		cmd := exec.Command("openssl", "req", "-new", "-newkey", "rsa:4096", "-nodes", "-subj", fmt.Sprintf("/CN=%s", hostname), "-out", serverCSRPath, "-keyout", serverKeyPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)

			return err
		}
		log.Printf("%s\n", out)

		ipAddress := "127.0.0.1"
		if ok, err := regexp.MatchString(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`, hostname); (err == nil) && ok {
			ipAddress = hostname
		}

		if err := ioutil.WriteFile(serverExtPath, []byte(fmt.Sprintf("subjectAltName = DNS:%s,IP:%s\nextendedKeyUsage = serverAuth\n", hostname, ipAddress)), 0666); err != nil {
			return err
		}

		cmd = exec.Command("openssl", "x509", "-req", "-days", "365", "-sha256", "-extfile", serverExtPath, "-in", serverCSRPath, "-CA", caPath, "-CAkey", caKeyPath, "-CAcreateserial", "-out", serverCertPath)
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)

			return err
		}
		log.Printf("%s\n", out)

		if err := removeAll(serverCSRPath, serverExtPath); err != nil {
			return err
		}
	}

	if _, err := os.Stat(clientCertPath); os.IsNotExist(err) {
		log.Printf("generating client cert %s %s\n", clientCertPath, clientKeyPath)
		cmd := exec.Command("openssl", "req", "-subj", "/CN=client", "-new", "-newkey", "rsa:4096", "-nodes", "-out", clientCSRPath, "-keyout", clientKeyPath)

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)

			return err
		}
		log.Printf("%s\n", out)

		if err := ioutil.WriteFile(clientExtPath, []byte("extendedKeyUsage = clientAuth\n"), 0666); err != nil {
			return err
		}

		cmd = exec.Command("openssl", "x509", "-req", "-days", "365", "-sha256", "-extfile", clientExtPath, "-in", clientCSRPath, "-CA", caPath, "-CAkey", caKeyPath, "-CAcreateserial", "-out", clientCertPath)
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)

			return err
		}
		log.Printf("%s\n", out)

		if err := removeAll(clientCSRPath, clientExtPath); err != nil {
			return err
		}
	}

	if err := allowAll(clientCertPath, clientKeyPath, serverCertPath, serverKeyPath, caPath, caKeyPath); err != nil {
		return err
	}

	return nil
}

func removeAll(paths ...string) error {
	for _, path := range paths {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

func allowAll(paths ...string) error {
	for _, path := range paths {
		if err := os.Chmod(path, 0666); err != nil {
			return err
		}
	}
	return nil
}

func clientCertCommand(hostname, clientCertPath, clientKeyPath, caPath string) (string, error) {
	var args []interface{}
	args = append(args, hostname)

	dataFiles, err := fileContents(clientCertPath, clientKeyPath, caPath)
	if err != nil {
		return "", err
	}

	scriptData := struct{ Hostname, CertData, KeyData, CAData interface{} }{
		hostname, dataFiles[0], dataFiles[1], dataFiles[2]}

	scriptTempl := template.Must(template.New("tmpl").Parse(`set -o errexit

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

echo "Load with: 'source ~/.docker/{{.Hostname}}/env.sh'"
`))

	envContent := &bytes.Buffer{}
	if err := scriptTempl.Execute(envContent, scriptData); err != nil {
		return "", nil
	}

	asciiContent := fmt.Sprintf("%+s", envContent.String())

	shasum := fmt.Sprintf("%x", sha1.Sum([]byte(asciiContent)))

	logData := struct{ ScriptContent, ShaSum interface{} }{
		envContent.String(), shasum}

	logTempl := template.Must(template.New("tmpl").Parse(`
##### COPY BELOW (including newlines) #####
{{.ScriptContent}}##### COPY ABOVE (including newlines) #####

##### INSTRUCTIONS #####
Copy to the clipboard all text and newlines between 'COPY ABOVE' and 'COPY BELOW' 

Run on MacOS with:
pbpaste | shasum  # should match {{.ShaSum}}
pbpaste | bash

or Linux:
xclip -o -selection clipboard | shasum
xclip -o -selection clipboard | bash

Expected SHA1: {{.ShaSum}}
##### END INSTRUCTIONS #####
`))

	logContent := &bytes.Buffer{}
	if err := logTempl.Execute(logContent, logData); err != nil {
		return "", nil
	}

	return logContent.String(), nil
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
