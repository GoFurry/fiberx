# 路线图

## 当前定位

`fiberx` 当前已经不是一个还在反复试探 preset 形态的仓库，而是一个边界比较清晰的生成器主线仓库：

- `cmd/fiberx` 负责 CLI 入口、诊断、inspect/diff/upgrade 与 build 编排
- `internal/*` 负责 planner、renderer、writer、metadata、upgrade、build 等核心流程
- `generator/*` 是模板、preset、capability 与规则的 source of truth
- `sample/` 是 reference-only 的参考输出，不再视作主维护面

当前主线已经稳定支持：

- 4 个官方 preset：`heavy`、`medium`、`light`、`extra-light`
- 3 个稳定 capability：`redis`、`swagger`、`embedded-ui`
- runtime 选项：`fiber-v2/v3`、`cobra/native`、`logger`、`db`、`data-access`、`json-lib`
- 项目级 metadata、`inspect`、`diff`、`upgrade inspect`、`upgrade plan`
- 项目级构建编排：`fiberx build`
- 生成器自检：`validate`、`doctor`

## 路线策略

当前阶段的优先级不是继续堆 capability，而是持续收紧维护边界：

1. 保持生成器主线是唯一可信维护面。
2. 保持 release 文案、CLI 输出、文档和测试契约一致。
3. 默认开发路径优先 fast lane，重黑盒与数据库矩阵放到显式 integration lane。
4. 在测试和文档边界稳定后，再继续扩 capability。

## 已完成里程碑

### v0.1.0 - 生成器主线定型

**Status:** Completed  
**Scope:** User-facing / Architecture / Documentation  
**Goal:** 建立 `fiberx` 作为 generator-first 仓库的主边界。

- [x] 固定 4 个官方 preset
- [x] 固定 3 个稳定 capability
- [x] 建立 runtime 选项与 metadata 基线
- [x] 将仓库主线收缩为 generator mainline

### v0.1.1 - 生命周期与构建安全基线

**Status:** Completed  
**Scope:** Stability / Testing / Documentation  
**Goal:** 为生成项目补齐生命周期、JSON backend 与 build hook 安全边界。

- [x] 补齐 Fiber v2 / v3 默认优雅关闭
- [x] 增加 `json-lib` 选项与 metadata 流转
- [x] 文档化 build hook trust boundary
- [x] 强化默认 middleware 组合

### v0.1.2 - 公共响应层与 timeout 收口

**Status:** Completed  
**Scope:** Stability / Developer-facing / Documentation  
**Goal:** 统一 `light / medium / heavy` 的公共错误、响应与 timeout 契约。

- [x] 引入共享 scaffold 常量
- [x] 加入基础错误模型与响应兼容层
- [x] 为业务路由引入 timeout 支持
- [x] 将 timeout、响应契约与验证要求写入文档

### v0.1.3 - CLI 预览、诊断与构建安全开关

**Status:** Completed  
**Scope:** User-facing / Testing / CI / Documentation  
**Goal:** 收口 CLI 体验、doctor/validate 输出与 build hook 控制面。

- [x] 增加 `new/init --print-plan [--json]`
- [x] 增加 `build --no-hooks` 与 `build --yes`
- [x] 为 `doctor` 增加 generator / project / standalone 分层输出
- [x] 增加 `explain matrix`
- [x] 补齐更完整的 CLI 与生成器回归测试

### v0.1.4 - 生成骨架公共层收口

**Status:** Completed  
**Scope:** Developer-facing / Stability / Testing / Architecture  
**Goal:** 收口默认生成骨架中的公共错误层、响应层和路由入口。

- [x] 统一 `APIError` / `AppError` 模型
- [x] 统一 `common.Success` / `common.Error` / `NewResponse` 入口
- [x] 清理 controller 中重复的错误桥接逻辑
- [x] 让 `api(...)` 接收 `AppModules`
- [x] 保持 `extra-light` 不被更重的公共层反向拖入
- [x] 通过生成回归覆盖上述收口行为

### v0.1.5 - release surface 对齐

**Status:** Completed  
**Scope:** Documentation / Release / Developer-facing  
**Goal:** 让 CLI、README、docs、changelog 与实际主线状态保持一致。

- [x] 同步 `currentRelease` / `nextRelease` 与 help/doctor/validate 文案
- [x] 同步 `cmd/fiberx/main_test.go` 中绑定旧 milestone 的断言
- [x] 更新 `CHANGELOG.md`
- [x] 更新 `docs/README.md`、`docs/guides/usage.md` 与 release snapshot 文案
- [x] 清理仓库内遗留的旧 milestone 叙事

