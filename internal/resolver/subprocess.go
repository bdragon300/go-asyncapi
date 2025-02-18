package resolver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/specurl"
)

const subprocessGracefulShutdownTimeout = 3 * time.Second

// SubprocessSpecFileResolver is the resolver based on user-defined command. Both local and remote specs are resolved
// by running this command as subprocess, which read specPath from stdin and write spec content to stdout.
type SubprocessSpecFileResolver struct {
	CommandLine string
	RunTimeout  time.Duration
	Logger      *log.Logger
}

type subprocessSpecCommand struct {
	command *exec.Cmd
	stdin   io.ReadWriter
	stdout  io.ReadWriter
	stderr  io.ReadWriter
}

func (r SubprocessSpecFileResolver) Resolve(specPath *specurl.URL) (io.ReadCloser, error) {
	r.Logger.Info(
		"Resolving spec by command",
		"specPath", specPath,
		"commandLine", r.CommandLine,
		"timeout", r.RunTimeout,
		"gracefulShutdownTimeout", subprocessGracefulShutdownTimeout,
	)

	ctx, cancel := context.WithTimeout(context.Background(), r.RunTimeout)
	defer cancel()

	cmd, err := r.getCommand(ctx)
	if err != nil {
		return nil, fmt.Errorf("get command: %w", err)
	}
	r.Logger.Debug("Run command", "cmd", cmd.command, "stdinData", specPath.SpecID)
	if err = cmd.command.Start(); err != nil {
		return nil, fmt.Errorf("start command: %w", err)
	}
	if _, err = fmt.Fprintln(cmd.stdin, specPath.SpecID); err != nil {
		return nil, fmt.Errorf("write to stdin: %w", err)
	}

	if err = cmd.command.Wait(); err != nil {
		return nil, fmt.Errorf("wait for command: %w", err)
	}
	r.Logger.Debug("Command finished")

	return io.NopCloser(cmd.stdout), nil
}

func (r SubprocessSpecFileResolver) getCommand(ctx context.Context) (subprocessSpecCommand, error) {
	res := subprocessSpecCommand{
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
	res.command.WaitDelay = subprocessGracefulShutdownTimeout

	return res, nil
}

// parseCommandLine splits the raw command line string into arguments
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
