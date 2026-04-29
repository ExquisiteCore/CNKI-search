---
name: cnki-search
description: 在中国知网（CNKI）上检索学术论文并导出参考文献。Use when Codex needs to find CNKI papers, search Chinese academic literature, retrieve paper metadata/details/references, or produce GB/T 7714-style citations from queries mentioning 知网, CNKI, 中国知网, 文献检索, 论文搜索, 参考文献, 期刊论文, 学位论文, or 会议论文.
---

# CNKI Reference Search

Use the local `cnki` Go CLI to search CNKI over HTTP, parse metadata, fetch details/references when available, and render results for reference-list workflows. This skill is for finding references, not downloading full text.

Core capability to preserve: direct citation export with `--format=citation`.

## Preconditions

Check the CLI first:

```bash
cnki --version
```

If it is unavailable, tell the user to install the release binary or run:

```bash
go install github.com/ExquisiteCore/cnki-search/cmd/cnki@latest
```

The CLI needs network access to `www.cnki.net`, `kns.cnki.net`, and `recsys.cnki.net`. In Codex sandboxed environments, rerun blocked live CNKI commands with network approval instead of silently falling back to invented results.

## Search Workflow

1. Convert the user's request into `cnki search` flags.
2. Prefer `--format=json` when you need to inspect, filter, or post-process results.
3. Use `--format=citation` directly when the user asks for references/citations.
4. Present concise results, then offer detail/reference follow-up only when useful.

Default command:

```bash
cnki search "<query>" --field=topic --sort=relevance --size=20 --format=json
```

Common examples:

```bash
cnki search "知识图谱" --size=20 --format=citation
cnki search "大语言模型" --from=2020 --to=2025 --sort=cited --size=30 --format=json
cnki search "张钹" --field=author --size=15 --format=table
```

## Flags

| Need | Flag | Values / Notes |
|---|---|---|
| Query | positional `<query>` | Required; join multi-word topics with spaces |
| Field | `--field` | `topic`, `keyword`, `title`, `author`, `abstract`, `fulltext`, `doi` |
| Year range | `--from`, `--to` | YYYY |
| Type | `--type` | `journal`, `master`, `phd`, `conference`, `newspaper`, `yearbook`; repeatable |
| Source filter | `--source` | Not supported in HTTP mode; do not use unless testing the explicit error |
| Sort | `--sort` | `relevance`, `date`, `cited`, `downloads` |
| Count | `--size` | 1-500; use small values for exploratory searches |
| Output | `--format` | `json`, `table`, `citation`, `markdown` |

If the query is missing, ask one short clarification question. Otherwise use conservative defaults.

## Details And References

Only call detail/reference commands with URLs returned by `cnki search`; CNKI detail URLs contain session/query parameters and must not be reconstructed.

```bash
cnki detail "<paper url>" --format=json
cnki detail "<paper url>" --with-refs --format=markdown
cnki refs "<paper url>" --format=json
```

Detail pages may trigger CNKI security verification more often than search. If that happens, report the block and keep the search metadata/citation output rather than fabricating detail fields.

## Output Guidance

For literature lists, prefer citation output:

```bash
cnki search "<topic>" --size=20 --format=citation
```

For screening papers, use JSON or table and summarize:

```markdown
| # | Title | Authors | Source | Year | Cited |
|---|---|---|---|---:|---:|
```

For JSON processing, use fields from `.results[]`: `seq`, `title`, `url`, `authors`, `source`, `year`, `issue`, `cited`, `downloads`.

## Errors

`cnki` exit codes:

| Code | Meaning | Response |
|---:|---|---|
| 0 | Success | Parse/render output |
| 1 | General HTTP/parse/network error | Report stderr and retry only if the fix is clear |
| 2 | Captcha or anti-bot challenge | Tell the user CNKI blocked the request; reduce frequency or retry later |
| 3 | No results | Relax query, year, field, or type filters |
| 4 | Invalid arguments | Correct flags and rerun |

Never present partial or invented CNKI results when the command failed.

## Rate And Scope

- Run CNKI commands serially; avoid parallel live searches.
- Keep requests focused; use `--size=10` to `--size=30` unless the user asks for more.
- Do not implement or imply login, OAuth, CAJ/PDF saving, or full-text download.
- Use `references/cnki.net.md` only when debugging CNKI HTTP behavior or anti-bot responses.

## Codex Handoff Pattern

When delegating or documenting a task for another Codex agent, phrase it like:

```text
Use $cnki-search to search CNKI for papers about "<topic>", prefer cited papers from 2020-2025, and return GB/T 7714 citations.
```