### v0.2.0 - CI / 测试瘦身与边界重置

**Status:** Completed  
**Scope:** Testing / CI / Documentation / Architecture  
**Goal:** 把默认本地回归、PR CI、黑盒矩阵、数据库矩阵和真实构建验证拆成 fast lane 与 integration lane 两层。

#### Focus

- 默认测试契约重定义
- 黑盒 / 数据库 / 真实构建路径显式迁移到 integration
- `sample/` reference-only 边界收口

#### Tasks

- [x] 将 `go test ./...` 收口为 fast lane，不依赖 Postgres / MySQL
- [x] 新增 `go test -tags=integration ./cmd/fiberx ./internal/core` 作为 integration lane
- [x] 将 `internal/core` 中的 capability 黑盒、runtime 数据库矩阵与 generated service helper 迁移到 `*_integration_test.go`
- [x] 将 `cmd/fiberx` 中真实 build 产物、profile 与 hook 执行测试迁移到 `*_integration_test.go`
- [x] 为 fast lane 保留代表性 generated project compile smoke，并移除广域矩阵里的重复 `runGeneratedProjectTests(...)`
- [x] 重写 `.github/workflows/ci.yml` 为 fast CI，并新增 `.github/workflows/integration.yml`
- [x] 在 `CONTRIBUTING.md`、`docs/guides/verification-matrix.md`、`docs/architecture/repository-rules.md`、`docs/guides/release-process.md` 中写清楚双车道契约
- [x] 明确 `sample/` 为 reference-only，而不是 generator source of truth

#### Acceptance Criteria

- 维护者可以明确回答 fast lane 与 integration lane 的边界
- `sample/` 不再被误认为主维护面
- PR CI 默认不再依赖数据库 service，release gate 仍保留 integration 覆盖

## 后续版本

### v0.3.0 - 下一批低耦合 capability

**Status:** Planned  
**Scope:** User-facing / Stability / Testing / Documentation  
**Goal:** 在测试与维护边界稳定后，再引入一批维护成本可控的新 capability。

- [ ] 从 `pprof`、`rate-limit`、`cors-profile` 中挑选一组可先落地的 capability
- [ ] 为新 capability 明确 allowed/default/optional preset 边界
- [ ] 为新 capability 补齐 explain、验证矩阵、生成回归与黑盒验证
- [ ] 更新 capability policy、usage、template selection 文档
- [ ] 确保 `extra-light` 继续保持最小化，不被新能力反向拖重

### v1.0.0-alpha.1 - 稳定候选冻结

**Status:** Planned  
**Scope:** Release / Testing / Documentation / Architecture  
**Goal:** 在正式 `v1.0.0` 前冻结公开产品面，确认架构、文档与回归契约足够稳定。

- [ ] 冻结 `preset`、`capability`、runtime 选项的公共语义
- [ ] 复核 upgrade / diff / metadata 是否足以支撑后续稳定演进
- [ ] 完成 release 流程、自检流程、回归矩阵与架构文档的一致性检查
- [ ] 评估是否需要额外的迁移说明、兼容策略和 semver 约束说明

### v1.0.0 - 首个稳定版

**Status:** Planned  
**Scope:** Release / Documentation / Testing  
**Goal:** 发布第一个正式稳定版，并给出长期维护边界。

- [ ] 完成 alpha 反馈收口
- [ ] 发布正式 changelog、release notes 与升级说明
- [ ] 明确后续 capability 扩展与兼容性承诺

## 短期 / 中期 / 长期方向

### 短期

- 稳定 fast lane / integration lane 的维护体验
- 持续观察 nightly integration 的耗时与失败模式
- 避免让 `sample/` 再次回到 source-of-truth 角色

### 中期

- 推进 `v0.3.0`
- 在不破坏 preset 边界的前提下扩 capability
- 继续保持 release 文案、CLI 输出和测试契约同步

### 长期

- 推进 `v1.0.0-alpha.1` 到 `v1.0.0`
- 冻结公开产品面
- 把维护重点放在稳定性、可验证性和可升级性，而不是平台化膨胀

## 暂不推进

这些方向当前仍不适合进入主线近期版本：

- GORM
- 完整 JWT auth
- 多租户 / RBAC
- Kubernetes
- 复杂 CI/CD 编排
- 完整前后端后台框架

这些事项要么耦合过重，要么会显著抬高模板维护成本，不符合当前阶段的产品边界。
