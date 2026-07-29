# Codex 项目认知与工作规则设计

## 目标

为后续 Codex 新对话提供稳定、快速且安全的项目上下文入口，使其无需重新通读整个仓库即可掌握：

- 项目提供的用户功能；
- 主要目录、模块接口及依赖关系；
- 启动、消息处理、状态持久化和关闭等核心逻辑；
- 本项目的构建、镜像发布和 Kubernetes 部署流程；
- 远端环境操作约束与敏感信息边界。

本次工作只新增项目认知与协作规则文档，不修改业务逻辑、配置、CI/CD、容器镜像或远端运行环境。

## 文档结构

采用两层文档结构：

1. 根目录 `AGENTS.md`
   - 作为 Codex 自动加载的短入口；
   - 保存必须遵守的开发、测试、部署和安全规则；
   - 强制新对话先阅读 `docs/CODEX_PROJECT_CONTEXT.md`；
   - 不复制大量代码细节，避免规则被长篇项目说明淹没。
2. `docs/CODEX_PROJECT_CONTEXT.md`
   - 保存可更新的详细项目认知；
   - 描述功能、目录、核心模块接口、运行链路、部署链路和验证边界；
   - 区分“代码已确认事实”和“需要连接实际环境验证的状态”。

## `AGENTS.md` 内容设计

### 上下文加载规则

- 开始处理本仓库任务前，先阅读 `docs/CODEX_PROJECT_CONTEXT.md`。
- 涉及具体模块时仍需读取对应源码；项目认知文档是导航和基线，不替代源码。
- 当代码、工作流或部署清单发生实质变化时，同步更新项目认知文档。

### 开发与验证规则

- 遵循仓库现有 Go 代码风格和模块组织，不进行无关重构。
- 修改代码时补充或更新测试。
- 优先运行与变更直接相关的测试，再运行可行的全量检查。
- 识别 `gotest` 中依赖网络、云服务或密钥的集成测试，不把缺少外部条件误判为纯单元测试失败。
- 保留用户现有未提交变更，不覆盖或回滚无关文件。

### 部署环境元数据

- 远端服务器：`1.15.231.125`
- SSH 用户：`root`
- 项目运行数据目录：`/root/qqbot`
- 编排方式：Kubernetes
- Kubernetes 清单：仓库实际文件名为 `qq_bot.yml`
- 容器镜像仓库：阿里云杭州地域
- 构建发布入口：`.github/workflows/docker.yml`
- 构建后更新入口：`.github/workflows/update_pod.yml`

### 敏感信息规则

- 服务器密码、QQ 凭据、API Key、Cookie、云密钥、镜像仓库凭据和 GitHub Secrets 不得写入仓库。
- 不在回复、日志、补丁、提交信息或可持久化命令历史中回显敏感值。
- 远端凭据只从安全渠道或当前获授权会话获取。
- 文档只记录环境变量名称、Secret 名称和安全获取方式，不记录真实值。
- 当前 `config.yml`、`qq_bot.yml` 及本地运行文件可能包含敏感信息，读取或展示时必须脱敏。

### 远端操作规则

- 默认先执行只读检查，例如工作负载、Pod、事件、日志和镜像状态。
- 未经明确授权，不修改远端文件、不应用清单、不重启或删除工作负载、不改变 Secret。
- 需要发布时优先沿用 GitHub Actions → 阿里云镜像 → Kubernetes 更新的现有链路。
- 不以手工远端构建替代 CI，除非用户明确要求应急处理。
- 执行变更前确认 Kubernetes 上下文、命名空间、目标 Deployment 和当前镜像。
- 变更后验证 rollout、Pod 就绪状态、容器日志以及对外端口。

## `docs/CODEX_PROJECT_CONTEXT.md` 内容设计

### 项目定位与功能

文档将说明该仓库是在 go-cqhttp/MiraiGo 基础上定制的 QQ 机器人，覆盖：

- QQ 登录、好友和群消息处理；
- OneBot 风格 HTTP、正反向 WebSocket 接口；
- QQBot gRPC 接口；
- 微博、36 氪、华尔街见闻、财新、知乎等资讯查询与定时推送；
- 币价、黄金和汇率查询；
- 数学表达式与凯利公式计算；
- 多翻译引擎轮询；
- ChatGPT、GPT-4、Bing Copilot、百度千帆和 DeepSeek 对话；
- 用户模型选择、短期会话上下文和推送状态持久化；
- 腾讯云 COS 备份及本地 LevelDB 消息存储。

需要明确 AI 作图代码仍存在，但命令列表中当前被注释，不应描述为默认启用功能。

### 目录与模块

文档按调用关系说明主要目录：

