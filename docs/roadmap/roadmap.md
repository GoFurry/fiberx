# 路线图

## 当前定位

`fiberx` 当前已经不是一个“还在摸索 preset 形态”的仓库，而是一个边界相对清晰的 GoFiber 生成器主线仓库：

- `cmd/fiberx` 负责 CLI 入口与命令面
- `internal/*` 负责 manifest、planner、renderer、writer、metadata、upgrade、build 等核心流程
- `generator/assets`、`generator/presets`、`generator/capabilities`、`generator/rules` 是生成器主线资产
- `sample/` 主要是参考快照与测试辅助，不应凌驾于生成器主线之上
- `.github/workflows/ci.yml`、`internal/core/generate_test.go`、`cmd/fiberx/main_test.go` 已承担主回归覆盖

当前仓库已经稳定支持这些产品面：

- 4 个官方 preset：`heavy`、`medium`、`light`、`extra-light`
- 3 个已实现 capability：`redis`、`swagger`、`embedded-ui`
- 运行时选项：`fiber v2/v3`、`cobra/native`、`logger`、`db`、`data-access`、`json-lib`
- 项目元数据、`inspect` / `diff` / `upgrade inspect` / `upgrade plan`
- 项目级构建编排：`fiberx build`
- 生成器自检：`validate`、`doctor`

从代码与测试来看，原本定义给 `v0.1.4` 的“公共错误/响应层收口、`AppModules` 路由收口、controller 默认代码去重、回归补强”已经基本落在生成器主线上，`v0.1.5` 也已经完成了第一轮 release-surface 对齐。当前更明显的剩余问题转向：

- `sample/v3` 与当前模板输出存在代际差异，不适合作为唯一事实来源
- 根级回归矩阵耗时较长，需要更清晰的分层与维护边界
- 后续 capability 扩展应建立在更稳定的样例与验证策略之上

## 路线策略

优先级顺序如下：

1. 先把“仓库对外叙事”与“生成器真实状态”对齐，避免文档、CLI 输出、测试期望继续误导维护者。
2. 再收口 `sample/` 与回归测试边界，明确谁是 source of truth，谁只是快照或示例。
3. 在发布面、测试面、样例面稳定之后，再引入新的低耦合 capability。
4. `v1.0.0` 之前不追求平台化扩张，继续坚持小而稳的生成器主线。

## 已完成里程碑

### v0.1.0 - 主线收缩为生成器仓库

**Status:** Completed  
**Scope:** User-facing / Architecture / Documentation  
**Goal:** 建立 `fiberx` 作为 CLI-first 生成器仓库的主边界。

#### Focus

- 预设体系定型
- capability 边界定型
- 生成器主线与历史内容分离

#### Tasks

- [x] 固定 4 个官方 preset：`heavy`、`medium`、`light`、`extra-light`
- [x] 固定 3 个稳定 capability：`redis`、`swagger`、`embedded-ui`
- [x] 建立 runtime 选项与生成器 metadata 基线
- [x] 将仓库主线收缩为 generator mainline

#### Acceptance Criteria

- 预设和 capability 可以被 CLI 直接枚举和解释
- 生成结果携带 `.fiberx/manifest.json`
- README 与架构文档能说明仓库边界

---

### v0.1.1 - 生命周期与构建安全基线

**Status:** Completed  
**Scope:** Stability / Testing / Documentation  
**Goal:** 为生成项目补齐生命周期、配置和构建安全边界。

#### Focus

- Fiber v3 生命周期补齐
- 构建 hook 信任边界
- JSON backend 选项扩展

#### Tasks

- [x] 补齐 Fiber v2 / v3 默认优雅关闭
- [x] 增加 `json-lib` 选项与 metadata 流转
- [x] 文档化 build hook trust boundary
- [x] 强化默认 middleware 组合

#### Acceptance Criteria

- 生成项目具备基础生命周期能力
- build hook 风险有明确文档说明
- 新 runtime 选项能被生成、检查和测试覆盖

---

### v0.1.2 - 共享骨架与超时/响应契约

**Status:** Completed  
**Scope:** Stability / Developer-facing / Documentation  
**Goal:** 统一 `light / medium / heavy` 的公共骨架行为。

#### Focus

- 共享常量和响应层
- timeout 路由能力
- release-facing 文档补齐

#### Tasks

- [x] 引入共享 scaffold 常量
- [x] 加入基础错误模型与响应兼容层
- [x] 为业务路由引入 timeout 支持
- [x] 将 timeout、响应契约与验证要求写入文档

