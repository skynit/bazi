# Database Migrations

本目录拥有应用启动时执行的版本化 schema runner。

## Local Rules

- `runner.go` 是 schema 版本和执行顺序的唯一 owner；命令入口不得直接调用 `AutoMigrate`。
- 新迁移追加版本；已记录版本不得原地改写，除非用户明确授权重建所有目标数据库。
- DDL、数据回填和索引变更必须同时说明 SQLite/MySQL 适用性、失败模式和回滚策略。
- 迁移验证只能使用一次性数据库，禁止对用户或生产数据试跑。
- 种子数据与结构迁移分离，重复执行语义必须明确。

## Local Verification

在 `src/` 运行 `go test ./migrations`。SQLite 通过不证明 MySQL 行为；交接必须记录实际验证的数据库。
