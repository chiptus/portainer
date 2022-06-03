package eksctl

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var (
	DefaultEksCtlVersion              = "v0.100.0-rc.0"
	DefaultAwsIamAuthenticatorVersion = "v0.5.7"
)

const (
	Eksctl              = "eksctl"
	AwsIamAuthenticator = "aws-iam-authenticator"
)

type Config struct {
	accessKeyId     string
	secretAccessKey string
	region          string
	binaryPath      string
}

func NewConfig(accessKeyId, secretAccessKey, region, binaryPath string) *Config {
	return &Config{
		accessKeyId:     accessKeyId,
		secretAccessKey: secretAccessKey,
		region:          region,
		binaryPath:      binaryPath,
	}
}

func (c *Config) Run(params ...string) error {
	err := ensureEksctl(c.binaryPath)
	if err != nil {
		log.Errorf("Cannot download eksctl and dependencies: %v", err)
		return fmt.Errorf("Failed to download eksctl or dependancy. Cannot create EKS cluster.")
	}

	// -C turns off colour output
	params = append([]string{"-C", "false"}, params...)

	// run eksctl with privided params
	cmd := exec.Command(Eksctl, params...)

	log.Debugf("exec: %v", cmd.Args)

	// add aws environment vars for authentication and region
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env,
		"AWS_ACCESS_KEY_ID="+c.accessKeyId,
		"AWS_SECRET_ACCESS_KEY="+c.secretAccessKey,
		"AWS_DEFAULT_REGION="+c.region,
	)

	stdout, _ := cmd.StdoutPipe()

	err = cmd.Start()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)

	// eksctl prefixes it's output with date/time and type in square brackets.
	// e.g. 2022-05-25 10:21:03 [ℹ]  waiting for CloudFormation stack "eksctl-matt-test2-cluster
	// We strip that off so it does not appear twice when run in a docker container
	stripPrefix := regexp.MustCompile(`.*\]  `)
	errorText := ""
	for scanner.Scan() {
		text := stripPrefix.ReplaceAllString(scanner.Text(), "")
		if strings.Contains(strings.ToLower(text), "error") {
			errorText = text
		}

		log.Infof("[cloud][eksctl][%s]", text)
	}

	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if _, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("eksctl error: %s, %v", errorText, err)
			}
		}
	}

	return err
}
