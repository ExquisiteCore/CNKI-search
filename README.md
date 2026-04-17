# cnki-search

独立的知网（CNKI）学术论文检索命令行工具 + Claude Code Skill。

- **`cnki` CLI**：Go 写的独立二进制，通过 chromedp 驱动本地 Chrome 访问中国知网执行检索，JSON 输出，可被任何工具/脚本调用。
- **Claude Code Skill**：`SKILL.md` 指引 Claude 把自然语言翻译成 `cnki` 命令，自动化完成文献检索 → 解析 → 格式化 → 输出的全流程。

相比 v1，**不再依赖** `web-access` skill 或系统 Chrome 远程调试。整个项目自包含。

## 功能

- **多字段检索**：主题 / 关键词 / 篇名 / 作者 / 摘要 / 全文 / DOI
- **高级筛选**：时间范围、文献类型（期刊/硕博/会议/报纸/年鉴）、来源类型（SCI/EI/核心/CSSCI/CSCD）
- **排序**：相关度 / 发表时间 / 被引 / 下载
- **元数据抽取**：标题、作者、单位、摘要、关键词、DOI、分类号、基金、被引、下载
- **参考文献导出**：GB/T 7714 格式
- **多种输出**：JSON（默认）/ table / citation / markdown
- **自动翻页**：按需抓取任意数量结果
- **登录态持久化**：内置 profile 目录，`cnki login` 一次后续复用
- **无头运行**：默认后台 Chrome，不打扰桌面

## 架构

```
┌─────────────────┐   CLI args      ┌─────────────────────┐
│  Claude Code    │ ───────────────▶│  cnki (Go binary)   │
│  (SKILL.md)     │ ◀─── JSON ───── │  chromedp + Chrome  │
└─────────────────┘                 └──────────┬──────────┘
                                               │ CDP
                                               ▼
                                     ┌───────────────────┐
                                     │  kns.cnki.net     │
                                     └───────────────────┘
```

## 安装

详见 [INSTALL.md](INSTALL.md)。简单来说：

```bash
# 1. 装 CLI（任选一种）
# A. 下载 Release：https://github.com/ExquisiteCore/cnki-search/releases
# B. go install：
go install github.com/ExquisiteCore/cnki-search/cmd/cnki@latest
# C. 源码构建：
git clone https://github.com/ExquisiteCore/cnki-search && cd cnki-search && go build -o cnki ./cmd/cnki

# 2. 装 Skill 文档（任选一种，详见 INSTALL.md）
# A. symlink（推荐）：
ln -s /path/to/cnki-search/skill ~/.claude/skills/cnki-search
# B. 复制：
cp -r /path/to/cnki-search/skill ~/.claude/skills/cnki-search

# 3. 首次登录（一次就好）
cnki login
```

## 使用

### 作为 CLI 直接用

```bash
# 基础检索（默认 JSON）
cnki search "深度学习" --size=10

# 核心期刊，按被引排序，近五年
cnki search "大语言模型" --source=core --from=2020 --to=2025 --sort=cited --size=30

# 人类可读表格
cnki search "Transformer" --size=10 --format=table

# GB/T 7714 引用格式
cnki search "知识图谱" --size=20 --format=citation

# 论文详情 + 参考文献
cnki detail "https://kns.cnki.net/kcms2/article/abstract?v=..." --with-refs --format=markdown

# 作者检索
cnki search "张钹" --field=author --size=15
```

### 作为 Claude Code Skill 用

装好之后在 Claude Code 里直接用自然语言触发：

```
帮我在知网上搜索"深度学习 图像识别"相关论文，要核心期刊，2020年以后的，按被引排序

在知网检索作者"张钹"的论文

帮我查一下知网上关于"大语言模型"的最新研究，需要 30 篇，给我 GB/T 7714 引用格式
```

Claude 会自动拼 `cnki` 命令、解析 JSON、格式化输出。

## 命令速查

### 全局 flag

| Flag | 说明 | 默认 |
|------|------|------|
| `--format` | json / table / citation / markdown | json |
| `--headed` | 有头模式（调试 / 登录 / 过验证码） | false |
| `--timeout` | 整体超时 | 90s |
| `--chrome` | Chrome 路径 | 自动探测 |
| `--profile-dir` | 用户数据目录 | 系统 cache 目录下 |
| `-v, --verbose` | 打印 chromedp 日志到 stderr | false |

### `cnki search <query>`

| Flag | 可选值 | 默认 |
|------|--------|------|
| `--field` | topic / keyword / title / author / abstract / fulltext / doi | topic |
| `--from` / `--to` | 年份 | 不限 |
| `--type` | journal / master / phd / conference / newspaper / yearbook（可重复） | 全部 |
| `--source` | sci / ei / core / cssci / cscd（可重复） | 全部 |
| `--sort` | relevance / date / cited / downloads | relevance |
| `--size` | 1-500 | 20 |

### `cnki detail <url>`

| Flag | 说明 |
|------|------|
| `--with-refs` | 同时抽取参考文献 |

### `cnki refs <url>`

只抽参考文献。

### `cnki login`

有头模式打开知网，用户手动登录后保存 cookie。**首次使用前跑一次。**

## 退出码

| 码 | 含义 |
|----|------|
| 0 | 成功 |
| 1 | 一般错误（网络 / DOM） |
| 2 | 验证码或反爬拦截（跑 `cnki login` 或 `--headed` 重试） |
| 3 | 无结果 |
| 4 | 参数非法 |

## 项目结构

```
cnki-search/
├── cmd/cnki/                  # Go CLI 入口
├── internal/                  # Go 内部包
│   ├── browser/               #   chromedp 封装、profile 管理、验证码检测
│   ├── cli/                   #   cobra 命令定义
│   ├── cnki/                  #   知网业务逻辑（search/detail/refs + selectors）
│   ├── model/                 #   数据结构
│   └── render/                #   json/table/citation/markdown 渲染
├── skill/                     # Claude Code Skill 资源（与 Go 源码隔离）
│   ├── SKILL.md               #   Claude 读取的 Skill 定义
│   └── references/
│       └── cnki.net.md        #   知网站点经验（DOM/反爬/陷阱）
├── .claude-plugin/            # Claude Code 插件元数据
│   ├── plugin.json            #   "skills": "./skill"
│   └── marketplace.json
├── .github/workflows/         # CI / Release
├── .goreleaser.yaml
├── go.mod / go.sum
├── INSTALL.md                 # 安装指南（CLI + Skill 两部分）
├── README.md                  # 本文件
└── LICENSE
```

两部分明确分工：

- **Go CLI 部分**（`cmd/`, `internal/`, `go.mod`, `.goreleaser.yaml`）——独立二进制，可被任何工具调用
- **Skill 部分**（`skill/`, `.claude-plugin/`）——Claude Code 读取的文档，告诉 Claude 怎么调用 CLI

## License

MIT
