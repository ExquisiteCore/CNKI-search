# cnki-search 安装指南

本仓库由两部分组成，**都需要装上才能在 Codex 或 Claude Code 里自然语言触发知网参考文献检索与引用导出**：

1. **Codex / Claude Code Skill**：`.codex-plugin/plugin.json`、`.claude-plugin/plugin.json` + `skills/cnki-search/SKILL.md`，告诉 agent 怎么调用 CLI
2. **`cnki` CLI 二进制**：Go 程序，通过 HTTP 访问知网 `kns8s` 接口，支持直接输出引用格式

当前版本不需要本地 Chrome、浏览器自动化或持久化会话目录，也没有登录初始化步骤。

## 一、安装 Skill / Plugin

### Codex

Codex 适配文件在：

```text
.codex-plugin/plugin.json
skills/cnki-search/SKILL.md
skills/cnki-search/agents/openai.yaml
```

将本仓库作为本地 Codex plugin/skill 加载后，Codex 可通过 `$cnki-search` 或相关自然语言请求触发。

### Claude Code（兼容）

在 Claude Code 里执行：

```
/plugin marketplace add ExquisiteCore/cnki-search
/plugin install cnki-search@cnki-search
```

装完后执行 `/reload-plugins` 或重启 Claude Code。

### 本地开发 / 调试

如果你在本地 clone 了仓库想实时改 `SKILL.md` 调试：

```bash
claude --plugin-dir /path/to/cnki-search
```

## 二、安装 `cnki` CLI

Skill/Plugin 只负责告诉 agent 怎么调 CLI，CLI 本身必须装到系统 `PATH`。三种方式任选其一：

### 方式 A：下载预编译二进制（推荐）

去 [GitHub Releases](https://github.com/ExquisiteCore/cnki-search/releases) 下载对应平台归档：

- Windows：`cnki_windows_amd64.zip`
- macOS：`cnki_darwin_amd64.tar.gz`（Intel）/ `cnki_darwin_arm64.tar.gz`（Apple Silicon）
- Linux：`cnki_linux_amd64.tar.gz` / `cnki_linux_arm64.tar.gz`

解压后把 `cnki` 可执行文件放入 `PATH`。

验证：

```bash
cnki --version
```

### 方式 B：`go install`（需要 Go 1.26+）

```bash
go install github.com/ExquisiteCore/cnki-search/cmd/cnki@latest
```

二进制会被放到 `$(go env GOBIN)` 或 `$(go env GOPATH)/bin`，确保该目录在 `PATH` 里。

### 方式 C：源码构建

```bash
git clone https://github.com/ExquisiteCore/cnki-search
cd cnki-search
go build -o cnki ./cmd/cnki    # Linux/macOS
# Windows PowerShell: go build -o cnki.exe .\cmd\cnki
```

## 三、验证安装

```bash
# 1. CLI 工作
cnki search "测试" --size=3 --format=table
cnki search "测试" --size=3 --format=citation

# 2. 在 Codex 或 Claude Code 里说："帮我在知网上搜深度学习相关的论文 5 篇"
#    agent 应该自动拼出 cnki 命令并把结果渲染成表格
```

## 前置要求

- **网络访问**：需要能访问 `www.cnki.net` 和 `kns.cnki.net`
- **Go 工具链**：仅方式 B / C 需要，要求 Go 1.26 或更高版本

## 常见问题

### 报错：退出码 2 "captcha or anti-bot challenge detected"

触发了知网风控。HTTP 模式不会弹浏览器过验证码，建议降低请求频率、稍后重试，或更换合规网络环境。

### 报错：退出码 3 "no results"

关键词太严格、年份太窄、文献类型限制太多。建议放宽年份、把 `--field` 从 `title` 改为 `topic`，或去掉 `--type`。

### 报错：`source filters are not supported by HTTP mode yet`

`--source=sci|ei|core|cssci|cscd` 属于旧浏览器 UI 路径下的来源筛选。HTTP 模式暂未启用该筛选，避免静默返回不符合条件的结果。

### `go install` 失败：`unknown revision` 或 `module not found`

先确认 Go 版本：

```bash
go version
```

如果低于 1.26，用方式 A（下载预编译二进制）绕过。

### Skill/Plugin 装好了但 agent 不识别 skill

1. Codex：确认本仓库以 Codex plugin/skill 形式加载，且 `.codex-plugin/plugin.json` 指向 `./skills/`
2. Claude Code：跑 `/reload-plugins` 或重启 Claude Code
3. Claude Code：跑 `/plugin` 确认 `cnki-search` 在已启用列表里
4. 确认仓库里 `skills/cnki-search/SKILL.md` 存在且 frontmatter 有 `description` 字段
