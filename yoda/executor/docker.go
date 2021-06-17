package executor

import (
	"bytes"
	"context"
	"github.com/GeoDB-Limited/odin-core/yoda/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"github.com/google/shlex"
)

type DockerExec struct {
	image   string
	timeout time.Duration
}

func NewDockerExec(image string, timeout time.Duration) *DockerExec {
	return &DockerExec{image: image, timeout: timeout}
}

// TODO: handle max data size (either in env, or in args)
func (e *DockerExec) Exec(code []byte, arg string, env interface{}) (ExecResult, error) {
	// TODO: Handle env if we are to revive Docker
	dir, err := ioutil.TempDir("/tmp", "executor")
	if err != nil {
		return ExecResult{}, sdkerrors.Wrap(err, "docker execution failed")
	}
	defer os.RemoveAll(dir)
	err = ioutil.WriteFile(filepath.Join(dir, "exec"), code, 0777)
	if err != nil {
		return ExecResult{}, sdkerrors.Wrap(err, "docker execution failed")
	}
	name := filepath.Base(dir)
	args, err := shlex.Split(arg)
	if err != nil {
		return ExecResult{}, sdkerrors.Wrap(err, "docker execution failed")
	}
	dockerArgs := append([]string{
		"run", "--rm",
		"-v", dir + ":/scratch:ro",
		"--name", name,
		e.image,
		"/scratch/exec",
	}, args...)
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		exec.Command("docker", "kill", name).Start()
		return ExecResult{}, sdkerrors.Wrap(errors.ErrExecutionTimeout, "docker execution failed")
	}
	exitCode := uint32(0)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = uint32(exitError.ExitCode())
		} else {
			return ExecResult{}, sdkerrors.Wrap(err, "docker execution failed")
		}
	}
	output, err := ioutil.ReadAll(io.LimitReader(&buf, types.MaxDataSize))
	if err != nil {
		return ExecResult{}, sdkerrors.Wrap(err, "failed to read docker output data")
	}
	return ExecResult{Output: output, Code: exitCode}, nil
}
