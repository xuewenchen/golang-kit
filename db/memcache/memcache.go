package memcache

import (
	"context"
	"fmt"
	opentracing "github.com/opentracing/opentracing-go"
	"golang-kit/config"
	"golang/gomemcache/memcache"
	"time"
)

const (
	_maxTTL = 30*24*60*60 - 1
)

type Conn struct {
	p   *Pool
	c   memcache.Conn
	ctx context.Context
}

// Pool memcache conn pool.
type Pool struct {
	*memcache.Pool
	c *config.Memcache
}

// NewPool new a memcache conn pool.
func NewMemcachePool(c *config.Memcache) (p *Pool) {
	p = &Pool{c: c}
	cnop := memcache.DialConnectTimeout(time.Duration(c.DialTimeout))
	rdop := memcache.DialReadTimeout(time.Duration(c.ReadTimeout))
	wrop := memcache.DialWriteTimeout(time.Duration(c.WriteTimeout))
	auop := memcache.DialPassword(c.Auth)
	p.Pool = memcache.NewPool(func() (memcache.Conn, error) {
		return memcache.Dial(c.Proto, c.Addr, cnop, rdop, wrop, auop)
	}, c.Idle)
	p.IdleTimeout = time.Duration(c.IdleTimeout)
	p.MaxActive = c.Active
	return
}

func (p *Pool) Get(ctx context.Context) *Conn {
	return &Conn{p: p, c: p.Pool.Get(), ctx: ctx}
}

func (p *Pool) Close() error {
	return p.Pool.Close()
}

func (c *Conn) Close() error {
	return c.c.Close()
}

func (c *Conn) Err() error {
	return c.c.Err()
}

func (c *Conn) Store(cmd, key string, value []byte, flags uint32, timeout int32, cas uint64) (err error) {
	if timeout > _maxTTL {
		timeout = _maxTTL
	}
	if c.ctx.Value("sync") == true {
		span, _ := opentracing.StartSpanFromContext(c.ctx, fmt.Sprintf("mc %s %s", cmd, key))
		span.LogEvent(fmt.Sprintf("%s %s", cmd, key))
		defer span.Finish()
	}
	err = c.c.Store(cmd, key, value, flags, timeout, cas)
	return
}

func (c *Conn) Get(cmd string, cb func(*memcache.Reply), keys ...string) (err error) {
	var (
		r   *memcache.Reply
		res []*memcache.Reply
	)
	if c.ctx.Value("sync") == true {
		span, _ := opentracing.StartSpanFromContext(c.ctx, fmt.Sprintf("mc %s", cmd))
		span.LogEvent(fmt.Sprintf("%s %v", cmd, keys))
		defer span.Finish()
	}
	if res, err = c.Gets(cmd, keys...); err != nil {
		return
	}
	for _, r = range res {
		cb(r)
	}
	return
}

func (c *Conn) Get2(cmd string, key string) (res *memcache.Reply, err error) {
	if c.ctx.Value("sync") == true {
		span, _ := opentracing.StartSpanFromContext(c.ctx, fmt.Sprintf("mc %s %s", cmd, key))
		span.LogEvent(fmt.Sprintf("%s %s", cmd, key))
		defer span.Finish()
	}
	res, err = c.c.Get(cmd, key)
	return
}

func (c *Conn) Gets(cmd string, keys ...string) (res []*memcache.Reply, err error) {
	if c.ctx.Value("sync") == true {
		span, _ := opentracing.StartSpanFromContext(c.ctx, fmt.Sprintf("mc %s", cmd))
		span.LogEvent(fmt.Sprintf("%s %v", cmd, keys))
		defer span.Finish()
	}
	res, err = c.c.Gets(cmd, keys...)
	return
}

func (c *Conn) Touch(key string, timeout int32) (err error) {
	if timeout > _maxTTL {
		timeout = _maxTTL
	}
	if c.ctx.Value("sync") == true {
		span, _ := opentracing.StartSpanFromContext(c.ctx, "mc touch")
		span.LogEvent(fmt.Sprintf("touch %s", key))
		defer span.Finish()
	}
	err = c.c.Touch(key, timeout)
	return
}

func (c *Conn) Delete(key string) (err error) {
	if c.ctx.Value("sync") == true {
		span, _ := opentracing.StartSpanFromContext(c.ctx, fmt.Sprintf("mc delete %s", key))
		span.LogEvent(fmt.Sprintf("delete %s", key))
		defer span.Finish()
	}
	err = c.c.Delete(key)
	return
}

func (c *Conn) IncrDecr(cmd string, key string, delta uint64) (res uint64, err error) {
	if c.ctx.Value("sync") == true {
		span, _ := opentracing.StartSpanFromContext(c.ctx, fmt.Sprintf("mc %s", cmd))
		span.LogEvent(fmt.Sprintf("%s %s", cmd, key))
		defer span.Finish()
	}
	res, err = c.c.IncrDecr(cmd, key, delta)
	return
}
