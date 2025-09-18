package compile

import (
	"fmt"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/log"
)

type Logger struct {
	ctx       *Context
	logger    *log.Logger
	callLevel int
}

func (c *Logger) Fatal(msg string, err error) {
	if err != nil {
		c.logger.Error(msg, "err", err, "path", c.ctx.CurrentRefPointer())
	}
	c.logger.Error(msg, "path", c.ctx.CurrentRefPointer())
}

func (c *Logger) Warn(msg string, args ...any) {
	args = append(args, "path", c.ctx.CurrentRefPointer())
	c.logger.Warn(msg, args...)
}

func (c *Logger) Info(msg string, args ...any) {
	args = append(args, "path", c.ctx.CurrentRefPointer())
	c.logger.Info(msg, args...)
}

func (c *Logger) Debug(msg string, args ...any) {
	l := c.logger
	if c.callLevel > 0 {
		msg = fmt.Sprintf("%s> %s", strings.Repeat("-", c.callLevel), msg) // Ex: prefix: --> Message...
	}
	args = append(args, "path", c.ctx.CurrentRefPointer())
	l.Debug(msg, args...)
}

func (c *Logger) Trace(msg string, args ...any) {
	l := c.logger
	if c.callLevel > 0 {
		msg = fmt.Sprintf("%s> %s", strings.Repeat("-", c.callLevel), msg) // Ex: prefix: --> Message...
	}
	args = append(args, "path", c.ctx.CurrentRefPointer())
	l.Trace(msg, args...)
}

func (c *Logger) NextCallLevel() {
	c.callLevel++
}

func (c *Logger) PrevCallLevel() {
	if c.callLevel > 0 {
		c.callLevel--
	}
}
