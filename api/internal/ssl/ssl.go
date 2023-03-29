package ssl

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"sync"
	"time"

	"github.com/portainer/libcrypto"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/ssl/revoke"
	portainer "github.com/portainer/portainer/api"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Service represents a service to manage SSL certificates
type Service struct {
	fileService     portainer.FileService
	dataStore       dataservices.DataStore
	rawCert         *tls.Certificate
	mtlsRawCert     *tls.Certificate
	certPool        *x509.CertPool
	shutdownTrigger context.CancelFunc
	crlService      *revoke.Service
	mu              sync.RWMutex
}

// NewService returns a pointer to a new Service
func NewService(fileService portainer.FileService, dataStore dataservices.DataStore, shutdownTrigger context.CancelFunc) *Service {
	return &Service{
		fileService:     fileService,
		dataStore:       dataStore,
		shutdownTrigger: shutdownTrigger,
		crlService:      revoke.NewService(),
	}
}

// Init initializes the service
func (service *Service) Init(host, certPath, keyPath, caCertPath, mTLSCertPath, mTLSKeyPath, mTLSCACertPath string) error {
	// mTLS
	mtlsSettings, err := service.dataStore.Settings().Settings()
	if err != nil {
		return errors.Wrap(err, "unable to retrieve the mTLS settings")
	}

	if mtlsSettings.Edge.MTLS.UseSeparateCert && mTLSCACertPath != "" && mTLSCertPath != "" && mTLSKeyPath != "" {
		ca, err := os.ReadFile(mTLSCACertPath)
		if err != nil {
			return errors.Wrap(err, "failed reading the mTLS CA certificate")
		}

		cert, err := os.ReadFile(mTLSCertPath)
		if err != nil {
			return errors.Wrap(err, "failed reading the mTLS certificate")
		}

		key, err := os.ReadFile(mTLSKeyPath)
		if err != nil {
			return errors.Wrap(err, "failed reading the mTLS key")
		}

		_, _, _, err = service.fileService.StoreMTLSCertificates(cert, ca, key)
		if err != nil {
			return errors.Wrap(err, "failed storing the mTLS files")
		}

		err = service.SetMTLSCertificates(ca, cert, key)
		if err != nil {
			return errors.Wrap(err, "unable to initialize mTLS")
		}

		if service.GetCACertificatePool() == nil {
			return errors.Wrap(err, "unable to initialize the mTLS CA certificate pool")
		}
	}

	// TLS
	certSupplied := certPath != "" && keyPath != ""
	caCertSupplied := caCertPath != ""

	if caCertSupplied && !certSupplied {
		return errors.Errorf("supplying a CA cert path (%s) requires an SSL cert and key file", caCertPath)
	}

	if certSupplied {
		newCertPath, newKeyPath, err := service.fileService.CopySSLCertPair(certPath, keyPath)
		if err != nil {
			return errors.Wrap(err, "failed copying supplied certs")
		}

		newCACertPath := ""
		if caCertSupplied {
			newCACertPath, err = service.fileService.CopySSLCACert(caCertPath)
			if err != nil {
				return errors.Wrap(err, "failed copying supplied CA cert")
			}
		}

		return service.cacheInfo(newCertPath, newKeyPath, &newCACertPath, false)
	}

	settings, err := service.GetSSLSettings()
	if err != nil {
		return errors.Wrap(err, "failed fetching SSL settings")
	}

	// certificates already exist
	if settings.CertPath != "" && settings.KeyPath != "" {
		err := service.cacheCertificate(settings.CertPath, settings.KeyPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		// continue if certs don't exist
		if err == nil {
			return nil
		}
	}

	// path not supplied and certificates don't exist - generate self-signed
	certPath, keyPath = service.fileService.GetDefaultSSLCertsPath()

	err = generateSelfSignedCertificates(host, certPath, keyPath)
	if err != nil {
		return errors.Wrap(err, "failed generating self signed certs")
	}

	return service.cacheInfo(certPath, keyPath, &caCertPath, true)
}

func generateSelfSignedCertificates(ip, certPath, keyPath string) error {
	if ip == "" {
		return errors.New("host can't be empty")
	}

	log.Info().Msg("no cert files found, generating self signed SSL certificates")

	return libcrypto.GenerateCertsForHost("localhost", ip, certPath, keyPath, time.Now().AddDate(5, 0, 0))
}

// GetRawCertificate gets the raw certificate
func (service *Service) GetRawCertificate() *tls.Certificate {
	service.mu.RLock()
	defer service.mu.RUnlock()

	return service.rawCert
}

func (service *Service) GetRawMTLSCertificate() *tls.Certificate {
	service.mu.RLock()
	defer service.mu.RUnlock()

	return service.mtlsRawCert
}

// GetSSLSettings gets the certificate info
func (service *Service) GetSSLSettings() (*portaineree.SSLSettings, error) {
	return service.dataStore.SSLSettings().Settings()
}

// SetCertificates sets the certificates
func (service *Service) SetCertificates(certData, keyData []byte) error {
	if len(certData) == 0 || len(keyData) == 0 {
		return errors.New("missing certificate files")
	}

	_, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return err
	}

	var certPath, keyPath string
	{
		service.mu.Lock()
		defer service.mu.Unlock()

		certPath, keyPath, err = service.fileService.StoreSSLCertPair(certData, keyData)
		if err != nil {
			return err
		}
	}

	err = service.cacheInfo(certPath, keyPath, nil, false)
	if err != nil {
		return err
	}

	service.shutdownTrigger()

	return nil
}

