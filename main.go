package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(log.Lshortfile)

	certPath := flag.String("certPath", filepath.Join("certs", "server-cert.pem"), "server cert")
	keyPath := flag.String("keyPath", filepath.Join("certs", "server-key.pem"), "server key")
	caPath := flag.String("caPath", filepath.Join("certs", "ca.pem"), "ca")
	listenAddr := flag.String("listenAddr", ":2376", "server key")
	flag.Parse()

	if err := run(*certPath, *keyPath, *caPath, *listenAddr); err != nil {
		log.Print(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run(certPath, keyPath, caPath, listenAddr string) error {
	serverCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return err
	}

	// Create a CA certificate pool and add cert.pem to it
	caCert, err := ioutil.ReadFile(caPath)
	if err != nil {
		return err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create the TLS Config with the CA pool and enable Client certificate validation
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	// Create TCP listener with TLS config
	listener, err := tls.Listen("tcp", listenAddr, tlsConfig)
	if err != nil {
		return err
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
