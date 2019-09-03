package util

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Get the current Git hash for a working directory.
func GetGitHash(workingDir string) (string, error) {
	var err error
	var out bytes.Buffer

	cmd := exec.Command("git", "describe", "--always")
	cmd.Stdout = &out
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// Get the date of the given Git hash for a working directory.
func GetGitCommitDate(hash, workingDir string) (time.Time, error) {
	var err error
	var out bytes.Buffer
	var stderr bytes.Buffer
	var commitDate string
	var commitTime time.Time

	rng := hash + "^.." + hash
	cmd := exec.Command("git", "log", "--pretty=fuller", rng)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	err = cmd.Run()
	if err != nil {
		return commitTime, fmt.Errorf("%s", stderr.String())
	}

	lines := strings.Split(out.String(), "\n")
	for _, s := range lines {
		if strings.HasPrefix(s, "CommitDate") {
			commitDate = s
			continue
		}
	}
	if commitDate == "" {
		return commitTime, fmt.Errorf("Could not find commit date")
	}
	commitDate = strings.TrimSpace(commitDate[12:])

	commitTime, err = time.Parse("Mon Jan 2 15:04:05 2006 -0700", commitDate)
	if err != nil {
		return commitTime, err
	}
	return commitTime.UTC(), nil
}
