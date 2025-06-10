// file: internal/adapters/input/udp_listener/listener.go
package udp

import (
	"context"
	"encoding/json"
	"log"
	"net"

	"github.com/shanth1/loggate/internal/core/domain"
	"github.com/shanth1/loggate/internal/core/ports"
)

type Listener struct {
	conn     *net.UDPConn
	ingester ports.LogIngester
}

func New(address string, ingester ports.LogIngester) (*Listener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	return &Listener{conn: conn, ingester: ingester}, nil
}

func (l *Listener) Start(ctx context.Context) {
	log.Printf("INFO: UDP listener started on %s", l.conn.LocalAddr())
	defer l.conn.Close()

	buffer := make([]byte, 65535)

	for {
		select {
		case <-ctx.Done():
			log.Println("INFO: Shutting down UDP listener...")
			return
		default:
			n, _, err := l.conn.ReadFromUDP(buffer)
			if err != nil {
				log.Printf("ERROR: reading from UDP: %v", err)
				continue
			}

			var msg domain.LogMessage
			if err := json.Unmarshal(buffer[:n], &msg); err != nil {
				log.Printf("WARN: failed to unmarshal log: %v. Raw: %s", err, string(buffer[:n]))
				continue
			}

			l.ingester.Ingest(ctx, msg)
		}
	}
}
