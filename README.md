# RSS Translator

RSS Translator 是一个将指定的 RSS 订阅信息中的文章标题进行翻译的工具，使用 Google 翻译 API 实现。

## 项目部署

1. 克隆本仓库或者下载 [release 版本](https://github.com/xxx/RSS-Translator/releases)
2. 使用 `go build -ldflags "-w -s" rss_translator.go` 命令编译二进制文件
3. 在项目目录下通过 `init` 命令初始化配置文件：`./RSS-Translator init`
4. 修改 `./config.json` 配置文件中的相关配置项：
   - `host`: 监听地址，默认为 `0.0.0.0`
   - `port`: 监听端口，默认为 `8855`
   - `language`: 要进行翻译的目标语言，默认为 `"zh"`
   - `cron`: 定时任务执行的时间间隔（Cron 表达式），默认为每笑死
   - `rss`: 要进行翻译的 RSS 订阅源列表，需要指定 `url`、`path`、`xml_item_path`、`xml_title_in_item_path` 四个键，分别表示 RSS 订阅源的 URL、要存储的文件路径、在 XML 中表示文章信息的路径以及在 XML 中表示文章标题的路径。
5. 启动服务：`./RSS-Translator run`