---
name: cnki-search
description:
  在中国知网（CNKI）上检索学术论文。触发场景：用户要求搜索知网论文、查找文献、检索期刊/学位论文/会议论文，
  或提到"知网"、"CNKI"、"中国知网"、"文献检索"、"论文搜索"等关键词时使用此 skill。
  支持关键词检索、主题检索、作者检索、高级检索、论文详情获取、参考文献/引用文献提取、批量结果导出。
metadata:
  author: xiaol
  version: "1.0.0"
---

# CNKI 知网检索 Skill

## 概述

通过浏览器 CDP 自动化访问中国知网（https://www.cnki.net），执行学术文献检索任务。本 skill 依赖 **web-access skill** 的 CDP Proxy 基础设施。

## 前置依赖

本 skill 必须在 **web-access skill** 已就绪的前提下使用。执行任何操作前：

```bash
node "$CLAUDE_SKILL_DIR/../web-access/scripts/check-deps.mjs"
```

如未通过，引导用户按 web-access skill 的指引完成 Chrome remote-debugging 设置。

## 检索流程

### Phase 1：明确检索需求

与用户确认以下信息（缺省时使用默认值）：

| 参数 | 说明 | 默认值 |
|------|------|--------|
| **关键词** | 检索词，支持多个（空格或 AND/OR 连接） | 必填 |
| **检索字段** | 主题 / 关键词 / 篇名 / 作者 / 摘要 / 全文 / 基金 / 参考文献 / DOI | 主题 |
| **时间范围** | 起止年份 | 不限 |
| **文献类型** | 期刊论文 / 硕士论文 / 博士论文 / 会议论文 / 报纸 / 年鉴 | 全部 |
| **来源类型** | 全部 / SCI / EI / 核心期刊 / CSSCI / CSCD | 全部 |
| **排序方式** | 相关度 / 发表时间（降序） / 被引频次 / 下载频次 | 相关度 |
| **需要数量** | 需要检索多少篇论文的信息 | 前 20 条 |

### Phase 2：执行检索

#### 2.1 打开知网并检索

```bash
# 创建新 tab 访问知网高级检索页
curl -s "http://localhost:3456/new?url=https://kns.cnki.net/kns8s/AdvSearch"
```

> **重要**：知网有多个入口域名，统一使用 `kns.cnki.net/kns8s/` 系列 URL。

#### 2.2 知网页面结构认知

知网高级检索页的核心 DOM 结构：

- **检索输入区**：`.search-box` 区域内有多个条件行
  - 每行包含：字段选择下拉框 + 输入框 + 逻辑运算符（AND/OR/NOT）
  - 字段选择器：`select` 元素，value 值对应检索字段代码
  - 输入框：`input.search-input` 或 `input[type="text"]`
- **时间范围**：年份起止选择器
- **来源类型**：复选框组（SCI、EI、核心期刊等）
- **检索按钮**：`input.btn-search` 或类似的提交按钮

#### 2.3 填写检索条件并提交

通过 `/eval` 操作 DOM 来填写检索条件。由于知网页面结构可能更新，**必须先探查当前页面实际 DOM 结构**，再决定操作方式：

```bash
# 1. 先探查搜索区域的 DOM 结构
curl -s -X POST "http://localhost:3456/eval?target=TAB_ID" \
  -d 'document.querySelector(".search-box, #advancedSearch, .advance-search, #ModuleSearch")?.innerHTML?.substring(0, 3000)'

# 2. 根据实际 DOM 构造填写逻辑（示例，实际选择器以探查结果为准）
curl -s -X POST "http://localhost:3456/eval?target=TAB_ID" \
  -d '(() => {
    const input = document.querySelector("input.search-input, input[name=\"txt_1_value1\"]");
    if (input) { input.value = "检索词"; input.dispatchEvent(new Event("input", {bubbles:true})); }
    return "filled";
  })()'

# 3. 点击检索
curl -s -X POST "http://localhost:3456/click?target=TAB_ID" \
  -d 'input.btn-search, .search-btn, button[type="submit"]'
```

**关键原则**：知网前端更新频繁，**不要硬编码选择器**。每次操作前先 `/eval` 探查目标元素的真实选择器，再执行操作。

#### 2.4 等待结果加载

```bash
# 等待搜索结果表格加载
curl -s -X POST "http://localhost:3456/eval?target=TAB_ID" \
  -d '(() => {
    const table = document.querySelector(".result-table-list, #gridTable, table.result-table");
    if (!table) return "LOADING";
    const rows = table.querySelectorAll("tbody tr");
    return "LOADED: " + rows.length + " results";
  })()'
```