- `main.go`、`cmd/gocq`：进程入口、配置、登录和重连；
- `coolq`：机器人核心、消息发送、事件、命令、AI、资讯和状态；
- `server`、`modules/api`、`modules/servers`：OneBot 通信适配器与动作路由；
- `cluster`、`bot_service`、`protocol`：gRPC 服务端、Spider 客户端和 protobuf；
- `db`：消息存储接口及 LevelDB/MongoDB 适配器；
- `util`：AI、资讯、金融、翻译、定时任务和文件工具；
- `internal`、`global`：配置运行态、缓存、自诊断、自更新和基础设施；
- `.github/workflows`、`Dockerfile`、`qq_bot.yml`：构建与部署。

### 核心运行链路

文档将覆盖以下数据流：

1. `main.main` 设置时区和日志，初始化 Spider gRPC 连接与每日备份任务。
2. `gocq.Main` 解析配置、初始化目录与数据库、恢复登录会话并加载好友/群列表。
3. `coolq.NewQQBot` 初始化状态、订阅消息事件并注册定时资讯任务。
4. `servers.Run` 根据 `config.yml` 启动 HTTP/WebSocket 通信适配器。
5. 机器人实例写入 `bot_service.QQBot` 后启动 `:9090` gRPC 服务。
6. 群聊和私聊消息进入高级事件处理，同时尝试数学计算、AI 问答和 `#命令` 分发。
7. 资讯、AI、翻译和金融模块调用外部 HTTP、SDK 或 Spider gRPC 服务。
8. 状态保存到本地文件，并按现有实现同步到腾讯云 COS；消息可写入 LevelDB。
9. 退出信号触发状态与资讯缓存保存。

同时会指出代码中大量标准 OneBot 事件订阅当前被注释，仅高级群聊/私聊消息和好友相关事件处于实际订阅状态。

### 接口与端口

- Kubernetes 清单声明容器端口 `8888` 和 gRPC 端口 `9090`。
- Service 以 NodePort 暴露：
  - `8888` → `30088`
  - `9090` → `30091`
- OneBot HTTP/WebSocket 的实际监听地址由挂载的 `config.yml` 决定；不能仅凭清单推断运行时协议配置。
- QQBot gRPC 提供发送消息、获取机器人信息、好友列表和群列表接口。
- Spider gRPC 地址来自 `SPIDER_SERVICE_ADDR`。

### 状态和持久化

文档将区分：

- Kubernetes `hostPath` 挂载：配置、地址、设备、会话、日志、数据及临时文件；
- 本地 LevelDB 消息库：`data/leveldb-v3`；
- JSON 状态：推送订阅、已发送新闻、全局状态和用户 AI 模型；
- 腾讯云 COS：状态与每日资讯备份；
- AI 会话：进程内存中的短期上下文，默认十分钟级清理。

### 构建与发布链路

文档将准确记录：

1. 推送到 `master`（仅 README 变化除外）或针对 `master` 的 PR 触发 Docker 工作流。
2. GitHub Actions 使用 Buildx 构建 `linux/amd64` 镜像。
3. 镜像推送到 `registry.cn-hangzhou.aliyuncs.com/lexmargin/qq-bot:latest`。
4. Docker 工作流完成后，更新工作流调用 `NoahAmethyst/kuber-opt-action@v1.0.0`，请求删除 `app=qq-bot` 的 Pod。
5. Kubernetes Deployment 使用 `imagePullPolicy: Always`，新 Pod 拉取 `latest`。

文档不会把“工作流完成”误写为“构建成功”：当前 `workflow_run` 未在 YAML 中显式检查 `conclusion == success`，这是需注意的现状。

### 测试与已知边界

- 记录仓库存在配置、HTTP 参数、CQCode、全局工具和内部 BTree 等本地测试。
- 标识 `gotest` 下多个测试会访问 AI、资讯、翻译、COS 或网络服务，依赖环境变量和外部可用性。
- GitHub Actions 当前仅包含镜像构建与 Pod 更新工作流，没有独立的 lint/test 工作流。
- 认知文档只陈述代码和仓库配置已确认的事实；集群当前状态、Secret、命名空间和服务器运行状态必须通过只读远端检查确认。

## 安全与错误处理

- 项目认知文档中不包含任何真实密码或密钥。
- 如自动扫描发现疑似凭据，只记录文件位置和风险类型，不复制值。
- 如果远端连接失败，区分网络、认证、Kubernetes 上下文和工作负载问题，不在失败日志中输出凭据。
- 如果仓库描述与远端状态不一致，以只读证据报告差异，在获得授权前不自动修正。

## 验证标准

实施完成后需要验证：

1. 根目录存在 `AGENTS.md`，且其中明确要求加载项目认知文档。
2. `docs/CODEX_PROJECT_CONTEXT.md` 包含功能、结构、核心逻辑和部署链路。
3. 文档引用的文件、端口、镜像和工作流名称与仓库一致。
4. 文档不包含服务器密码或已知 API/QQ/云服务敏感值。
5. 不修改用户已有的 `.dockerignore`、`.gitignore`、`config.yml`、`crun.sh`、`qq_bot.yml` 和 `qrcode.png` 变更。
6. Markdown 不含占位符、矛盾描述或未经验证的远端状态断言。
