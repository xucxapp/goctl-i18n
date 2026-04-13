package plugin

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	goctlplugin "github.com/zeromicro/go-zero/tools/goctl/plugin"
)

// PluginContext 是 goctl 通过标准输入传入的上下文。
type PluginContext struct {
	Api         *spec.ApiSpec
	ApiFilePath string
	Style       string
	Dir         string
}

// TargetProject 描述要被写入的目标项目。
type TargetProject struct {
	RootDir        string
	ModulePath     string
	TypesFilePath  string
	TypesPackage   string
	I18nDir        string
	LocaleDir      string
	I18nImportPath string
}

// LoadPluginContext 从 goctl 标准输入中读取插件上下文。
func LoadPluginContext() (*PluginContext, error) {
	info, err := goctlplugin.NewPlugin()
	if err != nil {
		return nil, err
	}

	return &PluginContext{
		Api:         info.Api,
		ApiFilePath: info.ApiFilePath,
		Style:       info.Style,
		Dir:         info.Dir,
	}, nil
}

// DiscoverTargetProject 解析目标项目的重要路径。
func DiscoverTargetProject(ctx *PluginContext, opts CommandOptions) (*TargetProject, error) {
	modulePath, err := readModulePath(filepath.Join(ctx.Dir, "go.mod"))
	if err != nil {
		return nil, err
	}

	typesFilePath, err := findTypesFile(ctx.Dir)
	if err != nil {
		return nil, err
	}

	project := &TargetProject{
		RootDir:       ctx.Dir,
		ModulePath:    modulePath,
		TypesFilePath: typesFilePath,
		TypesPackage:  "types",
		I18nDir:       filepath.Join(ctx.Dir, "internal", "i18n"),
		LocaleDir:     filepath.Join(ctx.Dir, filepath.FromSlash(opts.LocaleDir)),
	}

	i18nRelative, err := filepath.Rel(ctx.Dir, project.I18nDir)
	if err != nil {
		return nil, fmt.Errorf("计算 i18n 相对路径失败: %w", err)
	}
	project.I18nImportPath = modulePath + "/" + filepath.ToSlash(i18nRelative)

	return project, nil
}

// CollectRequestTypes 收集需要生成 Validate 方法的请求结构体名称。
func CollectRequestTypes(ctx *PluginContext) []string {
	if ctx == nil || ctx.Api == nil {
		return nil
	}

	typeSet := make(map[string]struct{})
	for _, group := range ctx.Api.Service.Groups {
		for _, route := range group.Routes {
			name := strings.TrimSpace(route.RequestTypeName())
			if name == "" {
				continue
			}
			if strings.HasSuffix(name, "Req") {
				typeSet[name] = struct{}{}
			}
		}
	}

	names := make([]string, 0, len(typeSet))
	for _, item := range ctx.Api.Types {
		name := strings.TrimSpace(item.Name())
		if name == "" {
			continue
		}
		if _, ok := typeSet[name]; ok {
			names = append(names, name)
		}
	}

	slices.Sort(names)
	return names
}

func readModulePath(goModPath string) (string, error) {
	file, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("读取 go.mod 失败: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("扫描 go.mod 失败: %w", err)
	}

	return "", fmt.Errorf("go.mod 中未找到 module 声明")
}

func findTypesFile(root string) (string, error) {
	defaultPath := filepath.Join(root, "internal", "types", "types.go")
	if fileExists(defaultPath) {
		return defaultPath, nil
	}

	var matches []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) != "types.go" {
			return nil
		}
		slashPath := filepath.ToSlash(path)
		if strings.Contains(slashPath, "/internal/types/") {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("扫描 types.go 失败: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("在 %s 下未找到 internal/types/types.go", root)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("找到多个 types.go，请显式指定唯一生成目录: %v", matches)
	}

	return matches[0], nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
