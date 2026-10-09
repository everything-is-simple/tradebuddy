// Demo 2: 验证 Go 读取 Excel（立花交易计划）
package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

func main() {
	fmt.Println("=== Demo 2: Go 读取 Excel 验证 ===")

	// 读取真实的立花交易计划文件
	excelPath := filepath.Join("..", "..", "tradebuddy-archived", "glm-docs-20261009", "证据③20_00立花交易计划_10-08（回档布网+跟踪池）.xlsx")

	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		log.Fatal("打开Excel失败:", err)
	}
	defer f.Close()

	// 读取"今晚交易计划"sheet
	sheetName := "今晚交易计划"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Printf("读取sheet失败: %v，尝试第一个sheet\n", err)
		sheets := f.GetSheetList()
		if len(sheets) > 0 {
			sheetName = sheets[0]
			rows, err = f.GetRows(sheetName)
			if err != nil {
				log.Fatal("读取第一个sheet也失败:", err)
			}
		}
	}

	fmt.Printf("\n文件: %s\n", filepath.Base(excelPath))
	fmt.Printf("Sheet: %s\n", sheetName)
	fmt.Printf("总行数: %d\n\n", len(rows))

	// 解析表头（通常在第4行，header=3）
	headerRow := 3
	if len(rows) <= headerRow {
		log.Fatal("数据行不足")
	}

	headers := rows[headerRow]
	fmt.Println("表头字段:")
	for i, h := range headers {
		if h != "" {
			fmt.Printf("  列%d: %s\n", i, h)
		}
	}

	// 解析数据行
	fmt.Println("\n前5条数据:")
	fmt.Println("代码\t名称\t决策\tz值")
	fmt.Println(string([]byte{'-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-'}))

	dataCount := 0
	for i := headerRow + 1; i < len(rows) && dataCount < 5; i++ {
		row := rows[i]
		if len(row) > 2 && row[0] != "" {
			// 假设列顺序：代码、名称、决策、z值...
			code := row[0]
			name := ""
			decision := ""
			zScore := ""

			if len(row) > 1 {
				name = row[1]
			}
			if len(row) > 2 {
				decision = row[2]
			}
			if len(row) > 3 {
				zScore = row[3]
			}

			fmt.Printf("%s\t%s\t%s\t%s\n", code, name, decision, zScore)
			dataCount++
		}
	}

	fmt.Println(string([]byte{'-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-'}))
	fmt.Printf("共找到 %d 条数据记录\n", dataCount)

	if dataCount > 0 {
		fmt.Println("\n✓✓✓ Demo 2 通过：Excel 读取正常 ✓✓✓")
	} else {
		fmt.Println("\n✗ Demo 2 失败：未读取到有效数据")
	}
}
