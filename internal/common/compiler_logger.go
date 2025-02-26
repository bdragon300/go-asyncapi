package common

import (
	"fmt"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/log"
)

type CompilerLogger struct {
	ctx       *CompileContext
	logger    *log.Logger
	callLevel int
}

func (c *CompilerLogger) Fatal(msg string, err error) {
	if err != nil {
		c.logger.Error(msg, "err", err, "path", c.ctx.CurrentPositionRef())
	}
	c.logger.Error(msg, "path", c.ctx.CurrentPositionRef())
}

func (c *CompilerLogger) Warn(msg string, args ...any) {
	args = append(args, "path", c.ctx.CurrentPositionRef())
	c.logger.Warn(msg, args...)
}

func (c *CompilerLogger) Info(msg string, args ...any) {
	args = append(args, "path", c.ctx.CurrentPositionRef())
	c.logger.Info(msg, args...)
}

func (c *CompilerLogger) Debug(msg string, args ...any) {
	l := c.logger
	if c.callLevel > 0 {
		msg = fmt.Sprintf("%s> %s", strings.Repeat("-", c.callLevel), msg) // Ex: prefix: --> Message...
	}
	args = append(args, "path", c.ctx.CurrentPositionRef())
	l.Debug(msg, args...)
}

func (c *CompilerLogger) Trace(msg string, args ...any) {
	l := c.logger
	if c.callLevel > 0 {
		msg = fmt.Sprintf("%s> %s", strings.Repeat("-", c.callLevel), msg) // Ex: prefix: --> Message...
	}
	args = append(args, "path", c.ctx.CurrentPositionRef())
	l.Trace(msg, args...)
}

func (c *CompilerLogger) NextCallLevel() {
	c.callLevel++
}

func (c *CompilerLogger) PrevCallLevel() {
	if c.callLevel > 0 {
		c.callLevel--
	}
}
