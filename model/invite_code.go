package model

import (
	"errors"
	"fmt"
	"one-api/common/config"
	"one-api/common/logger"
	"one-api/common/redis"
	"one-api/common/utils"
	"strings"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"gorm.io/gorm"
)

type InviteCode struct {
	ID          int            `json:"id"`
	Name        string         `json:"name" gorm:"type:varchar(100);default:'';index:idx_name_deleted"`             // 邀请码名称，用于搜索
	Code        string         `json:"code" gorm:"type:varchar(32);not null;index:idx_code_deleted_status"`         // 邀请码
	MaxUses     int            `json:"max_uses" gorm:"default:0;not null"`                                          // 可使用次数，0表示无限制
	UsedCount   int            `json:"used_count" gorm:"default:0;not null"`                                        // 已使用次数
	Status      int            `json:"status" gorm:"default:1;not null;index:idx_code_deleted_status"`              // 状态：1启用，2禁用
	StartsAt    int64          `json:"starts_at" gorm:"bigint;default:0"`                                           // 生效开始时间戳，0表示立即生效
	ExpiresAt   int64          `json:"expires_at" gorm:"bigint;default:0"`                                          // 生效结束时间戳，0表示永不过期
	CreatedTime int64          `json:"created_time" gorm:"bigint;not null;index:idx_deleted_created"`               // 创建时间，用于排序
	UpdatedTime int64          `json:"updated_time" gorm:"bigint;not null"`                                         // 更新时间
	CreatedBy   int            `json:"created_by" gorm:"not null"`                                                  // 创建者ID
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index:idx_code_deleted_status,idx_deleted_created,idx_name_deleted"` // 逻辑删除
}

const (
	InviteCodeStatusEnabled  = 1
	InviteCodeStatusDisabled = 2
)

// 邀请码锁机制
var inviteCodeLocks sync.Map
var inviteCodeCreateLock sync.Mutex
var redsyncInstance *redsync.Redsync
var redsyncOnce sync.Once

// InitInviteCodeLock 初始化邀请码锁系统（按照项目风格）
func InitInviteCodeLock() {
	redsyncOnce.Do(func() {
		if config.RedisEnabled && redis.RDB != nil {
			pool := goredis.NewPool(redis.RDB)
			redsyncInstance = redsync.New(pool)
			logger.SysLog("invite code distributed lock initialized with Redis")
		} else {
			logger.SysLog("invite code lock initialized with memory lock")
		}
	})
}

// LockInviteCode 对邀请码加锁
func LockInviteCode(code string) {
	lock, ok := inviteCodeLocks.Load(code)
	if !ok {
		inviteCodeCreateLock.Lock()
		defer inviteCodeCreateLock.Unlock()
		lock, ok = inviteCodeLocks.Load(code)
		if !ok {
			lock = new(sync.Mutex)
			inviteCodeLocks.Store(code, lock)
		}
	}
	lock.(*sync.Mutex).Lock()
}

// UnlockInviteCode 释放邀请码锁
func UnlockInviteCode(code string) {
	lock, ok := inviteCodeLocks.Load(code)
	if ok {
		lock.(*sync.Mutex).Unlock()
		// 可选：清理长时间未使用的锁（简单实现，避免内存泄漏）
		// 在生产环境中可以考虑更复杂的LRU清理策略
	}
}

// AcquireInviteCodeLock 获取邀请码分布式锁（使用redsync）
func AcquireInviteCodeLock(code string) (*redsync.Mutex, error) {
	if !config.RedisEnabled {
		return nil, errors.New("Redis not enabled")
	}

	if redsyncInstance == nil {
		return nil, errors.New("redsync not initialized")
	}

	lockKey := fmt.Sprintf("invite_code_lock:%s", code)
	mutex := redsyncInstance.NewMutex(lockKey,
		redsync.WithExpiry(30*time.Second),
		redsync.WithTries(1),
	)

	err := mutex.Lock()
	if err != nil {
		return nil, err
	}

	return mutex, nil
}

var allowedInviteCodeOrderFields = map[string]bool{
	"id":           true,
	"code":         true,
	"max_uses":     true,
	"used_count":   true,
	"status":       true,
	"starts_at":    true,
	"expires_at":   true,
	"created_time": true,
	"updated_time": true,
	"name":         true,
}

// InviteCodeSearchParams 邀请码搜索参数
type InviteCodeSearchParams struct {
	PaginationParams
	Keyword      string `form:"keyword"`
	Status       int    `form:"status"`
	StartsAtFrom int64  `form:"starts_at_from"`
	StartsAtTo   int64  `form:"starts_at_to"`
}

