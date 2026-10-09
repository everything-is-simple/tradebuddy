package datasource

import (
	"context"
	"fmt"
	"tradebuddy/internal/datasource/tdx"
	"tradebuddy/internal/datasource/tencent"
	"tradebuddy/internal/model"
)

// Manager 数据源管理器
type Manager struct {
	tdxReader   *tdx.Reader
	tencentAPI  *tencent.API
	preferTDX   bool // 优先使用TDX
}

// Config 数据源配置
type Config struct {
	TDXRoot     string // TDX根目录
	PreferTDX   bool   // 优先使用TDX本地数据
	EnableAPI   bool   // 启用在线API
}

// NewManager 创建数据源管理器
func NewManager(cfg *Config) *Manager {
	mgr := &Manager{
		preferTDX: cfg.PreferTDX,
	}

	// 初始化TDX读取器
	if cfg.TDXRoot != "" {
		mgr.tdxReader = tdx.NewReader(cfg.TDXRoot)
	}

	// 初始化腾讯API
	if cfg.EnableAPI {
		mgr.tencentAPI = tencent.NewAPI(nil)
	}

	return mgr
}

// GetDailyBars 获取日K线数据（自动选择数据源）
func (m *Manager) GetDailyBars(ctx context.Context, code string, limit int) ([]*model.DailyBar, error) {
	// 优先使用TDX
	if m.preferTDX && m.tdxReader != nil {
		if m.tdxReader.FileExists(code) {
			bars, err := m.tdxReader.ReadDayFile(ctx, code)
			if err == nil {
				// 只返回最后limit条
				if len(bars) > limit {
					bars = bars[len(bars)-limit:]
				}
				return bars, nil
			}
		}
	}

	// 回退到腾讯API
	if m.tencentAPI != nil {
		bars, err := m.tencentAPI.GetQFQKLine(ctx, code, "day", limit)
		if err == nil {
			return bars, nil
		}
	}

	// 如果之前没试过TDX，现在尝试
	if !m.preferTDX && m.tdxReader != nil {
		bars, err := m.tdxReader.ReadDayFile(ctx, code)
		if err == nil {
			if len(bars) > limit {
				bars = bars[len(bars)-limit:]
			}
			return bars, nil
		}
	}

	return nil, fmt.Errorf("failed to get daily bars for %s from all sources", code)
}

// GetMultipleDailyBars 批量获取日K线数据
func (m *Manager) GetMultipleDailyBars(ctx context.Context, codes []string, limit int) (map[string][]*model.DailyBar, error) {
	result := make(map[string][]*model.DailyBar)

	for _, code := range codes {
		bars, err := m.GetDailyBars(ctx, code, limit)
		if err != nil {
			fmt.Printf("failed to get bars for %s: %v\n", code, err)
			continue
		}
		result[code] = bars
	}

	return result, nil
}

// ListAvailableStocks 列出所有可用股票
func (m *Manager) ListAvailableStocks() ([]string, error) {
	if m.tdxReader == nil {
		return nil, fmt.Errorf("tdx reader not initialized")
	}

	return m.tdxReader.ListAllDayFiles()
}
