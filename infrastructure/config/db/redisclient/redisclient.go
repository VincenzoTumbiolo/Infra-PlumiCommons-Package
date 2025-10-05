package redisclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/amplifon-x/ax-go-application-layer/v5/slicex"
	"github.com/gomodule/redigo/redis"
)

// CacheErr indicates a missing key. Use the string value to discover the key
type CacheErr string

func (err CacheErr) Error() string {
	return fmt.Sprintf("cache yielded no value on hash %s", string(err))
}

// CacheErrs is a slice of CacheErr
type CacheErrs []CacheErr

// Hashes converts the CacheErrs in strings
func (errs CacheErrs) Hashes() []string {
	return slicex.Map(errs, func(err CacheErr) string { return string(err) })
}

func (errs CacheErrs) Is(err error) bool {
	_, ok := err.(CacheErrs)
	return ok
}

func (errs *CacheErrs) As(target any) bool {
	cast, ok := target.(CacheErrs)
	if !ok {
		return false
	}

	*errs = cast
	return true
}

func (errs CacheErrs) Unwrap() []error {
	return slicex.Map(errs, func(err CacheErr) error { return error(err) })
}

func (errs CacheErrs) Error() string {
	return errors.Join(errs.Unwrap()...).Error()
}

// ErrSkipCache indicates that the operation was skipped due to
// an explicit request
var ErrSkipCache = errors.New("cache skip requested")

const CACHE_CONTROL_CTX_KEY = "redis:cache-control"

// CacheControl is used to alter the behaviour during a cache operation
type CacheControl struct {
	// Skip the operation if true
	Skip bool
}

// RedisClient is a wrapper for a Redigo connection pool
type RedisClient struct {
	pool *redis.Pool
}

type RedisConfig struct {
	Host     string
	Password string
}

// New returns a RedisClient with the given configuration
func New(config RedisConfig) *RedisClient {
	connect := func() (redis.Conn, error) {
		opts := make([]redis.DialOption, 0)
		if config.Password != "" {
			opts = append(opts, redis.DialPassword(config.Password))
		}

		return redis.Dial("tcp", config.Host, opts...)
	}

	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        connect,
	}

	return &RedisClient{pool}
}

// Connection gets a connection from the pool
func (r *RedisClient) Connection(ctx context.Context) (redis.Conn, error) {
	return r.pool.GetContext(ctx)
}

// Ping the database
func (r *RedisClient) Ping(ctx context.Context) error {
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("error getting connection from pool: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		return fmt.Errorf("error during PING: %w", err)
	}

	return nil
}

// Close the connection pool
func (r *RedisClient) Close() error {
	return r.pool.Close()
}

// FlushWithPrefix flushes all entries with the given prefix
func (r RedisClient) FlushWithPrefix(ctx context.Context, prefix string) error {
	conn, err := r.Connection(ctx)
	if err != nil {
		return fmt.Errorf("error getting connection from pool: %w", err)
	}
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", prefix+"*"))
	if err != nil {
		return fmt.Errorf("error building KEYS: %w", err)
	}

	for _, key := range keys {
		_, err = conn.Do("DEL", key)
		if err != nil {
			return fmt.Errorf("error DELeting key %s: %w", key, err)
		}
	}
	return err
}

// FlushAll the entries
func (r RedisClient) FlushAll(ctx context.Context) error {
	conn, err := r.Connection(ctx)
	if err != nil {
		return fmt.Errorf("error getting connection from pool: %w", err)
	}
	defer conn.Close()

	if _, err = conn.Do("FLUSHALL"); err != nil {
		return fmt.Errorf("error during FLUSHALL: %w", err)
	}

	return nil
}

// KeyVal is a generic key-value pair
type KeyVal[T any] struct {
	Key   string
	Value T
}

// KeyIdentity indicates how to extract the key of a given type
type KeyIdentity[T any] func(T) string

// HSetFromValue builds a KeyVal with the given value and identity function
func HSetFromValue[T any](val T, id KeyIdentity[T]) KeyVal[T] {
	return KeyVal[T]{
		Key:   id(val),
		Value: val,
	}
}

// HSet the given KeyVal
func HSet[T any](ctx context.Context, r *RedisClient, val KeyVal[T]) error {
	conn, err := r.Connection(ctx)
	if err != nil {
		return fmt.Errorf("error getting connection from pool: %w", err)
	}
	defer conn.Close()

	args := redis.Args{}.Add(val.Key).AddFlat(val.Value)
	if _, err = conn.Do("HSET", args...); err != nil {
		return fmt.Errorf("error HSETting key %s: %w", val.Key, err)
	}

	return nil
}

