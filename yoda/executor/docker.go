package executor

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"time"

	"github.com/levigross/grequests"
)

// Only use in testnet. No intensive testing, use at your own risk
type DockerExec struct {
	image     string
	name      string
	timeout   time.Duration
	portLists chan string
	maxTry    int
}

func NewDockerExec(image string, timeout time.Duration, maxTry int, startPort int, endPort int) *DockerExec {
	portLists := make(chan string, endPort-startPort+1)
	name := "docker-runtime-executor-"
	for i := startPort; i <= endPort; i++ {
		port := strconv.Itoa(i)
		StartContainer(name, port, image)
		portLists <- port
	}

	return &DockerExec{image: image, name: name, timeout: timeout, portLists: portLists, maxTry: maxTry}
}

func StartContainer(name string, port string, image string) error {
	err := exec.Command("docker", "restart", name+port).Run()
	if err != nil {
		dockerArgs := append([]string{
			"run",
			"--name", name + port,
			"-p", port + ":5000",
			"--restart=always",
			"--memory=512m",
			image,
		})

		cmd := exec.CommandContext(context.Background(), "docker", dockerArgs...)
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err = cmd.Start()
	}
	return err
}

func (e *DockerExec) PostRequest(
	code []byte,
	arg string,
	env interface{},
	name string,
	port string,
) (ExecResult, error) {
	executable := base64.StdEncoding.EncodeToString(code)
	resp, err := grequests.Post(
		"http://localhost:"+port,
		&grequests.RequestOptions{
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			JSON: map[string]interface{}{
				"executable": executable,
				"calldata":   arg,
				"timeout":    e.timeout.Milliseconds(),
				"env":        env,
			},
			RequestTimeout: e.timeout,
		},
	)

	if err != nil {
		urlErr, ok := err.(*url.Error)
		if !ok || !urlErr.Timeout() {
			return ExecResult{}, err
		}
		// Return timeout code
		return ExecResult{Output: []byte{}, Code: 111}, nil
	}

	if !resp.Ok {
		return ExecResult{}, ErrRestNotOk
	}

	r := externalExecutionResponse{}
	err = resp.JSON(&r)

	if err != nil {
		return ExecResult{}, err
	}

	go func() {
		// StartContainer(name, port, e.image)
		err := exec.Command("docker", "restart", name+port).Run()
		for err != nil {
			err = StartContainer(name, port, e.image)
		}
		e.portLists <- port
	}()
	if r.Returncode == 0 {
		return ExecResult{Output: []byte(r.Stdout), Code: 0, Version: r.Version}, nil
	} else {
		return ExecResult{Output: []byte(r.Stderr), Code: r.Returncode, Version: r.Version}, nil
	}
}

func (e *DockerExec) Exec(code []byte, arg string, env interface{}) (ExecResult, error) {
	port := <-e.portLists
	errs := []error{}
	for i := 0; i < e.maxTry; i++ {
		execResult, err := e.PostRequest(code, arg, env, e.name, port)
		if err == nil {
			return execResult, err
		}
		errs = append(errs, err)
		time.Sleep(500 * time.Millisecond)
	}
	return ExecResult{}, fmt.Errorf(ErrReachMaxTry.Error()+", tried errors: %#q", errs)
}
