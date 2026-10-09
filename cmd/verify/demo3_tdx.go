// Demo 3: 验证通达信 .day 文件读取
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
)

const DayRecordSize = 32

type DayBar struct {
	Date   int32   // YYYYMMDD
	Open   float32
	High   float32
	Low    float32
	Close  float32
	Amount float32
	Volume int32
	Count  int32 // 成交笔数（可能未使用）
}

func main() {
	fmt.Println("=== Demo 3: 通达信 .day 文件验证 ===")

	// 尝试几个常见路径
	tdxPaths := []string{
		"H:\\new_tdx64\\vipdoc",
		"C:\\new_tdx64\\vipdoc",
		"C:\\tdx\\vipdoc",
		"D:\\tdx\\vipdoc",
	}

	var tdxRoot string
	for _, path := range tdxPaths {
		if _, err := os.Stat(path); err == nil {
			tdxRoot = path
			break
		}
	}

	if tdxRoot == "" {
		fmt.Println("⚠ 未找到通达信安装目录，跳过Demo 3")
		fmt.Println("如需验证，请手动指定TDX路径")
		return
	}

	fmt.Printf("找到通达信目录: %s\n\n", tdxRoot)

	// 读取贵州茅台 sh600519.day
	dayFilePath := filepath.Join(tdxRoot, "sh", "lday", "sh600519.day")
	bars, err := readDayFile(dayFilePath)
	if err != nil {
		log.Printf("读取失败: %v\n", err)
		// 尝试深市
		dayFilePath = filepath.Join(tdxRoot, "sz", "lday", "sz000001.day")
		bars, err = readDayFile(dayFilePath)
		if err != nil {
			log.Fatal("读取深市也失败:", err)
		}
	}

	fmt.Printf("文件: %s\n", filepath.Base(dayFilePath))
	fmt.Printf("K线数量: %d\n\n", len(bars))

	// 显示最近5根K线
	fmt.Println("最近5根K线:")
	fmt.Println("日期\t\t开盘\t最高\t最低\t收盘\t成交量")
	fmt.Println(string([]byte{'-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-'}))

	start := len(bars) - 5
	if start < 0 {
		start = 0
	}

	for i := start; i < len(bars); i++ {
		bar := bars[i]
		fmt.Printf("%d\t%.2f\t%.2f\t%.2f\t%.2f\t%.0f\n",
			bar.Date, bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)
	}

	fmt.Println(string([]byte{'-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-'}))

	if len(bars) > 0 {
		fmt.Println("\n✓✓✓ Demo 3 通过：TDX .day 文件读取正常 ✓✓✓")
	} else {
		fmt.Println("\n✗ Demo 3 失败：未读取到K线数据")
	}
}

func readDayFile(path string) ([]DayBar, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var bars []DayBar
	buf := make([]byte, DayRecordSize)

	for {
		n, err := io.ReadFull(f, buf)
		if err == io.EOF {
			break
		}
		if err != nil || n != DayRecordSize {
			return nil, fmt.Errorf("读取记录失败: %v", err)
		}

		bar := DayBar{
			Date:   int32(binary.LittleEndian.Uint32(buf[0:4])),
			Open:   float32(binary.LittleEndian.Uint32(buf[4:8])) / 100.0,
			High:   float32(binary.LittleEndian.Uint32(buf[8:12])) / 100.0,
			Low:    float32(binary.LittleEndian.Uint32(buf[12:16])) / 100.0,
			Close:  float32(binary.LittleEndian.Uint32(buf[16:20])) / 100.0,
			Amount: math.Float32frombits(binary.LittleEndian.Uint32(buf[20:24])),
			Volume: int32(binary.LittleEndian.Uint32(buf[24:28])),
			Count:  int32(binary.LittleEndian.Uint32(buf[28:32])),
		}

		bars = append(bars, bar)
	}

	return bars, nil
}
