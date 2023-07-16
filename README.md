## QQ bot

#### [主要依赖项目](https://github.com/Mrs4s/go-cqhttp)

在 **config.yml** 中设置你自己的账号

在 **device.json**中 protol字段配置设备信息(device文件会在第一次启动自动生成)：

| 值 | 类型            | 限制                               |
|---|---------------|----------------------------------|
| 0 | Default/Unset | 当前版本下默认为iPad                     |
| 1 | Android Phone | 无                                |
| 2 | Android Watch | 无法接收 notify 事件、无法接收口令红包、无法接收撤回消息 |
| 3 | MacOS         | 无                                |
| 4 | 企点            | 只能登录企点账号或企点子账号                   |
| 5 | iPad          | 无                                |
| 6 | aPad          | 无                                |

#### 本qq机器人已开发以下功能（所有功能均支持群聊与私聊）：

* 通过命令触发微博热搜推送
* 通过命令触发36氪每日热榜推送
* 通过命令触发华尔街最新资讯推送
* 通过命令触发知乎热榜资讯推送
* 定时推送 & 推送管理：可在群与私聊中 通过命令 开启/关闭
* 通过命令获得BTC/ETH/BNB最近币价(数据来自币安)，可以获取指定币种价格
* 集成ChatGPT & BingChat的对话(支持十分钟滚动窗口的上下文，chatGpt3.5和4.0共享上下文)
* 通过命令切换 chatGpt3.5 chatGpt4.0和 bingChat模式
* 支持ChatGpt4，感谢 [ChimeraGPT API](https://chimeragpt.adventblocks.cc/) 提供的支持
* 集成DELL.2的AI作图，通过命令与描述词
* 通过命令触发翻译 负载均衡(有道、腾讯、百度、火山、阿里)
* 基于tencent oss的状态存储
* 定时推送当天新闻数据到tencent oss
* 通过命令设置环境变量（只有bot owner可以使用）
* 停止运行bot(只有bot owner可以使用)

如果要运行本机器需要设置以下环境变量

```shell
# openai api key
ENV OPENAI_API_KEY=

# chimera api key，提供免费的chatgpt支持包括4.0
ENV CHIMERA_KEY=

# 临时文件存储目录
ENV FILE_ROOT=

# 百度云智能平台api
ENV BAIDU_API_KEY=
ENV BAIDU_SECRET_KEY=

# 阿里云api 用于翻译
ENV ALI_ACCESS_ID=
ENV ALI_ACCESS_SECRET=

# 腾讯云api 用于翻译 & 状态文件/新闻文件存储
ENV TC_SECRET_ID=
ENV TC_SECRET_KEY=

# 有道云api 用于翻译
ENV YD_APP_KEY=
ENV YD_SECRET_KEY=

# 火山引擎api 用于翻译
ENV VOLC_ACCESS_KEY=
ENV VOLC_SECRET_KEY=

# BingChat cookie
# 如何获取请参考:https://github.com/NoahAmethyst/bingchat-api/blob/master/README.md
ENV COOKIE=

```

如果是无法访问openai域名的大陆ip,可以添加如下环境变量：

```shell
# 请求中转服务的host
ENV REMOTE_PROXY=
```

也可以使用[ChimeraGPT API](https://chimeragpt.adventblocks.cc/)提供的apikey替换openai的域名，目前大陆可访问

关于请求中转项目可以查看[该项目
](https://github.com/NoahAmethyst/openai-proxy)






