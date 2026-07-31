# QQ Bot

本项目基于 [go-cqhttp](https://github.com/Mrs4s/go-cqhttp) 和 MiraiGo 定制，提供 QQ 群聊/私聊命令、定时资讯、金融工具、翻译、Jarvis 文本助手、OpenAI 图片生成，以及 OneBot HTTP/WebSocket 和 QQBot gRPC 接口。

## 主要功能

- 微博、36Kr、华尔街见闻、财新、知乎资讯查询，以及群聊/私聊定时推送管理。
- BTC、ETH、BNB 等币价、黄金价格与波动提醒、汇率、数学表达式和凯利策略。
- 有道、腾讯、阿里和火山翻译适配器轮转。
- Jarvis gRPC 文本助手。
- OpenAI DALL·E 3 图片生成。OpenAI 在本项目中只用于图片，不再承担文本对话或新闻总结。
- Tencent COS 状态/新闻记录同步、OneBot HTTP/WebSocket，以及 QQBot gRPC 服务。

## Jarvis 文本助手

Jarvis gRPC 是唯一文本助手协议和唯一文本模型入口。程序在读取配置后阻塞连接 Jarvis，连接失败会停止启动；默认集群内地址为 `jarvis:9090`。Jarvis 负责模型选择、对话历史和对话生命周期，go-cqhttp 不再保存模型选择，也没有本地会话窗口。

普通问答使用以下稳定身份隔离上下文：

- 私聊：`qq:private:<user>`
- 群聊：`qq:group:<group>:<user>`

`#模式` 命令只反映这一单一模式：

- `#模式` 返回“当前助手模式：Jarvis”。
- `#模式 <任意参数>` 返回“模型已由 Jarvis 服务统一管理，不支持在机器人端切换”。

Jarvis 的权威 Proto 位于开发机目录 `/Users/amethyst/Jarvis/jarvis/api/grpc`。修改或重新生成本仓库的 `protocol/jarvis.proto` 和 Go stub 前必须先核对该目录；它是开发工作规则，不是容器运行时依赖。

## 华尔街见闻知识与每日总结

当前知识库改造只覆盖华尔街见闻。白天定时抓取任务在每轮识别新增或内容已变化的新闻后，将上海时区新闻日期、标题、摘要和通过校验的华尔街见闻 HTTPS 原文 URL 写入 Jarvis。同一 URL、标题和摘要在同一天生成稳定的知识输入；抓取阶段不会让 Jarvis 读取原文 URL。

每天 23:56 只生成一次总结，再将同一结果推送给全部群聊和私聊订阅者。总结阶段使用 Jarvis `Generate` 无记忆接口，允许 Jarvis 选择性读取与核心结论直接相关的原文 URL，并要求结合同一知识身份下的历史新闻，按明确日期串联跨日时间线和因果传导，而不是只汇总当天条目。

总结严格按以下五章逐章发送：

1. 今日全球经济形势概览
2. 重点事件、原文补充与影响分析
3. 跨日时间线及因果传导
4. 未来展望、风险与机会
5. 投资建议/注意事项

只有单章超过 QQ 消息长度限制时，才在该章内部继续分片；不会把整份总结作为一次长消息或普通连续分页发送。

## 配置

在 `config.yml` 中设置 QQ 账号、消息行为、数据库和 OneBot 服务。Jarvis 的完整配置如下：

```yaml
jarvis:
  grpc-target: ${JARVIS_GRPC_TARGET:jarvis:9090}
  connect-timeout-seconds: 10
  chat-timeout-seconds: 300
  ingest-timeout-seconds: 120
  summary-timeout-seconds: 600
  wallstreet-user-id: go-cqhttp:wallstreet
```

`${JARVIS_GRPC_TARGET:jarvis:9090}` 表示环境变量 `JARVIS_GRPC_TARGET` 非空时覆盖目标地址，否则使用 Kubernetes Service DNS `jarvis:9090`。四类超时依次用于启动连接、普通对话、知识写入和每日总结，单位均为秒。`wallstreet-user-id` 是华尔街知识与总结共用的稳定 Jarvis 身份。

## 运行环境变量

只在运行环境或 Secret 管理系统中提供值，不要提交凭证：

| 类别 | 环境变量 | 用途 |
| --- | --- | --- |
| Jarvis | `JARVIS_GRPC_TARGET` | 可选覆盖 Jarvis gRPC 目标；集群内默认无需设置。 |
| OpenAI 图片 | `OPENAI_API_KEY`, `OPENAI_PROXY`, `NOT_MIRROR` | DALL·E 3 图片生成；`NOT_MIRROR` 非空时直连 OpenAI。`OPENAI_PROXY` 只填写主机名，不含 `https://` 和 `/v1`。 |
| 翻译 | `YD_APP_KEY`, `YD_SECRET_KEY`, `TC_SECRET_ID`, `TC_SECRET_KEY`, `ALI_ACCESS_ID`, `ALI_ACCESS_SECRET`, `VOLC_ACCESS_KEY`, `VOLC_SECRET_KEY` | 初始化有道、腾讯、阿里和火山翻译适配器。 |
| 百度翻译预留 | `BAIDU_API_KEY`, `BAIDU_SECRET_KEY` | 代码保留配置读取，但当前百度翻译适配器初始化被禁用。 |
| COS | `TC_SECRET_ID`, `TC_SECRET_KEY` | Tencent COS 状态和新闻记录同步；与腾讯翻译共用变量名。 |
| 运行身份/文件 | `OWNER`, `FILE_ROOT` | Bot owner 身份和临时/状态文件根目录。 |
| 其他服务 | `SPIDER_SERVICE_ADDR`, `REMOTE_PROXY`, `NOT_MIRROR` | Spider gRPC 地址，以及币安行情镜像/直连控制。 |

旧的 ChatGPT、GPT-4、Bing Copilot、Ernie/Qianfan、DeepSeek 和 Chimera 文本模型及其凭证配置已移除。

## 构建与运行

项目声明 Go 1.19，Docker 构建使用 Go 1.19.7：

```shell
go mod download
go build
./go-cqhttp
```

首次运行会生成设备配置；也可以通过 `-c` 指定配置文件。设备协议等基础登录配置沿用上游 go-cqhttp。

本次 Jarvis 改造的本地验证范围是格式、Proto 一致性和 Go 构建。真实 Jarvis 调用和 QQ 消息发送需要在远端部署启动后验证，本文不声明已经完成实时调用测试。

## CI/CD 与部署核验

推送 `master` 后，GitHub Actions 的 `Docker Image CI` 构建并推送 `linux/amd64` 镜像，随后 `Update App` 更新 `qq-bot` Pod。`Update App` 由上游工作流的 `completed` 事件触发，当前 workflow 没有额外判断构建结论，因此不能只看到它被触发就认定部署成功。

每次推送后必须：

1. 监控 `Docker Image CI` 和 `Update App` 两个 workflow 到结束，并调查任何失败。
2. 对集群进行只读核验：Deployment rollout、实际运行镜像、Pod readiness、重启次数和相关日志。

仓库的部署约束、端口和敏感文件说明见 [`docs/CODEX_PROJECT_CONTEXT.md`](docs/CODEX_PROJECT_CONTEXT.md)。