#### Acceptance Criteria

- 三个主要 preset 默认行为一致且可回归验证
- 缺失路由保持 JSON `404` envelope
- 相关文档与生成结果一致

---

### v0.1.3 - CLI 预览、诊断与构建安全收口

**Status:** Completed  
**Scope:** User-facing / Testing / CI / Documentation  
**Goal:** 收口 CLI 体验、doctor/validate 诊断层与构建安全开关。

#### Focus

- 生成前预览
- 构建 hooks 安全控制
- 诊断与 explain 能力

#### Tasks

- [x] 增加 `new/init --print-plan [--json]`
- [x] 增加 `build --no-hooks` 与 `build --yes`
- [x] 为 `doctor` 增加 generator / project / standalone 分层输出
- [x] 增加 `explain matrix`
- [x] 补齐更完整的 CLI 与生成器回归测试

#### Acceptance Criteria

- CLI 可在不落盘的前提下预览计划
- 非交互环境下 hook 行为可显式控制
- 诊断输出能区分 generator 与 generated project

---

### v0.1.4 - 生成骨架公共层收口

**Status:** Completed  
**Scope:** Developer-facing / Stability / Testing / Architecture  
**Goal:** 收口默认生成骨架的公共错误层、响应层和业务路由入口。

#### Focus

- `pkg/common/error.go` 与 `response.go` 收口
- controller 默认错误路径去重
- `AppModules` 路由入口替代膨胀式顶层路由拼装

#### Tasks

- [x] 在 `light / medium / heavy` 中统一 `APIError` / `AppError` 模型
- [x] 统一 `common.Success` / `common.Error` / `NewResponse` 双入口写法
- [x] 清理 controller 中重复的错误桥接逻辑
- [x] 让 `api(...)` 接收 `AppModules`
- [x] 保持 `extra-light` 不被更重的公共层拖入耦合
- [x] 通过生成回归覆盖上述骨架收口行为

#### Acceptance Criteria

- 生成器模板中不再依赖旧的路由注册模式
- `light / medium / heavy` 的公共错误与响应契约保持一致
- `internal/core` 与 `cmd/fiberx` 回归测试能覆盖主路径

#### Notes

这一版在生成器主线层面已经基本完成，但仓库中的 release 文案、changelog、usage 文档和样例快照尚未同步，因此后续版本的重点不应再描述为“继续做公共层收口”，而应转向发布叙事与样例/测试边界治理。

## 版本计划

### v0.1.5 - 发布叙事与仓库状态对齐

**Status:** In progress  
**Scope:** Documentation / CI / Release / Developer-facing  
**Goal:** 把仓库对外表达的版本状态同步到当前真实主线，消除“代码已前进、文档仍停留在旧 milestone”的错位。

#### Focus

- release 文案同步
- changelog 与 usage 文档补齐
- 旧 milestone 文本从 CLI、测试和 docs 中清理

#### Tasks

- [x] 同步 `cmd/fiberx/main.go` 中的 `currentRelease`、`nextRelease` 与 help/doctor/validate 文案
- [x] 同步 `cmd/fiberx/main_test.go` 中绑定旧 milestone 的断言
- [x] 更新 `CHANGELOG.md`，补上 `v0.1.3` 与 `v0.1.4` 的实际收口内容
- [x] 更新 `docs/README.md`、`docs/guides/usage.md`、相关 release snapshot 文案
- [x] 检查仓库中所有非历史文档里的 `v0.1.2` / `v0.1.3` 旧状态描述，避免继续误导

#### Acceptance Criteria

- 除历史记录外，不再存在把当前 release 叙述成 `v0.1.2` / `v0.1.3` 进行中的文本
- CLI help、`validate`、`doctor`、README、docs 的版本叙事一致
- `go test ./internal/... ./cmd/...` 仍通过

#### Notes

仓库内这批 release-surface 对齐工作已经完成；在正式打出 `v0.1.5` tag 之前，当前已完成 release 仍然保持为 `v0.1.4`。

---

### v0.2.0 - 样例、快照与回归边界收口

**Status:** Planned  
**Scope:** Testing / CI / Documentation / Architecture  
**Goal:** 明确 `generator mainline`、`sample` 快照、黑盒回归三者的职责边界，降低后续维护歧义。

#### Focus

- `sample/` 的定位澄清
- 样例同步策略
- 回归测试分层

#### Tasks