// GetCACertificatePool gets the CA Certificate pem file and returns it as a CertPool
func (service *Service) GetCACertificatePool() *x509.CertPool {
	if service == nil {
		return nil
	}

	service.mu.RLock()
	defer service.mu.RUnlock()

	if service.certPool != nil {
		return service.certPool
	}

	settings, err := service.GetSSLSettings()
	if err != nil {
		log.Error().Err(err).Msg("unable to retrieve the TLS settings")
		return nil
	}

	if settings.CACertPath == "" {
		return nil
	}

	caCert, err := os.ReadFile(settings.CACertPath)
	if err != nil {
		log.Debug().Str("path", settings.CACertPath).Err(err).Msg("error reading CA cert")
		return nil
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)

	return certPool
}

func (service *Service) SetHTTPEnabled(httpEnabled bool) error {
	settings, err := service.dataStore.SSLSettings().Settings()
	if err != nil {
		return err
	}

	if settings.HTTPEnabled == httpEnabled {
		return nil
	}

	settings.HTTPEnabled = httpEnabled

	err = service.dataStore.SSLSettings().UpdateSettings(settings)
	if err != nil {
		return err
	}

	service.shutdownTrigger()

	return nil
}

func (service *Service) ValidateCACert(tlsConn *tls.ConnectionState) error {
	// if a caCert is set, then reject any requests that don't have a client Auth cert signed with it
	if tlsConn == nil || len(tlsConn.PeerCertificates) == 0 {
		log.Error().Msg("no clientAuth Agent certificate offered")

		return errors.New("no clientAuth Agent certificate offered")
	}

	serverCACertPool := service.GetCACertificatePool()
	if serverCACertPool == nil {
		log.Error().Msg("CA Certificate not found")

		return errors.New("no CA Certificate was found")
	}

	opts := x509.VerifyOptions{
		Roots:     serverCACertPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	agentCert := tlsConn.PeerCertificates[0]

	revoked, err := service.crlService.VerifyCertificate(agentCert)
	if err != nil {
		log.Error().Err(err).Msg("failed verifying certificate with CRL list")

		return errors.Wrap(err, "failed verifying certificate with CRL list")
	}

	if revoked {
		return errors.New("client certificate is revoked")
	}

	if _, err := agentCert.Verify(opts); err != nil {
		log.Error().Err(err).Msg("agent certificate not signed by the CACert")

		return errors.New("agent certificate wasn't signed by required CA Cert")
	}

	log.Debug().
		Stringer("subject", agentCert.Subject).
		Strs("dns_names", agentCert.DNSNames).
		Msg("successfully validated TLS Client Chain")

	return nil
}

func (service *Service) cacheCertificate(certPath, keyPath string) error {
	rawCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return err
	}

	service.rawCert = &rawCert

	return nil
}

func (service *Service) cacheInfo(certPath string, keyPath string, caCertPath *string, selfSigned bool) error {
	service.mu.Lock()
	defer service.mu.Unlock()

	err := service.cacheCertificate(certPath, keyPath)
	if err != nil {
		return err
	}

	settings, err := service.dataStore.SSLSettings().Settings()
	if err != nil {
		return err
	}

	settings.CertPath = certPath
	settings.KeyPath = keyPath
	settings.SelfSigned = selfSigned
	if caCertPath != nil {
		settings.CACertPath = *caCertPath
	}

	return service.dataStore.SSLSettings().UpdateSettings(settings)
}

func (service *Service) SetMTLSCertificates(ca []byte, cert []byte, key []byte) error {
	service.mu.Lock()
	defer service.mu.Unlock()

	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return err
	}

	tlsCert.Leaf, err = x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return err
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return errors.New("could not parse the mTLS CA certificate")
	}

	service.mtlsRawCert = &tlsCert
	service.certPool = certPool

	return nil
}

func (service *Service) DisableMTLS() {
	service.mu.Lock()
	service.mtlsRawCert = nil
	service.certPool = nil
	service.mu.Unlock()
}
