package queue

import (
	"fmt"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type Queue struct {
	conn   *nats.Conn
	server *server.Server
}

func RunEmbeddedNATS(inProcess, enableLogging bool) (*Queue, error) {
	opts := &server.Options{
		ServerName: "embedded_server",
		DontListen: inProcess,
		//JetStream:       true,
		//JetStreamDomain: "embedded",
	}

	natsServer, err := server.NewServer(opts)
	if err != nil {
		return nil, err
	}

	if enableLogging {
		natsServer.ConfigureLogger()
	}

	go natsServer.Start()
	if !natsServer.ReadyForConnections(5 * time.Second) {
		return nil, fmt.Errorf("server not ready: timeout")
	}

	var clientOpts []nats.Option
	if inProcess {
		clientOpts = append(clientOpts, nats.InProcessServer(natsServer))
	}

	natsConn, err := nats.Connect(natsServer.ClientURL(), clientOpts...)
	if err != nil {
		return nil, err
	}

	return &Queue{conn: natsConn, server: natsServer}, nil
}

func (q *Queue) Subscribe(topicName string, handler nats.MsgHandler) error {
	_, err := q.conn.Subscribe(topicName, handler)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic '%s': %v", topicName, err)
	}

	return nil
}

func (q *Queue) Publish(topicName string, data []byte) error {
	return q.conn.Publish(topicName, data)
}

func (q *Queue) Shutdown() {
	q.server.Shutdown()
}

func (q *Queue) WaitForShutdown() {
	q.server.WaitForShutdown()
}
