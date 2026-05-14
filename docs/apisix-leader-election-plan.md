# APISIX Leader Election 框架层开发计划

## 背景与目标

当前多服务在 `main.go` 内直接调用 APISIX `RegisterService/DeregisterService`。
在 K8s 多副本场景下，多个 Pod 会并发写网关配置，存在以下风险：

- 多 Pod 互相覆盖 upstream/routes
- 非预期退出触发反注册，误删共享路由
- 服务间实现不一致，维护成本高

本计划目标：在 `stardust.mini` 框架层提供统一的 APISIX 注册控制能力，支持 Leader Election，保证同一服务同一时刻仅一个 Pod 写 APISIX。

## 范围

- 框架目录：`stardust.mini/register`
- 能力对象：APISIX 网关注册流程（register/reconcile/deregister）
- 使用场景：K8s 多副本部署
- 不在本期范围：业务路由规则设计、网关插件策略变更

## 设计原则

- 单一写者：仅 Leader 执行 APISIX 写操作
- 幂等优先：使用 PUT + Reconcile，避免重复写导致漂移
- 安全默认：默认不在 Pod 正常退出时删除路由
- 渐进兼容：保留 single/off 模式，兼容非 K8s 与本地调试

## 配置模型

新增统一配置项（建议挂载到 `apisix` 或 `gateway_register` 段）：

- `mode`: `off | single | leader`
- `lease_name`: 例如 `user-service-apisix-lock`
- `lease_namespace`: 默认读取 `POD_NAMESPACE`
- `lease_duration_seconds`: 默认 15
- `renew_deadline_seconds`: 默认 10
- `retry_period_seconds`: 默认 2
- `deregister_on_shutdown`: 默认 `false`
- `reconcile_interval_seconds`: 默认 30

默认策略：

- K8s 环境：`mode=leader`
- 非 K8s 环境：`mode=single`

## 分阶段实施计划

### Phase 1：接口与状态机冻结（0.5 天）

输出统一接口与行为契约：

- `GatewayRegistrar`：`Start(ctx)`、`Stop(ctx)`、`Reconcile(ctx)`
- 状态机：`Follower -> Leader -> Lost`
- 回调约定：`OnStartedLeading` 触发 register/reconcile，`OnStoppedLeading` 停止写网关

交付物：

- 接口定义
- 配置结构定义
- 状态流转说明

### Phase 2：框架能力实现（1.5 天）

在 `stardust.mini/register` 实现：

- `leader` 模式：基于 K8s Lease 的选主
- `single` 模式：直接注册（兼容当前）
- `off` 模式：禁用 APISIX 写操作
- 幂等 Reconcile：周期性 PUT upstream/routes
- 退出策略：`deregister_on_shutdown=false` 时不执行删除

交付物：

- 新增框架代码
- 单元测试（选主、丢锁、重复注册、退出行为）

### Phase 3：样板服务接入（1 天）

先接入 1 个服务（建议 `userServer`）：

- 去除 `main.go` 内直接 `RegisterService/DeregisterService`
- 使用框架 `GatewayRegistrar`
- 增加日志：当前模式、leader 身份、reconcile 成功/失败

交付物：

- 样板服务改造完成
- 样板服务测试通过

### Phase 4：14 服务批量迁移（1.5 天）

按统一模板迁移其余服务：

- 统一注册入口
- 统一配置键
- 统一默认策略 `deregister_on_shutdown=false`

交付物：

- 14 服务全部完成接入
- 编译与关键测试通过

### Phase 5：K8s 配套与验收（1 天）

补齐部署侧能力：

- ServiceAccount + RBAC（`leases.coordination.k8s.io`）
- 注入环境变量：`POD_NAME`、`POD_NAMESPACE`
- 场景验证：滚动升级、Leader 崩溃切换、全量重启

交付物：

- 验证报告
- 上线检查清单

## 验收标准

- 任一时刻同一服务仅 1 个 Pod 执行 APISIX 写操作
- 非 Leader 不触发 Register/Deregister
- 单 Pod 退出不删除共享路由
- Leader 切换后 30 秒内完成 Reconcile
- 滚动升级期间 APISIX 路由数量稳定

## 风险与应对

1. 双写冲突（bootstrap + 自注册）

- 风险：两套来源并存导致覆盖
- 应对：迁移窗口二选一，先 `mode=off` 再切 `mode=leader`

2. RBAC 缺失导致无法选主

- 风险：一直无法成为 Leader
- 应对：启动阶段输出权限检测日志，失败告警

3. 退出误删路由

- 风险：影响在线流量
- 应对：默认 `deregister_on_shutdown=false`，删除仅走显式下线流程

## 里程碑建议

- M1：框架接口与 leader 基础能力完成
- M2：样板服务接入 + 可运行验证
- M3：14 服务完成迁移 + K8s 验收通过

## 后续可扩展项

- 将 APISIX route/upstream 定义改为声明式并支持差异比对
- 引入指标：leader 切换次数、reconcile 延迟、写入失败率
- 支持按服务分组或租户分组的锁粒度策略
