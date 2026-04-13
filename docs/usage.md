# 使用说明

## 前置条件

- 已安装 Go
- 已安装 `goctl`
- 目标项目已通过 `goctl api go` 生成基础 go-zero API 代码

## 1. 安装插件

```bash
go install github.com/xucxapp/goctl-i18n/cmd/goctl-i18n@latest
```

## 2. 生成 go-zero API 项目

```bash
goctl api go -api demo.api -dir .
```

## 3. 执行插件

仅生成 `Validate()`：

```bash
goctl api plugin -api demo.api -dir . -p goctl-i18n="validate"
```

生成翻译层与 i18n 资源：

```bash
goctl api plugin -api demo.api -dir . -p goctl-i18n="validate --translator"
```

额外生成自定义校验骨架：

```bash
goctl api plugin -api demo.api -dir . -p goctl-i18n="validate --translator --custom"
```

指定语言资源目录：

```bash
goctl api plugin -api demo.api -dir . -p goctl-i18n="validate --translator --locale-dir internal/i18n/locales"
```

## 4. 业务中使用 Translate

```go
if err := req.Validate(); err != nil {
	return nil, types.Translate(err, "zh")
}
```

如果业务已经拿到语言参数，也可以直接传入：

```go
lang := "en"
if err := req.Validate(); err != nil {
	return nil, types.Translate(err, lang)
}
```

## 5. 文件说明

- `internal/types/types.go`
  - 自动追加 `Validate()` 方法
- `internal/types/validation.go`
  - 自定义校验注册入口
- `internal/types/translator.go`
  - 校验错误翻译入口
- `internal/i18n/i18n.go`
  - `go-i18n` 初始化与本地化工具
- `internal/i18n/locales/*.yaml`
  - 语言资源文件

## 6. 本地验证

```bash
go test ./...
go build ./...
```
