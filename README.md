## QQ bot

#### [主要依赖项目](https://github.com/Mrs4s/go-cqhttp)

在 **config.yml** 中设置你自己的账号

在 **device.json**中 protol字段配置设备信息：

| 值 | 类型 | 限制 |
| --- | --- | --- |
| 0 | Default/Unset | 当前版本下默认为iPad |
| 1 | Android Phone | 无 |
| 2 | Android Watch | 无法接收 notify 事件、无法接收口令红包、无法接收撤回消息 |
| 3 | MacOS | 无 |
| 4 | 企点 | 只能登录企点账号或企点子账号 |
| 5 | iPad | 无 |
| 6 | aPad | 无 |

#### 本qq机器人已开发以下功能（所有功能均支持群聊与私聊）：
* 通过命令触发微博热搜推送
* 通过命令触发36氪每日热榜推送 
* 通过命令触发华尔街最新资讯推送
* 通过命令触发知乎热榜资讯推送
* 定时推送 & 推送管理：可在群与私聊中 通过命令 开启/关闭
* 通过命令获得BTC/ETH/BNB最近币价（数据来自币安）
* 集成ChatGPT & BingChat的对话(支持十分钟滚动窗口的上下文)
* 通过命令切换chatgpt和bingchat模式
* 集成DELL.2的AI作图，通过命令与描述词
* 通过命令触发翻译 负载均衡(有道、腾讯、百度、火山、阿里)

如果要运行本机器需要设置以下环境变量
```shell
# openai api key
ENV OPENAI_API_KEY=

# 临时文件存储目录
ENV FILE_ROOT=

# 百度云智能平台api
ENV BAIDU_API_KEY=
ENV BAIDU_SECRET_KEY=

# 阿里云api
ENV ALI_ACCESS_ID=
ENV ALI_ACCESS_SECRET=

# 腾讯云api
ENV TC_SECRET_ID=
ENV TC_SECRET_KEY=

# 有道云api
ENV YD_APP_KEY=
ENV YD_SECRET_KEY=

# 火山引擎api
ENV VOLC_ACCESS_KEY=
ENV VOLC_SECRET_KEY=

```

如果是无法访问openai域名的大陆ip,可以添加如下环境变量：
```shell
# 请求中转服务的host
ENV REMOTE_PROXY=
```

关于请求中转项目可以查看[该项目
](https://github.com/NoahAmethyst/openai-proxy)






