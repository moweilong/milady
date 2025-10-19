// Package gobash provides the ability to execute commands, scripts, executables in the go environment with live log output.
package gobash

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

// Exec suitable for executing a single non-blocking command, outputting standard and error logs,
// but the log output is not real time, no execution, command name must be in system path,
// Note: If the execution of a command blocks permanently, it can cause a concurrent leak.
func Exec(name string, args ...string) ([]byte, error) {
	cmdName, err := exec.LookPath(name) // cmdName is absolute path
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(cmdName, args...)
	return getResult(cmd)
}

// Result of the execution of the command
type Result struct {
	StdOut chan string
	Err    error // If nil after the command is executed, the command is executed successfully
	Pid    int   // Process ID of the command
}

// Run execute the command, no execution, command name must be in system path,
// you can actively end the command, the execution results are returned in real time in Result.StdOut
func Run(ctx context.Context, name string, args ...string) *Result {
	result := &Result{StdOut: make(chan string), Err: error(nil)}

	go func() {
		defer func() { close(result.StdOut) }() // execution complete, channel closed
		cmdName, err := exec.LookPath(name)     // cmdName is absolute path
		if err != nil {
			result.Err = err
			return
		}
		cmd := exec.CommandContext(ctx, cmdName, args...)
		handleExec(ctx, cmd, result)
	}()

	return result
}

func handleExec(ctx context.Context, cmd *exec.Cmd, result *Result) {
	stdout, stderr, err := getCmdReader(cmd)
	if err != nil {
		result.Err = err
		return
	}

	result.StdOut <- strings.Join(cmd.Args, " ") + fmt.Sprintf(" [pid]=%d", cmd.Process.Pid) + "\n" // command name and pid
	result.Pid = cmd.Process.Pid

	reader := bufio.NewReader(stdout)
	// reads each line in real time
	line := ""
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) { // determine if it has been read
				break
			}
			result.Err = err
			break
		}
		select {
		case result.StdOut <- line:
		case <-ctx.Done():
			result.Err = fmt.Errorf("%v", ctx.Err())
			return
		}
	}

	// capture error logs
	bytesErr, err := io.ReadAll(stderr)
	if err != nil {
		result.Err = err
		return
	}

	err = cmd.Wait()
	if err != nil {
		if len(bytesErr) != 0 {
			result.Err = errors.New(string(bytesErr))
			return
		}
		result.Err = err
	}
}

func getCmdReader(cmd *exec.Cmd) (stdout io.ReadCloser, stderr io.ReadCloser, err error) {
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	stderr, err = cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, nil, err
	}

	return stdout, stderr, nil
}

func getResult(cmd *exec.Cmd) ([]byte, error) {
	stdout, stderr, err := getCmdReader(cmd)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	bytesErr, err := io.ReadAll(stderr)
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		if len(bytesErr) != 0 {
			return nil, errors.New(string(bytesErr))
		}
		return nil, err
	}

	return bytes, nil
}

// ParsePid extracts the process ID from the command output string.
func ParsePid(s string) int {
	const key = "[pid]="
	idx := strings.Index(s, key)
	if idx == -1 {
		return 0
	}

	part := s[idx+len(key):]
	part = strings.TrimSpace(part)

	pid, err := strconv.Atoi(part)
	if err != nil {
		return 0
	}
	return pid
}