// HSetPipe runs multiple HSET commands on a batch of KeyVal
func HSetPipe[T any](ctx context.Context, r *RedisClient, vals []KeyVal[T]) error {
	conn, err := r.Connection(ctx)
	if err != nil {
		return fmt.Errorf("error getting connection from pool: %w", err)
	}
	defer conn.Close()

	var retErr error
	for _, val := range vals {
		args := redis.Args{}.Add(val.Key).AddFlat(val.Value)
		err := conn.Send("HSET", args...)

		retErr = errors.Join(retErr, err)
	}

	if retErr != nil {
		retErr = fmt.Errorf("errors during HSET batch: %w", retErr)
	}

	if err := conn.Flush(); err != nil {
		return fmt.Errorf("error flushing commands: %w\nbatch errors: %w", err, retErr)
	}

	return retErr
}

// HGetAll returns a struct from a given hash
func HGetAll[T any](ctx context.Context, r *RedisClient, hash string) (T, error) {
	var result T
	if shouldSkip(ctx) {
		return result, ErrSkipCache
	}

	conn, err := r.Connection(ctx)
	if err != nil {
		return result, fmt.Errorf("error getting connection from pool: %w", err)
	}
	defer conn.Close()

	values, err := redis.Values(conn.Do("HGETALL", hash))
	if err != nil {
		return result, fmt.Errorf("error during HGETALL with hash %s: %w", hash, err)
	}

	if len(values) == 0 {
		return result, CacheErr(hash)
	}

	if err := redis.ScanStruct(values, &result); err != nil {
		return result, fmt.Errorf("error scanning struct: %w", err)
	}

	return result, nil
}

// HGetAllPipe runs multiple HGETALL commands from a batch of hashes
func HGetAllPipe[T any](ctx context.Context, r *RedisClient, hashes []string) ([]T, error) {
	if shouldSkip(ctx) {
		return nil, ErrSkipCache
	}

	results := make([]T, 0, len(hashes))

	conn, err := r.Connection(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting connection from pool: %w", err)
	}
	defer conn.Close()

	for _, h := range hashes {
		err := conn.Send("HGETALL", h)
		if err != nil {
			return nil, fmt.Errorf("error during HGETALL with hash %s: %w", h, err)
		}
	}

	if err := conn.Flush(); err != nil {
		return nil, fmt.Errorf("error flushing commands: %w", err)
	}

	var reterr CacheErrs
	for _, hash := range hashes {
		values, err := redis.Values(conn.Receive())
		if err != nil {
			return results, fmt.Errorf("error receiving values for hash %s: %w", hash, err)
		}

		if len(values) == 0 {
			reterr = append(reterr, CacheErr(hash))
			continue
		}

		var r T
		if err := redis.ScanStruct(values, &r); err != nil {
			return results, fmt.Errorf("error scanning struct: %w", err)
		}

		results = append(results, r)
	}

	return results, reterr
}

// JsonSet the given KeyVal
func JsonSet[T any](ctx context.Context, r *RedisClient, val KeyVal[T]) error {
	conn, err := r.Connection(ctx)
	if err != nil {
		return fmt.Errorf("error getting connection from pool: %w", err)
	}
	defer conn.Close()

	jsonDoc, err := json.Marshal(val.Value)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	if _, err := conn.Do("JSON.SET", val.Key, "$", string(jsonDoc)); err != nil {
		return fmt.Errorf("error during JSON.SET with key %s: %w", val.Key, err)
	}

	return nil
}

// JsonGet the from the given document key
func JsonGet[T any](ctx context.Context, r *RedisClient, docKey string) (T, error) {
	var result T
	if shouldSkip(ctx) {
		return result, ErrSkipCache
	}

	conn, err := r.Connection(ctx)
	if err != nil {
		return result, fmt.Errorf("error getting connection from pool: %w", err)
	}
	defer conn.Close()

	doc, err := conn.Do("JSON.GET", docKey)
	if err != nil {
		return result, fmt.Errorf("error during JSON.GET with key %s: %w", docKey, err)
	}

	if doc == nil {
		return result, CacheErr(docKey)
	}

	jsonDoc, ok := doc.([]byte)
	if !ok {
		return result, errors.New("JSON.GET result is not []byte")
	}

	if err := json.Unmarshal(jsonDoc, &result); err != nil {
		return result, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return result, nil
}

func shouldSkip(ctx context.Context) bool {
	cc, ok := ctx.Value(CACHE_CONTROL_CTX_KEY).(CacheControl)
	if !ok {
		return false
	}

	return cc.Skip
}
