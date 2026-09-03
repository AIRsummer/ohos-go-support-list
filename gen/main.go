// gen 从 support_list.md 生成 docs/data.json。
//
// 用法：
//
//	go run ./gen
//
// 数据校验失败时以非零码退出并打印所有错误行。
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

type Package struct {
	Name    string `json:"name"`
	Module  string `json:"module"`
	Version string `json:"version"`
	// Source: "official"=官方源直接可用；"repo"=需通过鸿蒙制品仓下载
	Source string `json:"source"`
}

type Data struct {
	LastUpdated string    `json:"last_updated"`
	Total       int       `json:"total"`
	Official    int       `json:"official"`
	Repo        int       `json:"repo"`
	Packages    []Package `json:"packages"`
}

var versionRe = regexp.MustCompile(`^v\d+(\.\d+)*(\.\d+)?([-+~][0-9A-Za-z.-]+)*$|^v\d+\.\d+\.\d+-\d{14}-[0-9a-f]+$`)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	if !fileExists(filepath.Join(root, "support_list.md")) {
		// 兼容：从 gen/ 目录直接运行（go run main.go / go run .），自动回退到上级仓库根目录
		if fileExists(filepath.Join(root, "..", "support_list.md")) {
			root = filepath.Join(root, "..")
		}
	}
	if !fileExists(filepath.Join(root, "support_list.md")) {
		fatal("在 %s 及其上级目录均未找到 support_list.md，请指定仓库根目录路径", root)
	}
	src := filepath.Join(root, "support_list.md")

	raw, err := os.ReadFile(src)
	if err != nil {
		fatal("读取 %s 失败: %v", src, err)
	}

	pkgs, errs := parse(string(raw))
	if len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "数据校验失败：")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "  "+e)
		}
		os.Exit(1)
	}

	// 校验通过后再按 module 路径去重检查（表格行可能重复收录同一 module）
	seen := map[string]int{}
	for i, p := range pkgs {
		if prev, ok := seen[p.Module]; ok {
			errs = append(errs, fmt.Sprintf("第 %d 行 module 路径 %q 与第 %d 行重复", i+1, p.Module, prev+1))
		} else {
			seen[p.Module] = i
		}
	}
	if len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "数据校验失败：")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "  "+e)
		}
		os.Exit(1)
	}

	data := Data{LastUpdated: time.Now().Format("2006-01-02"), Total: len(pkgs), Packages: pkgs}
	for _, p := range pkgs {
		if p.Source == "repo" {
			data.Repo++
		} else {
			data.Official++
		}
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fatal("序列化失败: %v", err)
	}
	out = append(out, '\n')

	dst := filepath.Join(root, "docs", "data.json")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		fatal("创建 docs 目录失败: %v", err)
	}
	if err := os.WriteFile(dst, out, 0o644); err != nil {
		fatal("写入 %s 失败: %v", dst, err)
	}

	fmt.Printf("已生成 %s：%d 个包（官方源 %d / 制品仓 %d）\n", dst, data.Total, data.Official, data.Repo)
}

// parse 解析 markdown 表格。表格列：包名 | 版本 | 仓库地址(module路径) | 获取方式
func parse(content string) ([]Package, []string) {
	var pkgs []Package
	var errs []string
	inTable := false
	lineNo := 0

	for _, line := range strings.Split(content, "\n") {
		lineNo++
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "|") {
			inTable = false
			continue
		}

		cols := splitRow(trimmed)
		// 表头行
		if !inTable {
			if len(cols) >= 4 && strings.Contains(cols[0], "包名") {
				inTable = true
			}
			continue
		}
		// 跳过分隔行 |---|---|
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
			// 官方源(非Go包) 等带备注的形式，保留备注由前端展示
			src = "official"
		default:
			errs = append(errs, fmt.Sprintf("第 %d 行获取方式无法识别：%q（仅支持 官方源 / 制品仓）", lineNo, source))
			continue
		}

		if !versionRe.MatchString(version) {
			errs = append(errs, fmt.Sprintf("第 %d 行版本号格式异常：%q", lineNo, version))
		}
		if module == "" || strings.ContainsAny(module, " ") {
			errs = append(errs, fmt.Sprintf("第 %d 行 module 路径为空或含空格：%q", lineNo, module))
		}

		pkgs = append(pkgs, Package{Name: name, Module: module, Version: version, Source: src})
	}

	if len(pkgs) == 0 {
		errs = append(errs, "未解析到任何数据行，请检查表格格式")
	}

	sort.Slice(pkgs, func(i, j int) bool {
		return strings.ToLower(pkgs[i].Module) < strings.ToLower(pkgs[j].Module)
	})
	return pkgs, errs
}

func splitRow(line string) []string {
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func isSeparator(cols []string) bool {
	for _, c := range cols {
		if c != "" && !strings.HasPrefix(c, "-") {
			return false
		}
	}
	return true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
