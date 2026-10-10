// gen 从 support_list.md 生成 docs/data/ 下的分类数据文件。
//
// 用法：go run ./gen
//
// 产出结构（参照 ohos-python/ohos-python-support-list）：
//   docs/data/index.json          — 总索引（分类统计）
//   docs/data/all.json            — 全部包
//   docs/data/categories/{id}.json — 按分类分文件
//
// 分类由 classifyPackage() 根据包名 + module 路径自动推断。
// 分类 ID 与 Python 仓（ohos-python-support-list）保持一致。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ============================================================
// 分类定义（ID 与 ohos-python/ohos-python-support-list 一致）
//
// Tokens:   对包名分词后做完整词匹配（精确，防误报）
// Fragments: 对包名+module 路径做子串匹配（宽泛，覆盖面大）
// ============================================================

type Category struct {
	ID        string
	Label     string
	Tokens    map[string]bool
	Fragments []string
}

var Categories = []Category{
	{
		ID:    "ai-ml",
		Label: "AI 与机器学习",
		Tokens: map[string]bool{
			"ai": true, "ml": true, "llm": true, "nlp": true,
			"agent": true, "agents": true, "mcp": true,
		},
		Fragments: []string{
			"tensorflow", "torch", "keras", "scikit", "sklearn",
			"xgboost", "lightgbm", "huggingface", "transformer",
			"onnx", "langchain", "llama", "machine-learning",
			"deep-learning", "neural", "embedding", "openai",
		},
	},
	{
		ID:    "database",
		Label: "数据库与存储",
		Tokens: map[string]bool{
			"db": true, "sql": true, "orm": true,
		},
		Fragments: []string{
			"database", "sqlalchemy", "sqlite", "postgres", "psycopg",
			"mysql", "mariadb", "mongodb", "pymongo", "redis",
			"cassandra", "clickhouse", "elasticsearch", "neo4j",
			"influxdb", "dynamodb", "couchdb", "duckdb", "lmdb",
			"leveldb", "rocksdb", "pebble", "gorm", "entgo",
			"sqlx", "mgo", "bbolt", "badger", "chromadb",
		},
	},
	{
		ID:    "data-science",
		Label: "数据科学与计算",
		Tokens: map[string]bool{
			"math": true, "stats": true, "plot": true,
			"data": true, "dataset": true, "etl": true,
		},
		Fragments: []string{
			"numpy", "scipy", "pandas", "polars", "pyarrow",
			"matplotlib", "seaborn", "plotly", "sympy",
			"jupyter", "notebook", "dask", "ray", "spark",
			"parquet", "hdf5", "h5py", "netcdf", "xarray",
			"scientific", "analytics", "visualization",
			"dataframe", "timeseries", "time-series",
		},
	},
	{
		ID:    "web-network",
		Label: "Web 与网络",
		Tokens: map[string]bool{
			"api": true, "http": true, "web": true, "rpc": true,
			"oauth": true, "jwt": true, "rest": true, "grpc": true,
			"dns": true, "proxy": true, "socket": true, "tcp": true,
			"udp": true, "websocket": true, "crawler": true,
		},
		Fragments: []string{
			"gin-gonic", "echo", "fiber", "chi", "gorilla/mux",
			"fasthttp", "resty", "go-resty", "protobuf",
			"colly", "crawler", "scrape", "martian",
			"websocket", "netlink", "packet", "pcap",
			"libpfring", "sarama", "amqp", "kafka",
		},
	},
	{
		ID:    "dev-test",
		Label: "开发工具与测试",
		Tokens: map[string]bool{
			"test": true, "testing": true, "lint": true, "debug": true,
			"build": true, "docs": true, "types": true, "mock": true,
			"assert": true, "bench": true, "stub": true, "stubs": true,
		},
		Fragments: []string{
			"pytest", "unittest", "coverage", "hypothesis",
			"ruff", "flake8", "pylint", "mypy", "black", "isort",
			"pre-commit", "sphinx", "mkdocs", "setuptools",
			"cookiecutter", "profiler", "debugger", "tree-sitter",
			"testify", "golangci", "goimports", "gofmt", "staticcheck",
			"govet", "benchmark", "esbuild", "vite", "webpack",
		},
	},
	{
		ID:    "system-cloud",
		Label: "基础设施与云服务",
		Tokens: map[string]bool{
			"cloud": true, "docker": true, "kubernetes": true, "k8s": true,
			"container": true, "deploy": true, "terraform": true,
			"ansible": true, "helm": true, "monitor": true,
			"keyboard": true, "mouse": true, "clipboard": true,
			"screen": true, "window": true, "tty": true, "signal": true,
			"process": true, "crypto": true, "security": true,
			"tls": true, "ssl": true, "cert": true, "hash": true,
		},
		Fragments: []string{
			"boto", "aws", "azure", "google-cloud", "openstack",
			"aliyun", "tencentcloud", "huaweicound", "docker",
			"podman", "argocd", "argo", "terraform", "ansible",
			"prometheus", "grafana", "opentelemetry", "kurtosis",
			"dapr", "istio", "envoy", "nginx", "caddy",
			"keyboard", "clipboard", "screen", "term", "tty",
			"signal", "process", "syscall", "user", "termios",
			"console", "cursor", "cryptography", "openssl",
			"bcrypt", "argon2", "securejoin", "x509", "pem",
			"ssh", "systemd", "filesystem", "watchdog", "cron",
			"scheduler", "celery", "supervisor",
		},
	},
	{
		ID:    "office",
		Label: "通用办公",
		Tokens: map[string]bool{
			"browser": true, "pdf": true, "ppt": true, "pptx": true,
			"docx": true, "excel": true, "xls": true, "xlsx": true,
			"office": true, "spreadsheet": true, "selenium": true,
			"webdriver": true, "chrome": true, "playwright": true,
		},
		Fragments: []string{
			"selenium", "playwright", "browser", "chromedp",
			"rod", "stealth", "puppeteer", "pdf",
			"openpyxl", "xlsx", "spreadsheet", "docx",
			"wordprocessing", "document-convert", "gmailctl",
			"lark-cli", "dingtalk",
		},
	},
	{
		ID:    "other",
		Label: "其他",
		Tokens: map[string]bool{},
		Fragments: []string{},
	},
}

