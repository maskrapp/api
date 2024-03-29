package global

import (
	"context"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/maskrapp/api/internal/config"
	"github.com/maskrapp/api/internal/domains"
	"github.com/maskrapp/api/internal/jwt"
	"github.com/maskrapp/api/internal/mailer"
	"github.com/maskrapp/api/internal/ratelimit"
	"github.com/maskrapp/api/internal/recaptcha"
	"gorm.io/gorm"
)

type Instances struct {
	Gorm        *gorm.DB
	Redis       *redis.Client
	RateLimiter *ratelimit.RateLimiter
	Recaptcha   *recaptcha.Recaptcha
	JWT         *jwt.JWTHandler
	Mailer      *mailer.Mailer
	Domains     *domains.Domains
}

type Context interface {
	context.Context
	Instances() *Instances
	Config() *config.Config
}

type globalContext struct {
	context.Context
	instances *Instances
	config    *config.Config
}

func NewContext(ctx context.Context, instances *Instances, config *config.Config) Context {
	return &globalContext{
		Context:   ctx,
		instances: instances,
		config:    config,
	}
}

func (r *globalContext) Instances() *Instances {
	return r.instances
}

func (r *globalContext) Config() *config.Config {
	return r.config
}

func WithCancel(ctx Context) (Context, context.CancelFunc) {

	c, cancel := context.WithCancel(ctx)

	return &globalContext{
		Context:   c,
		config:    ctx.Config(),
		instances: ctx.Instances(),
	}, cancel
}

func (r *globalContext) WithDeadline(ctx Context, deadline time.Time) (Context, context.CancelFunc) {
	c, cancel := context.WithDeadline(ctx, deadline)
	return &globalContext{
		Context:   c,
		instances: ctx.Instances(),
		config:    ctx.Config(),
	}, cancel
}

func WithValue(ctx Context, key interface{}, value interface{}) Context {
	return &globalContext{
		Context:   context.WithValue(ctx, key, value),
		instances: ctx.Instances(),
		config:    ctx.Config(),
	}
}

func WithTimeout(ctx Context, timeout time.Duration) (Context, context.CancelFunc) {
	c, cancel := context.WithTimeout(ctx, timeout)
	return &globalContext{
		Context:   c,
		instances: ctx.Instances(),
		config:    ctx.Config(),
	}, cancel
}
