# CNKI-search Skill 安装指南

## 前置依赖

本 skill 依赖 **web-access skill** 提供的 CDP 浏览器自动化能力。新设备必须先安装 web-access skill。

### 1. 安装 web-access skill

```bash
claude install-skill https://github.com/eze-is/web-access
```

> 如果 `install-skill` 不可用，手动克隆：
>
> ```bash
> git clone https://github.com/eze-is/web-access ~/.claude/skills/web-access
> ```

### 2. 配置 Chrome 远程调试

web-access skill 通过 Chrome DevTools Protocol (CDP) 控制用户日常 Chrome 浏览器。需要一次性开启：

1. 打开 Chrome，地址栏输入 `chrome://inspect/#remote-debugging`
2. 勾选 **"Allow remote debugging for this browser instance"**
3. 可能需要重启 Chrome

### 3. 确认 Node.js 版本

```bash
node -v
```

要求 **Node.js 22+**（使用原生 WebSocket）。低于 22 的版本需额外安装 `ws` 模块。

### 4. 安装 cnki-search skill

```bash
# 方式一：从 Git 仓库安装（推荐）
git clone <本仓库地址> ~/.claude/skills/cnki-search

# 方式二：手动复制
cp -r . ~/.claude/skills/cnki-search/
```

### 5. 验证安装

启动 Claude Code，输入任意知网检索请求即可触发：

```
帮我在知网上搜索"深度学习"相关论文
```

Skill 会自动：
1. 检查 web-access 依赖是否就绪
2. 启动 CDP Proxy 连接 Chrome
3. 打开知网执行检索

## 常见问题

### CDP 连接失败

**现象**：提示 Chrome 端口不可达

**解决**：
- 确认 Chrome 已打开并启用了远程调试
- 确认没有其他程序占用 9222 端口
- 重启 Chrome 后重新勾选远程调试选项

### 验证码拦截

**现象**：知网弹出滑块验证码

**解决**：Skill 会自动截图并提示你手动完成验证码，完成后告知 Agent 继续即可。短时间内频繁检索更容易触发，建议控制检索频率。

### 需要登录

**现象**：提示内容不可访问

**解决**：基本检索和元数据查看通常无需登录。如果需要全文下载等功能，在 Chrome 中登录知网账号（机构 IP 或个人账号），完成后告知 Agent 继续。
