package forwarders

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/bombsimon/logrusr/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
	"github.com/kong/kubernetes-telemetry/pkg/serializers"
	"github.com/kong/kubernetes-telemetry/pkg/telemetry"
	"github.com/kong/kubernetes-telemetry/pkg/types"
)

func TestTLSForwarder(t *testing.T) {
	log := logrusr.New(logrus.New())
	telemetryServer := newTelemetryTestServer(t, "localhost:0")

	// This is the time limit for the whole test.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	receivedChan := telemetryServer.Run(ctx, t)

	telemetryServerAddr := telemetryServer.Addr()
	tf, err := NewTLSForwarder(
		telemetryServerAddr,
		log,
		func(c *tls.Config) {
			c.InsecureSkipVerify = true
		},
	)
	require.NoError(t, err)

	serializer := serializers.NewSemicolonDelimited()
	require.NotNil(t, serializer)

	consumer := telemetry.NewConsumer(serializer, tf)

	m, err := telemetry.NewManager(
		"test-ping",
		telemetry.OptManagerPeriod(time.Hour),
		telemetry.OptManagerLogger(log),
	)
	require.NoError(t, err)

	w := telemetry.NewWorkflow("test1")
	p, err := provider.NewFixedValueProvider("test1-provider", types.ProviderReport{
		"key": "value",
	},
	)
	require.NoError(t, err)
	w.AddProvider(p)
	m.AddWorkflow(w)
	require.NoError(t, m.AddConsumer(consumer))
	require.NoError(t, m.Start())

	require.NoError(t, m.TriggerExecute(ctx, "test-signal"))
	assertData(ctx, t, receivedChan, []string{
		"<14>signal=test-signal;key=value;\n",
	})
	require.NoError(t, m.TriggerExecute(ctx, "test-signal-2"))
	assertData(ctx, t, receivedChan, []string{
		"<14>signal=test-signal-2;key=value;\n",
	})
	require.NoError(t, telemetryServer.Close())

	t.Log("Recreate server with the same address to simulate unreliable network/server")
	telemetryServer = newTelemetryTestServer(t, telemetryServerAddr)
	receivedChan = telemetryServer.Run(ctx, t)

	require.NoError(t, m.TriggerExecute(ctx, "test-signal-3"))
	assertData(ctx, t, receivedChan, []string{
		"<14>signal=test-signal-3;key=value;\n",
	})
	require.NoError(t, m.TriggerExecute(ctx, "test-signal-4"))
	assertData(ctx, t, receivedChan, []string{
		"<14>signal=test-signal-4;key=value;\n",
	})
	require.NoError(t, telemetryServer.Close())

	m.Stop()
}

type telemetryServer struct {
	listener net.Listener
}