func GetInviteCodesList(params *InviteCodeSearchParams) (*DataResult[InviteCode], error) {
	var inviteCodes []*InviteCode
	db := DB

	// 关键词搜索 - 使用索引优化
	if keyword := strings.TrimSpace(params.Keyword); keyword != "" {
		db = db.Where("code LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态过滤 - 使用索引
	if params.Status > 0 {
		db = db.Where("status = ?", params.Status)
	}

	// 生效时间范围过滤 - 查询生效期与指定时间区间有交集的邀请码
	if params.StartsAtFrom > 0 && params.StartsAtTo > 0 {
		// 验证时间范围的合理性
		if params.StartsAtFrom >= params.StartsAtTo {
			return nil, errors.New("开始时间必须小于结束时间")
		}
		// 两个时间区间有交集的条件：
		// 查询开始时间 <= 邀请码结束时间 AND 查询结束时间 >= 邀请码开始时间
		// 特殊处理：starts_at=0表示立即生效，expires_at=0表示永不过期
		db = db.Where(
			"(? <= CASE WHEN expires_at = 0 THEN ? ELSE expires_at END) AND (? >= CASE WHEN starts_at = 0 THEN 0 ELSE starts_at END)",
			params.StartsAtFrom, params.StartsAtTo, params.StartsAtTo,
		)
	} else if params.StartsAtFrom > 0 {
		// 只有开始时间：邀请码的结束时间 >= 查询开始时间
		db = db.Where("expires_at = 0 OR expires_at >= ?", params.StartsAtFrom)
	} else if params.StartsAtTo > 0 {
		// 只有结束时间：邀请码的开始时间 <= 查询结束时间
		db = db.Where("starts_at = 0 OR starts_at <= ?", params.StartsAtTo)
	}

	return PaginateAndOrder[InviteCode](db, &params.PaginationParams, &inviteCodes, allowedInviteCodeOrderFields)
}

func GetInviteCodeById(id int) (*InviteCode, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	inviteCode := InviteCode{ID: id}
	err := DB.First(&inviteCode, "id = ?", id).Error
	return &inviteCode, err
}

// CheckInviteCode 只验证邀请码有效性，不增加使用次数
func CheckInviteCode(code string) error {
	if code == "" {
		return errors.New("邀请码不能为空")
	}

	var inviteCode InviteCode
	err := DB.Where("code = ?", code).First(&inviteCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("邀请码无效")
		}
		return err
	}

	// 检查邀请码状态
	if inviteCode.Status != InviteCodeStatusEnabled {
		return errors.New("邀请码无效")
	}

	now := utils.GetTimestamp()

	// 检查是否已生效
	if inviteCode.StartsAt > 0 && inviteCode.StartsAt > now {
		return errors.New("邀请码无效")
	}

	// 检查是否过期
	if inviteCode.ExpiresAt > 0 && inviteCode.ExpiresAt < now {
		return errors.New("邀请码无效")
	}

	// 检查使用次数（MaxUses = 0 表示无限使用）
	if inviteCode.MaxUses > 0 && inviteCode.UsedCount >= inviteCode.MaxUses {
		return errors.New("邀请码无效")
	}

	return nil
}

// UseInviteCode 使用邀请码，增加使用次数
func UseInviteCode(code string) error {
	if code == "" {
		return errors.New("邀请码不能为空")
	}

	// 优先使用Redis分布式锁，失败时使用内存锁
	if config.RedisEnabled {
		mutex, err := AcquireInviteCodeLock(code)
		if err == nil && mutex != nil {
			defer func() {
				unlockOk, unlockErr := mutex.Unlock()
				if unlockErr != nil || !unlockOk {
					logger.SysError(fmt.Sprintf("failed to unlock invite code %s: ok=%v, err=%v", code, unlockOk, unlockErr))
				}
			}()
		} else {
			// Redis锁失败，记录警告并降级到内存锁
			logger.SysError(fmt.Sprintf("failed to acquire Redis lock for invite code %s, fallback to memory lock: %v", code, err))
			LockInviteCode(code)
			defer UnlockInviteCode(code)
		}
	} else {
		// 无Redis时使用内存锁
		LockInviteCode(code)
		defer UnlockInviteCode(code)
	}

	var inviteCode InviteCode
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 查询邀请码
		err := tx.Where("code = ?", code).First(&inviteCode).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("邀请码无效")
			}
			return err
		}

		// 检查邀请码状态
		if inviteCode.Status != InviteCodeStatusEnabled {
			return errors.New("邀请码无效")
		}

		now := utils.GetTimestamp()

		// 检查是否已生效
		if inviteCode.StartsAt > 0 && inviteCode.StartsAt > now {
			return errors.New("邀请码无效")
		}

		// 检查是否过期
		if inviteCode.ExpiresAt > 0 && inviteCode.ExpiresAt < now {
			return errors.New("邀请码无效")
		}

		// 检查使用次数（MaxUses = 0 表示无限使用）
		if inviteCode.MaxUses > 0 && inviteCode.UsedCount >= inviteCode.MaxUses {
			return errors.New("邀请码无效")
		}

		// 增加使用次数
		inviteCode.UsedCount++
		inviteCode.UpdatedTime = utils.GetTimestamp()

		// 如果使用次数达到上限，自动禁用（MaxUses = 0 表示无限使用，不禁用）
		if inviteCode.MaxUses > 0 && inviteCode.UsedCount >= inviteCode.MaxUses {
			inviteCode.Status = InviteCodeStatusDisabled
		}

		return tx.Save(&inviteCode).Error
	})

	return err
}

