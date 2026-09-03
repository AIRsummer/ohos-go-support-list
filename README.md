# OpenHarmony Go 包支持列表

[通过 GitHub Pages 在线查看支持列表](https://airsummer.github.io/ohos-go-support-list/)

本项目展示已在 OpenHarmony/HarmonyOS 平台验证的 Go 三方库，可通过 GitHub Pages 按包名或 module 路径搜索，并按获取方式（制品仓 / 官方源）筛选。

获取方式为「制品仓」的包经过鸿蒙适配后发布在鸿蒙制品仓，需将 GOPROXY 指向制品仓拉取；「官方源」的包无需配置直接使用。详见主仓库 [Go_Package_For_HarmonyOS](https://gitcode.com/OpenHarmonyPCDeveloper/Goland_Package_For_HarmonyOS) 的[社区源章节](https://gitcode.com/OpenHarmonyPCDeveloper/Goland_Package_For_HarmonyOS#社区源鸿蒙制品仓)。

## 项目结构

```text
├── index.html          # GitHub Pages 查询页面（docs/）
├── support_list.md     # 基础数据源（人工维护，与主仓库保持同步）
├── gen/                # 静态数据生成器（Go，仅标准库）
│   ├── main.go         # 解析校验 support_list.md，输出 docs/data.json
│   └── go.mod
└── docs/
    ├── index.html      # 查询页面（原生 JS，无框架依赖）
    ├── data.json       # 生成产物（git 提交，Pages 发布目录）
    └── go-logo-blue.svg
```

## 更新数据

编辑 `support_list.md` 后执行：

```bash
go -C gen run .   # 在仓库根目录执行
```

生成器会校验重复 module 路径、版本号格式、获取方式枚举值（官方源 / 制品仓），校验失败时列出所有错误行并以非零码退出。生成完成后提交 `support_list.md` 和 `docs/data.json`，推送后 GitHub Pages 自动更新。

## 与主仓库的关系

- 数据与工具链维护在主仓库 [Go_Package_For_HarmonyOS](https://gitcode.com/OpenHarmonyPCDeveloper/Goland_Package_For_HarmonyOS)（GitCode）
- 本仓库用于 GitHub Pages 托管查询页面（GitCode 暂无 Pages 功能），更新时将主仓库的 `support_list.md`、`docs/`、`gen/` 同步推送至此
