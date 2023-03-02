package sshtest

import (
	"sort"
	"strings"

	"github.com/fvbommel/sortorder"
	"golang.org/x/crypto/ssh"
)

type SSHTestResult struct {
	Address string `json:"address"`
	Error   string `json:"error,omitempty"`
}

func SSHTest(config *ssh.ClientConfig, ips []string) []SSHTestResult {
	maxWorkers := 50
	numJobs := len(ips)

	if len(ips) < maxWorkers {
		maxWorkers = len(ips)
	}

	jobsChan := make(chan int, numJobs)
	resultsChan := make(chan SSHTestResult, numJobs)

	worker := func(id int, jobs <-chan int, results chan<- SSHTestResult) {
		for j := range jobs {
			results <- sshLogin(ips[j], config)
		}
	}

	for w := 1; w <= maxWorkers; w++ {
		go worker(w, jobsChan, resultsChan)
	}

	for j := 0; j < numJobs; j++ {
		jobsChan <- j
	}
	close(jobsChan)

	results := []SSHTestResult{}
	for i := 0; i < numJobs; i++ {
		results = append(results, <-resultsChan)
	}

	sort.Slice(results, func(i, j int) bool {
		return sortorder.NaturalLess(strings.ToLower(results[i].Address), strings.ToLower(results[j].Address))
	})

	return results
}

func sshLogin(target string, config *ssh.ClientConfig) SSHTestResult {
	conn, err := ssh.Dial("tcp", target+":22", config)
	if err != nil {
		return SSHTestResult{
			Address: target,
			Error:   err.Error(),
		}
	}

	defer conn.Close()
	return SSHTestResult{
		Address: target,
	}
}
