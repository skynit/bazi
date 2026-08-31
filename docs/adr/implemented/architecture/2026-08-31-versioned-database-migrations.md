# ADR: Use one versioned database migration runner
Status: implemented

## Problem

应用启动通过 GORM `AutoMigrate` 建表，但 `src/migrations` 同时保存一套没有执行入口且已与模型字段漂移的 MySQL SQL 文件。新字段可能只进入其中一个路径，SQLite 与 MySQL 也无法共享可观察的 schema 版本。

## Decision

`src/migrations/runner.go` 是数据库 schema 的唯一 owner。应用启动只调用 `migrations.Apply`；runner 按递增版本执行迁移，并在 `schema_migrations` 记录已完成版本。首个版本建立当前持久化模型，后续结构或数据变化追加新版本，不在命令入口直接调用 `AutoMigrate`。

迁移必须能够针对一次性 SQLite 数据库运行；涉及 SQL 方言、锁、索引或并发行为时，再增加隔离的 MySQL 验证 lane。服务内置的命理知识表不写入数据库种子表。

## Alternatives considered

**继续在命令入口直接使用 AutoMigrate。** 实现较短，但没有可查询版本，模型变化何时应用无法审计，也无法表达一次性数据回填。

**执行原有 MySQL SQL 文件。** 这些文件已经与当前 GORM 模型和 SQLite 模式漂移，维护两套 schema 定义会继续产生双 owner。

## Consequences

启动迁移幂等且有明确版本，SQLite 与 MySQL 共用同一执行入口。新增持久化字段必须同步增加迁移版本和聚焦测试。GORM 在首个基线版本中仍负责生成方言相关 DDL，因此 SQLite 测试不构成 MySQL 运行证明；MySQL 专属变化需要单独验证。