func newTelemetryTestServer(t *testing.T, addr string) telemetryServer {
	t.Helper()
	const (
		rootPEM = `-----BEGIN CERTIFICATE-----
MIIDzDCCArSgAwIBAgIJAP5AVMhOiD+WMA0GCSqGSIb3DQEBCwUAMHExCzAJBgNV
BAYTAlVTMRMwEQYDVQQIDApDYWxpZm9ybmlhMRYwFAYDVQQHDA1TYW4gRnJhbmNp
c2NvMQ0wCwYDVQQKDARLb25nMRgwFgYDVQQLDA9LdWJlcm5ldGVzIHRlYW0xDDAK
BgNVBAMMA0tHTzAeFw0yMjA4MjMxMzA3NTJaFw0zMjA4MjAxMzA3NTJaMHExCzAJ
BgNVBAYTAlVTMRMwEQYDVQQIDApDYWxpZm9ybmlhMRYwFAYDVQQHDA1TYW4gRnJh
bmNpc2NvMQ0wCwYDVQQKDARLb25nMRgwFgYDVQQLDA9LdWJlcm5ldGVzIHRlYW0x
DDAKBgNVBAMMA0tHTzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMmb
QFXQoXIC56ZLMDL+qZAHfd92tBs0pxJv3ZKalIKInVy8YwnDmEINLkemsTlPgK32
NxI1UvW1nTo42CDp3GXE3dmKOxv5C1rWvR0EbnYyDK3idX4sS3nT3jqRJiSP0iJq
+r3LJjUg+LUVL+v0Yw41i1yhdEFyelTAKulykt9AVCDYA65j9cAhOjmFJjHTUR/J
/UveNARa/x9fu7rxfCBpMk+yPcXHBvJYm+GP4RF6izDRkJF+WBC9iPhri+feILk9
8W0VW77TC4tLBGUNBSy7JyRg6lK0VpCAZjPUG/mOnztn9E4YoVVm/qOoVLizGmOS
MGFAHoX68IOSLJkQHkkCAwEAAaNnMGUwCwYDVR0PBAQDAgQwMBMGA1UdJQQMMAoG
CCsGAQUFBwMBMEEGA1UdEQQ6MDiCCWxvY2FsaG9zdIIJMTI3LjAuMC4xgiB0ZXN0
LmdhdGV3YXktb3BlcmF0b3Iua29uZ2hxLmNvbTANBgkqhkiG9w0BAQsFAAOCAQEA
pseJcAOULR0NBziL9nvKjSSyhhUZCrj12QwXX/dFVUCu6RNGcvGTmv5tLh8W3CIs
vJMHKZbZ1+afRM+nHwMfKqqpkJU0TJaKBbf+Ky+Zrk9xu8TVO0cprB6RMUhQrdB9
10L6y6I0qseNMSWg/1wtcE3tiTFC7/Zc+ywK2wU8lwKfSgFBOxa+jgfMJQoYfCTs
Q5jdmdhJTzXrY8YscOFvqe6zH4fpH58HiG0v9Q5/YMb5kTwQsoAIR78TJN+6Klg8
kKklxFcttse5LppQ5mwfRFXFl3/UErOq3YED5cVgraCj0H2xe1kTxIPQR4ex/DgF
/zyvypUrraeZJTYrFthkgQ==
-----END CERTIFICATE-----`
		keyPEM = `-----BEGIN PRIVATE KEY-----
MIIEugIBADANBgkqhkiG9w0BAQEFAASCBKQwggSgAgEAAoIBAQDJm0BV0KFyAuem
SzAy/qmQB33fdrQbNKcSb92SmpSCiJ1cvGMJw5hCDS5HprE5T4Ct9jcSNVL1tZ06
ONgg6dxlxN3Zijsb+Qta1r0dBG52Mgyt4nV+LEt50946kSYkj9Iiavq9yyY1IPi1
FS/r9GMONYtcoXRBcnpUwCrpcpLfQFQg2AOuY/XAITo5hSYx01Efyf1L3jQEWv8f
X7u68XwgaTJPsj3FxwbyWJvhj+EReosw0ZCRflgQvYj4a4vn3iC5PfFtFVu+0wuL
SwRlDQUsuyckYOpStFaQgGYz1Bv5jp87Z/ROGKFVZv6jqFS4sxpjkjBhQB6F+vCD
kiyZEB5JAgMBAAECggEAS1F5A5Zh+lojePj2FNcXOfvShr2uI8vT7wtj1/VwLiQj
xhWLWoZ8R5DtDU+1Phf5lwQ5JtBNIgarqqi59fHoqQyXZUJDOvwbxeAb3s9dBUNF
gWDtTCn4OJdymqbHfTlN5BXbfzR6Hbcns18q/BfdOd2/Jugaqqi+ExOH9JcdT9Hq
2M1E4j+9OYQ4IreQsPzQzxmMAJnkDnDhXlG4Xv7RwCT8v1Uytkk965BD38oqp5p8
D2NrQZgY0nUK4oFiUEZV6kz1y4YfkNuEUoQ4porrk2m9nyCqbHVGINGVTlvmlfDC
qZt0Mt4+iDJwk+KKinKSJFrUYuCzZsq/bNiYYCPq0QKBgQDmPKtOGehjOneLIc/Z
ObovG15ftA2Ln1VGdNnoxAns0MUm+PI+w1H8Ox+8+yZAtBjKiQmUZ1zNdYsMG/LZ
T+O4HI3BudBMaAQbs+ve1LMZjwLAqejJQaPf6NxMSmi1D+wi2GWUR1VCTktXeMU3
VCb+QUQvTfd1tBgKkIlJXuQMNwKBgQDgKm+7++ew32Z/omWs5+WWfvAUQ9BxGdB7
fLGrJCptZz4HMezZ4D8LwLwo4oBvtLJkKiZPgfdxJqh/tb2HtR9GGesyqic1KG+c
AZuBoTXjb6i/YrCTBfFHNJacpt28oYnNyrVxe/fYP28KWRG47Es8ZBUd6l6NDfa9
uDsld7DpfwJ/Udc/DYQoFx2xYMOkHpNmm1gfM+XM6tS4e1MOIq+M16Fec3wKoETN
39skbQjZkCZ0qYoM3bPgSSh/RM6qhJThXZDI0xQ59u8ChtZuAceZ4nvzfojnNqMe
nXko1fWdQr9mMPy7Hvo8VFWAcpd7gy9mrPqGQkp0rGJYKWk3Y91XMwKBgBotOnEE
QJEJ9WkwKJlhVxEU76oeJSgf8JWLASBQD9hItxiV/ueOZS5VKmPH12G0AyTpOyIL
tj5zWjfXnDNNtkI0Yp+++OcfOrFICsW/cpCFiHoY5y+0APHktTXD0p7lajcq1bdT
16Rb+/aEYiprBXoe4cxlgvcLy2VqLxX3/SO3AoGAFyqu1k2B9hc+IWQYa03EWyx4
mM1T33W8gj2ADVwcLsfbK0inSk1jtGvd/ZksB7Hy9htBYOpU6a5HZpTZwm+Ek/T+
1VI1t2doPiYvTIo3Qy6JKeL6DmR73jZA41XtzrcQJ6Kfi5PFewqkIGe9Y6iSHRJX
f3cb9gYaLWdmvkx8p3g=
-----END PRIVATE KEY-----`
	)

	tlsCert, err := tls.X509KeyPair([]byte(rootPEM), []byte(keyPEM))
	require.NoError(t, err)
	listener, err := tls.Listen(
		"tcp4",
		addr,
		&tls.Config{
			Certificates:       []tls.Certificate{tlsCert},
			InsecureSkipVerify: true, //nolint:gosec
		},
	)
	require.NoError(t, err)
	t.Logf("TLS server address %s", listener.Addr().String())

	return telemetryServer{
		listener: listener,
	}
}

