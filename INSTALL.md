# cnki-search 安装指南

本仓库由两部分组成，**都需要装上才能在 Claude Code 里自然语言触发知网检索**：

1. **Claude Code Plugin** —— `.claude-plugin/plugin.json` + `skills/cnki-search/SKILL.md`，告诉 Claude 怎么调用 CLI
2. **`cnki` CLI 二进制** —— Go 程序，真正驱动本地 Chrome 访问知网

## 一、安装 Plugin（推荐：Claude Code Marketplace）

在 Claude Code 里执行：

```
/plugin marketplace add ExquisiteCore/cnki-search
/plugin install cnki-search@cnki-search
```

第一条把本仓库注册为一个 marketplace；第二条从该 marketplace 装 `cnki-search` 插件。装完 `/reload-plugins` 或重启 Claude Code，skill 会自动加载到命名空间 `cnki-search`。

### 本地开发 / 调试

如果你在本地 clone 了仓库想实时改 SKILL.md 调试，不需要走 marketplace，直接：

```bash
claude --plugin-dir /path/to/cnki-search
```

这会优先加载本地副本，改完跑 `/reload-plugins` 生效。

## 二、安装 `cnki` CLI

Plugin 只负责告诉 Claude 怎么调 CLI，CLI 本身必须装到系统 PATH。三种方式任选其一：

### 方式 A：下载预编译二进制（推荐）

去 [GitHub Releases](https://github.com/ExquisiteCore/cnki-search/releases) 下载对应平台归档：

- Windows：`cnki_windows_amd64.zip`
- macOS：`cnki_darwin_amd64.tar.gz`（Intel）/ `cnki_darwin_arm64.tar.gz`（Apple Silicon）
- Linux：`cnki_linux_amd64.tar.gz` / `cnki_linux_arm64.tar.gz`

解压后把 `cnki` 可执行文件放入 PATH。

**Windows 示例**：把 `cnki.exe` 放到 `C:\Users\<你>\AppData\Local\Microsoft\WindowsApps\` 或自己的 `bin` 目录。

**macOS / Linux 示例**：

```bash
tar xf cnki_darwin_arm64.tar.gz
sudo mv cnki /usr/local/bin/
```

验证：

```bash
cnki --version
```

### 方式 B：`go install`（需要 Go 1.25+）

```bash
go install github.com/ExquisiteCore/cnki-search/cmd/cnki@latest
```

二进制会被放到 `$(go env GOBIN)` 或 `$(go env GOPATH)/bin`，确保该目录在 PATH 里。

### 方式 C：源码构建

```bash
git clone https://github.com/ExquisiteCore/cnki-search
cd cnki-search
go build -o cnki ./cmd/cnki    # Linux/macOS
# Windows PowerShell: go build -o cnki.exe .\cmd\cnki
```

## 三、首次使用：登录

知网部分功能需要账号（机构 IP 或个人账号）。首次使用前跑一次：

```bash
cnki login
```

这会弹出一个有头 Chrome 窗口，请你在窗口里手动完成知网登录（可以扫码、账号密码、机构认证都行），完成后回到终端按 Enter。cookie 会保存到默认 profile 目录：

- Windows：`%LOCALAPPDATA%\cnki-search\chrome\`
- macOS：`~/Library/Caches/cnki-search/chrome/`
- Linux：`~/.cache/cnki-search/chrome/`

之后所有的 `cnki search`、`cnki detail` 都会复用这个登录态。

## 四、前置要求

- **浏览器**：系统已安装 Chrome、Edge 或 Chromium（任一即可）。`cnki` 会自动探测；若装在非标准路径，用 `--chrome` 指定：

  ```bash
  cnki --chrome="/path/to/chrome" search "..."
  ```

- **Go 工具链**（仅方式 B / C 需要）：Go 1.25 或更高版本

## 五、验证安装

```bash
# 1. CLI 工作
cnki search "测试" --size=3 --format=table

# 2. 在 Claude Code 里说："帮我在知网上搜深度学习相关的论文 5 篇"
#    Claude 应该自动拼出 cnki 命令并把结果渲染成表格
```

## 常见问题

### 报错：`launch chrome: exec: "chrome": executable file not found`

系统找不到 Chrome。两种解决：

1. 安装 Chrome 或 Edge
2. 用 `--chrome` 明确指定路径：
   ```bash
   cnki --chrome="C:\Program Files\Google\Chrome\Application\chrome.exe" search "..."
   ```

### 报错：退出码 2 "captcha or anti-bot challenge detected"

触发了知网风控：

1. 先 `cnki login` 更新登录态
2. 还不行就加 `--headed` 手动过一次验证码
3. 短期频繁检索容易触发，建议降低频率

### 报错：退出码 3 "no results"

关键词太严格、年份太窄、来源限制太多。参考 `SKILL.md` "错误处理"部分的建议。

### `go install` 失败：`unknown revision` 或 `module not found`

确保你的 Go 版本 ≥ 1.25：

```bash
go version
```

如果低于 1.25，用方式 A（下载预编译二进制）绕过。

### 每次 search 都跳出 Chrome 窗口

默认应该是无头模式。如果每次都跳窗口，检查是不是不小心在全局加了 `--headed`。无头模式下 Chrome 完全不可见。

### 登录态丢失

profile 目录被清了，或换机器了。重新跑 `cnki login` 即可。

### Plugin 装好了但 Claude 不识别 skill

1. 跑 `/reload-plugins` 或重启 Claude Code
2. 跑 `/plugin` 确认 `cnki-search` 在已启用列表里
3. 确认仓库里 `skills/cnki-search/SKILL.md` 存在且 frontmatter 有 `description` 字段
