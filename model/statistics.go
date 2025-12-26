package model

import (
	"fmt"
	"one-api/common"
	"os"
	"strings"
	"time"
)

type Statistics struct {
	Date             time.Time `gorm:"primary_key;type:date" json:"date"`
	UserId           int       `json:"user_id" gorm:"primary_key"`
	ChannelId        int       `json:"channel_id" gorm:"primary_key"`
	ModelName        string    `json:"model_name" gorm:"primary_key;type:varchar(255)"`
	RequestCount     int       `json:"request_count"`
	Quota            int       `json:"quota"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	RequestTime      int       `json:"request_time"`
}

func GetUserModelStatisticsByPeriod(userId int, startTime, endTime string) (LogStatistic []*LogStatisticGroupModel, err error) {
	dateStr := "date"
	if common.UsingPostgreSQL {
		dateStr = "TO_CHAR(date, 'YYYY-MM-DD') as date"
	} else if common.UsingSQLite {
		dateStr = "strftime('%Y-%m-%d', date) as date"
	} else {
		// MySQL/TiDB - 显式格式化日期以确保兼容性
		dateStr = "DATE_FORMAT(date, '%Y-%m-%d') as date"
	}

	err = DB.Raw(`
		SELECT `+dateStr+`,
		model_name, 
		sum(request_count) as request_count,
		sum(quota) as quota,
		sum(prompt_tokens) as prompt_tokens,
		sum(completion_tokens) as completion_tokens,
		sum(request_time) as request_time
		FROM statistics
		WHERE user_id= ?
		AND date BETWEEN ? AND ?
		GROUP BY date, model_name
		ORDER BY date, model_name
	`, userId, startTime, endTime).Scan(&LogStatistic).Error
	return
}

func GetChannelExpensesStatisticsByPeriod(startTime, endTime, groupType string, userID int) (LogStatistics []*LogStatisticGroupChannel, err error) {

	var whereClause strings.Builder
	whereClause.WriteString("WHERE date BETWEEN ? AND ?")
	args := []interface{}{startTime, endTime}

	if userID > 0 {
		whereClause.WriteString(" AND user_id = ?")
		args = append(args, userID)
	}

	dateStr := "date"
	if common.UsingPostgreSQL {
		dateStr = "TO_CHAR(date, 'YYYY-MM-DD') as date"
	} else if common.UsingSQLite {
		dateStr = "strftime('%%Y-%%m-%%d', date) as date"
	} else {
		// MySQL/TiDB - 显式格式化日期以确保兼容性
		dateStr = "DATE_FORMAT(date, '%%Y-%%m-%%d') as date"
	}

	baseSelect := `
        SELECT ` + dateStr + `,
        sum(request_count) as request_count,
        sum(quota) as quota,
        sum(prompt_tokens) as prompt_tokens,
        sum(completion_tokens) as completion_tokens,
        sum(request_time) as request_time,`

	var sql string
	if groupType == "model" {
		sql = baseSelect + `
            model_name as channel
            FROM statistics
            %s
            GROUP BY date, model_name
            ORDER BY date, model_name`
	} else if groupType == "model_type" {
		sql = baseSelect + `
            model_owned_by.name as channel
            FROM statistics
            JOIN prices ON statistics.model_name = prices.model
			JOIN model_owned_by ON prices.channel_type = model_owned_by.id
            %s
            GROUP BY date, model_owned_by.name
            ORDER BY date, model_owned_by.name`

	} else {
		sql = baseSelect + `
            MAX(channels.name) as channel
            FROM statistics
            JOIN channels ON statistics.channel_id = channels.id
            %s
            GROUP BY date, channel_id
            ORDER BY date, channel_id`
	}

	sql = fmt.Sprintf(sql, whereClause.String())
	err = DB.Raw(sql, args...).Scan(&LogStatistics).Error
	if err != nil {
		return nil, err
	}

	return LogStatistics, nil
}

type StatisticsUpdateType int

const (
	StatisticsUpdateTypeToDay     StatisticsUpdateType = 1
	StatisticsUpdateTypeYesterday StatisticsUpdateType = 2
	StatisticsUpdateTypeALL       StatisticsUpdateType = 3
)

func UpdateStatistics(updateType StatisticsUpdateType) error {
	sql := `
	%s statistics (date, user_id, channel_id, model_name, request_count, quota, prompt_tokens, completion_tokens, request_time)
	SELECT 
		%s as date,
		user_id,
		channel_id,
		model_name, 
		count(1) as request_count,
		sum(quota) as quota,
		sum(prompt_tokens) as prompt_tokens,
		sum(completion_tokens) as completion_tokens,
		sum(request_time) as request_time
	FROM logs
	WHERE
		type = 2
		%s
	GROUP BY date, channel_id, user_id, model_name
	ORDER BY date, model_name
	%s
	`

	sqlPrefix := ""
	sqlWhere := ""
	sqlDate := ""
	sqlSuffix := ""

	// 统一获取时区信息
	location := time.Local
	if tzEnv := os.Getenv("TZ"); tzEnv != "" {
		if loc, err := time.LoadLocation(tzEnv); err == nil {
			location = loc
		}
	}
	now := time.Now().In(location)
	_, offsetSeconds := now.Zone()

	// SQLite 需要特殊格式的偏移字符串
	getSqliteOffset := func() string {
		hours := offsetSeconds / 3600
		minutes := (offsetSeconds % 3600) / 60
		if hours >= 0 {
			offset := fmt.Sprintf("+%d hours", hours)
			if minutes != 0 {
				offset += fmt.Sprintf(" %d minutes", minutes)
			}
			return offset
		}
		offset := fmt.Sprintf("%d hours", hours)
		if minutes != 0 {
			offset += fmt.Sprintf(" %d minutes", -minutes)
		}
		return offset
	}

	if common.UsingSQLite {
		sqlPrefix = "INSERT OR REPLACE INTO"
		sqlDate = fmt.Sprintf("strftime('%%Y-%%m-%%d', datetime(created_at, 'unixepoch', '%s'))", getSqliteOffset())
		sqlSuffix = ""
	} else if common.UsingPostgreSQL {
		sqlPrefix = "INSERT INTO"
		tzName := "UTC"
		if tzEnv := os.Getenv("TZ"); tzEnv != "" {
			tzName = tzEnv
		}
		sqlDate = fmt.Sprintf("DATE_TRUNC('day', TO_TIMESTAMP(created_at) AT TIME ZONE '%s')::DATE", tzName)
		sqlSuffix = `ON CONFLICT (date, user_id, channel_id, model_name) DO UPDATE SET
		request_count = EXCLUDED.request_count,
		quota = EXCLUDED.quota,
		prompt_tokens = EXCLUDED.prompt_tokens,
		completion_tokens = EXCLUDED.completion_tokens,
		request_time = EXCLUDED.request_time`
	} else {
		sqlPrefix = "INSERT INTO"
		// MySQL: 检测 MySQL 时区，决定是否需要转换
		var mysqlTz string
		DB.Raw("SELECT @@session.time_zone").Scan(&mysqlTz)
		mysqlIsUTC := mysqlTz == "UTC" || mysqlTz == "+00:00"

		if mysqlIsUTC {
			// MySQL 是 UTC，需要转换为本地时区
			hours := offsetSeconds / 3600
			minutes := (offsetSeconds % 3600) / 60
			var tzOffset string
			if hours >= 0 {
				tzOffset = fmt.Sprintf("+%02d:%02d", hours, minutes)
			} else {
				tzOffset = fmt.Sprintf("-%02d:%02d", -hours, -minutes)
			}
			sqlDate = fmt.Sprintf("DATE(CONVERT_TZ(FROM_UNIXTIME(created_at), '+00:00', '%s'))", tzOffset)
		} else {
			// MySQL 是本地时区（SYSTEM 或 +08:00 等），直接使用
			sqlDate = "DATE(FROM_UNIXTIME(created_at))"
		}
		sqlSuffix = `ON DUPLICATE KEY UPDATE
		request_count = VALUES(request_count),
		quota = VALUES(quota),
		prompt_tokens = VALUES(prompt_tokens),
		completion_tokens = VALUES(completion_tokens),
		request_time = VALUES(request_time)`
	}

	todayTimestamp := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).Unix()

	switch updateType {
	case StatisticsUpdateTypeToDay:
		sqlWhere = fmt.Sprintf("AND created_at >= %d", todayTimestamp)
	case StatisticsUpdateTypeYesterday:
		yesterdayTimestamp := todayTimestamp - 86400
		sqlWhere = fmt.Sprintf("AND created_at >= %d AND created_at < %d", yesterdayTimestamp, todayTimestamp)
	}

	err := DB.Exec(fmt.Sprintf(sql, sqlPrefix, sqlDate, sqlWhere, sqlSuffix)).Error
	return err
}
