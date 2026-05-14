# fiberx

![License](https://img.shields.io/badge/License-MIT-6C757D?style=flat&color=3B82F6)
![Release](https://img.shields.io/github/v/release/gofurry/fiberx?style=flat&color=blue)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/gofurry/fiberx)](https://goreportcard.com/report/github.com/gofurry/fiberx)

[English](./README.md)

`fiberx` 是一个以 CLI 为入口的 Fiber 项目生成器仓库。

仓库当前只维护生成器主线本身：模板资源、规划规则、校验、渲染、升级评估、构建辅助和回归验证。

## 版本

- `v0.1.0`：已完成
- `v0.1.1`：已完成
- `v0.1.2`：已完成
- `v0.1.3`：已完成
- `v0.1.4`：已完成
- `v0.1.5`：进行中

## 文档入口

- [文档索引](./docs/README.md)
- [使用指南](./docs/guides/usage.md)
- [发布流程](./docs/guides/release-process.md)
- [Build Hook 安全说明](./docs/guides/build-hook-safety.md)
- [生成器架构](./docs/architecture/fiberx-generator-architecture.md)
- [模板边界](./docs/architecture/template-boundaries.md)
- [仓库规则](./docs/architecture/repository-rules.md)
- [贡献指南](./CONTRIBUTING.md)
- [变更记录](./CHANGELOG.md)
- [路线图](./docs/roadmap/roadmap.md)

## 当前生成器能力

- `medium`：稳定生产基线，默认带 Swagger 和 embedded UI
- `heavy`：生产导向轨道，默认带 Swagger、embedded UI、metrics、scheduler，可选 Redis
- `light`：轻量 HTTP 服务，保留 SQLite-first CRUD 和可选 Swagger / embedded UI
- `extra-light`：最小可启动骨架，保留 SQLite 启动、健康检查和 recover-only 中间件
- 默认栈：`Fiber v3 + Cobra + Viper`
- 兼容栈：`Fiber v2 + native-cli`
- `medium / heavy / light` 支持运行时参数：`--logger`、`--db`、`--data-access`
- 生成项目支持配置 profiles、运行时元信息、升级评估和项目级构建自动化

## 快速开始

```bash
go run ./cmd/fiberx new demo --preset medium
cd demo
go run . serve
```

兼容栈示例：

```bash
go run ./cmd/fiberx new demo-legacy --preset medium --fiber-version v2 --cli-style native
```

运行时参数示例：

```bash
go run ./cmd/fiberx new demo-data --preset medium --logger slog --db pgsql --data-access sqlx
```

构建示例：

```bash
go run ./cmd/fiberx build
go run ./cmd/fiberx build --dry-run
go run ./cmd/fiberx build --profile prod
```

## 仓库目录定位

- `sample/`：参考快照和测试对照，不是当前正式维护的 generator 主线
- `output/`：本地生成产物与临时二进制目录，除 `.gitkeep` 外默认不纳入 Git

## v0.1.4 发布范围

`v0.1.4` 已完成默认骨架公共层收口：

- 统一 `light / medium / heavy` 的 `pkg/common/error.go`
- 统一 `pkg/common/response.go` 的响应写出路径
- 清理默认 controller 的重复错误分支
- 调整业务路由入口，让 `api(...)` 接收 `AppModules`
- 保持 `extra-light` 最小面，不跟随这一版变重
- 补强共享骨架路径上的生成回归

## v0.1.5 当前范围

`v0.1.5` 当前聚焦发布表述与仓库状态对齐：

- 统一 CLI help、`validate`、`doctor` 的 release 文案
- 补齐 `CHANGELOG.md` 与发布面文档
- 让 README、文档索引、使用指南、路线图保持一致
- 继续澄清 generator 主线与参考快照之间的边界

## Build Hook 安全提示

- `fiberx build` 可能执行项目自定义 hooks
- 只应在你信任的仓库中执行这些 hooks
- 可以先用 `fiberx build --dry-run` 查看计划
- 有 hooks 的构建默认会要求确认；非交互环境下请显式使用 `--yes` 或 `--no-hooks`

## License

本项目采用 MIT License，详见 [LICENSE](./LICENSE)。
