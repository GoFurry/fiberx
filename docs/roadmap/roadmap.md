# 路线图

## 定位

`fiberx` 的目标不是演变成一个大而全的平台，而是成为一个面向 GoFiber 项目的可组合生成器，以及一个轻量、可验证的构建辅助工具。

## 当前状态

- `v0.1.0`：已完成
- `v0.1.1`：已完成
- `v0.1.2`：已完成
- `v0.1.3`：已完成
- `v0.1.4`：进行中

当前主线已经具备这些稳定能力：

- 项目生成：`new`、`init`
- 结构说明：`list`、`explain`
- 项目检查：`inspect`、`diff`
- 升级评估：`upgrade inspect`、`upgrade plan`
- 构建辅助：`build`
- 生成器自检：`validate`、`doctor`

## v0.1.3

`v0.1.3` 已完成，主要收口了 CLI 体验、构建安全边界，以及一批高优先级骨架修补：

- 生成前预览：`new/init --print-plan [--json]`
- 构建安全开关：`build --no-hooks`、`build --yes`
- `doctor` 自动区分 generator / project / standalone
- `explain matrix` 输出 preset 与 capability 支持矩阵
- `validate --verbose`、`doctor --verbose`、`explain matrix` 输出分段优化
- 生成骨架默认错误响应脱敏
- `timeout` 对多 handler 链完整生效
- `medium / heavy` 默认示例列表移除内置缓存
- 配置加载改为“默认配置 + YAML 覆盖”，显式 `false` 生效
- `light / medium / heavy` 的业务初始化改为轻量显式实例传递
- 默认 `fiberx.yaml` 不再预置空 `pre_hooks` / `post_hooks`
- SQLite 默认路径会自动创建父目录与数据库文件

## v0.1.4

`v0.1.4` 聚焦默认骨架公共层收口，不引入新的 capability：

- 收口 `light / medium / heavy` 的 `pkg/common/error.go`
- 统一 `pkg/common/response.go` 的双入口写法与错误桥接
- 清理默认 controller 的错误分支，统一走共享错误/响应层
- 调整业务路由入口，让 `api(...)` 接收 `AppModules`，避免顶层路由随业务数量膨胀
- 保持 `extra-light` 最小化，不强行吸收更重的公共层
- 持续补强生成回归，确保默认 CRUD、health、swagger、embedded-ui、metrics 行为不回退

明确不在这一版推进：

- `pprof`
- `rate-limit`
- `cors-profile`
- 模板系统内部的大规模重构

## 后续能力方向

优先考虑这些与主线耦合适中、维护成本可控的能力：

- `pprof`
- `rate-limit`
- `cors-profile`
- `sentry`
- `otel-lite`

暂不推进这些高耦合或高维护项：

- GORM
- 完整 JWT auth
- 多租户 / RBAC
- Kubernetes
- 复杂 CI/CD
- 完整前后端后台

## 模板系统后续打磨

后续会继续减少脆弱的字符串替换，逐步收口为更稳定的模板体系：

- `*.tmpl`：模板渲染
- `*.snippet`：仅注入片段，不直接输出文件
- 普通文件：原样复制

注入点会继续收敛为显式标记，例如：

```go
// fiberx:inject imports
// fiberx:inject bootstrap
// fiberx:inject routes
```
