# cnki-search

`cnki` 是一个用于查找参考文献的 CNKI CLI。核心目标是快速检索论文、查看详情/参考文献，并直接导出 GB/T 7714 风格引用。

- **`cnki` CLI**：Go 写的独立二进制，通过 HTTP 访问知网 `kns8s` 接口执行检索，支持 JSON 和直接引用格式输出，可被任何工具/脚本调用。
- **Claude Code Skill**：`SKILL.md` 指引 Claude 把自然语言翻译成 `cnki` 命令，自动化完成文献检索、解析、格式化、输出。

相比 v2 的无头浏览器实现，当前版本不再依赖本地 Chrome、浏览器自动化或持久化会话目录。

## 功能

- **多字段检索**：主题 / 关键词 / 篇名 / 作者 / 摘要 / 全文 / DOI
- **筛选**：时间范围、文献类型（期刊/硕博/会议/报纸/年鉴）
- **排序**：相关度 / 发表时间 / 被引 / 下载
- **元数据抽取**：标题、作者、单位、摘要、关键词、DOI、分类号、基金、被引、下载
- **参考文献导出**：GB/T 7714 格式
- **多种输出**：JSON（默认）/ table / citation / markdown
- **自动翻页**：按需抓取指定数量结果

说明：来源类型筛选（SCI/EI/核心/CSSCI/CSCD）在 HTTP 模式下暂未启用，避免静默返回不符合限制条件的结果。

## 架构

```
┌─────────────────┐   CLI args      ┌─────────────────────┐
│  Claude Code    │ ───────────────▶│  cnki (Go binary)   │
│  (SKILL.md)     │ ◀─── JSON ───── │  HTTP client        │
└─────────────────┘                 └──────────┬──────────┘
                                               │ HTTPS
                                               ▼
                                     ┌───────────────────┐
                                     │  kns.cnki.net     │
                                     └───────────────────┘
```

## 安装

详见 [INSTALL.md](INSTALL.md)。简单来说：

```bash
# 1. 装 Plugin（在 Claude Code 里执行）
/plugin marketplace add ExquisiteCore/cnki-search
/plugin install cnki-search@cnki-search

# 2. 装 CLI（任选一种）
# A. 下载 Release：https://github.com/ExquisiteCore/cnki-search/releases
# B. go install：
go install github.com/ExquisiteCore/cnki-search/cmd/cnki@latest
# C. 源码构建：
git clone https://github.com/ExquisiteCore/cnki-search && cd cnki-search && go build -o cnki ./cmd/cnki
```

## 使用

### 作为 CLI 直接用

```bash
# 基础检索（默认 JSON）
cnki search "深度学习" --size=10

# 按被引排序，近五年
cnki search "大语言模型" --from=2020 --to=2025 --sort=cited --size=30

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
帮我在知网上搜索"深度学习 图像识别"相关论文，2020年以后的，按被引排序

在知网检索作者"张钹"的论文

帮我查一下知网上关于"大语言模型"的最新研究，需要 30 篇，给我 GB/T 7714 引用格式
```

Claude 会自动拼 `cnki` 命令、解析 JSON、格式化输出。

## 命令速查

### 全局 flag

| Flag | 说明 | 默认 |
|------|------|------|
| `--format` | json / table / citation / markdown | json |
| `--timeout` | 单次命令整体超时 | 90s |
| `--user-agent` | HTTP User-Agent | 模拟 Chrome |

### `cnki search <query>`

| Flag | 可选值 | 默认 |
|------|--------|------|
| `--field` | topic / keyword / title / author / abstract / fulltext / doi | topic |
| `--from` / `--to` | 年份 | 不限 |
| `--type` | journal / master / phd / conference / newspaper / yearbook（可重复） | 全部 |
| `--source` | 暂不支持 HTTP 模式 | 无 |
| `--sort` | relevance / date / cited / downloads | relevance |
| `--size` | 1-500 | 20 |

### `cnki detail <url>`

| Flag | 说明 |
|------|------|
| `--with-refs` | 同时抽取参考文献 |

### `cnki refs <url>`

只抽参考文献。

## 退出码

| 码 | 含义 |
|----|------|
| 0 | 成功 |
| 1 | 一般错误（网络 / HTTP / 解析） |
| 2 | 验证码或反爬拦截 |
| 3 | 无结果 |
| 4 | 参数非法 |

## 项目结构

```
cnki-search/
├── cmd/cnki/                  # Go CLI 入口
├── internal/                  # Go 内部包
│   ├── cli/                   #   cobra 命令定义
│   ├── cnki/                  #   HTTP client + 知网业务逻辑
│   ├── model/                 #   数据结构
│   └── render/                #   json/table/citation/markdown 渲染
├── skills/cnki-search/        # Claude Code Skill 资源
├── .claude-plugin/            # Claude Code 插件元数据
├── .github/workflows/         # CI / Release
├── .goreleaser.yaml
├── go.mod / go.sum
├── INSTALL.md
├── README.md
└── LICENSE
```

## License

MIT
