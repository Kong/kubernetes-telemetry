package forwarders

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
)

const (
	defaultTimeout  = time.Second * 30
	defaultDeadline = time.Minute
)

var defaultTLSConf = tls.Config{
	MinVersion: tls.VersionTLS13,
	MaxVersion: tls.VersionTLS13,
}

type tlsForwarder struct {
	logger logr.Logger

	tlsConf *tls.Config
	address string
}

// TLSOpt defines an option type that manipulates *tls.Config.
type TLSOpt func(*tls.Config)

// NewTLSForwarder creates a TLS forwarder which forwards received serialized reports
// to a TLS endpoint specified by the provided address.
func NewTLSForwarder(address string, logger logr.Logger, tlsOpts ...TLSOpt) (*tlsForwarder, error) {
	tlsConf := defaultTLSConf.Clone()
	for _, opt := range tlsOpts {
		opt(tlsConf)
	}

	return &tlsForwarder{
		logger:  logger,
		tlsConf: tlsConf,
		address: address,
	}, nil
}

// Name returns the name of the forwarder.
func (tf *tlsForwarder) Name() string {
	return "TLSForwarder"
}

// Forward forwards the received payload to the configured TLS endpoint.
func (tf *tlsForwarder) Forward(ctx context.Context, payload []byte) (err error) {
	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout: defaultTimeout,
		},
		"tcp",
		tf.address,
		tf.tlsConf,
	)
	if err != nil {
		return fmt.Errorf("failed to connect to reporting server: %w", err)
	}
	// Set named return value in defer to not swallow the error returned by Close()
	// and use errors.Join() to preserve any previously returned error. Go assigns
	// explicitly returned value to named value before executing the deferred function.
	defer func() {
		err = errors.Join(err, conn.Close())
	}()

	var deadline time.Time
	if d, ok := ctx.Deadline(); ok {
		deadline = d
	} else {
		deadline = time.Now().Add(defaultDeadline)
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("failed to set report connection deadline: %w", err)
	}

	if _, err := conn.Write(payload); err != nil {
		return fmt.Errorf("failed to send report: %w", err)
	}
	return nil
}
