# Jarvis gRPC 接入设计

## 目标

将 go-cqhttp 中除图片生成以外的文本大模型调用统一替换为 Jarvis gRPC 服务，并让 Jarvis 接管对话历史和知识检索。同时增强每日华尔街见闻总结：实时写入新闻知识、在总结时按需读取相关原文、关联过往新闻形成时间线，并按固定章节向 QQ 用户发送结果。

## 范围

本次包含：

- 使用 Jarvis gRPC `Chat` 替换 ChatGPT、GPT-4、Bing Copilot、百度千帆和 DeepSeek 文本对话调用；
- 删除旧模型选择及 go-cqhttp 本地对话上下文生命周期；
- 使用 Jarvis gRPC `Ingest` 写入华尔街见闻新闻知识；
- 保留每天 23:56 的经济总结任务并改为调用 Jarvis；
- 将总结按四个业务章节发送；
- 增加 Jarvis gRPC 配置、Proto 和生成的 Go Stub；
- 更新项目认知文档；
- 通过现有 GitHub Actions 发布，并在远端 Kubernetes 环境进行启动与集成检查。

本次不包含：

- 修改 Jarvis 服务端；
- 为 Jarvis 增加 30 分钟对话 TTL；
- HTTP 客户端或 HTTP/gRPC 自动降级；
- OpenAI 图片生成迁移；
- 微博、财新、36 氪或知乎知识入库；
- 新增或运行单元测试。

## 已确认的环境事实

2026-07-29 对远端集群进行了只读检查：

- Kubernetes context 为 `kubernetes-admin@kubernetes`；
- `jarvis` 与 `qq-bot` 均位于 `default` namespace；
- Jarvis Service 名称为 `jarvis`；
- Jarvis 集群内 gRPC 端口为 `9090`；
- Jarvis 外部 gRPC NodePort 为 `30094`；
- Jarvis Service Endpoint 暴露 Pod 的 `9090` 端口；
- 检查时 Jarvis 与 qq-bot Pod 均为 Ready。

这些是带日期的运行状态，不作为永久事实。后续部署和故障排查仍需按 `AGENTS.md` 重新执行只读检查。

## 协议与 Proto

只实现 gRPC 单协议。集群内默认目标为：

```text
jarvis:9090
```

Jarvis Proto 的权威来源是：

```text
/Users/amethyst/Jarvis/jarvis/api/grpc
```

接入前从该目录核对 `jarvis.proto`。go-cqhttp 仓库保存用于构建的 Proto 副本和生成的 Go Stub；生成时通过 Go package 映射参数确定本地包路径，不改变 RPC、字段编号或字段语义。权威 Proto 发生变化时，先核对兼容性，再更新副本和 Stub。

本次只使用：

- `jarvis.JarvisService/Chat`
- `jarvis.JarvisService/Ingest`

不实现 HTTP fallback。`Chat` 没有幂等键，deadline、连接中断或结果未知时不得自动切换协议或重试。

## 配置

在 `config.yml`、`modules/config/default_config.yml` 和配置结构中增加：

```yaml
jarvis:
  grpc-target: ${JARVIS_GRPC_TARGET:jarvis:9090}
  connect-timeout-seconds: 10
  chat-timeout-seconds: 300
  ingest-timeout-seconds: 60
  summary-timeout-seconds: 600
  wallstreet-user-id: go-cqhttp:wallstreet
```

字段语义：

- `grpc-target`：Jarvis gRPC 目标；集群内默认使用 Service DNS；
- `connect-timeout-seconds`：进程启动时建立 Jarvis gRPC 连接的 deadline；
- `chat-timeout-seconds`：普通私聊和群聊 `Chat` 的 deadline；
- `ingest-timeout-seconds`：单条华尔街新闻 `Ingest` 的 deadline；
- `summary-timeout-seconds`：23:56 跨日总结 `Chat` 的 deadline，允许 Jarvis 完成相关 URL 抓取和反思；
- `wallstreet-user-id`：华尔街新闻知识和每日总结共享的稳定隔离键。

四类 timeout 必须大于零，目标和知识用户 ID 必须非空。运行时建立一个复用的明文 gRPC 连接，并在进程关闭时释放连接。

## 组件边界

### 生成协议包

