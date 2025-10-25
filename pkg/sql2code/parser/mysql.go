package parser

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// GetMysqlTableInfo 从MySQL数据库获取表的CREATE TABLE DDL语句
// 该函数连接到MySQL数据库，使用SHOW CREATE TABLE命令查询指定表的完整创建语句，
// 为后续的代码生成提供原始的表结构信息。
//
// 参数:
//   dsn - 数据库连接字符串，格式为"user:password@(host:port)/dbname"
//   tableName - 要获取表结构的表名
//
// 返回值:
//   string - 表的CREATE TABLE DDL语句
//   error - 查询过程中的错误，如果成功则为nil
func GetMysqlTableInfo(dsn, tableName string) (string, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return "", fmt.Errorf("GetMysqlTableInfo error, %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SHOW CREATE TABLE `" + tableName + "`")
	if err != nil {
		return "", fmt.Errorf("query show create table error, %v", err)
	}

	defer rows.Close()
	if !rows.Next() {
		return "", fmt.Errorf("not found found table '%s'", tableName)
	}

	var table string
	var info string
	err = rows.Scan(&table, &info)
	if err != nil {
		return "", err
	}

	return info, nil
}

// GetTableInfo get table info from mysql
// Deprecated: replaced by GetMysqlTableInfo
func GetTableInfo(dsn, tableName string) (string, error) {
	return GetMysqlTableInfo(dsn, tableName)
}
