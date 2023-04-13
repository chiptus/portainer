package helm

import "path/filepath"

const SSLCertDir = "SSL_CERT_DIR"

func (handler *Handler) defaultHelmEnv() []string {
	return []string{SSLCertDir + "=" + filepath.Dir(handler.fileService.GetSSLClientCertPath())}
}
