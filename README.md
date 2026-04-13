# goctl-i18n

`goctl-i18n` 是一个面向 `goctl api plugin` 的插件，用于为 go-zero API 生成结果补充两类能力：

- `validate`：为请求结构体自动生成 `Validate() error`
- `i18n`：基于 `go-i18n` 生成校验错误翻译层与中英文 YAML 语言资源

## 功能说明

- 自动识别 `.api` 中被路由引用且以 `Req` 结尾的请求结构体
- 自动为各服务的 `internal/types/types.go` 追加 `Validate()` 方法
- 将共享校验器、校验扩展、翻译器和 i18n 资源统一生成到单仓根目录 `tools/goctli18n`
- 可选生成共享 `validation.go`，提供自定义校验注册骨架
- 可选生成共享 `translator.go`，提供 `Translate(err, lang)` 多语言翻译入口
- 可选生成中英文两套 YAML 语言资源
- 重复执行保持幂等，不会重复追加同一批 `Validate()` 方法

## 安装

```bash
go install github.com/xucxapp/goctl-i18n/cmd/goctl-i18n@latest
```

Windows 下请确认 `%GOPATH%\bin` 已加入 `PATH`。

## 使用方式

先用 `goctl` 生成 go-zero API 项目：

```bash
goctl api go -api demo.api -dir .
```

然后执行插件：

```bash
goctl api plugin -api demo.api -dir . -p goctl-i18n="validate"
```

生成校验翻译与 i18n 文件：

```bash
goctl api plugin -api demo.api -dir . -p goctl-i18n="validate --translator"
```

同时生成自定义校验骨架：

```bash
goctl api plugin -api demo.api -dir . -p goctl-i18n="validate --translator --custom"
```

开启调试日志：

```bash
goctl api plugin -api demo.api -dir . -p goctl-i18n="validate --translator --custom --debug"
```

## 生成结果

默认会修改：

- `internal/types/types.go`

按选项生成：

- `tools/goctli18n/validator.go`
- `tools/goctli18n/validation.go`
- `tools/goctli18n/translator.go`
- `tools/goctli18n/i18n.go`
- `tools/goctli18n/locales/active.zh.yaml`
- `tools/goctli18n/locales/active.en.yaml`

## Translate 示例

当开启 `--translator` 后，可在业务中显式传入语言：

```go
err := req.Validate()
if err != nil {
	return nil, goctli18n.Translate(err, "zh")
}
```

或：

```go
err := req.Validate()
if err != nil {
	return nil, goctli18n.Translate(err, "en")
}
```

## 示例文件

- API 示例：`examples/demo.api`
- 预期输出说明：`examples/expected/README.md`

## 开发验证

```bash
go test ./...
go build ./...
```

完整设计与使用说明见：

- `docs/design.md`
- `docs/usage.md`

## 致谢

`validate` 部分的设计与接入方式借鉴了 [linabellbiu/goctl-validate](https://github.com/linabellbiu/goctl-validate)。
