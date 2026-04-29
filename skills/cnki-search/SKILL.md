---
name: cnki-search
description:
  在中国知网（CNKI）上检索学术论文。触发场景：用户要求搜索知网论文、查找文献、检索期刊/学位论文/会议论文，
  或提到"知网"、"CNKI"、"中国知网"、"文献检索"、"论文搜索"等关键词时使用此 skill。
  支持关键词检索、主题检索、作者检索、高级检索、论文详情获取、参考文献提取、批量结果导出。
metadata:
  author: xiaol
  version: "2.0.0"
---

# CNKI 参考文献检索 Skill

## 概述

通过独立的 `cnki` Go CLI 以 HTTP 方式访问中国知网 `kns8s` 接口，执行参考文献检索、详情查看、参考文献抽取和引用格式导出任务。当前版本不依赖外部 web-access skill、本地 Chrome、浏览器自动化或持久化会话目录。

Claude 在这里的职责是：

1. 把用户的自然语言需求翻译成 `cnki` 命令行参数
2. 执行命令，解析返回的 JSON
3. 按用户要求渲染成表格 / 引用格式 / 详细卡片，其中 `--format=citation` 是核心能力，不能丢

## 前置依赖

### 工具链检查

执行任何操作前，先确认 `cnki` 命令可用：

```bash
cnki --version
```

如果未安装或返回错误，引导用户按 `INSTALL.md` 步骤安装（推荐：去 GitHub Release 下载预编译二进制并加入 PATH；或 `go install github.com/ExquisiteCore/cnki-search/cmd/cnki@latest`）。

### 风控处理

当前 HTTP 模式不会弹出浏览器登录或人工过验证码。若命令返回退出码 `2`（ErrCaptcha），说明触发了验证码或反爬拦截，按下文"错误处理"降低频率或稍后重试。

## 检索流程

### Phase 1：明确检索需求

与用户确认以下信息（缺省时使用默认值）：

| 参数 | CLI flag | 可选值 | 默认 |
|------|---------|--------|------|
| 关键词 | `<query>` 位置参数 | 检索词，多词用空格 | 必填 |
| 检索字段 | `--field` | topic / keyword / title / author / abstract / fulltext / doi | topic |
| 起始年份 | `--from` | YYYY | 不限 |
| 截止年份 | `--to` | YYYY | 不限 |
| 文献类型 | `--type` | journal / master / phd / conference / newspaper / yearbook（可重复） | 全部 |
| 来源类型 | `--source` | HTTP 模式暂不支持 | 不使用 |
| 排序方式 | `--sort` | relevance / date / cited / downloads | relevance |
| 结果数量 | `--size` | 任意正整数（≤500） | 20 |

缺少关键词时必须向用户追问，其余可用默认值。

### Phase 2：执行检索

组装命令，默认 JSON 输出方便后续解析：

```bash
cnki search "深度学习 图像识别" \
  --field=topic \
  --from=2020 \
  --sort=cited \
  --size=30 \
  --format=json
```

Bash 示例（把 JSON 存到变量中）：

```bash
RESULT=$(cnki search "大语言模型" --size=20 --sort=cited --format=json)
# 用 jq 拿到前 5 条标题
echo "$RESULT" | jq -r '.results[:5] | .[] | .title'
```

#### JSON 输出结构

```json
{
  "query": {"q":"深度学习","field":"topic","sort":"cited","size":20},
  "total_hits": 12345,
  "fetched": 20,
  "results": [
    {
      "seq": 1,
      "title": "...",
      "url": "https://kns.cnki.net/kcms2/article/abstract?v=...",
      "authors": ["张三","李四"],
      "source": "计算机学报",
      "year": 2024,
      "issue": "2024-03",
      "cited": 45,
      "downloads": 230
    }
  ]
}
```

### Phase 3：获取论文详情（可选）

当用户需要某篇论文的完整信息时，**必须使用上一步返回的 `url`**（它带会话参数，不能手工拼接）：

```bash
cnki detail "<paper url from step 2>" --format=json
```

加 `--with-refs` 可同时抽取参考文献：

```bash
cnki detail "<paper url>" --with-refs --format=json
```

返回字段：`title / authors / institutions / abstract / keywords / doi / clc / source / issue / year / fund / cited / downloads / references`。

### Phase 4：单独获取参考文献（可选）

```bash
cnki refs "<paper url>" --format=json
```

返回 `[{seq, text}, ...]` 数组。