// 分类优先级（先匹配先返回）
var CategoryPriority = []string{
	"dev-test", "database", "data-science", "office", "ai-ml",
	"web-network", "system-cloud",
}

var CategoriesByID = func() map[string]Category {
	m := make(map[string]Category)
	for _, c := range Categories {
		m[c.ID] = c
	}
	return m
}()

func classifyPackage(name, module string) string {
	normalized := strings.ToLower(strings.NewReplacer("_", "-", ".", "-").Replace(name + " " + module))
	tokens := make(map[string]bool)
	for _, t := range strings.FieldsFunc(normalized, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	}) {
		tokens[t] = true
	}
	for _, catID := range CategoryPriority {
		cat := CategoriesByID[catID]
		for t := range tokens {
			if cat.Tokens[t] {
				return cat.ID
			}
		}
		for _, f := range cat.Fragments {
			if strings.Contains(normalized, f) {
				return cat.ID
			}
		}
	}
	return "other"
}

// ============================================================
// 数据结构与生成
// ============================================================

type Package struct {
	Name     string `json:"name"`
	Module   string `json:"module"`
	Version  string `json:"version"`
	Category string `json:"category"`
	Source   string `json:"source"`
}

type CategoryStat struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type Index struct {
	LastUpdated string         `json:"last_updated"`
	Total       int            `json:"total"`
	Official    int            `json:"official"`
	Repo        int            `json:"repo"`
	Categories  []CategoryStat `json:"categories"`
}

