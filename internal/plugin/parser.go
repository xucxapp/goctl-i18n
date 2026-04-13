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
	ServiceDir       string
	ModuleRootDir    string
	ModulePath       string
	TypesFilePath    string
	SharedDir        string
	LocaleDir        string
	SharedImportPath string
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
	moduleRootDir, modulePath, err := findModuleRoot(ctx.Dir)
	if err != nil {
		return nil, err
	}

	typesFilePath, err := findTypesFile(ctx.Dir)
	if err != nil {
		return nil, err
	}

	normalizedLocaleDir := strings.TrimRight(opts.LocaleDir, "/")
	sharedRelativeDir := strings.TrimSuffix(normalizedLocaleDir, "/locales")
	if sharedRelativeDir == normalizedLocaleDir {
		sharedRelativeDir = normalizedLocaleDir
	}

	sharedDir := filepath.Join(moduleRootDir, filepath.FromSlash(sharedRelativeDir))
	localeDir := filepath.Join(moduleRootDir, filepath.FromSlash(normalizedLocaleDir))
	sharedRelative, err := filepath.Rel(moduleRootDir, sharedDir)
	if err != nil {
		return nil, fmt.Errorf("计算共享目录相对路径失败: %w", err)
	}

	project := &TargetProject{
		ServiceDir:       ctx.Dir,
		ModuleRootDir:    moduleRootDir,
		ModulePath:       modulePath,
		TypesFilePath:    typesFilePath,
		SharedDir:        sharedDir,
		LocaleDir:        localeDir,
		SharedImportPath: modulePath + "/" + filepath.ToSlash(sharedRelative),
	}

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

func findModuleRoot(startDir string) (string, string, error) {
	current := startDir
	for {
		goModPath := filepath.Join(current, "go.mod")
		if fileExists(goModPath) {
			modulePath, err := readModulePath(goModPath)
			if err != nil {
				return "", "", err
			}
			return current, modulePath, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", "", fmt.Errorf("从 %s 向上查找 go.mod 失败", startDir)
		}
		current = parent
	}
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
