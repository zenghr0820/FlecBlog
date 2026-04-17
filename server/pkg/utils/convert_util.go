package utils

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"
)

// ToString 将任意类型转换为字符串
func ToString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprint(val)
	}
}

// ToStringArray 将任意类型转换为字符串数组
func ToStringArray(value any) []string {
	if value == nil {
		return []string{}
	}

	switch v := value.(type) {
	case []string:
		return v
	case string:
		return splitAndTrim(v)
	default:
		// 尝试通过反射处理切片类型
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Slice {
			res := make([]string, 0, val.Len())
			for i := 0; i < val.Len(); i++ {
				if str := ToString(val.Index(i).Interface()); str != "" {
					res = append(res, str)
				}
			}
			return res
		}

		// 其他类型转为单个元素的数组
		str := ToString(value)
		if str == "" {
			return []string{}
		}
		return []string{str}
	}
}

// ToDate 将任意类型转换为 *time.Time
func ToDate(value any) *time.Time {
	if value == nil {
		return nil
	}

	// 如果已经是 time.Time 类型，直接返回
	if t, ok := value.(time.Time); ok {
		return &t
	}

	str := ToString(value)
	if str == "" {
		return nil
	}

	// 常用日期格式列表
	formats := []string{
		time.RFC3339,                // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,            // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02 15:04:05",       // 标准日期时间
		"2006-01-02T15:04:05Z",      // ISO 8601 UTC
		"2006-01-02T15:04:05-07:00", // ISO 8601 带时区
		"2006-01-02 15:04",          // 不含秒
		"2006-01-02",                // 仅日期
		"2006/01/02 15:04:05",       // 斜杠分隔
		"2006/01/02",                // 斜杠分隔仅日期
		"01/02/2006 15:04:05",       // 美式日期
		"01/02/2006",                // 美式日期仅日期
	}

	for _, f := range formats {
		if t, err := time.Parse(f, str); err == nil {
			return &t
		}
	}

	return nil
}

// ToFieldName 将 snake_case 转换为 PascalCase（用于反射字段名匹配）
// 例如: "publish_time" -> "PublishTime"
func ToFieldName(s string) string {
	if s == "" {
		return ""
	}

	parts := strings.Split(s, "_")
	var res strings.Builder

	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		// 使用 unicode.ToUpper 替代已弃用的 strings.Title
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		res.WriteString(string(runes))
	}

	return res.String()
}

// splitAndTrim 按逗号分割字符串并去除空白
func splitAndTrim(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			res = append(res, t)
		}
	}
	return res
}
