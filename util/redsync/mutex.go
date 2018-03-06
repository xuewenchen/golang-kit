package redsync

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"golang-kit/config"
	"golang-kit/db/redis"
	xredis "golang/redigo/redis"
)

const (
	DefaultTries  = 16                     // default try
	DefaultExpiry = 3 * time.Second        // default expire
	DefaultDelay  = 512 * time.Millisecond // default delay
)

var (
	ErrFailed = errors.New("failed to acquire lock") // ErrFailed is returned when lock cannot be acquired
)

var (
	DEFALUTCONFIG = &MutexConfig{
		try:    DefaultTries,
		delay:  DefaultDelay,
		expire: DefaultExpiry,
	}
)

// just hold the redis pool
type RedSync struct {
	pool *redis.Pool
}

// A Mutex is a mutual exclusion lock.
type Mutex struct {
	// must have
	Name  string // key name
	value string // key value
	pool  *redis.Pool

	config *MutexConfig
}

type MutexConfig struct {
	try    int           // try number
	delay  time.Duration // try internal sleep time
	expire time.Duration // expire time
}

func NewMutexConfig(try int, delay, expire time.Duration) *MutexConfig {
	return &MutexConfig{
		try:    try,
		delay:  delay,
		expire: expire,
	}
}

// New RedSync
func NewRedSync(c *config.Redis) *RedSync {
	pool := redis.NewRedisPool(c)
	return &RedSync{pool}
}

// New Mutex
// the name is must set
// if the config is nil, will use default config
func (r *RedSync) NewMutex(name string, c *MutexConfig) (mux *Mutex) {
	mux = &Mutex{
		Name: name,
		pool: r.pool,
	}
	if c == nil {
		mux.config = DEFALUTCONFIG
	}
	return
}

// lock the key, this key has expire time
// so if you forget Unlock it, it will ok
func (m *Mutex) Lock() error {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}
	value := base64.StdEncoding.EncodeToString(b)

	n := 0 // try number, because the first lock man perhase unlock it soon, you can try for some times

	for i := 0; i < m.config.try; i++ {
		conn := m.pool.Get(context.Background())
		reply, err := xredis.String(conn.Do("set", m.Name, value, "nx", "px", int(m.config.expire/time.Millisecond)))
		conn.Close()
		if err != nil {
			continue
		}
		if reply != "OK" {
			continue
		}
		if n > m.config.try {
			return ErrFailed
		}
		n++
		time.Sleep(m.config.delay) // you must sleep for sometime
	}
	return ErrFailed
}

// if you you want to lock this key for some time, you need touch it sometimes
func (m *Mutex) Touch() bool {
	value := m.value
	if value == "" {
		return false
	}
	conn := m.pool.Get(context.Background())
	reply, err := xredis.String(conn.Do("set", m.Name, value, "xx", "px", int(m.config.expire/time.Millisecond)))
	conn.Close()
	if err != nil {
		return false
	}
	if reply == "OK" {
		return true
	}
	return false
}

// del this key
func (m *Mutex) Unlock() (b bool) {
	conn := m.pool.Get(context.Background())
	status, err := xredis.Int(conn.Do("del", m.Name))
	conn.Close()
	if err != nil {
		return false
	}
	if status != 1 {
		return false
	}
	return true
}
