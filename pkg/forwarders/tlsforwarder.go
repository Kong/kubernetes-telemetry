package forwarders

import (
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

func (tf *tlsForwarder) Name() string {
	return "TLSForwarder"
}

func (tf *tlsForwarder) Forward(payload []byte) error {
	err := tf.conn.SetDeadline(time.Now().Add(defaultDeadline))
	if err != nil {
		return fmt.Errorf("failed to set report connection deadline: %w", err)
	}

	_, err = tf.conn.Write(payload)
	if err != nil {
		return fmt.Errorf("failed to send report: %w", err)
	}
	return nil
}
