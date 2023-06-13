package eksctl

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	// Vars set during build time.  Don't make const!
	DefaultEksCtlVersion              = "v0.143.0"
	DefaultAwsIamAuthenticatorVersion = "v0.6.10"
)

const (
	Eksctl              = "eksctl"
	AwsIamAuthenticator = "aws-iam-authenticator"
)

type Config struct {
	id              string
	accessKeyId     string
	secretAccessKey string
	region          string
	binaryPath      string
}

func NewConfig(id, accessKeyId, secretAccessKey, region, binaryPath string) *Config {
	return &Config{
		id:              id,
		accessKeyId:     accessKeyId,
		secretAccessKey: secretAccessKey,
		region:          region,
		binaryPath:      binaryPath,
	}
}

func (c *Config) Run(params ...string) error {
	err := ensureEksctl(c.binaryPath)
	if err != nil {
		log.Error().Err(err).Msg("cannot download eksctl and dependencies")

		return fmt.Errorf("Failed to download eksctl or dependancy. Cannot create EKS cluster.")
	}

	// -C turns off colour output
	params = append([]string{"-C", "false"}, params...)

	// run eksctl with privided params
	cmd := exec.Command(Eksctl, params...)

	log.Debug().Strs("args", cmd.Args).Msg("exec")

	// add aws environment vars for authentication and region
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env,
		"AWS_ACCESS_KEY_ID="+c.accessKeyId,
		"AWS_SECRET_ACCESS_KEY="+c.secretAccessKey,
		"AWS_DEFAULT_REGION="+c.region,
	)

	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start eksctl, error: %w", err)
	}

	scanner := bufio.NewScanner(stdout)

	// eksctl prefixes it's output with date/time and type in square brackets.
	// e.g. 2022-05-25 10:21:03 [ℹ]  waiting for CloudFormation stack "eksctl-matt-test2-cluster
	// We strip that off so it does not appear twice when run in a docker container
	stripPrefix := regexp.MustCompile(`.*\]  `)
	ekserr := regexp.MustCompile(`✖|(?i)error|CREATE_FAILED`)
	errorText := ""
	logText := []string{}
	ticker := time.NewTicker(500 * time.Millisecond)
	ch := make(chan string)
	run := true

	// scanner.Scan blocks.  This logic allows us to read as a block of output from eksctl and output to the
	// portainer log in one block at a time.  Rather than line by line.  This should make reading the log output
	// easier when a lot of things are going on.
	go func() {
		for scanner.Scan() {
			ch <- scanner.Text()
		}
		run = false
	}()

	// We periodically poll the output of the command rather than running the
	// whole command and then printing the output. This is because calling
	// provision for a provider is not async. For the other providers it doesn't
	// matter because they return very quickly, but for AWS it can sometimes
	// take 30 minutes and it would be bad UX to show nothing in the logs while
	// that happens.
	for run {
		select {
		case <-ticker.C:
			if len(logText) > 0 {
				log.Info().
					Str("cluster_id", c.id).
					Str("output", strings.Join(logText, "\n  ")).
					Msg("")

				logText = []string{}
			}

		case text := <-ch:
			if errorText == "" && ekserr.MatchString(text) {
				// drop the first error seen into here to be returned once the process exits
				errorText = stripPrefix.ReplaceAllString(text, "")
			}

			logText = append(logText, text)
		}
	}

	// Dump out any remaining log text.
	log.Info().
		Str("cluster_id", c.id).
		Str("output", strings.Join(logText, "\n  ")).
		Msg("")

	if err = cmd.Wait(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			if _, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("eksctl error: %s, %w. See Portainer and AWS CloudFormation logs for more detail.", errorText, err)
			}
		}
	}

	return err
}
