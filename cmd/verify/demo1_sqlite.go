// Demo 1: 验证 Go + SQLite 读写
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func main() {
	fmt.Println("=== Demo 1: Go + SQLite 验证 ===")

	// 1. 创建数据库
	dbPath := filepath.Join("data", "test_demo1.db")
	os.MkdirAll("data", 0755)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("打开数据库失败:", err)
	}
	defer db.Close()

	// 2. 创建表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS instruments (
			code TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			board TEXT,
			list_date TEXT
		)
	`)
	if err != nil {
		log.Fatal("创建表失败:", err)
	}

	// 3. 插入数据
	testStocks := []struct {
		code      string
		name      string
		board     string
		listDate  string
	}{
		{"sh600519", "贵州茅台", "主板", "2001-08-27"},
		{"sz000001", "平安银行", "主板", "1991-04-03"},
		{"sz300750", "宁德时代", "创业板", "2018-06-11"},
		{"sh688981", "中芯国际", "科创板", "2020-07-16"},
	}

	for _, stock := range testStocks {
		_, err = db.Exec(
			"INSERT OR REPLACE INTO instruments (code, name, board, list_date) VALUES (?, ?, ?, ?)",
			stock.code, stock.name, stock.board, stock.listDate,
		)
		if err != nil {
			log.Printf("插入 %s 失败: %v", stock.code, err)
		}
	}

	// 4. 查询数据
	rows, err := db.Query("SELECT code, name, board FROM instruments ORDER BY code")
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	defer rows.Close()

	fmt.Println("\n查询结果:")
	fmt.Println("代码\t\t名称\t\t板块")
	fmt.Println(string([]byte{'-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-'}))

	count := 0
	for rows.Next() {
		var code, name, board string
		if err := rows.Scan(&code, &name, &board); err != nil {
			log.Fatal("扫描失败:", err)
		}
		fmt.Printf("%s\t%s\t%s\n", code, name, board)
		count++
	}

	// 5. 验证结果
	fmt.Println(string([]byte{'-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-', '-'}))
	if count == 4 {
		fmt.Printf("✓ 验证成功: 插入4条，查询%d条\n", count)
		fmt.Printf("✓ 数据库文件: %s (%.2f KB)\n", dbPath, float64(getFileSize(dbPath))/1024)
	} else {
		fmt.Printf("✗ 验证失败: 预期4条，实际%d条\n", count)
		os.Exit(1)
	}

	fmt.Println("\n✓✓✓ Demo 1 通过：Go + SQLite 读写正常 ✓✓✓")
}

func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
