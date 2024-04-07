package compiler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

const HangedSubprocessTimeout = 3 * time.Second

type SpecFileResolver interface {
	Resolve(specPath string) (io.ReadCloser, error)
}

// DefaultSpecFileResolver is the default file resolver. Local spec files are read from the filesystem, remote spec
// files are downloaded via http(s).
type DefaultSpecFileResolver struct {
	Client  *http.Client
	Timeout time.Duration
	BaseDir string // Where to search local specs
	Logger  *types.Logger
}

func (r DefaultSpecFileResolver) Resolve(specPath string) (io.ReadCloser, error) {
	if IsRemoteSpecPath(specPath) {
		return r.resolveRemote(specPath)
	}
	return r.resolveLocal(specPath)
}

func (r DefaultSpecFileResolver) resolveLocal(specPath string) (io.ReadCloser, error) {
	absPath, err := filepath.Abs(filepath.Join(r.BaseDir, specPath))
	if err != nil {
		return nil, fmt.Errorf("resolve path %q: %w", filepath.Join(r.BaseDir, specPath), err)
	}
	r.Logger.Info("Reading local file", "baseDir", r.BaseDir, "specPath", specPath, "absolutePath", absPath)
	return os.Open(absPath)
}

func (r DefaultSpecFileResolver) resolveRemote(specPath string) (io.ReadCloser, error) {
	r.Logger.Info("Downloading remote file", "specPath", specPath)
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, specPath, nil)
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

// SubprocessSpecFileResolver is the resolver based on user-defined command. Both local and remote specs are resolved
// by running this command as subprocess, which read specPath from stdin and write spec content to stdout.
type SubprocessSpecFileResolver struct {
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

func (r SubprocessSpecFileResolver) Resolve(specPath string) (io.ReadCloser, error) {
	r.Logger.Info("Resolving spec by command", "specPath", specPath, "commandLine", r.CommandLine, "timeout", r.RunTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), r.RunTimeout)
	defer cancel()

	cmd, err := r.getCommand(ctx)
	if err != nil {
		return nil, fmt.Errorf("get command: %w", err)
	}
	r.Logger.Debug("Run command", "cmd", cmd.command, "stdinData", specPath)
	if err = cmd.command.Start(); err != nil {
		return nil, fmt.Errorf("start command: %w", err)
	}
	if _, err = fmt.Fprintln(cmd.stdin, specPath); err != nil {
		return nil, fmt.Errorf("write to stdin: %w", err)
	}

	if err = cmd.command.Wait(); err != nil {
		return nil, fmt.Errorf("wait command: %w", err)
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
	res.command.WaitDelay = HangedSubprocessTimeout

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

func IsRemoteSpecPath(specPath string) bool {
	_, _, remote := utils.SplitRefToPathPointer(specPath)
	return remote
}
