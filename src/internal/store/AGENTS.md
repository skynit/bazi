# Stores

Store 使用 GORM 实现查询、分页和持久化，并隔离数据库细节。

## Local Rules

- 查询作用域必须包含用户/命盘所有权约束；分页的排序、总数和边界保持稳定。
- 使用 GORM/SQL 参数绑定，不拼接用户输入。
- SQLite 通过不证明 MySQL 行为；涉及 SQL 方言、并发或迁移时明确验证后端。
- Store 返回领域可理解的错误或底层错误，不在此层生成 HTTP 响应。

## Local Verification

在 `src/` 运行 `go test ./internal/store`；数据库专属行为使用明确命名的隔离数据库 lane。
