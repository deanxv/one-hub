package redis

import (
	"context"
	"one-api/common/config"
	"one-api/common/logger"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

var RDB *redis.Client

const Nil = redis.Nil

// InitRedisClient This function is called after init()
func InitRedisClient() (err error) {
	redisConn := viper.GetString("redis_conn_string")

	if redisConn == "" {
		logger.SysLog("REDIS_CONN_STRING not set, Redis is not enabled")
		return nil
	}
	if viper.GetInt("sync_frequency") == 0 {
		logger.SysLog("SYNC_FREQUENCY not set, Redis is disabled")
		return nil
	}
	logger.SysLog("Redis is enabled")
	opt, err := redis.ParseURL(redisConn)
	if err != nil {
		logger.FatalLog("failed to parse Redis connection string: " + err.Error())
		return
	}

	opt.DB = viper.GetInt("redis_db")
	RDB = redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = RDB.Ping(ctx).Result()
	if err != nil {
		logger.FatalLog("Redis ping test failed: " + err.Error())
	} else {
		config.RedisEnabled = true
		// for compatibility with old versions
		config.MemoryCacheEnabled = true
	}

	return err
}

func ParseRedisOption() *redis.Options {
	opt, err := redis.ParseURL(viper.GetString("redis_conn_string"))
	if err != nil {
		logger.FatalLog("failed to parse Redis connection string: " + err.Error())
	}
	return opt
}

func RedisSet(key string, value string, expiration time.Duration) error {
	ctx := context.Background()
	return RDB.Set(ctx, key, value, expiration).Err()
}

func RedisGet(key string) (string, error) {
	ctx := context.Background()
	return RDB.Get(ctx, key).Result()
}

func RedisDel(key string) error {
	ctx := context.Background()
	return RDB.Del(ctx, key).Err()
}

func RedisDecrease(key string, value int64) error {
	ctx := context.Background()
	return RDB.DecrBy(ctx, key, value).Err()
}

func NewScript(script string) *redis.Script {
	return redis.NewScript(script)
}

func GetRedisClient() *redis.Client {
	return RDB
}

func ScriptRunCtx(ctx context.Context, script *redis.Script, keys []string, args ...interface{}) (interface{}, error) {
	return script.Run(ctx, RDB, keys, args...).Result()
}

func RedisExists(key string) (bool, error) {
	ctx := context.Background()
	exists, err := RDB.Exists(ctx, key).Result()
	return exists > 0, err
}

func RedisSAdd(key string, members ...interface{}) error {
	ctx := context.Background()
	return RDB.SAdd(ctx, key, members...).Err()
}

func RedisSIsMember(key string, member interface{}) (bool, error) {
	ctx := context.Background()
	return RDB.SIsMember(ctx, key, member).Result()
}

const (
	// StickySessionKeyPrefixClaudeCode ClaudeCode 渠道的 sticky session key 前缀
	StickySessionKeyPrefixClaudeCode = "sticky_session:"
	// StickySessionKeyPrefixGeminiCli GeminiCli 渠道的 sticky session key 前缀
	StickySessionKeyPrefixGeminiCli = "sticky_session_gemini:"
	// StickySessionKeyPrefixCodex Codex 渠道的 sticky session key 前缀
	StickySessionKeyPrefixCodex = "sticky_session_codex:"
	// DefaultStickySessionTTL 默认 TTL（1小时）
	DefaultStickySessionTTL = 1 * time.Hour
)

// GetStickySessionKeyPrefix 根据渠道类型获取对应的 Redis key 前缀
// 不同渠道类型使用不同的前缀，避免 session hash 冲突
func GetStickySessionKeyPrefix(channelType int) string {
	switch channelType {
	case config.ChannelTypeClaudeCode:
		return StickySessionKeyPrefixClaudeCode
	case config.ChannelTypeGeminiCli:
		return StickySessionKeyPrefixGeminiCli
	case config.ChannelTypeCodex:
		return StickySessionKeyPrefixCodex
	default:
		return StickySessionKeyPrefixClaudeCode // 默认使用 ClaudeCode 前缀
	}
}

// SetStickySessionMapping 设置 sticky session 映射
// sessionHash: session 的哈希值
// channelID: 渠道 ID
// channelType: 渠道类型（用于确定 key 前缀）
// ttl: 过期时间，如果为 0 则使用默认值（1小时）
func SetStickySessionMapping(sessionHash string, channelID int, channelType int, ttl time.Duration) error {
	if !config.RedisEnabled || RDB == nil {
		return nil // Redis 未启用，静默返回
	}

	if sessionHash == "" {
		return nil // 空 session hash，不需要设置
	}

	if ttl == 0 {
		ttl = DefaultStickySessionTTL
	}

	ctx := context.Background()
	keyPrefix := GetStickySessionKeyPrefix(channelType)
	key := keyPrefix + sessionHash
	return RDB.Set(ctx, key, channelID, ttl).Err()
}

// GetStickySessionMapping 获取 sticky session 映射
// sessionHash: session 的哈希值
// channelType: 渠道类型（用于确定 key 前缀）
// 返回渠道 ID，如果不存在返回 0
func GetStickySessionMapping(sessionHash string, channelType int) (int, error) {
	if !config.RedisEnabled || RDB == nil {
		return 0, nil // Redis 未启用，返回 0
	}

	if sessionHash == "" {
		return 0, nil
	}

	ctx := context.Background()
	keyPrefix := GetStickySessionKeyPrefix(channelType)
	key := keyPrefix + sessionHash
	result, err := RDB.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // key 不存在
		}
		return 0, err
	}
	return result, nil
}

