---

<div align="center">

# Cloudflare VLESS Node Manager

A lightweight, robust tool for managing Cloudflare VLESS nodes with a Web Dashboard.  
一个轻量级、健壮的 Cloudflare VLESS 节点管理工具，带 Web 控制面板。

<p align="center">
  <a href="#-english-readme">
    <img src="https://img.shields.io/badge/Language-English-blue?style=for-the-badge" alt="English">
  </a>
  &nbsp;&nbsp;
  <a href="#-中文说明">
    <img src="https://img.shields.io/badge/语言-中文-red?style=for-the-badge" alt="Chinese">
  </a>
</p>

![Dashboard Preview](https://via.placeholder.com/800x400?text=Dashboard+Preview+Image)

</div>

---

<div id="-english-readme"></div>

## 🇺🇸 English Readme

### Introduction
**Cloudflare VLESS Node Manager** is a Go-based utility designed to generate and manage VLESS node configurations using Cloudflare's preferred IPs. It features a modern, responsive Web Dashboard that allows you to manage UUIDs, domains, and IP sources without touching configuration files manually.

### ✨ Key Features

*   **Web Dashboard**: Visual management of UUIDs, Domains, Ports, and IP sources.
*   **IP Auto-Fetch & Filtering**:
    *   Automatically fetch Cloudflare preferred IPs from a remote URL.
    *   **Prefix Allowlist**: Only keep IPs starting with specific prefixes (e.g., `104.16|172.`) to ensure quality.
*   **Smart Node Generation**:
    *   **Hourly Rotation**: To balance load and prevent blocking, the program automatically rotates through your list of UUIDs/Domains **every hour**.
    *   **Batch Generation**: Combines the *currently selected* UUID/Domain with all filtered IPs to generate multiple entry points.
*   **Seamless Experience**:
    *   **AJAX Saving**: Save configurations without refreshing the page.
    *   **Auto-Repair**: Automatically generates/repairs `config.json` and `dashboard.html` if missing.
    *   **Full TLS Support**: Correctly handles HTTP `Host` headers and TLS **SNI (Server Name Indication)** to ensure successful handshakes.
*   **Subscription API**: Provides a `/sub` endpoint returning a JSON list of nodes, compatible with Xray/Sing-box clients.

### 🛠️ Installation & Usage

#### Prerequisites
*   [Go (Golang)](https://go.dev/dl/) installed (v1.18+ recommended).

#### Steps
1.  **Run Directly**:
    ```bash
    go run main.go
    ```
2.  **Or Build & Run**:
    ```bash
    go build -o cf-manager main.go
    ./cf-manager
    ```
3.  **Access Dashboard**:
    Open your browser and visit: `http://localhost:1111` (Default Port)

### ⚙️ Configuration

The program automatically creates a working directory in your user home folder (e.g., `~/cloudflare_nodes/` or `C:\Users\Name\cloudflare_nodes\`).

| Parameter | Description | Default |
| :--- | :--- | :--- |
| **Web Port** | The port for the Web Dashboard. | `1111` |
| **Node Port** | The VLESS port (usually CF HTTPS port). | `443` |
| **IP URL** | URL to a text file containing Cloudflare IPs (one per line). | *Example URL* |
| **IP Allow Prefix** | **Allowlist**. Filter IPs by prefix. Separate multiple prefixes with `\|`. <br>Example: `104.16\|172.67` (Leave empty to allow all). | Empty |
| **UUID / Domain** | Your VLESS server credentials. One pair per row. | - |

### 🔄 Hourly Rotation Logic
To prevent excessive traffic on a single domain:
1.  The system checks the current **hour** (0-23).
2.  Formula: `Index = Current_Hour % Total_UUID_Configs`.
3.  The `/sub` API only returns nodes generated using the **currently selected** UUID/Domain pair.
4.  Clients updating subscriptions hourly will automatically switch to the next domain.

### 🔌 API Endpoints
*   `GET /`: The Web Dashboard.
*   `GET /sub`: Returns the generated node list (JSON).
*   `POST /save`: Saves settings and refreshes IPs (AJAX).

---
**Disclaimer**: This project is for educational and technical research purposes only.

[↑ Back to Top](#cloudflare-vless-node-manager)

---

<div id="-中文说明"></div>

## 🇨🇳 中文说明

### 简介
**Cloudflare VLESS Node Manager** 是一个基于 Go 语言开发的轻量级工具，用于管理和生成基于 Cloudflare 优选 IP 的 VLESS 节点配置。它内置了一个现代化的 Web 控制面板，支持 IP 自动获取、前缀白名单过滤、多域名轮询以及实时预览。

### ✨ 核心功能

*   **Web 可视化管理**：通过浏览器轻松配置 UUID、域名、端口和 IP 源。
*   **IP 自动获取与过滤**：
    *   支持从远程 URL 拉取 Cloudflare 优选 IP 列表。
    *   **前缀白名单 (Prefix Allowlist)**：支持通过前缀（如 `104.16|172.`）只保留指定网段的 IP，过滤掉质量差的 IP。
*   **智能节点生成**：
    *   **每小时轮询**：为了负载均衡和防封锁，程序**每小时**自动从配置的 UUID/域名列表中轮换选择一组。
    *   **批量生成**：将当前时段选中的 UUID/域名与所有优选 IP 组合，生成多个节点入口。
*   **无缝体验**：
    *   **AJAX 无刷新保存**：保存配置更流畅，无需重新加载页面。
    *   **自动修复**：自动生成或修复缺失的 `config.json` 和 `dashboard.html` 文件。
    *   **完整 TLS 支持**：自动处理 HTTP `Host` 头和 TLS **SNI (Server Name Indication)**，完美解决握手失败问题。
*   **订阅接口**：提供 `/sub` 接口输出 JSON 格式的节点列表，可直接被 Xray/Sing-box 等客户端解析。

### 🛠️ 安装与运行

#### 前置要求
*   已安装 [Go (Golang)](https://go.dev/dl/) 环境 (建议 1.18+)。

#### 运行步骤
1.  **直接运行**:
    ```bash
    go run main.go
    ```
2.  **编译运行**:
    ```bash
    go build -o cf-manager main.go
    ./cf-manager
    ```
3.  **访问控制面板**:
    打开浏览器访问：`http://localhost:1111` (默认端口)

### ⚙️ 配置说明

程序启动后会自动在用户主目录下创建工作目录（例如 Windows 下为 `C:\Users\用户名\cloudflare_nodes\`）。

| 参数项 | 说明 | 默认值 |
| :--- | :--- | :--- |
| **Web Port** | 管理面板的访问端口。 | `1111` |
| **Node Port** | 生成的 VLESS 节点连接端口（通常是 CF 的 HTTPS 端口）。 | `443` |
| **IP URL** | 获取 Cloudflare 优选 IP 的远程文本文件地址（每行一个 IP）。 | *示例地址* |
| **IP Allow Prefix** | **IP 白名单前缀**。多个前缀用 `\|` 分隔。<br>例如：`104.16\|172.67` 表示只保留这两个网段开头的 IP。留空则不过滤。 | 空 |
| **UUID / Domain** | 你的 VLESS 服务器凭证列表。每行一对。 | - |

### 🔄 每小时轮询机制 (Hourly Rotation)
为了避免单一域名流量过大或被针对，本程序采用时间片轮询机制：
1.  程序读取当前系统时间的**小时数** (0-23)。
2.  计算公式：`当前索引 = 当前小时 % UUID域名列表总数`。
3.  `/sub` 接口只会返回**当前时段被选中**的那一组 UUID/Domain 生成的节点。
4.  **效果**：客户端每小时刷新订阅时，会自动切换到下一个域名/UUID 组合，实现流量分摊。

### 🔌 API 接口
*   `GET /`: Web 控制面板页面。
*   `GET /sub`: 获取生成的节点列表（JSON 格式）。
*   `POST /save`: 保存配置并刷新 IP 缓存（AJAX 调用）。

---
**免责声明**: 本项目仅供网络技术研究和学习使用，请勿用于任何非法用途。

[↑ 回到顶部](#cloudflare-vless-node-manager)