// UseInviteCodeWithTx 在指定事务中使用邀请码，增加使用次数
// 注意：此函数在事务中调用，外层已有锁保护
func UseInviteCodeWithTx(tx *gorm.DB, code string) error {
	if code == "" {
		return errors.New("邀请码不能为空")
	}

	var inviteCode InviteCode
	// 查询邀请码（在事务中已经有外部锁保护）
	err := tx.Where("code = ?", code).First(&inviteCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("邀请码无效")
		}
		return err
	}

	// 检查邀请码状态
	if inviteCode.Status != InviteCodeStatusEnabled {
		return errors.New("邀请码无效")
	}

	now := utils.GetTimestamp()

	// 检查是否已生效
	if inviteCode.StartsAt > 0 && inviteCode.StartsAt > now {
		return errors.New("邀请码无效")
	}

	// 检查是否过期
	if inviteCode.ExpiresAt > 0 && inviteCode.ExpiresAt < now {
		return errors.New("邀请码无效")
	}

	// 检查使用次数（MaxUses = 0 表示无限使用）
	if inviteCode.MaxUses > 0 && inviteCode.UsedCount >= inviteCode.MaxUses {
		return errors.New("邀请码无效")
	}

	// 增加使用次数
	inviteCode.UsedCount++
	inviteCode.UpdatedTime = utils.GetTimestamp()

	// 如果使用次数达到上限，自动禁用（MaxUses = 0 表示无限使用，不禁用）
	if inviteCode.MaxUses > 0 && inviteCode.UsedCount >= inviteCode.MaxUses {
		inviteCode.Status = InviteCodeStatusDisabled
	}

	return tx.Save(&inviteCode).Error
}

func (inviteCode *InviteCode) Insert() error {
	inviteCode.CreatedTime = utils.GetTimestamp()
	inviteCode.UpdatedTime = utils.GetTimestamp()
	return DB.Create(inviteCode).Error
}

func (inviteCode *InviteCode) Update() error {
	inviteCode.UpdatedTime = utils.GetTimestamp()
	return DB.Model(inviteCode).Select("name", "max_uses", "status", "starts_at", "expires_at", "updated_time").Updates(inviteCode).Error
}

func (inviteCode *InviteCode) Delete() error {
	return DB.Delete(inviteCode).Error
}

func DeleteInviteCodeById(id int) error {
	if id == 0 {
		return errors.New("id 为空！")
	}
	inviteCode := InviteCode{ID: id}
	err := DB.Where(inviteCode).First(&inviteCode).Error
	if err != nil {
		return err
	}
	return inviteCode.Delete()
}

// GenerateInviteCode 生成唯一的邀请码
func GenerateInviteCode() (string, error) {
	const maxRetries = 10
	const codeLength = 8

	for i := 0; i < maxRetries; i++ {
		code := utils.GetRandomString(codeLength) // 生成8位随机字符串

		// 使用唯一索引进行快速查询
		var count int64
		err := DB.Model(&InviteCode{}).Where("code = ?", code).Count(&count).Error
		if err != nil {
			return "", errors.New("查询邀请码失败: " + err.Error())
		}

		if count == 0 {
			return code, nil
		}
	}
	return "", errors.New("生成邀请码失败，请重试")
}

// BatchDeleteInviteCodes 批量删除邀请码
func BatchDeleteInviteCodes(ids []int) error {
	return DB.Where("id IN ?", ids).Delete(&InviteCode{}).Error
}