如果返回 `LOADING`，等待 2-3 秒后重试（最多 5 次）。

### Phase 3：提取检索结果

#### 3.1 提取结果列表

```bash
curl -s -X POST "http://localhost:3456/eval?target=TAB_ID" \
  -d '(() => {
    const rows = document.querySelectorAll(".result-table-list tbody tr, #gridTable tbody tr");
    const results = [];
    rows.forEach((row, i) => {
      const title = row.querySelector(".name a, td.name a");
      const authors = row.querySelector(".author, td.author");
      const source = row.querySelector(".source a, td.source a");
      const date = row.querySelector(".date, td.date");
      const cite = row.querySelector(".quote, td.quote");
      const download = row.querySelector(".download, td.download");
      if (title) {
        results.push({
          seq: i + 1,
          title: title.textContent.trim(),
          link: title.href || "",
          authors: authors?.textContent?.trim() || "",
          source: source?.textContent?.trim() || "",
          date: date?.textContent?.trim() || "",
          cited: cite?.textContent?.trim() || "0",
          downloads: download?.textContent?.trim() || "0"
        });
      }
    });
    return JSON.stringify(results, null, 2);
  })()'
```

#### 3.2 翻页获取更多结果

当用户需要的数量超出当前页时，逐页获取：

```bash
# 点击下一页
curl -s -X POST "http://localhost:3456/click?target=TAB_ID" \
  -d '#PageNext, a.next, .pagebar a:last-child'
```

每页提取后与已有结果合并，直到达到用户要求的数量。

### Phase 4：获取论文详情（可选）

当用户需要某篇论文的详细信息时，在新 tab 中打开论文详情页：

```bash
# 用论文详情链接打开新 tab
curl -s "http://localhost:3456/new?url=PAPER_DETAIL_URL"
```

从详情页提取完整元数据：

```bash
curl -s -X POST "http://localhost:3456/eval?target=DETAIL_TAB_ID" \
  -d '(() => {
    const info = {};
    info.title = document.querySelector("h1, .wx-tit h1")?.textContent?.trim();
    info.authors = [...document.querySelectorAll(".author a, h3:first-of-type a")].map(a => a.textContent.trim());
    info.institutions = [...document.querySelectorAll(".orgn a, h3:nth-of-type(2) a")].map(a => a.textContent.trim());
    info.abstract = document.querySelector("#ChDivSummary, .abstract-text")?.textContent?.trim();
    info.keywords = [...document.querySelectorAll(".keywords a, p.keywords a")].map(a => a.textContent.trim().replace(/;$/, ""));
    const doiEl = [...document.querySelectorAll(".top-tip span, .doi")].find(el => el.textContent.includes("DOI"));
    info.doi = doiEl?.textContent?.replace(/.*DOI[：:]?\s*/, "").trim() || "";
    const clcEl = [...document.querySelectorAll(".top-tip span")].find(el => el.textContent.includes("分类号"));
    info.clc = clcEl?.textContent?.replace(/.*分类号[：:]?\s*/, "").trim() || "";
    info.source = document.querySelector(".top-tip a:first-child, .sourinfo a")?.textContent?.trim();
    info.issue = document.querySelector(".top-tip .year, .sourinfo .year")?.textContent?.trim();
    info.fund = document.querySelector(".fund, .funds")?.textContent?.trim();
    info.cited = document.querySelector("#annotationcount, .cited")?.textContent?.trim();
    info.downloads = document.querySelector("#downloadcount, .download")?.textContent?.trim();
    return JSON.stringify(info, null, 2);
  })()'
```

### Phase 5：获取参考文献列表（可选）

在论文详情页中，参考文献通常需要展开加载：

```bash
curl -s -X POST "http://localhost:3456/eval?target=DETAIL_TAB_ID" \
  -d '(() => {
    const refSection = document.querySelector("#CataLogContent .essayBox, .ref-list, #references");
    if (!refSection) return "NO_REF_SECTION";
    const refs = [...refSection.querySelectorAll("li, .essayLi")].map((li, i) => ({
      seq: i + 1,
      text: li.textContent.trim()
    }));
    return JSON.stringify(refs, null, 2);
  })()'
```

### Phase 6：格式化输出

根据用户需求选择输出格式：

#### 简洁列表（默认）

