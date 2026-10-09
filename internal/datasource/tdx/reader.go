package tdx

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tradebuddy/internal/model"
)

// Reader TDX数据文件读取器
type Reader struct {
	rootPath string // TDX根目录，如 H:\new_tdx64\vipdoc
}

// NewReader 创建TDX读取器
func NewReader(rootPath string) *Reader {
	return &Reader{
		rootPath: rootPath,
	}
}

// ReadDayFile 读取单个.day文件
// code: 股票代码（如 sh600519, sz000001）
func (r *Reader) ReadDayFile(ctx context.Context, code string) ([]*model.DailyBar, error) {
	// 构建文件路径
	filePath := r.buildFilePath(code)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// 获取文件大小
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// 计算记录数（每条记录32字节）
	const recordSize = 32
	recordCount := stat.Size() / recordSize

	// 读取所有记录
	bars := make([]*model.DailyBar, 0, recordCount)
	for i := int64(0); i < recordCount; i++ {
		select {
		case <-ctx.Done():
			return bars, ctx.Err()
		default:
		}

		bar, err := r.readRecord(file, code)
		if err != nil {
			if err == io.EOF {
				break
			}
			// 跳过损坏的记录
			continue
		}

		bars = append(bars, bar)
	}

	return bars, nil
}

// ReadMultipleDayFiles 批量读取.day文件
func (r *Reader) ReadMultipleDayFiles(ctx context.Context, codes []string) (map[string][]*model.DailyBar, error) {
	result := make(map[string][]*model.DailyBar)

	for _, code := range codes {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		bars, err := r.ReadDayFile(ctx, code)
		if err != nil {
			fmt.Printf("failed to read day file for %s: %v\n", code, err)
			continue
		}

		result[code] = bars
	}

	return result, nil
}

// readRecord 读取单条记录（32字节）
func (r *Reader) readRecord(file *os.File, code string) (*model.DailyBar, error) {
	var record dayRecord
	err := binary.Read(file, binary.LittleEndian, &record)
	if err != nil {
		return nil, err
	}

	// 转换为DailyBar
	bar := &model.DailyBar{
		Code:      code,
		Date:      formatDate(record.Date),
		Open:      float64(record.Open) / 100.0,
		High:      float64(record.High) / 100.0,
		Low:       float64(record.Low) / 100.0,
		Close:     float64(record.Close) / 100.0,
		Amount:    float64(record.Amount),
		Volume:    int64(record.Volume),
		AdjFactor: 1.0, // TDX原始数据，未复权
		Source:    "tdx",
	}

	return bar, nil
}

// buildFilePath 构建.day文件路径
func (r *Reader) buildFilePath(code string) string {
	// 代码格式：sh600519 或 sz000001
	code = strings.ToLower(code)

	var market string
	var stockCode string

	if strings.HasPrefix(code, "sh") {
		market = "sh"
		stockCode = code // sh600519.day
	} else if strings.HasPrefix(code, "sz") {
		market = "sz"
		stockCode = code // sz000001.day
	} else {
		// 默认深圳
		market = "sz"
		stockCode = "sz" + code
	}

	// 构建路径：rootPath/市场/lday/代码.day
	// 例：H:\new_tdx64\vipdoc\sh\lday\sh600519.day
	return filepath.Join(r.rootPath, market, "lday", stockCode+".day")
}

// dayRecord TDX日线记录结构（32字节）
type dayRecord struct {
	Date   uint32  // 日期（YYYYMMDD格式）
	Open   uint32  // 开盘价×100
	High   uint32  // 最高价×100
	Low    uint32  // 最低价×100
	Close  uint32  // 收盘价×100
	Amount float32 // 成交额（元）
	Volume uint32  // 成交量（手）
	Count  uint32  // 成交笔数（保留字段）
}

// formatDate 格式化日期：20241008 -> 2024-10-08
func formatDate(date uint32) string {
	dateStr := fmt.Sprintf("%d", date)
	if len(dateStr) != 8 {
		return ""
	}

	year := dateStr[0:4]
	month := dateStr[4:6]
	day := dateStr[6:8]

	return fmt.Sprintf("%s-%s-%s", year, month, day)
}

// ListAllDayFiles 列出所有.day文件
func (r *Reader) ListAllDayFiles() ([]string, error) {
	var files []string

	// 遍历sh和sz目录
	markets := []string{"sh", "sz"}
	for _, market := range markets {
		ldayDir := filepath.Join(r.rootPath, market, "lday")

		entries, err := os.ReadDir(ldayDir)
		if err != nil {
			// 目录不存在或无权限，跳过
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if strings.HasSuffix(name, ".day") {
				// 提取股票代码：sh600519.day -> sh600519
				code := strings.TrimSuffix(name, ".day")
				files = append(files, code)
			}
		}
	}

	return files, nil
}

// GetFileModTime 获取文件最后修改时间
func (r *Reader) GetFileModTime(code string) (time.Time, error) {
	filePath := r.buildFilePath(code)

	stat, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to stat file: %w", err)
	}

	return stat.ModTime(), nil
}

// FileExists 检查.day文件是否存在
func (r *Reader) FileExists(code string) bool {
	filePath := r.buildFilePath(code)
	_, err := os.Stat(filePath)
	return err == nil
}
