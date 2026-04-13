# 设计说明

## 目标

`goctl-i18n` 的目标是作为 `goctl api plugin` 的增强插件，在不侵入 go-zero 既有生成流程的前提下，为请求参数校验和多语言翻译补齐基础设施。

## 执行流程

1. `goctl api plugin` 将 `.api` 路径、输出目录和样式通过标准输入传给插件
2. 插件读取目标项目 `go.mod`，识别模块路径
3. 插件定位目标项目中的 `internal/types/types.go`
4. 插件从 `.api` AST 中收集被路由引用且以 `Req` 结尾的请求结构体
5. 插件更新 `types.go`
6. 插件按参数选择是否生成 `validation.go`、`translator.go`、`i18n.go` 与语言资源

## 核心分层

- `cmd/goctl-i18n/main.go`
  - 负责程序入口
- `internal/plugin/options.go`
  - 负责子命令与 flags 解析
- `internal/plugin/parser.go`
  - 负责读取 goctl 上下文、识别模块路径和目标文件路径
- `internal/plugin/writer.go`
  - 负责幂等更新 `types.go` 与模板文件写入
- `internal/plugin/generator.go`
  - 负责模板渲染

## Validate 接入策略

- 通过为 `*Req` 类型追加 `Validate() error` 接入 go-zero 请求解析流程
- 共享 `validator.New()` 实例命名为 `validate`
- 自定义校验函数注册拆到 `validation.go`，减少对 `types.go` 的重复改写

## i18n 设计

- 使用 `go-i18n` 作为翻译底座
- 默认生成：
  - `internal/i18n/i18n.go`
  - `internal/i18n/locales/active.zh.yaml`
  - `internal/i18n/locales/active.en.yaml`
- 语言资源使用 YAML 格式，首版聚焦 validator 常用标签
- 对外翻译入口为 `Translate(err, lang string) error`
- 插件不负责从 HTTP Header 自动取语言，语言由业务显式传入

## 幂等策略

- `types.go` 使用 Go AST 检查 import、共享变量和已有 `Validate()` 方法
- 已存在的 `validation.go`、`translator.go`、`i18n.go` 与语言资源默认不覆盖
- 重复执行插件时，同一个请求结构体不会重复追加 `Validate()` 方法

## 当前边界

- 当前只处理被路由引用且名称以 `Req` 结尾的请求结构体
- 当前字段展示采用 `validator.FieldError.Field()` 返回值
- 当前语言资源只覆盖首批常用 validator 标签
