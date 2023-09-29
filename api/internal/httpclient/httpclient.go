package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

const defaultClientTimeout = 30 * time.Second

func New() *http.Client {
	return &http.Client{
		Timeout:   defaultClientTimeout,
		Transport: http.DefaultTransport,
	}
}

type ClientOptions struct {
	Timeout            time.Duration
	ExtraClientCert    string
	InsecureSkipVerify bool
}

type HttpClientOption func(*ClientOptions)

func NewWithOptions(options ...HttpClientOption) *http.Client {
	conf := &ClientOptions{}
	for _, option := range options {
		option(conf)
	}

	client := &http.Client{
		Timeout:   conf.Timeout,
		Transport: http.DefaultTransport,
	}

	if conf.ExtraClientCert != "" || conf.InsecureSkipVerify {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: conf.InsecureSkipVerify,
				RootCAs:            getRoots(conf.ExtraClientCert),
			},
			Proxy: http.ProxyFromEnvironment,
		}
	}

	return client
}

func WithClientTimeout(timeout time.Duration) HttpClientOption {
	return func(client *ClientOptions) {
		client.Timeout = timeout
	}
}

func WithInsecureSkipVerify(insecureSkipVerify bool) HttpClientOption {
	return func(client *ClientOptions) {
		client.InsecureSkipVerify = insecureSkipVerify
	}
}

func WithClientCertificate(clientCert string) HttpClientOption {
	return func(client *ClientOptions) {
		client.ExtraClientCert = clientCert
	}
}

func getRoots(extra string) *x509.CertPool {
	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	cert, err := os.ReadFile(extra)
	if err != nil {
		return rootCAs
	}

	// Append our extra cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM(cert); !ok {
		log.Warn().Msgf("local certs not appended, using system certs only")
	}

	return rootCAs
}