在 `protocol` 下保存 Jarvis Proto 副本，在 `protocol/pb` 下保存生成的 Jarvis Go 消息和客户端 Stub。生成文件只由 Proto 生成，不手工修改。

### Jarvis 客户端适配器

新增与现有 `cluster/spider_svc` 类似的 Jarvis 适配模块，职责仅为：

- 初始化和关闭 gRPC 连接；
- 按连接、普通聊天、知识写入和总结调用类型创建对应 deadline；
- 将本地请求转换为生成的 `ChatRequest` 或 `IngestRequest`；
- 返回有类型的响应；
- 将 gRPC 错误保留为可用 `status.Code` 分类的错误。

该模块不得记录完整提示词、新闻摘要、回答正文或服务端 error details。

### 聊天助手

聊天助手只保留 Jarvis 模式：

- `#模式` 不带参数时只显示“当前助手模式：Jarvis”；
- `#模式 <编号>` 不再切换模型，统一返回“模型已由 Jarvis 服务统一管理，不支持在机器人端切换”；
- 帮助文本不再显示 ChatGPT、GPT-4、Bing Copilot、百度千帆或 DeepSeek 的模式编号；
- 删除持久化的用户模型选择；
- 删除 OpenAI、千帆和 Bing 本地会话对象、上下文数组、活动 channel 和十分钟清理 goroutine；
- 删除旧文本模型 handler；
- OpenAI 图片客户端和 `#图片` 行为保持不变。

Jarvis 对话身份：

```text
私聊：qq:private:<用户号>
群聊：qq:group:<群号>:<用户号>
```

同一用户在不同群不共享 Jarvis 对话记忆，私聊也不与群聊共享。

## 华尔街新闻知识入库

现有华尔街见闻任务在 09:05–23:55 每十分钟抓取一次。每次抓取成功后：

1. 按原始 URL 对本批新闻去重；
2. 为每条新闻构造知识文本；
3. 调用 Jarvis `Ingest`；
4. 入库失败不阻断原有 QQ 新闻推送，也不阻断其他新闻入库。

知识文本固定包含：

```text
新闻日期：<Asia/Shanghai 的 YYYY-MM-DD>
标题：<新闻标题>
摘要：<当前抓取结果中的摘要>
```

请求字段：

```text
content    = 上述知识文本
source_url = 新闻原始 URL
user_id    = jarvis.wallstreet-user-id
```

同一天重复抓取同一 URL 且标题、摘要未变化时，三项输入保持一致，使 Jarvis 的确定性知识 ID 执行覆盖式写入。新闻内容发生变化时允许形成更新后的知识点。

本阶段不要求 Jarvis 读取每条新闻原文。原文读取只在 23:56 总结时按相关性触发。

## 每日跨日总结

保留 `SummaryReporter("0 56 23 * * *")`。

总结前从当日华尔街记录中按 URL 去重，整理新闻日期、采集时间、标题、摘要和 URL。调用 `Chat` 时使用与知识入库相同的 `wallstreet-user-id`，使 Jarvis 能检索已积累的历史新闻知识和过往总结对话。

总结提示必须要求 Jarvis：

- 结合当天材料和相关历史知识；
- 识别核心事件，仅对核心事件使用 `web_scrape` 读取相关原文 URL；
- 按明确日期串联事件演化、政策变化、市场反应和因果传导；
- 区分新闻事实、分析和推测；
- 不把单日新闻简单改写为摘要；
- 严格输出四个固定章节标记。

只有满足以下条件的 URL 才进入提示词：

- scheme 为 `https`；
- host 为 `wallstreetcn.com` 或其子域名；
- URL 解析成功且不包含凭据。

这项客户端过滤只能缩小输入范围，不能修复 Jarvis `web_scrape` 缺少 DNS、私网地址和重定向校验的服务端风险。由于本次明确不修改 Jarvis，剩余风险作为已知限制保留。

## 章节输出与发送

Jarvis 必须按以下四章输出：

1. 今日全球经济形势概览
2. 重点事件、原文补充与影响分析
3. 跨日时间线及因果传导
4. 未来展望、风险与机会

提示词使用稳定、机器可识别的章节标记。本地解析器按标记提取章节，而不是按固定字符数把整篇总结分页。

