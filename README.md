# ever-token-vscode

## 功能

ever-token-vscode根据提供的参数，从印象笔记开放平台，<https://app.yinxiang.com/api/DeveloperToken.action> 获取 accessToken,并自动写入vscode配置文件中，方便在vscode中使用印象笔记插件。  

## 用法

```bash
$ ./ever-token-vscode -h
Usage of ./ever-token-vscode:
  -password string
        evernote password
  -settings string
        evernote settings file path, default ~/.config/Code/User/settings.json
  -username string
        evernote username
```
