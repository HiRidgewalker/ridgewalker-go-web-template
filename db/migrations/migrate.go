// Package migrations 嵌入 goose 迁移文件，使二进制无需随包携带 db/ 目录即可
// 执行迁移。
package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib" // 注册 "pgx" database/sql 驱动
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var files embed.FS

// versionTable 是本项目的 goose 版本表名，避免与其他项目共用默认表。
const versionTable = "learning_go_web_notebook_goose_db_version"

// Run 使用短生命周期的 database/sql 连接应用所有待执行的 up 迁移，迁移记录写入
// 本项目独立的版本表。goose 会把迁移包在事务中执行。
func Run(databaseURL string, log *slog.Logger) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open database for migration: %w", err)
	}
	defer func() { _ = db.Close() }()

	db.SetMaxOpenConns(1)

	goose.SetBaseFS(files)
	goose.SetTableName(versionTable)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Info("数据库迁移已应用")
	return nil
}