```
## 知网检索结果：「{关键词}」

共找到 XX 条结果，以下为前 N 条（按{排序方式}排序）：

| # | 标题 | 作者 | 来源 | 年份 | 被引 |
|---|------|------|------|------|------|
| 1 | xxx  | xxx  | xxx  | 2024 | 15   |
| 2 | xxx  | xxx  | xxx  | 2023 | 8    |
...
```

#### 详细信息

对用户指定的论文，输出完整元数据：

```
### 《论文标题》

- **作者**：张三, 李四
- **单位**：XX大学XX学院
- **来源**：《期刊名》2024年第3期
- **DOI**：10.xxxx/xxxx
- **摘要**：...
- **关键词**：关键词1; 关键词2; 关键词3
- **分类号**：TP311
- **基金**：国家自然科学基金(No.xxx)
- **被引**：15次 | **下载**：230次
```

#### 参考文献格式导出

当用户需要将检索结果用于论文写作时，按 GB/T 7714 格式输出：

```
[1] 张三, 李四. 论文标题[J]. 期刊名, 2024, 30(3): 45-52.
[2] 王五. 学位论文标题[D]. 北京: 北京大学, 2023.
```

## 特殊场景处理

### 登录提示

知网部分功能（如全文下载）需要机构或个人登录。检索和元数据获取通常无需登录。如果遇到登录弹窗：

```bash
# 关闭登录弹窗
curl -s -X POST "http://localhost:3456/eval?target=TAB_ID" \
  -d '(() => {
    const close = document.querySelector(".login-mask .close, .modal .close-btn, [class*=\"close\"]");
    if (close) { close.click(); return "closed"; }
    return "no_popup";
  })()'
```

如果登录弹窗阻挡了核心内容获取，告知用户在 Chrome 中登录知网后继续。

### 反爬/验证码

知网对自动化操作有检测。如果遇到验证码：
1. 截图让用户看到当前页面状态
2. 告知用户手动完成验证码
3. 用户确认后刷新页面继续

```bash
# 截图当前状态
curl -s "http://localhost:3456/screenshot?target=TAB_ID&file=/tmp/cnki-captcha.png"
```

### 搜索结果为空

如果检索无结果：
1. 建议用户调整关键词（更宽泛/更精确）
2. 尝试同义词替换
3. 调整检索字段（如从「篇名」换为「主题」）
4. 放宽时间范围

### 知网域名和 URL 模式

| 功能 | URL |
|------|-----|
| 首页 | `https://www.cnki.net` |
| 高级检索 | `https://kns.cnki.net/kns8s/AdvSearch` |
| 简单检索 | `https://kns.cnki.net/kns8s/search` |
| 论文详情 | `https://kns.cnki.net/kcms2/article/abstract?v=...` |
| 作者页面 | `https://kns.cnki.net/kcms2/author/detail?v=...` |

## 操作节奏

- **GUI 交互优先**：知网对自动化有检测，优先使用 GUI 交互（点击、输入）模拟用户行为
- **操作间隔**：相邻操作间保持 1-2 秒自然间隔，避免触发风控
- **不要并行多 tab**：在知网上避免同时打开过多 tab，串行操作更安全
- **先探查后操作**：每个操作前先 eval 查看当前 DOM 状态，确认元素存在再操作

## 与其他 Skill 协作

### 与 lunwen skill 协作

当 lunwen（毕业论文写作）skill 需要文献检索时，本 skill 可被调用来：
1. 根据论文主题检索相关文献
2. 提取参考文献的完整元数据
3. 输出 GB/T 7714 格式的参考文献列表
4. 筛选高被引/核心期刊文献

### 与 research-writing-skill 协作

当 research-writing-skill（科研写作）需要文献综述支持时，本 skill 可：
1. 按主题批量检索文献
2. 提取关键论文的摘要和关键词
3. 为文献综述章节提供素材

## 子 Agent 使用指南

在子 Agent prompt 中调用本 skill：

```
必须加载 cnki-search skill 和 web-access skill 并遵循指引。
任务：在知网上获取关于「{主题}」的学术文献，需要 {N} 篇，按被引频次排序，仅限核心期刊，时间范围 2020-2024。
将结果以表格形式返回，包含标题、作者、来源、年份、被引次数。
```

## 任务结束

完成检索后：
1. 关闭自己创建的所有 tab
2. 向用户呈现格式化的检索结果
3. 询问是否需要查看某篇论文的详细信息
4. 询问是否需要调整检索条件重新搜索

```bash
# 关闭 tab
curl -s "http://localhost:3456/close?target=TAB_ID"
```
