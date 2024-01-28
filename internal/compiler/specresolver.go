package compiler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

const HangedProcessTimeout = 3 * time.Second

type SpecResolver interface {
	Resolve(specID string) (io.ReadCloser, error)
}

// DefaultSpecResolver is the default spec resolver. Local specs are read from the filesystem, remote specs are
// downloaded via http.
type DefaultSpecResolver struct {
	Client  *http.Client
	Timeout time.Duration
	Logger  *types.Logger
}

func (r DefaultSpecResolver) Resolve(specID string) (io.ReadCloser, error) {
	if IsRemoteSpecID(specID) {
		return r.resolveRemote(specID)
	}
	return r.resolveLocal(specID)
}

func (r DefaultSpecResolver) resolveLocal(specID string) (io.ReadCloser, error) {
	r.Logger.Info("Reading local file", "specID", specID)
	return os.Open(specID)
}

func (r DefaultSpecResolver) resolveRemote(specID string) (io.ReadCloser, error) {
	r.Logger.Info("Downloading remote file", "specID", specID)
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, specID, nil)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	r.Logger.Debug(
		"HTTP request",
		"method", req.Method,
		"url", req.URL,
		"headers", strings.Join(lo.MapToSlice(req.Header, func(key string, values []string) string {
			return fmt.Sprintf("%s: %s", key, strings.Join(values, ", "))
		}), "\n"),
		"host", req.Host,
	)
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	r.Logger.Debug(
		"HTTP response",
		"status", resp.Status,
		"headers", strings.Join(lo.MapToSlice(resp.Header, func(key string, values []string) string {
			return fmt.Sprintf("%s: %s", key, strings.Join(values, ", "))
		}), "\n"),
	)
	if resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("error http code: %d", resp.StatusCode)
	}
	return resp.Body, nil
}

// SubproccessSpecResolver is the resolver based on user-defined command. Both local and remote specs are resolved
// by running this command as subprocess, which read specID from stdin and write spec content to stdout.
type SubproccessSpecResolver struct {
	CommandLine string
	RunTimeout  time.Duration
	Logger      *types.Logger
}

type subprocessSpecCommand struct {
	command *exec.Cmd
	stdin   io.ReadWriter
	stdout  io.ReadWriter
	stderr  io.ReadWriter
}

func (r SubproccessSpecResolver) Resolve(specID string) (io.ReadCloser, error) {
	r.Logger.Info("Resolving spec by command", "specID", specID, "commandLine", r.CommandLine, "timeout", r.RunTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), r.RunTimeout)
	defer cancel()

	cmd, err := r.getCommand(ctx)
	if err != nil {
		return nil, fmt.Errorf("get command: %w", err)
	}
	r.Logger.Debug("Run command", "cmd", cmd.command, "stdinData", specID)
	if err = cmd.command.Start(); err != nil {
		return nil, fmt.Errorf("start command: %w", err)
	}
	if _, err = fmt.Fprintln(cmd.stdin, specID); err != nil {
		return nil, fmt.Errorf("write to stdin: %w", err)
	}

	if err = cmd.command.Wait(); err != nil {
		return nil, fmt.Errorf("wait command: %w", err)
	}
	r.Logger.Debug("Command finished")

	return io.NopCloser(cmd.stdout), nil
}

func (r SubproccessSpecResolver) getCommand(ctx context.Context) (subprocessSpecCommand, error) {
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
	res.command.WaitDelay = HangedProcessTimeout

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

func IsRemoteSpecID(specID string) bool {
	_, _, remote := utils.SplitRefToPathPointer(specID)
	return remote
}
