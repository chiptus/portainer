package ssh

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/sftp"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Password string
	Config   ssh.ClientConfig
}

type SSHConnection struct {
	IP        string
	SSHConfig *SSHConfig
	*ssh.Client
}

func NewSSHConfig(user, password, passphrase, privateKey string) (*SSHConfig, error) {
	auth := ssh.Password(password)
	if privateKey != "" {
		// Create signer with the private key.
		key, err := base64.StdEncoding.DecodeString(privateKey)
		if err != nil {
			log.Err(err).Msg("failed to decode private key")
			return nil, err
		}
		var signer ssh.Signer
		if passphrase == "" {
			signer, err = ssh.ParsePrivateKey(key)
			if err != nil {
				log.Err(err).Msg("failed to parse private key")
				return nil, err
			}
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
			if err != nil {
				log.Err(err).Msg("failed to parse private key")
				return nil, err
			}
		}
		auth = ssh.PublicKeys(signer)
	}

	return &SSHConfig{
		Password: password,
		Config: ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				auth,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         time.Duration(5) * time.Second,
		},
	}, nil
}

func NewConnection(user, password, passphrase, privateKey, ip string) (*SSHConnection, error) {
	config, err := NewSSHConfig(user, password, passphrase, privateKey)
	if err != nil {
		return nil, err
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", ip), &config.Config)
	if err != nil {
		return nil, err
	}

	s := SSHConnection{
		IP:        ip,
		SSHConfig: config,
		Client:    conn,
	}

	return &s, nil
}

func NewConnectionWithCredentials(ip string, credentials *models.CloudCredential) (*SSHConnection, error) {
	username, ok := credentials.Credentials["username"]
	if !ok {
		log.Debug().Msg("credentials are missing ssh username")
		return nil, fmt.Errorf("missing ssh username")
	}
	password := credentials.Credentials["password"]
	passphrase, passphraseOK := credentials.Credentials["passphrase"]
	privateKey, privateKeyOK := credentials.Credentials["privateKey"]

	if passphraseOK && !privateKeyOK {
		log.Debug().Msg("passphrase provided, but we are missing a private key")
	}

	return NewConnection(username, password, passphrase, privateKey, ip)
}

func (s *SSHConnection) RunCommand(command string, out io.Writer) error {
	log.Debug().Str("node", s.IP).Msgf("Running command: %s", command)

	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	sftpClient, err := sftp.NewClient(s.Client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	passSFTP, _ := sftpClient.Create(".password")
	defer sftpClient.Remove(".password")
	err = sftpClient.Chmod(".password", 0o600)
	if err != nil {
		return err
	}
	_, err = passSFTP.Write([]byte(s.SSHConfig.Password))
	if err != nil {
		return errors.New("failed to write password to file")
	}
	passSFTP.Close()

	session.Stdout = out
	session.Stderr = os.Stderr

	return session.Run(fmt.Sprintf("cat '.password' | sudo -S %s", command))
}
