## README

### 项目简介
本项目是一个自动发送Jira任务到期提醒到钉钉群的工具。它会每天早上8点和下午6点检查当前用户在Jira中当天及未来三天内到期的任务，并将这些任务信息发送到指定的钉钉Webhook。

### 功能描述
1. **加载配置文件**：从`config.json`中读取必要的配置信息，包括Jira URL、Token、SOCKS5代理地址以及钉钉Webhook URL。
2. **设置HTTP客户端**：根据配置文件中的SOCKS5代理信息设置HTTP客户端，并使用Bearer Token进行身份验证以访问Jira API。
3. **获取任务列表**：
   - 搜索当天到期的任务。
   - 搜索接下来三天内到期的任务。
4. **构建消息内容**：将找到的任务信息格式化为字符串。
5. **发送消息到钉钉**：通过钉钉Webhook发送构建好的消息内容。
6. **定时任务**：程序会在每天早上8点和下午6点执行一次任务检查与消息发送操作。

### 使用方法
1. **准备配置文件**：创建一个名为`config.json`的文件，内容如下：
    ```json
    {
      "jira_url": "https://your-jira-instance.atlassian.net",
      "jira_token": "your_jira_api_token",
      "socks5_proxy": "socks5://localhost:1080",
      "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=your_dingtalk_webhook_token"
    }
    ```

2. **运行程序**：确保Go环境已安装，然后在命令行中进入项目目录并执行以下命令启动程序：
    ```bash
    go run main.go
    ```


### 依赖库
- `github.com/andygrunwald/go-jira`: 用于与Jira API交互。
- `golang.org/x/net/proxy`: 支持通过SOCKS5代理发起HTTP请求。

### 注意事项
- 确保你的Jira实例允许通过API访问，并且提供的Token具有足够的权限来读取任务信息。
- 如果你不需要使用SOCKS5代理，请修改代码移除相关的代理配置部分。
- 钉钉机器人需要提前创建好，并获取其Webhook URL。

### 错误处理
程序中对可能出现的错误进行了捕获并在控制台输出错误信息，便于调试和问题排查。

### 未来发展
可以考虑增加更多自定义选项，例如调整通知时间、支持更多的消息模板等。
