package ratelimiter

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"math"
	"time"
)

// redis

// 时间窗口

const script = `
local tokens_key = KEYS[1]
local timestamp_key = KEYS[2]

local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local fill_time = capacity/rate
local ttl = math.floor(fill_time*2)

local last_tokens = tonumber(redis.call("get", tokens_key))
if last_tokens == nil then
    last_tokens = capacity
end

local last_refreshed = tonumber(redis.call("get", timestamp_key))
if last_refreshed == nil then
    last_refreshed = 0
end

local delta = math.max(0, now-last_refreshed)
local filled_tokens = math.min(capacity, last_tokens+(delta*rate))
local allowed = filled_tokens >= requested
local new_tokens = filled_tokens
if allowed then
    new_tokens = filled_tokens - requested
end

redis.call("setex", tokens_key, ttl, new_tokens)
redis.call("setex", timestamp_key, ttl, now)

return { allowed, new_tokens }
`

// Limit defines the maximum frequency of some events.
// Limit is represented as number of events per second.
// A zero Limit allows no events.
type Limit float64

// Inf is the infinite rate limit; it allows all events (even if burst is zero).
const Inf = Limit(math.MaxFloat64)

type Limiter struct {
	client *redis.Client

	limit Limit
	burst int

	scriptHash string

	// mu sync.Mutex

	key string
}

// CreateLimiter returns a new Limiter that allows events up to rate r and permits
// bursts of at most b tokens.
func CreateLimiter(r Limit, b int, key string, client *redis.Client) (lmt *Limiter, err error) {

	//创建limiter
	lmt = new(Limiter)
	lmt.client = client
	lmt.limit = r
	lmt.burst = b
	lmt.key = key

	//load script
	go func() {
		timer := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timer.C:
				lmt.loadScript()
			}
		}
	}()
	return lmt, lmt.loadScript()

}

func (lmt *Limiter) loadScript() error {
	if lmt.client == nil {
		return errors.New("redis client is nil")
	}

	lmt.scriptHash = fmt.Sprintf("%x", sha1.Sum([]byte(script)))
	exists, err := lmt.client.ScriptExists(lmt.scriptHash).Result()
	if err != nil {
		log.Println(err)
		return err
	}

	// load script when missing.
	if !exists[0] {
		_, err := lmt.client.ScriptLoad(script).Result()
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

// Every converts a minimum time interval between events to a Limit.
func Every(interval time.Duration) Limit {
	if interval <= 0 {
		return Inf
	}
	return 1 / Limit(interval.Seconds())
}

// Allow is shorthand for AllowN(time.Now(), 1).
func (lim *Limiter) Allow() bool {
	return lim.AllowN(time.Now(), 1)
}

// AllowN reports whether n events may happen at time now.
// Use this method if you intend to drop / skip events that exceed the rate limit.
// Otherwise use Reserve or Wait.
func (lim *Limiter) AllowN(now time.Time, n int) bool {
	return lim.reserveN(now, n).ok
}

// A Reservation holds information about events that are permitted by a Limiter to happen after a delay.
// A Reservation may be canceled, which may enable the Limiter to permit additional events.
type Reservation struct {
	ok     bool
	tokens int
}

func (lmt *Limiter) reserveN(now time.Time, n int) Reservation {
	if lmt.client == nil {
		return Reservation{
			ok:     true,
			tokens: n,
		}
	}

	results, err := lmt.client.EvalSha(
		lmt.scriptHash,
		[]string{lmt.key + ".tokens", lmt.key + ".ts"},
		float64(lmt.limit),
		lmt.burst,
		now.Unix(),
		n,
	).Result()
	if err != nil {
		log.Println("fail to call rate limit: ", err)
		return Reservation{
			ok:     true,
			tokens: n,
		}
	}

	rs, ok := results.([]interface{})
	if ok {
		newTokens, _ := rs[1].(int64)
		return Reservation{
			ok:     rs[0] == int64(1),
			tokens: int(newTokens),
		}
	}

	log.Println("fail to transform results")
	return Reservation{
		ok:     true,
		tokens: n,
	}
}