- [ ] 明确 `sample/` 是“参考快照”还是“需要跟当前模板同步的演示输出”，并在文档中写清楚
- [ ] 如果继续保留同步要求，为 `sample/v3` 建立更新流程或自动校验策略
- [ ] 把生成测试、黑盒启动测试、外部依赖测试的职责边界写入 `docs/guides/verification-matrix.md`
- [ ] 评估并收口当前根级回归耗时，避免 CI 因黑盒矩阵继续膨胀
- [ ] 保证新增维护者能清楚知道应以 `generator/assets` 还是 `sample/` 作为修改依据

#### Acceptance Criteria

- 维护者可以明确回答“哪一层才是 source of truth”
- `sample/` 的漂移不再是隐性问题
- CI 与本地回归路径有清晰分层，且文档可执行

---

### v0.3.0 - 下一批低耦合 capability

**Status:** Planned  
**Scope:** User-facing / Stability / Testing / Documentation  
**Goal:** 在主线边界稳定后，再引入一批维护成本可控的低耦合能力。

#### Focus

- 小而清晰的 capability 增量
- preset 边界与默认策略
- 文档与测试同步交付

#### Tasks

- [ ] 从 `pprof`、`rate-limit`、`cors-profile` 中挑选一组可先落地的 capability
- [ ] 为新 capability 明确 allowed/default/optional preset 边界
- [ ] 为新 capability 补齐 explain、验证矩阵、生成回归与黑盒验证
- [ ] 更新 capability policy、usage、template selection 文档
- [ ] 确保 `extra-light` 继续保持最小化，不被新能力反向拖重

#### Acceptance Criteria

- 新 capability 有明确边界而不是“所有 preset 都塞进去”
- 新增能力不会破坏现有四个 preset 的可预测性
- 文档、CLI、测试矩阵同步完成

#### Notes

`sentry`、`otel-lite` 可以在这一阶段之后再评估；前提是 capability 模型、样例同步策略和测试分层已经收口。

---

### v1.0.0-alpha.1 - 稳定候选冻结

**Status:** Planned  
**Scope:** Release / Testing / Documentation / Architecture  
**Goal:** 在正式 `v1.0.0` 前冻结公开产品面，验证架构与文档是否已足够稳定。

#### Focus

- CLI 产品面冻结
- preset / capability 边界冻结
- release 工程与文档完成度审查

#### Tasks

- [ ] 冻结 `preset`、`capability`、runtime 选项的公共语义
- [ ] 复核 upgrade / diff / metadata 是否足以支撑后续稳定演进
- [ ] 完成 release 流程、自检流程、回归矩阵、架构文档的一致性检查
- [ ] 评估是否需要额外的迁移说明、兼容策略和 semver 约束说明

#### Acceptance Criteria

- 公开产品面不再频繁改名或改语义
- 文档、测试、发布流程可以支撑第一次稳定发布评审
- 不存在阻塞 `v1.0.0` 的架构级歧义

---

### v1.0.0 - 首个稳定版

**Status:** Planned  
**Scope:** Release / Documentation / Testing  
**Goal:** 发布第一个正式稳定版，并给出长期维护边界。

#### Focus

- 稳定发布
- 文档收官
- 维护承诺

#### Tasks

- [ ] 完成 alpha 反馈收口
- [ ] 发布正式 changelog、release notes 与升级说明
- [ ] 明确后续 capability 扩展与兼容性承诺

#### Acceptance Criteria

- README、docs、CLI 输出、测试矩阵、release notes 保持一致
- 核心生成路径、检查路径、构建路径都具备可重复验证的方法
- 仓库边界与维护策略足够清晰，适合长期演进

## 短期 / 中期 / 长期方向

### 短期

- 完成 `v0.1.5`
- 修正 release / milestone 叙事滞后
- 补齐 changelog 与 usage 文档
- 停止让旧版本文案继续污染 CLI 和测试

### 中期

- 完成 `v0.2.0`
- 收口 `sample/` 与 generator mainline 的关系
- 给测试矩阵和 CI 时间成本建立清晰边界
- 在边界稳定后推进第一批低耦合 capability

### 长期

- 完成 `v1.0.0-alpha.1` 到 `v1.0.0`
- 冻结公开产品面
- 把维护重点放在稳定性、验证性和可升级性，而不是平台化扩张

## 暂不推进

这些方向当前仍不适合进入主线近期版本：

- GORM
- 完整 JWT auth
- 多租户 / RBAC
- Kubernetes
- 复杂 CI/CD 编排
- 完整前后端后台框架

这些事项要么耦合过重，要么会显著抬高模板维护成本，不符合当前阶段的产品边界。
