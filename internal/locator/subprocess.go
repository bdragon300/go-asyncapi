package locator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
)

// Subprocess locator reads the document contents from the shell command output.
//
// The locator runs the command specified in the CommandLine field and passes the document url as the first line of the
// command's stdin, and expects the document contents from the command's stdout. After that the command should exit.
type Subprocess struct {
	// CommandLine is the command line to run.
	CommandLine string
	// If set, RunTimeout is the maximum time the command is allowed to run. If the command didn't finish within timeout,
	// it receives a SIGTERM signal and the ShutdownTimeout start to tick. If shutdown timeout is reached and the
	// command still runs, the command is killed with a SIGKILL signal.
	RunTimeout time.Duration
	// If set, ShutdownTimeout is the time to wait for the command to finish after the SIGTERM signal is sent before
	// killing it with a SIGKILL signal.
	ShutdownTimeout time.Duration
	// RootDirectory is the base directory to locate the file paths. Not used if empty.
	RootDirectory string
	Logger        *log.Logger
}

type subprocessCommand struct {
	command *exec.Cmd
	stdin   io.ReadWriter
	stdout  io.ReadWriter
	stderr  io.ReadWriter
}

// ResolveURL joins the [jsonpointer.JSONPointer] to the base document and ref inside it and returns the pointer to the
// referenced document.
//
// If target is absolute (absolute filesystem path or URL), it is returned as is. If target is relative, it is joined to the base
// document location (filesystem path or URL).
//
// Examples:
//
//	Base: http://example.com/schemas/root.json#/components/schemas/A
//	Target: ../common.json#/components/schemas/B
//	Result: http://example.com/schemas/common.json#/components/schemas/B
//
//	Base: /home/user/schemas/root.json#/components/schemas/A
//	Target: ../common.json#/components/schemas/B
//	Result: /home/user/schemas/common.json#/components/schemas/B
//
//	Base: http://example.com/schemas/root.json#/components/schemas/A
//	Target: /home/user/schemas/common.json#/components/schemas/B
//	Result: /home/user/schemas/common.json#/components/schemas/B
//
//	Base: /home/user/schemas/root.json#/components/schemas/A
//	Target: http://example.com/schemas/common.json#/components/schemas/B
//	Result: http://example.com/schemas/common.json#/components/schemas/B
func (r Subprocess) ResolveURL(base, target *jsonpointer.JSONPointer) (*jsonpointer.JSONPointer, error) {
	return joinBase(r.RootDirectory, base, target)
}

// Locate reads the given document URI by running the command specified in the CommandLine field. Function blocks
// until the command finishes or terminates. Returns an io.ReadCloser to the command output contents.
func (r Subprocess) Locate(docURL *jsonpointer.JSONPointer) (io.ReadCloser, error) {
	r.Logger.Info(
		"Run the command",
		"document", docURL,
		"commandLine", r.CommandLine,
		"runTimeout", r.RunTimeout,
		"shutdownTimeout", r.ShutdownTimeout,
	)

	ctx, cancel := context.WithTimeout(context.Background(), r.RunTimeout)
	defer cancel()

	cmd, err := r.getCommand(ctx)
	if err != nil {
		return nil, fmt.Errorf("get command: %w", err)
	}
	r.Logger.Debug("Run command", "cmd", cmd.command, "stdinData", docURL.Location())
	if err = cmd.command.Start(); err != nil {
		return nil, fmt.Errorf("start command: %w", err)
	}
	if _, err = fmt.Fprintln(cmd.stdin, docURL.Location()); err != nil {
		return nil, fmt.Errorf("write to stdin: %w", err)
	}

	if err = cmd.command.Wait(); err != nil {
		return nil, fmt.Errorf("wait for command: %w", err)
	}
	r.Logger.Debug("Command finished")

	return io.NopCloser(cmd.stdout), nil
}

func (r Subprocess) getCommand(ctx context.Context) (subprocessCommand, error) {
	res := subprocessCommand{
		stdin:  bytes.NewBuffer(make([]byte, 0)),
		stdout: bytes.NewBuffer(make([]byte, 0)),
		stderr: os.Stderr,
	}

	args := parseCommandLine(r.CommandLine)
	if len(args) == 0 || args[0] == "" {
		return res, fmt.Errorf("command line is empty")
	}

	res.command = exec.CommandContext(ctx, args[0], args[1:]...)
	res.command.Stdout = res.stdout
	res.command.Stdin = res.stdin
	res.command.Stderr = res.stderr
	res.command.WaitDelay = r.ShutdownTimeout

	return res, nil
}

// parseCommandLine splits the raw shell command line string into arguments
func parseCommandLine(commandLine string) []string {
	var args []string
	var arg string
	var firstQuote rune
	var escapeNext bool

	for _, c := range commandLine {
		if escapeNext {
			arg += string(c)
			escapeNext = false
			continue
		}

		switch c {
		case ' ':
			if firstQuote > 0 {
				arg += string(c)
			} else if arg != "" {
				args = append(args, arg)
				arg = ""
			}
		case '"', '\'':
			switch {
			case c == firstQuote:
				firstQuote = 0
			case firstQuote == 0:
				firstQuote = c
			default:
				arg += string(c)
			}
		case '\\':
			escapeNext = true
		default:
			arg += string(c)
		}
	}
	if arg != "" {
		args = append(args, arg)
	}
	return args
}
