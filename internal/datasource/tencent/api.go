package tencent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tradebuddy/internal/model"
)

// API 腾讯股票数据API客户端
type API struct {
	client      *http.Client
	rateLimiter *RateLimiter
}

// Config API配置
type Config struct {
	RequestInterval time.Duration // 请求间隔（默认100ms）
	Timeout         time.Duration // 请求超时（默认10s）
}

// NewAPI 创建腾讯API客户端
func NewAPI(cfg *Config) *API {
	if cfg == nil {
		cfg = &Config{
			RequestInterval: 100 * time.Millisecond,
			Timeout:         10 * time.Second,
		}
	}

	return &API{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		rateLimiter: NewRateLimiter(cfg.RequestInterval),
	}
}

// GetQFQKLine 获取前复权K线数据
// code: 股票代码（如 sh600519, sz000001）
// period: 周期（day/week/month）
// limit: 获取数量（最多1000）
func (api *API) GetQFQKLine(ctx context.Context, code string, period string, limit int) ([]*model.DailyBar, error) {
	// 限流
	api.rateLimiter.Wait()

	// 转换代码格式：sh600519 -> sh600519
	// 腾讯API使用的格式
	code = strings.ToLower(code)

	// 构建URL
	// 腾讯API: http://web.ifzq.gtimg.cn/appstock/app/fqkline/get
	// 参数格式：code,period,开始日期,结束日期,limit,复权类型
	url := fmt.Sprintf(
		"http://web.ifzq.gtimg.cn/appstock/app/fqkline/get?_var=kline_day%s&param=%s,%s,,,,%d,qfq",
		code, code, period, limit,
	)

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "http://gu.qq.com")

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 去掉JSONP前缀和后缀
	// 响应格式：kline_dayshxxxxxx={"code":0,...}
	bodyStr := string(body)

	// 找到等号后的JSON部分
	startIdx := strings.Index(bodyStr, "={")
	if startIdx == -1 {
		return nil, fmt.Errorf("invalid response format: %s", bodyStr[:min(100, len(bodyStr))])
	}

	// 提取JSON（去掉最后的分号或括号）
	jsonStr := strings.TrimSpace(bodyStr[startIdx+1:])
	jsonStr = strings.TrimSuffix(jsonStr, ";")
	jsonStr = strings.TrimSuffix(jsonStr, ")")

	// 解析响应
	var result TencentResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查返回码
	if result.Code != 0 {
		return nil, fmt.Errorf("tencent api error: code=%d, msg=%s", result.Code, result.Msg)
	}

	// 提取K线数据
	bars, err := api.parseKLineData(code, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kline data: %w", err)
	}

	return bars, nil
}

// GetMultipleQFQKLine 批量获取前复权K线（带并发控制）
func (api *API) GetMultipleQFQKLine(ctx context.Context, codes []string, period string, limit int) (map[string][]*model.DailyBar, error) {
	result := make(map[string][]*model.DailyBar)

	for _, code := range codes {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		bars, err := api.GetQFQKLine(ctx, code, period, limit)
		if err != nil {
			// 记录错误但继续处理其他股票
			fmt.Printf("failed to get kline for %s: %v\n", code, err)
			continue
		}

		result[code] = bars
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseKLineData 解析K线数据
func (api *API) parseKLineData(code string, resp *TencentResponse) ([]*model.DailyBar, error) {
	// 从响应中提取数据
	dataMap, ok := resp.Data[code]
	if !ok {
		return nil, fmt.Errorf("no data for code %s", code)
	}

	// 提取qfqday或qfqweek或qfqmonth数组
	var klineData [][]string
	if data, ok := dataMap["qfqday"].([]interface{}); ok {
		klineData = convertToStringArray(data)
	} else if data, ok := dataMap["qfqweek"].([]interface{}); ok {
		klineData = convertToStringArray(data)
	} else if data, ok := dataMap["qfqmonth"].([]interface{}); ok {
		klineData = convertToStringArray(data)
	} else {
		return nil, fmt.Errorf("no kline data found")
	}

	// 解析每一行K线数据
	bars := make([]*model.DailyBar, 0, len(klineData))
	for _, row := range klineData {
		if len(row) < 6 {
			continue
		}

		// 格式：[日期, 开盘, 收盘, 最高, 最低, 成交量]
		// 例：["2024-10-08", "45.60", "46.20", "46.50", "45.30", "1234567"]
		bar, err := parseBarRow(code, row)
		if err != nil {
			continue
		}

		bars = append(bars, bar)
	}

	return bars, nil
}

// parseBarRow 解析单行K线数据
func parseBarRow(code string, row []string) (*model.DailyBar, error) {
	if len(row) < 6 {
		return nil, fmt.Errorf("invalid row length: %d", len(row))
	}

	date := row[0]
	open, _ := strconv.ParseFloat(row[1], 64)
	close, _ := strconv.ParseFloat(row[2], 64)
	high, _ := strconv.ParseFloat(row[3], 64)
	low, _ := strconv.ParseFloat(row[4], 64)
	volume, _ := strconv.ParseInt(row[5], 10, 64)

	// 成交额（如果有第7列）
	var amount float64
	if len(row) >= 7 {
		amount, _ = strconv.ParseFloat(row[6], 64)
	}

	return &model.DailyBar{
		Code:      code,
		Date:      date,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
		Amount:    amount,
		AdjFactor: 1.0, // 前复权，默认因子为1
		Source:    "tencent",
	}, nil
}

// convertToStringArray 转换interface{}数组为string数组
func convertToStringArray(data []interface{}) [][]string {
	result := make([][]string, 0, len(data))
	for _, item := range data {
		if arr, ok := item.([]interface{}); ok {
			strArr := make([]string, 0, len(arr))
			for _, v := range arr {
				strArr = append(strArr, fmt.Sprintf("%v", v))
			}
			result = append(result, strArr)
		}
	}
	return result
}

// TencentResponse 腾讯API响应结构
type TencentResponse struct {
	Code int                       `json:"code"`
	Msg  string                    `json:"msg"`
	Data map[string]map[string]interface{} `json:"data"`
}

// RateLimiter 限流器
type RateLimiter struct {
	interval time.Duration
	lastReq  time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{
		interval: interval,
		lastReq:  time.Now().Add(-interval), // 初始化为可立即请求
	}
}

// Wait 等待直到可以发送请求
func (rl *RateLimiter) Wait() {
	elapsed := time.Since(rl.lastReq)
	if elapsed < rl.interval {
		time.Sleep(rl.interval - elapsed)
	}
	rl.lastReq = time.Now()
}