发送规则：

- 每章作为独立 QQ 消息发送；
- 消息保留章节标题；
- 章节之间保持原顺序；
- 只有单章超过 QQ 单条消息长度时，才在该章内部调用现有长消息分割逻辑；
- 章内分段必须保留章名和分段序号；
- 返回内容缺少必要章节标记时，不发送整段长文，改为发送简短的“总结结构异常”提示。

`low_confidence=true` 时，在第一章之前额外发送低置信度提示，但仍按章节发送有效总结。

## 错误处理与日志

### Chat

- 每次调用设置配置的 deadline；
- 不自动重试；
- 不进行协议 fallback；
- 失败时仅向用户发送简短失败提示；
- 日志记录操作名、gRPC status code 和必要的非敏感定位字段；
- 不记录完整提示词、回答、服务端 error details 或用户对话。

### Ingest

- 单条失败后继续处理下一条；
- 失败不影响原有新闻推送；
- 日志可记录经过校验的新闻 URL 和 gRPC status code；
- 不记录摘要正文或服务端 error details；
- 不做隐式无限重试。

### 启动和关闭

- 配置无效时启动失败并输出不含敏感内容的明确错误；
- gRPC 目标不可达时不得伪装为初始化成功；
- 进程关闭时释放 Jarvis ClientConn；
- 本次不增加远端 Jarvis 或 Kubernetes 配置变更。

## 验证与发布

按用户明确要求，不新增或运行单元测试。删除或调整因旧文本模型移除而失效的旧供应商测试代码，但不创建新的测试套件。

本地验证只包含：

- `gofmt`；
- Proto 副本与生成 Stub 的一致性检查；
- `go build` 编译检查；
- `git diff --check`；
- 敏感信息和提交范围检查。

发布使用项目既有路径：

1. 只提交本次相关文件，不包含用户已有的 `.dockerignore`、`.gitignore`、`config.yml`、`crun.sh`、`qq_bot.yml` 或 `qrcode.png` 改动，除非某个文件本身属于本次已确认范围；
2. 推送到 GitHub，触发 `.github/workflows/docker.yml` 构建并推送镜像；
3. 监控构建结果；
4. 由现有更新工作流刷新 Kubernetes Pod；
5. 对远端集群执行只读验证。

远端验证包括：

- 当前 context、namespace、Deployment 和回滚镜像；
- rollout 状态、Pod readiness、重启次数和实际镜像；
- qq-bot 启动日志中不存在 Jarvis 配置、Proto、连接或初始化错误；
- qq-bot Pod 能解析 `jarvis` 并连接 `jarvis:9090`；
- Jarvis 与 qq-bot Pod 均保持 Ready；
- 部署后的实际机器人私聊、群聊隔离和回答行为；
- 后续华尔街抓取周期的知识入库日志；
- 23:56 实际总结的跨日内容、相关 URL 读取和四章节发送结果。

没有新鲜远端证据时，不宣称部署或运行验证通过。

## 文档同步

实现完成后更新 `docs/CODEX_PROJECT_CONTEXT.md`：

- 助手模式从五种旧模型改为单一 Jarvis；
- 删除本地模型选择和短期对话上下文描述；
- 增加 Jarvis gRPC 客户端、Proto 和配置说明；
- 增加华尔街知识入库与跨日总结链路；
- 更新状态持久化、定时任务和运行依赖描述；
- 保留 OpenAI 图片生成现状。

## 验收标准

- 所有用户文本对话只调用 Jarvis gRPC；
- 不再存在可切换的旧文本模型模式；
- go-cqhttp 不再维护文本对话生命周期；
- 私聊、不同群聊之间的 Jarvis `user_id` 正确隔离；
- 华尔街新闻实时写入专用 Jarvis 知识空间；
- 23:56 总结使用同一知识身份并明确要求跨日时间线；
- Jarvis 只接收经过校验的相关华尔街原文 URL；
- 总结以四章独立消息发送，整篇不再做简单分页；
- OpenAI 图片生成行为未被迁移或删除；
- 配置默认指向远端已确认的集群内目标 `jarvis:9090`；
- 项目认知文档与新运行链路一致；
- GitHub Actions 构建成功，远端 rollout、启动和连通性检查具有新鲜证据。
