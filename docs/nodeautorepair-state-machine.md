# Node AutoRepair 状态机设计

## 目的
把现有的 Node AutoRepair 控制器改造成状态机（state machine），对检测到的不健康节点按序执行：Cordon -> Drain -> Reboot -> Uncordon，保证幂等、可恢复与可观测。

## 存储方式
- 使用 Node annotation 存储状态（轻量、无需 CRD）。建议 annotation keys：
  - `nodeautorepair.oracle.com/state` — 当前状态（Detected/Cordoning/Draining/Rebooting/Uncordoning/Succeeded/Failed）
  - `nodeautorepair.oracle.com/repair-id` — 当前修复任务唯一 ID，防止并发冲突
  - `nodeautorepair.oracle.com/last-transition` — ISO8601 时间戳
  - `nodeautorepair.oracle.com/attempts` — 重试次数

## 状态定义
- Detected
- Cordoning
- Draining
- Rebooting
- Uncordoning
- Succeeded
- Failed

## 状态转换（核心逻辑）
1. Reconcile 获取 Node 并读取 annotation 状态（若无则为 `Detected`）
2. switch state:
   - Detected -> 把状态写为 `Cordoning` (记录 repair-id)
   - Cordoning -> 执行 cordon(node)；成功则 setState(`Draining`)；失败重试或 setState(`Failed`)。
   - Draining -> 执行 drain(node)；成功则 setState(`Rebooting`)；失败重试或 setState(`Failed`)。
   - Rebooting -> 调用云重启 API（或写注解由节点 agent 执行）；若操作触发成功则 setState(`Uncordoning`)；否则重试/失败。
   - Uncordoning -> 执行 uncordon(node)；成功则 setState(`Succeeded`)；否则重试/失败。
   - Succeeded/Failed -> no-op（可保留历史并触发告警或人工介入）。

每次状态写回都更新 `last-transition` 与 `attempts`。

## 幂等与重试
- 所有操作必须幂等：再次 cordon/uncordon/再调用 drain 不应出错或导致不一致。
- 默认重试策略：最大 3 次，指数退避（base=10s）。
- 每个状态设置超时：Cordoning 30s，Draining 10m，Rebooting 5m，Uncordoning 30s（可配置）。超时后进入 `Failed` 或根据策略延长。

## 安全约束
- 在 `Draining` 前检查并尊重 PodDisruptionBudget（PDB）。若 PDB 阻止迁移，则延迟重试或把任务标记为 `Failed`（可配置）。
- 忽略 DaemonSet pods 与使用本地卷的 pods（参考 `kubectl drain` 的行为）。
- 仅 leader 控制器执行实际操作（复用现有 leader election）。
- 限制并发：
  - 单节点：通过 `repair-id` 防止对同一 Node 的并发修复。
  - 全局：可配置最大并发修复数（建议通过 controller flags）。

## 实现建议（调用点）
- Cordon/Uncordon: 使用 `client-go` 更新 `Node.Spec.Unschedulable`（幂等）。
- Drain: 建议复用 `k8s.io/kubectl/pkg/drain` 的 `drain.Helper` 或参考其实现，正确处理 PDB/DaemonSet/local PV。
- Reboot: 使用仓库已有的 OCI 客户端（`pkg/oci`）调用实例重启 API；若不可用，作为后备在 Node 上写注解触发节点 agent 重启。
- Annotation 写入需实现乐观并发重试（处理 resourceVersion 冲突）。

## 监控与告警
- 为每次状态转换生成 Kubernetes Event，便于审计。
- 导出 Prometheus 指标：repair_total、repair_failures_total、repair_duration_seconds（按状态分段）。

## 测试策略
- 单元测试：每个 state handler 的成功路径、错误路径、幂等性。
- 集成测试：模拟 API server 与云端重启接口，验证从 `Detected` 到 `Succeeded` 的完整流程。
- e2e：在可选环境中注入不健康 Node 并观察自动修复行为。

## 日志与可观测字段
- 在 Node annotation 中保留 `repair-id`、`last-transition`、`attempts`。
- 详细日志包含 nodeName、repair-id、state、error、attempts 与 timestamps。

## 失败处理与人工介入
- 当状态为 `Failed` 时，保留 annotation 并生成 Event，便于运维干预。
- 提供可配置的自动重试或将失败交给人工流程（通过事件/外部系统）。

## 安装与迁移注意
- 此改造向后兼容：若未设置注解，仍可按旧逻辑检测不健康并写入初始 `Detected`。
- 文档中记录 annotation keys 与可配置参数。

## 下一步（建议实现顺序）
1. 在 `controllers/` 添加状态机 runner 与状态枚举（skeleton）。
2. 实现 `Cordoning` 与 `Draining`，并编写单元测试。
3. 实现 `Rebooting`（调用 `pkg/oci`），然后 `Uncordoning`。
4. 测试、添加 metrics 与事件、文档更新。

---

文件由开发工作记录：状态机设计与实现要点（用于实现 `controllers/nodeautorepair_controller.go` 的重构）。