var versionRe = regexp.MustCompile(`^v\d+(\.\d+)*(\.\d+)?([-+~][0-9A-Za-z.-]+)*$|^v\d+\.\d+\.\d+-\d{14}-[0-9a-f]+$`)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	src := filepath.Join(root, "support_list.md")
	data, err := os.ReadFile(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取 %s: %v\n", src, err)
		os.Exit(1)
	}

	var pkgs []Package
	var errs []string
	inTable := false
	lineNo := 0

	for _, line := range strings.Split(string(data), "\n") {
		lineNo++
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ">") {
			continue
		}
		cols := splitRow(trimmed)
		if !inTable {
			if len(cols) >= 4 && strings.Contains(cols[0], "包名") {
				inTable = true
			}
			continue
		}
		if isSeparator(cols) {
			continue
		}
		if len(cols) != 4 {
			errs = append(errs, fmt.Sprintf("第 %d 行列数不是 4：%q", lineNo, trimmed))
			continue
		}
		name, version, module, source := cols[0], cols[1], cols[2], cols[3]
		var src string
		switch {
		case source == "官方源":
			src = "official"
		case source == "制品仓":
			src = "repo"
		case strings.HasPrefix(source, "官方源"):
			src = "official"
		default:
			errs = append(errs, fmt.Sprintf("第 %d 行获取方式无法识别：%q", lineNo, source))
			continue
		}
		if !versionRe.MatchString(version) {
			errs = append(errs, fmt.Sprintf("第 %d 行版本号格式异常：%q", lineNo, version))
		}
		if module == "" || strings.ContainsAny(module, " ") {
			errs = append(errs, fmt.Sprintf("第 %d 行 module 路径为空或含空格：%q", lineNo, module))
		}
		category := classifyPackage(name, module)
		pkgs = append(pkgs, Package{Name: name, Module: module, Version: version, Category: category, Source: src})
	}

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(1)
	}
	if len(pkgs) == 0 {
		fmt.Fprintln(os.Stderr, "未解析到任何数据行")
		os.Exit(1)
	}

	seen := make(map[string]bool)
	for _, p := range pkgs {
		if seen[p.Module] {
			errs = append(errs, fmt.Sprintf("重复 module：%s（包 %s）", p.Module, p.Name))
		}
		seen[p.Module] = true
	}
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(1)
	}

	sort.Slice(pkgs, func(i, j int) bool {
		return strings.ToLower(pkgs[i].Name) < strings.ToLower(pkgs[j].Name)
	})

	dataDir := filepath.Join(root, "docs", "data")
	catDir := filepath.Join(dataDir, "categories")
	os.RemoveAll(catDir)
	os.MkdirAll(catDir, 0755)

	writeJSON(filepath.Join(dataDir, "all.json"), pkgs)

	officialCount, repoCount := 0, 0
	catCounts := make(map[string]int)
	catPkgs := make(map[string][]Package)
	for _, p := range pkgs {
		catCounts[p.Category]++
		catPkgs[p.Category] = append(catPkgs[p.Category], p)
		if p.Source == "repo" {
			repoCount++
		} else {
			officialCount++
		}
	}
	for _, cat := range Categories {
		if ps, ok := catPkgs[cat.ID]; ok {
			writeJSON(filepath.Join(catDir, cat.ID+".json"), ps)
		} else {
			writeJSON(filepath.Join(catDir, cat.ID+".json"), []Package{})
		}
	}

	var catStats []CategoryStat
	for _, cat := range Categories {
		catStats = append(catStats, CategoryStat{ID: cat.ID, Label: cat.Label, Count: catCounts[cat.ID]})
	}
	index := Index{
		LastUpdated: time.Now().Format("2006-01-02"),
		Total:       len(pkgs),
		Official:    officialCount,
		Repo:        repoCount,
		Categories:  catStats,
	}
	writeJSON(filepath.Join(dataDir, "index.json"), index)

	fmt.Printf("已生成 %d 个包（官方源 %d / 制品仓 %d）\n", index.Total, index.Official, index.Repo)
	for _, cat := range Categories {
		fmt.Printf("  %-16s %s (%d)\n", cat.ID, cat.Label, catCounts[cat.ID])
	}
}

func splitRow(line string) []string {
	line = strings.TrimPrefix(strings.TrimSuffix(line, "|"), "|")
	parts := strings.Split(line, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func isSeparator(cols []string) bool {
	for _, c := range cols {
		if !strings.HasPrefix(c, "---") {
			return false
		}
	}
	return len(cols) > 0
}

func writeJSON(path string, v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}
