package plugin

import (
	"fmt"
	"strings"
)

// Execute 执行插件命令。
func Execute(args []string) error {
	opts, err := ParseOptions(args)
	if err != nil {
		return err
	}

	ctx, err := LoadPluginContext()
	if err != nil {
		return fmt.Errorf("读取 goctl 插件上下文失败: %w", err)
	}

	project, err := DiscoverTargetProject(ctx, opts)
	if err != nil {
		return fmt.Errorf("解析目标项目失败: %w", err)
	}

	requestTypes := CollectRequestTypes(ctx)
	if len(requestTypes) == 0 {
		return fmt.Errorf("未找到需要处理的请求结构体，请确认 .api 中存在被路由引用且以 Req 结尾的类型")
	}

	if opts.Debug {
		debugf(opts, "api 文件: %s", ctx.ApiFilePath)
		debugf(opts, "服务目录: %s", ctx.Dir)
		debugf(opts, "模块根目录: %s", project.ModuleRootDir)
		debugf(opts, "模块路径: %s", project.ModulePath)
		debugf(opts, "types 文件: %s", project.TypesFilePath)
		debugf(opts, "共享目录: %s", project.SharedDir)
		debugf(opts, "请求结构体: %s", strings.Join(requestTypes, ", "))
	}

	if err := UpdateTypesFile(project, requestTypes); err != nil {
		return fmt.Errorf("更新 types.go 失败: %w", err)
	}

	files, err := GenerateFiles(project, opts)
	if err != nil {
		return fmt.Errorf("生成模板文件失败: %w", err)
	}

	if err := WriteGeneratedFiles(files, opts); err != nil {
		return fmt.Errorf("写入生成文件失败: %w", err)
	}

	fmt.Printf("goctl-i18n 已完成，处理了 %d 个请求结构体\n", len(requestTypes))
	return nil
}

func debugf(opts CommandOptions, format string, args ...any) {
	if opts.Debug {
		fmt.Printf("[debug] "+format+"\n", args...)
	}
}