// DeleteStickySessionMapping 删除 sticky session 映射
// sessionHash: session 的哈希值
// channelType: 渠道类型（用于确定 key 前缀）
func DeleteStickySessionMapping(sessionHash string, channelType int) error {
	if !config.RedisEnabled || RDB == nil {
		return nil
	}

	if sessionHash == "" {
		return nil
	}

	ctx := context.Background()
	keyPrefix := GetStickySessionKeyPrefix(channelType)
	key := keyPrefix + sessionHash
	return RDB.Del(ctx, key).Err()
}

// ExtendStickySessionMappingTTL 智能续期 sticky session 映射的 TTL
// 只有当剩余 TTL 小于阈值时才续期，避免频繁操作 Redis
// sessionHash: session 的哈希值
// channelType: 渠道类型（用于确定 key 前缀）
// fullTTL: 新的 TTL
// renewalThresholdMinutes: 续期阈值（分钟），剩余 TTL 小于此值时才续期，0 表示不续期
func ExtendStickySessionMappingTTL(sessionHash string, channelType int, fullTTL time.Duration, renewalThresholdMinutes int) error {
	if !config.RedisEnabled || RDB == nil {
		return nil
	}

	if sessionHash == "" {
		return nil
	}

	if fullTTL == 0 {
		fullTTL = DefaultStickySessionTTL
	}

	// 如果阈值为 0，不执行续期（与 code-relay-demo 保持一致）
	if renewalThresholdMinutes == 0 {
		return nil
	}

	ctx := context.Background()
	keyPrefix := GetStickySessionKeyPrefix(channelType)
	key := keyPrefix + sessionHash

	// 获取当前剩余 TTL
	remainingTTL, err := RDB.TTL(ctx, key).Result()
	if err != nil {
		return err
	}

	// key 不存在或已过期
	if remainingTTL == -2*time.Second {
		return nil
	}

	// key 存在但没有 TTL（永不过期，不需要处理）
	if remainingTTL == -1*time.Second {
		return nil
	}

	// 智能续期策略：仅在剩余时间少于阈值时才续期
	renewalThreshold := time.Duration(renewalThresholdMinutes) * time.Minute
	if remainingTTL < renewalThreshold {
		return RDB.Expire(ctx, key, fullTTL).Err()
	}

	return nil
}
