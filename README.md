## QQ bot

#### [主要依赖项目](https://github.com/Mrs4s/go-cqhttp)

在 **config.yml** 中设置你自己的账号

#### 本qq机器人已开发以下功能：
* 微博热搜推送
* 36氪每日热榜推送 
* 华尔街最新资讯推送
* 获得BTC/ETH/BNB最近币价（数据来自币安）
* 集成ChatGPT的对话
* 集成DELL.2的AI作图
* 翻译 负载均衡(有道、腾讯、百度、火山、阿里)

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






