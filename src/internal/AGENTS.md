# Internal Packages

本目录承载不可被外部模块导入的应用实现。根级和 `src/AGENTS.md` 规则继续适用。

## Package Boundaries

- `handler/` 只处理 HTTP 合同和依赖调用。
- `service/` 拥有领域计算和解释逻辑。
- `store/` 拥有查询、分页和持久化行为。
- `model/` 拥有数据库实体与共享 DTO。
- `config/`、`middleware/` 是配置和认证的横切边界。

## Local Rules

- 跨包共享类型放到实际 owner 包，禁止建立无归属的 `common`/`utils` 包。
- 错误在最了解语义的层产生；handler 只映射为稳定 HTTP 状态和响应。
- 修改跨层合同前，先定位所有生产者、序列化边界和消费者。

## Local Verification

运行发生变化的直接包；公共类型变化再补充直接消费者包。不得以全量测试掩盖 owner 不清。