func (ts telemetryServer) Addr() string {
	return ts.listener.Addr().String()
}

func (ts telemetryServer) Close() error {
	return ts.listener.Close()
}

func (ts telemetryServer) Run(ctx context.Context, t *testing.T) <-chan string {
	receivedData := make(chan string)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			conn, err := ts.listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					t.Logf("server: closed, not accepting more connections")
					return
				}
				t.Logf("server: Accept() returned error: %v", err)
				return
			}
			go handleConnection(t, conn, receivedData)
		}
	}()
	return receivedData
}

func assertData(ctx context.Context, t *testing.T, receivedData <-chan string, expectedData []string) {
	for _, expected := range expectedData {
		select {
		case received := <-receivedData:
			assert.Equal(t, expected, received)
		case <-ctx.Done():
			assert.Failf(t, "timeout waiting for data", "expected data: %q", expected)
		}
	}
}

// handleConnection reads data from the connection in the loop (client or error ends the connection)
// and writes it to the channel receivedData. In case of error, it logs the error and returns.
func handleConnection(t *testing.T, conn net.Conn, receivedData chan<- string) {
	defer conn.Close()
	t.Logf("server: accepted from %s", conn.RemoteAddr())
	for {
		buf := make([]byte, 512)
		n, err := conn.Read(buf)
		if err != nil {
			t.Logf("server: conn: read: %s", err)
			return
		}

		data := string(buf[:n])
		t.Logf("server: conn: echo %q", data)
		receivedData <- data
	}
}
