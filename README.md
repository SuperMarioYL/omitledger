[English](README.en.md) | **简体中文**

<picture>
  <source media="(max-width: 640px) and (prefers-color-scheme: dark)" srcset="assets/presentation/hero-mobile-dark.svg">
  <source media="(max-width: 640px)" srcset="assets/presentation/hero-mobile-light.svg">
  <source media="(prefers-color-scheme: dark)" srcset="assets/presentation/hero-dark.svg">
  <img src="assets/presentation/hero-light.svg" width="1000" alt="把省略的事项、理由、文件与会话记录到本地台账，供审阅者决定哪些需要重新处理。">
</picture>

**把省略的事项、理由、文件与会话记录到本地台账，供审阅者决定哪些需要重新处理。**

`v0.1.0` · `Go 1.24+` · [MIT](LICENSE)

[Website](https://omitledger.lei6393.com) · [Demo record](docs/demo-results.json)

## 为什么使用

审阅一次 Agent 修改时，已经写出的内容容易看到，主动没做的部分却常常消失。OmitLedger 让调用方在做出省略决定时记录事项和自述理由，随后通过列表和 reopen 形成可追踪的复核动作。

## 架构

<picture>
  <source media="(max-width: 640px) and (prefers-color-scheme: dark)" srcset="assets/presentation/architecture-mobile-dark.svg">
  <source media="(max-width: 640px)" srcset="assets/presentation/architecture-mobile-light.svg">
  <source media="(prefers-color-scheme: dark)" srcset="assets/presentation/architecture-dark.svg">
  <img src="assets/presentation/architecture-light.svg" width="1000" alt="Cobra CLI 把 add/list/reopen/report/export 映射到 ledger.Store。SQLite 保存 Omission 的类别、目标文件、会话、状态与复核备注。reopen 更新台账状态并把条目追加到 .omitledger/reopen.jsonl 供下一会话重注入；mcp（MCP server + TUI）仍为占位。">
</picture>

Cobra CLI 把 add/list/reopen/report/export 映射到 ledger.Store。SQLite 保存 Omission 的类别、目标文件、会话、状态与复核备注。reopen 更新台账状态并把条目追加到 `.omitledger/reopen.jsonl` 供下一会话重注入，不会自动执行被省略的任务；mcp（MCP server + TUI）仍为占位。

源码入口：[cmd/omitledger/root.go](cmd/omitledger/root.go) · [cmd/omitledger/add.go](cmd/omitledger/add.go) · [cmd/omitledger/list.go](cmd/omitledger/list.go) · [cmd/omitledger/reopen.go](cmd/omitledger/reopen.go) · [cmd/omitledger/report.go](cmd/omitledger/report.go) · [cmd/omitledger/export.go](cmd/omitledger/export.go) · [cmd/omitledger/mcp.go](cmd/omitledger/mcp.go) · [internal/ledger/store.go](internal/ledger/store.go)

## 安装

需要 Go 1.24+。使用纯 Go SQLite 驱动。下方脚本只操作临时数据库，不修改被提及的 retry.go。

```bash
git clone https://github.com/SuperMarioYL/omitledger.git
cd omitledger
go build -o bin/omitledger ./cmd/omitledger
```

## 快速开始

完整 shell 脚本创建临时台账，记录一条合成省略、查看列表、reopen 后再次查看。ID 随运行生成。

```bash
bash examples/presentation-demo.sh
```

完整输入与执行步骤见上方命令及 [Demo 记录](docs/demo-results.json)。

## 使用

```bash
./bin/omitledger init
./bin/omitledger add --item "API example" --reason "deferred for review" --file README.md --category doc
./bin/omitledger list --status open
./bin/omitledger reopen 1 --note "include a complete request"
./bin/omitledger report                      # 会话后 markdown 摘要（按类别分组 + 重请求清单）
./bin/omitledger report --session <id>       # 只汇总某个会话
./bin/omitledger export json > omissions.json  # CI merge-gate 用的 JSON 导出
```
`--item` 与 `--reason` 必填。类别约定为 test/file/section/refactor/doc/log，仓库级条目可用 `--file "*"`。reopen 可接受稳定 ID 或 # 列行号；筛选视图（如 `--status open`）中的行号同样是全量台账位置，可直接用于 reopen。

## 实际 Demo

<picture>
  <source media="(max-width: 640px) and (prefers-color-scheme: dark)" srcset="assets/presentation/process-mobile-dark.svg">
  <source media="(max-width: 640px)" srcset="assets/presentation/process-mobile-light.svg">
  <source media="(prefers-color-scheme: dark)" srcset="assets/presentation/process-dark.svg">
  <img src="assets/presentation/process-light.svg" width="1000" alt="完整 shell 脚本创建临时台账，记录一条合成省略、查看列表、reopen 后再次查看。ID 随运行生成。">
</picture>

### 记录并重新请求

同一事项从 open 变为 reopened，备注要求覆盖重试分支。

```text
$ bash examples/presentation-demo.sh
recorded omission 000T0ZPHK7EB5WN6SCHAH9546H
  item:     retry-path test
  file:     retry.go
  category: test
  reason:   deferred during initial implementation
  session:  presentation-demo
#  ID                          STATUS  CATEGORY  FILE      ITEM             REASON
1  000T0ZPHK7EB5WN6SCHAH9546H  open    test      retry.go  retry-path test  deferred during initial implementation

1 omission(s)
reopened 000T0ZPHK7EB5WN6SCHAH9546H: retry-path test
  note: cover the retry branch
#  ID                          STATUS    CATEGORY  FILE      ITEM             REASON
1  000T0ZPHK7EB5WN6SCHAH9546H  reopened  test      retry.go  retry-path test  deferred during initial implementation

1 omission(s)
```

## 能力与接入

<picture>
  <source media="(max-width: 640px) and (prefers-color-scheme: dark)" srcset="assets/presentation/integrations-mobile-dark.svg">
  <source media="(max-width: 640px)" srcset="assets/presentation/integrations-mobile-light.svg">
  <source media="(prefers-color-scheme: dark)" srcset="assets/presentation/integrations-dark.svg">
  <img src="assets/presentation/integrations-light.svg" width="1000" alt="可以在 Agent 工作流中显式调用 CLI，也可参考自带的技能片段。它依赖调用方主动记录，不会扫描代码推断遗漏，也不判断理由是否合理。">
</picture>

可以在 Agent 工作流中显式调用 CLI，也可参考自带的技能片段。它依赖调用方主动记录，不会扫描代码推断遗漏，也不判断理由是否合理。



## 配置

台账位置优先级：`--store` > `OMITLEDGER_STORE` > `.omitledger/ledger.db`（支持 `~/` 前缀展开）。会话优先读取 `OMITLEDGER_SESSION`、CLAUDE_SESSION_ID、CURSOR_SESSION_ID，最终回退 local。SQLite 使用 WAL 与 5000ms busy timeout。每次 reopen 会向台账目录下的 `.omitledger/reopen.jsonl` 追加一条 JSON 重请求事件（append-only），技能片段在下一会话开始时读取它并重新请求对应条目。

## 路线图与范围

当前可用入口为 init/add/list/reopen/report/export，reopen 经 `.omitledger/reopen.jsonl` 在下一会话重注入。MCP server、TUI 查看器与团队聚合仍未实现，不能因仓库存在内部模块就视为 CLI 已支持。

- 记录是调用方自述；缺少条目不表示没有遗漏，reopened 不表示工作已完成。
- mcp 占位命令会成功返回，但没有提供其描述的完整功能。

![Terminal recording](assets/demo.gif) · [Recording script](docs/demo.tape)

## 许可证

[MIT](LICENSE)
