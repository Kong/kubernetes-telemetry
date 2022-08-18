package forwarders

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"

	"github.com/kong/kubernetes-telemetry/pkg/log"
)

const (
	defaultTimeout  = time.Second * 30
	defaultDeadline = time.Minute
)

var tlsConf = tls.Config{
	MinVersion: tls.VersionTLS13,
	MaxVersion: tls.VersionTLS13,
}

type tlsForwarder struct {
	logger logr.Logger
	conn   *tls.Conn
}

// NewTLSForwarder creates a TLS forwarder which forwards received serialized reports
// to a TLS endpoint specified by the provided address.
func NewTLSForwarder(address string, logger logr.Logger) *tlsForwarder {
	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout: defaultTimeout,
		},
		"tcp",
		address,
		&tlsConf,
	)
	if err != nil {
		logger.V(log.DebugLevel).Info("failed to connect to reporting server", "error", err)
		return nil
	}

	return &tlsForwarder{
		logger: logger,
		conn:   conn,
	}
}

// Name returns the name of the forwarder.
func (tf *tlsForwarder) Name() string {
	return "TLSForwarder"
}

// Forward forwards the received payload to the configured TLS endpoint.
func (tf *tlsForwarder) Forward(ctx context.Context, payload []byte) error {
	var deadline time.Time
	if d, ok := ctx.Deadline(); ok {
		deadline = d
	} else {
		deadline = time.Now().Add(defaultDeadline)
	}

	err := tf.conn.SetDeadline(deadline)
	if err != nil {
		return fmt.Errorf("failed to set report connection deadline: %w", err)
	}

	_, err = tf.conn.Write(payload)
	if err != nil {
		return fmt.Errorf("failed to send report: %w", err)
	}
	return nil
}