### Phase 5：格式化输出

根据用户需求选择 CLI 自带的输出格式，或把 JSON 重新渲染：

#### 人类可读表格

CLI 直出，无需二次处理：

```bash
cnki search "深度学习" --size=10 --format=table
```

#### GB/T 7714 引用格式

```bash
cnki search "深度学习" --size=10 --format=citation
cnki search "知识图谱" --size=20 --format=citation
```

输出：

```
[1] 张三, 李四. 基于深度学习的图像识别研究[J]. 计算机学报, 2024.
[2] 王五. 卷积神经网络综述[J]. 软件学报, 2023.
```

#### Markdown 详细卡片

```bash
cnki detail "<paper url>" --with-refs --format=markdown
```

输出（对话中可直接展示）：

```markdown
### 《论文标题》

- **作者**：张三, 李四
- **单位**：XX大学XX学院
- **来源**：《期刊名》2024年第3期
- **DOI**：10.xxxx/xxxx
- **被引**：15 次 | **下载**：230 次
- **关键词**：关键词1; 关键词2

**摘要**：……
```

#### 在对话中直接呈现

默认把 JSON 里的 `results` 渲染成如下 Markdown 表格给用户看：

```
## 知网检索结果：「{关键词}」

命中 XX 条，以下为前 N 条（按{排序方式}排序）：

| # | 标题 | 作者 | 来源 | 年份 | 被引 |
|---|------|------|------|------|------|
| 1 | ... | ... | ... | 2024 | 15 |
```

## 错误处理

`cnki` 退出码约定：

| 退出码 | 含义 | 应对 |
|--------|------|------|
| 0 | 成功 | 解析 JSON 继续 |
| 1 | 一般错误（网络/DOM 异常） | 读 stderr 提示用户 |
| 2 | 验证码或反爬拦截 | 降低频率、稍后重试，或更换合规网络环境 |
| 3 | 检索结果为空 | 建议用户调整关键词、放宽时间、换检索字段 |
| 4 | 参数非法 | 读 stderr 的校验提示，重新询问用户 |

### 遇到退出码 2（验证码）

优先级从高到低：

1. 告知用户短期频繁检索容易触发风控，建议降低频率
2. 稍后重试，必要时调整 `--timeout`
3. 更换合规网络环境后重试

### 遇到退出码 3（无结果）

- 放宽 `--from/--to` 年份
- 把 `--field` 从 `title` 改为 `topic` 或 `keyword`
- 去掉 `--type` 的限制
- 建议同义词替换（"大语言模型" → "LLM" / "预训练语言模型"）

## 操作节奏

- **串行调用**：避免在同一秒发多个 `cnki search`；必要时用户可以连续追问但别并行触发
- **先 search 后 detail**：detail URL 必须来自 search 返回的 `url` 字段，不要凭空构造
- **批量详情要节制**：如果用户要 20 篇的完整详情，分批执行或直接用 search 自带的信息 + 摘要抽取

## 与其他 Skill 协作

### 与 lunwen skill 协作

当 lunwen（毕业论文写作）skill 需要文献检索时，本 skill 可被调用来：

1. 根据论文主题检索相关文献：`cnki search "主题" --sort=cited`
2. 提取参考文献元数据：`cnki detail <url>`
3. 输出 GB/T 7714 格式：`cnki search ... --format=citation`
4. 筛选高被引文献

### 与 research-writing-skill 协作

1. 按主题批量检索：`cnki search "..." --size=30 --format=json`
2. 提取摘要和关键词：迭代 `cnki detail <url>` for top N
3. 为文献综述提供素材

## 子 Agent 使用指南

在子 Agent prompt 中调用本 skill：

```
必须加载 cnki-search skill 并遵循指引。
任务：用 `cnki` 命令行在知网上获取关于「{主题}」的学术文献，需要 {N} 篇，
按被引频次排序，时间范围 2020-2025。
将结果以 GB/T 7714 格式返回。
```

## 任务结束

完成检索后：

1. 向用户呈现格式化结果（默认 Markdown 表格）
2. 如果用户要了引用格式，附上 `--format=citation` 的输出
3. 询问是否需要查看某篇的详情（`cnki detail <url>`）
4. 询问是否需要调整检索条件重新搜索

## 参考：知网站点特性

参见同目录下的 `references/cnki.net.md` —— 记录了 CNKI 的 kns8s 架构、反爬行为、HTTP 接口和编码注意事项。
