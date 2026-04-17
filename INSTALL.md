# cnki-search Skill 安装指南

本 skill 包含两部分：

1. **`cnki` CLI 二进制** —— 独立的 Go 程序，负责驱动本地 Chrome 访问知网
2. **Skill 文档** —— Claude Code 读取的 `SKILL.md`，把自然语言翻译成 `cnki` 命令

两者都要安装才能用。

## 一、安装 `cnki` CLI

### 方式 A：下载预编译二进制（推荐）

去 [GitHub Releases](https://github.com/ExquisiteCore/cnki-search/releases) 下载对应平台的归档：

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

## 二、安装 Skill 文档

Claude Code 读取 skill 文档的路径是 `~/.claude/skills/<skill-name>/`。本仓库的 Skill 资源放在 `skill/` 子目录下（与 Go CLI 源码隔离），所以不能直接把整个仓库 clone 过去——需要让 `~/.claude/skills/cnki-search/` 指向 `<repo>/skill/`。

**推荐做法：symlink**

```bash
# Linux / macOS
git clone https://github.com/ExquisiteCore/cnki-search ~/src/cnki-search
ln -s ~/src/cnki-search/skill ~/.claude/skills/cnki-search
```

```powershell
# Windows PowerShell（需要开发者模式或管理员）
git clone https://github.com/ExquisiteCore/cnki-search $env:USERPROFILE\src\cnki-search
New-Item -ItemType SymbolicLink `
  -Path $env:USERPROFILE\.claude\skills\cnki-search `
  -Target $env:USERPROFILE\src\cnki-search\skill
```

symlink 的好处是 `git pull` 一次即可同时更新 CLI 源码和 Skill 文档。

**备选：直接复制**

如果不想用 symlink：

```bash
git clone https://github.com/ExquisiteCore/cnki-search ~/src/cnki-search
cp -r ~/src/cnki-search/skill ~/.claude/skills/cnki-search
```

注意每次仓库 Skill 文档更新时需要手动重新复制。

**作为 Claude Code Plugin 安装**

如果你的环境支持 Claude Code plugin marketplace，`.claude-plugin/plugin.json` 已配置 `"skills": "./skill"`，可按 plugin 流程加载。

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
