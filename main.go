package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ================= 1. 结构体定义 =================

type UUIDDomain struct {
	UUID   string `json:"uuid"`
	Domain string `json:"domain"`
}

type TLSUTLS struct {
	Enabled     bool   `json:"enabled"`
	Fingerprint string `json:"fingerprint"`
}

type TLSType struct {
	Enabled        bool    `json:"enabled"`
	ServerName     string  `json:"server_name"` // [修复] 新增 SNI 字段
	Insecure       bool    `json:"insecure"`
	RecordFragment bool    `json:"record_fragment"`
	UTLS           TLSUTLS `json:"utls"`
}

type Transport struct {
	EarlyDataHeaderName string            `json:"early_data_header_name"`
	Headers             map[string]string `json:"headers"`
	MaxEarlyData        int               `json:"max_early_data"`
	Path                string            `json:"path"`
	Type                string            `json:"type"`
}

type NodeTemplate struct {
	PacketEncoding string    `json:"packet_encoding"`
	NodePort       int       `json:"node_port"`
	Type           string    `json:"type"`
	TLS            TLSType   `json:"tls"`
	Transport      Transport `json:"transport"`
}

type Config struct {
	WebPort     int          `json:"web_port"`
	IPURL       string       `json:"ip_url"`
	IPPrefix    string       `json:"ip_prefix"`
	UUIDDomains []UUIDDomain `json:"uuid_domains"`
	NodeTpl     NodeTemplate `json:"node_tpl"`
}

// ================= 2. 全局变量 =================

var (
	config       Config
	mu           sync.RWMutex
	tmpl         *template.Template
	workDir      string
	configFile   string
	templateFile string
	ips          []string
	ipsTime      time.Time
)

// ================= 3. HTML 模板 (含 AJAX 逻辑) =================

var defaultHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Node Dashboard</title>
<style>
body{font-family:Arial,sans-serif;background:#1e1e1e;color:#cfcfcf;margin:0;padding:20px;line-height:1.5}
h1{color:#f0f0f0;margin-bottom:20px;border-bottom:1px solid #444;padding-bottom:10px}
.panel{border:1px solid #444;background:#252525;border-radius:6px;margin-bottom:20px}
.panel-header{display:flex;align-items:center;cursor:pointer;padding:12px;background:#2e2e2e;border-radius:4px;font-weight:bold;margin-bottom:1px;transition:background 0.2s}
.panel-header:hover{background:#3e3e3e}
.panel-header svg{width:16px;height:16px;margin-right:8px;transition:transform 0.3s;fill:#cfcfcf}
.panel-content{max-height:0;overflow:hidden;transition:max-height 0.3s ease,padding 0.3s ease;padding:0 10px;}
.panel-content.open {padding: 10px; border-top: 1px solid #444;}
.form-group{margin-bottom:12px}
label{display:inline-block;width:180px;font-weight:bold}
input[type="text"],input[type="number"]{padding:8px;background:#2e2e2e;border:1px solid #555;color:#fff;border-radius:4px;width:200px}
input[name="ip_url"], input[name="ip_prefix"]{width:400px}
.uuid-row{display:flex;align-items:center;margin-bottom:10px;background:#2a2a2a;padding:10px;border-radius:4px}
.uuid-row input{margin-right:10px;flex:1}
.btn{cursor:pointer;padding:6px 12px;border:none;border-radius:4px;font-weight:bold;color:#fff;transition:background 0.2s;margin-right:5px}
.btn-red{background:#c0392b}
.btn-red:hover{background:#e74c3c}
.btn-green{background:#27ae60}
.btn-green:hover{background:#2ecc71}
.btn-blue{background:#2980b9;font-size:16px;width:100%;margin-top:10px;padding:12px}
.btn-blue:hover{background:#3498db}
.btn:disabled{background:#555;cursor:not-allowed;color:#aaa}
pre#preview{background:#111;padding:15px;border:1px solid #333;color:#0f0;border-radius:4px;overflow:auto;min-height:200px;max-height:600px;font-family:Consolas,monaco,monospace}
.control-bar {padding: 10px; border-top: 1px solid #444; background: #2a2a2a;}

/* Toast Notification */
#toast {
    visibility: hidden;
    min-width: 250px;
    background-color: #27ae60;
    color: #fff;
    text-align: center;
    border-radius: 4px;
    padding: 16px;
    position: fixed;
    z-index: 1;
    right: 30px;
    top: 30px;
    font-size: 16px;
    box-shadow: 0 4px 6px rgba(0,0,0,0.3);
    opacity: 0;
    transition: opacity 0.5s, top 0.5s;
}
#toast.show {
    visibility: visible;
    opacity: 1;
    top: 50px;
}
</style>
<script>
// UI Interactions
function togglePanel(header){
    const panel = header.nextElementSibling;
    const svg = header.querySelector('svg');
    if(panel.style.maxHeight && panel.style.maxHeight !== "0px"){
        panel.style.maxHeight = "0";
        panel.classList.remove('open');
        if(svg) svg.style.transform = "rotate(0deg)";
    } else {
        panel.classList.add('open');
        panel.style.maxHeight = panel.scrollHeight + "px";
        if(svg) svg.style.transform = "rotate(90deg)";
    }
}
function expandPanel(id) {
    const panel = document.getElementById(id);
    if(panel) {
        panel.classList.add('open');
        panel.style.maxHeight = panel.scrollHeight + "px";
        const header = panel.previousElementSibling;
        const svg = header.querySelector('svg');
        if(svg) svg.style.transform = "rotate(90deg)";
    }
}
function addRow(){
    const c = document.getElementById('uuid-container');
    const i = c.children.length;
    const div = document.createElement('div');
    div.className = 'uuid-row';
    div.innerHTML = '<input type="text" name="uuid'+i+'" placeholder="UUID"><input type="text" name="domain'+i+'" placeholder="Domain"><button type="button" class="btn btn-red" onclick="removeRow(this)">Remove</button>';
    c.appendChild(div);
    expandPanel('uuid-container');
}
function removeRow(btn) {
    const row = btn.parentNode;
    const container = row.parentNode;
    row.remove();
    container.style.maxHeight = container.scrollHeight + "px";
}

// Data Handling
function fetchPreview(){
    const pre = document.getElementById('preview');
    pre.style.opacity = '0.5';
    fetch('/sub')
    .then(r => r.json())
    .then(d => { 
        pre.textContent = JSON.stringify(d, null, 2); 
        pre.style.opacity = '1';
    })
    .catch(e => { 
        pre.textContent = "Error: " + e; 
        pre.style.opacity = '1';
    });
}

function showToast(message) {
    const x = document.getElementById("toast");
    x.innerText = message;
    x.className = "show";
    setTimeout(function(){ x.className = x.className.replace("show", ""); }, 3000);
}

function saveSettings(e) {
    e.preventDefault(); // Prevent default form submission
    
    const btn = document.getElementById('save-btn');
    const originalText = btn.innerText;
    btn.innerText = "Saving...";
    btn.disabled = true;

    const form = document.getElementById('config-form');
    // Convert form data to URLSearchParams for standard POST body
    const formData = new URLSearchParams(new FormData(form));

    fetch('/save', {
        method: 'POST',
        body: formData,
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
    })
    .then(response => response.json())
    .then(data => {
        if(data.status === 'ok') {
            showToast("Settings Saved Successfully!");
            fetchPreview(); // Refresh preview
        } else {
            alert("Error saving: " + data.message);
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert("Network Error");
    })
    .finally(() => {
        btn.innerText = originalText;
        btn.disabled = false;
    });
    
    return false;
}

window.onload = function(){
    fetchPreview();
    document.querySelectorAll('.panel-header').forEach(h => {
        h.addEventListener('click', () => togglePanel(h));
        togglePanel(h); 
    });
}
</script>
</head>
<body>
<div id="toast">Saved</div>
<h1>Node Dashboard</h1>

<form id="config-form" onsubmit="saveSettings(event)">
    <div class="panel">
        <div class="panel-header"><svg viewBox="0 0 24 24"><path d="M8 5l8 7-8 7V5z"/></svg>Base Settings</div>
        <div class="panel-content">
            <div class="form-group">
                <label>Web Port:</label>
                <input type="number" name="web_port" value="{{.WebPort}}" min="1" max="65535">
            </div>
            <div class="form-group">
                <label>Node Port:</label>
                <input type="number" name="node_port" value="{{.NodeTpl.NodePort}}" min="1" max="65535">
            </div>
            <div class="form-group">
                <label>IP URL:</label>
                <input type="text" name="ip_url" value="{{.IPURL}}">
            </div>
            <div class="form-group">
                <label>IP Allow Prefix:</label>
                <input type="text" name="ip_prefix" value="{{.IPPrefix}}" placeholder="e.g. 104.16|172.67 (Empty = All)">
            </div>
        </div>
    </div>

    <div class="panel">
        <div class="panel-header"><svg viewBox="0 0 24 24"><path d="M8 5l8 7-8 7V5z"/></svg>UUID / Domain List</div>
        <div class="panel-content" id="uuid-container">
            {{range $i,$v := .UUIDDomains}}
            <div class="uuid-row">
                <input type="text" name="uuid{{$i}}" value="{{$v.UUID}}" placeholder="UUID">
                <input type="text" name="domain{{$i}}" value="{{$v.Domain}}" placeholder="Domain">
                <button type="button" class="btn btn-red" onclick="removeRow(this)">Remove</button>
            </div>
            {{end}}
        </div>
        <div class="control-bar">
            <button type="button" class="btn btn-green" onclick="addRow()">+ Add New Node</button>
        </div>
    </div>

    <button type="submit" id="save-btn" class="btn btn-blue">Save Configuration & Refresh</button>
</form>

<div class="panel">
    <div class="panel-header"><svg viewBox="0 0 24 24"><path d="M8 5l8 7-8 7V5z"/></svg>Node Preview (Changes Every Hour)</div>
    <div class="panel-content">
        <pre id="preview">Loading nodes...</pre>
    </div>
</div>

</body>
</html>`

// ================= 4. 主函数 =================

func main() {
	usr, _ := user.Current()
	workDir = filepath.Join(usr.HomeDir, "cloudflare_nodes")
	configFile = filepath.Join(workDir, "config.json")
	templateFile = filepath.Join(workDir, "dashboard.html")

	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		_ = os.MkdirAll(workDir, 0755)
	}

	err := ioutil.WriteFile(templateFile, []byte(defaultHTML), 0644)
	if err != nil {
		log.Println("Warning: Could not write template file:", err)
	}

	loadConfig()
	updateIPs()

	var parseErr error
	tmpl, parseErr = template.New("dashboard").ParseFiles(templateFile)
	if parseErr != nil {
		tmpl = template.Must(template.New("dashboard").Parse(defaultHTML))
	}

	// 注册路由
	http.HandleFunc("/", dashboardHandler) // 只渲染页面
	http.HandleFunc("/save", saveHandler)  // [新] 处理 AJAX 保存
	http.HandleFunc("/sub", subHandler)    // 获取节点 JSON

	fmt.Printf("\n------------------------------------------------\n")
	fmt.Printf("Dashboard running at: http://localhost:%d\n", config.WebPort)
	fmt.Printf("Work Directory: %s\n", workDir)
	fmt.Printf("------------------------------------------------\n\n")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.WebPort), nil))
}

// ================= 5. 核心逻辑 =================

func updateIPs() {
	mu.Lock()
	defer mu.Unlock()

	if config.IPURL == "" || !strings.HasPrefix(config.IPURL, "http") {
		ips = []string{"127.0.0.1"}
		ipsTime = time.Now()
		return
	}

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(config.IPURL)
	if err != nil {
		log.Println("Error fetching IPs:", err)
		if len(ips) == 0 {
			ips = []string{"127.0.0.1"}
		}
		return
	}
	defer resp.Body.Close()

	var prefixes []string
	if config.IPPrefix != "" {
		parts := strings.Split(config.IPPrefix, "|")
		for _, p := range parts {
			if strings.TrimSpace(p) != "" {
				prefixes = append(prefixes, strings.TrimSpace(p))
			}
		}
	}

	var newIPs []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" { continue }

		shouldKeep := false
		if len(prefixes) == 0 {
			shouldKeep = true
		} else {
			for _, p := range prefixes {
				if strings.HasPrefix(line, p) {
					shouldKeep = true
					break
				}
			}
		}
		if shouldKeep {
			newIPs = append(newIPs, line)
		}
	}

	if len(newIPs) > 0 {
		ips = newIPs
		log.Printf("Fetched %d IPs. Prefix Filter: %v\n", len(ips), prefixes)
	} else {
		log.Println("Fetched list empty, keeping old IPs")
	}
	ipsTime = time.Now()
}

func loadConfig() {
	mu.Lock()
	defer mu.Unlock()

	data, err := ioutil.ReadFile(configFile)
	fileLoaded := false
	if err == nil {
		if json.Unmarshal(data, &config) == nil {
			fileLoaded = true
		}
	}

	if config.WebPort == 0 { config.WebPort = 1111 }
	if config.IPURL == "" { config.IPURL = "https://raw.githubusercontent.com/example/ip-list/main/ips.txt" }
	
	if len(config.UUIDDomains) == 0 {
		config.UUIDDomains = []UUIDDomain{{UUID: "", Domain: ""}}
	} else if len(config.UUIDDomains) == 1 && config.UUIDDomains[0].UUID == "" && config.UUIDDomains[0].Domain == "" {
		// keep empty row
	}

	if config.NodeTpl.NodePort == 0 { config.NodeTpl.NodePort = 443 }
	if config.NodeTpl.PacketEncoding == "" { config.NodeTpl.PacketEncoding = "xudp" }
	if config.NodeTpl.Type == "" { config.NodeTpl.Type = "vless" }
	if config.NodeTpl.TLS.UTLS.Fingerprint == "" { config.NodeTpl.TLS.UTLS.Fingerprint = "chrome" }
	if config.NodeTpl.Transport.EarlyDataHeaderName == "" { config.NodeTpl.Transport.EarlyDataHeaderName = "Sec-WebSocket-Protocol" }
	if config.NodeTpl.Transport.MaxEarlyData == 0 { config.NodeTpl.Transport.MaxEarlyData = 2560 }
	if config.NodeTpl.Transport.Path == "" { config.NodeTpl.Transport.Path = "/" }
	if config.NodeTpl.Transport.Type == "" { config.NodeTpl.Transport.Type = "ws" }
	
	if config.NodeTpl.Transport.Headers == nil {
		config.NodeTpl.Transport.Headers = make(map[string]string)
	}
	if _, ok := config.NodeTpl.Transport.Headers["Host"]; !ok {
		config.NodeTpl.Transport.Headers["Host"] = "example.com"
	}
	if _, ok := config.NodeTpl.Transport.Headers["User-Agent"]; !ok {
		config.NodeTpl.Transport.Headers["User-Agent"] = "Mozilla/5.0"
	}

	if !fileLoaded || err != nil {
		saveConfigInternalLocked()
	} else {
		saveConfigInternalLocked()
	}
}

func saveConfigInternalLocked() {
	data, _ := json.MarshalIndent(config, "", "  ")
	ioutil.WriteFile(configFile, data, fs.FileMode(0644))
}

// ================= 6. HTTP Handlers =================

// dashboardHandler: 仅处理 GET，渲染 HTML
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	
	t, err := template.ParseFiles(templateFile)
	if err != nil { t = tmpl }
	t.Execute(w, config)
}

// saveHandler: [新增] 处理 AJAX POST 请求
func saveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	r.ParseForm()
	mu.Lock()
	
	// 更新配置
	config.WebPort = atoi(r.FormValue("web_port"), config.WebPort)
	config.NodeTpl.NodePort = atoi(r.FormValue("node_port"), config.NodeTpl.NodePort)
	config.IPURL = r.FormValue("ip_url")
	config.IPPrefix = r.FormValue("ip_prefix")

	var uuids []UUIDDomain
	for i := 0; ; i++ {
		u := r.FormValue("uuid" + itoa(i))
		d := r.FormValue("domain" + itoa(i))
		if u == "" && d == "" && r.FormValue("uuid"+itoa(i+1)) == "" { break }
		if u != "" || d != "" {
			uuids = append(uuids, UUIDDomain{UUID: u, Domain: d})
		}
	}
	if len(uuids) == 0 { uuids = []UUIDDomain{{UUID: "", Domain: ""}} }
	config.UUIDDomains = uuids
	
	// 保存文件
	saveConfigInternalLocked()
	mu.Unlock()
	
	// 触发更新 IP
	updateIPs()

	// 返回 JSON 响应
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok"}`))
}

func subHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	needUpdate := time.Since(ipsTime) > 1*time.Hour
	mu.RUnlock()

	if needUpdate {
		updateIPs()
	}

	mu.RLock()
	defer mu.RUnlock()

	var nodes []map[string]interface{}
	currentIPs := ips
	if len(currentIPs) == 0 {
		currentIPs = []string{"127.0.0.1"}
	}
	if len(config.UUIDDomains) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodes)
		return
	}

	currentHour := time.Now().Hour()
	selectIndex := currentHour % len(config.UUIDDomains)
	selectedConfig := config.UUIDDomains[selectIndex]

	if selectedConfig.UUID == "" {
		found := false
		for _, u := range config.UUIDDomains {
			if u.UUID != "" {
				selectedConfig = u
				found = true
				break
			}
		}
		if !found {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(nodes)
			return
		}
	}

	for _, ip := range currentIPs {
		// 1. 设置 TLS SNI
		tlsConfig := config.NodeTpl.TLS
		tlsConfig.ServerName = selectedConfig.Domain // [修复] 显式设置 SNI 确保 TLS 握手成功

		// 2. 设置 Transport Host (深拷贝防止并发污染)
		transportConfig := config.NodeTpl.Transport
		newHeaders := make(map[string]string)
		for k, v := range transportConfig.Headers {
			newHeaders[k] = v
		}
		newHeaders["Host"] = selectedConfig.Domain
		transportConfig.Headers = newHeaders

		node := map[string]interface{}{
			"packet_encoding": config.NodeTpl.PacketEncoding,
			"server":          ip,
			"server_port":     config.NodeTpl.NodePort,
			"tag":             fmt.Sprintf("%s-%s", selectedConfig.Domain, ip),
			"type":            config.NodeTpl.Type,
			"uuid":            selectedConfig.UUID,
			"tls":             tlsConfig,       // 包含 ServerName
			"transport":       transportConfig, // 包含 Host Header
		}
		nodes = append(nodes, node)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

// ================= 7. 工具函数 =================

func atoi(s string, def int) int {
	var v int
	if _, err := fmt.Sscan(s, &v); err != nil { return def }
	return v
}

func itoa(i int) string { return fmt.Sprintf("%d", i) }