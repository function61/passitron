package main

import (
	"io"
	"log"
	"net"
	"os"
)

/*	Linux / Mac: use ENV variable

	Windows: use https://github.com/131/pageantbridge
*/

// this is essentially a domain socket -> TCP bridge

const (
	sourceSocket  = "/tmp/ssh-agent.sock"
	targetTcpAddr = "127.0.0.1:8096"
)

func checkSocketExistence() {
	_, err := os.Stat(sourceSocket)
	if err == nil { // socket exists
		if err := os.Remove(sourceSocket); err != nil {
			log.Fatalf("sshagentproxy: remove error: %s", err.Error())
		}
	} else if !os.IsNotExist(err) { // some other error than not exists
		log.Fatalf("sshagentproxy: unexpected Stat() error: %s", err.Error())
	}
}

func handleOneClient(client net.Conn) {
	log.Printf("sshagentproxy: client connected")

	server, err := net.Dial("tcp", targetTcpAddr)
	if err != nil {
		log.Printf("sshagentproxy: failed to connect to endpoint: %s", err.Error())
		client.Close()
		return
	}

	writerAndReaderDone := make(chan bool, 2)

	go func() {
		if _, err := io.Copy(server, client); err != nil {
			log.Printf("sshagentproxy: client->server copy error: %s", err.Error())
		}

		log.Printf("sshagentproxy: client->server done")

		server.Close()

		writerAndReaderDone <- true
	}()

	go func() {
		if _, err := io.Copy(client, server); err != nil {
			log.Printf("sshagentproxy: server->client copy error: %s", err.Error())
		}

		log.Printf("sshagentproxy: server->client done")

		writerAndReaderDone <- true
	}()

	<-writerAndReaderDone
	<-writerAndReaderDone

	log.Printf("sshagentproxy: client disconnected")
}

func main() {
	checkSocketExistence()

	log.Printf("sshagentproxy: listening at %s", sourceSocket)
	log.Printf("sshagentproxy: pro tip $ export SSH_AUTH_SOCK=\"%s\"", sourceSocket)

	socketListener, err := net.Listen("unix", sourceSocket)
	if err != nil {
		log.Fatalf("sshagentproxy: sock listen error: %s", err.Error())
	}

	for {
		// intentionally only supporting sequential connections for now
		client, err := socketListener.Accept()
		if err != nil {
			log.Printf("sshagentproxy: Accept() error: %s", err.Error())
			continue
		}

		go handleOneClient(client)
	}
}
