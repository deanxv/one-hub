package controller

import (
	"errors"
	"net/http"
	"one-api/model"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// validateInviteCodeFormat 验证邀请码格式
func validateInviteCodeFormat(code string) error {
	if code == "" {
		return nil // 空邀请码允许，会自动生成
	}

	// 去除前后空格
	trimmedCode := strings.TrimSpace(code)
	if trimmedCode != code {
		return errors.New("邀请码不能包含前后空格")
	}

	// 检查长度
	if len(code) < 3 || len(code) > 32 {
		return errors.New("邀请码长度必须在3-32个字符之间")
	}

	// 检查字符格式：只允许字母、数字、下划线和短横线
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", code)
	if !matched {
		return errors.New("邀请码只能包含字母、数字、下划线和短横线")
	}

	return nil
}

func GetInviteCodesList(c *gin.Context) {
	params := model.InviteCodeSearchParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	inviteCodes, err := model.GetInviteCodesList(&params)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    inviteCodes,
	})
}

func GetInviteCode(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的ID参数",
		})
		return
	}

	inviteCode, err := model.GetInviteCodeById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取邀请码失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    inviteCode,
	})
}

type CreateInviteCodeRequest struct {
	Name      string `json:"name"`
	Code      string `json:"code"` // 手动指定邀请码
	MaxUses   int    `json:"max_uses"`
	StartsAt  int64  `json:"starts_at"`  // 生效开始时间
	ExpiresAt int64  `json:"expires_at"` // 生效结束时间
	Count     int    `json:"count"`      // 批量创建数量
}

func CreateInviteCode(c *gin.Context) {
	var req CreateInviteCodeRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 邀请码格式验证
	if err := validateInviteCodeFormat(req.Code); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 参数验证（MaxUses = 0 表示无限使用，负数不允许，强制改为0）
	if req.MaxUses < 0 {
		req.MaxUses = 0
	}
	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 100 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "批量创建数量不能超过100个",
		})
		return
	}

	// 时间验证
	if req.StartsAt > 0 && req.ExpiresAt > 0 && req.StartsAt >= req.ExpiresAt {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "过期时间必须大于生效时间",
		})
		return
	}

	userId := c.GetInt("id")
	var createdCodes []string

	for i := 0; i < req.Count; i++ {
		var code string
		var err error

		// 如果手动指定了邀请码且只创建一个
		if req.Code != "" && req.Count == 1 {
			code = req.Code
			// 检查邀请码是否已存在
			var count int64
			model.DB.Model(&model.InviteCode{}).Where("code = ?", code).Count(&count)
			if count > 0 {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "邀请码已存在",
				})
				return
			}
		} else {
			// 自动生成邀请码
			code, err = model.GenerateInviteCode()
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
		}

		inviteCode := model.InviteCode{
			Code:      code,
			Name:      req.Name,
			MaxUses:   req.MaxUses,
			StartsAt:  req.StartsAt,
			ExpiresAt: req.ExpiresAt,
			Status:    model.InviteCodeStatusEnabled,
			CreatedBy: userId,
		}

		if req.Count > 1 {
			inviteCode.Name = req.Name + "_" + strconv.Itoa(i+1)
		}

		err = inviteCode.Insert()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "创建邀请码失败",
			})
			return
		}
		createdCodes = append(createdCodes, inviteCode.Code)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"codes": createdCodes,
			"count": len(createdCodes),
		},
	})
}

type UpdateInviteCodeRequest struct {
	Name      string `json:"name"`
	MaxUses   int    `json:"max_uses"`
	Status    int    `json:"status"`
	StartsAt  int64  `json:"starts_at"`
	ExpiresAt int64  `json:"expires_at"`
}

func UpdateInviteCode(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的ID参数",
		})
		return
	}

	var req UpdateInviteCodeRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	inviteCode, err := model.GetInviteCodeById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 参数验证（MaxUses = 0 表示无限使用）
	if req.MaxUses < 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "最大使用次数不能小于0",
		})
		return
	}

	// 只有当MaxUses发生变化时才检查使用次数限制
	if req.MaxUses != inviteCode.MaxUses && req.MaxUses > 0 && req.MaxUses < inviteCode.UsedCount {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "最大使用次数不能小于已使用次数",
		})
		return
	}

	if req.Status != model.InviteCodeStatusEnabled && req.Status != model.InviteCodeStatusDisabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的状态值",
		})
		return
	}

	// 时间验证
	if req.StartsAt > 0 && req.ExpiresAt > 0 && req.StartsAt >= req.ExpiresAt {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "过期时间必须大于生效时间",
		})
		return
	}

	// 如果要启用邀请码，检查使用次数是否已达上限（MaxUses = 0 表示无限使用）
	// 使用数据库中的MaxUses值，而不是请求中的值（状态切换时请求中的MaxUses为0）
	if req.Status == model.InviteCodeStatusEnabled && inviteCode.MaxUses > 0 && inviteCode.UsedCount >= inviteCode.MaxUses {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "邀请码使用次数已达上限，无法启用",
		})
		return
	}

	// 只更新非零值字段，避免状态切换时重置其他字段
	if req.Name != "" {
		inviteCode.Name = req.Name
	}
	if req.MaxUses != 0 || (req.MaxUses == 0 && req.Name != "") { // 如果传了name说明是完整更新
		inviteCode.MaxUses = req.MaxUses
	}
	inviteCode.Status = req.Status           // 状态总是更新
	if req.StartsAt != 0 || req.Name != "" { // 如果传了name说明是完整更新
		inviteCode.StartsAt = req.StartsAt
	}
	if req.ExpiresAt != 0 || req.Name != "" { // 如果传了name说明是完整更新
		inviteCode.ExpiresAt = req.ExpiresAt
	}

	err = inviteCode.Update()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// GenerateRandomInviteCode 生成随机邀请码
func GenerateRandomInviteCode(c *gin.Context) {
	code, err := model.GenerateInviteCode()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    gin.H{"code": code},
	})
}

func DeleteInviteCode(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的ID参数",
		})
		return
	}

	err = model.DeleteInviteCodeById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// BatchDeleteInviteCodes 批量删除邀请码
type BatchDeleteRequest struct {
	Ids []int `json:"ids" binding:"required"`
}

func BatchDeleteInviteCodes(c *gin.Context) {
	var req BatchDeleteRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if len(req.Ids) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请选择要删除的邀请码",
		})
		return
	}

	err = model.BatchDeleteInviteCodes(req.Ids)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "批量删除失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
