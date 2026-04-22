package plugin

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"
)

const defaultLocaleDir = "tools/i18n/locales"

// CommandOptions 描述 validate 子命令的参数。
type CommandOptions struct {
	Command    string
	Custom     bool
	Translator bool
	LocaleDir  string
	Debug      bool
}

// ParseOptions 解析插件命令行参数。
func ParseOptions(args []string) (CommandOptions, error) {
	if len(args) == 0 {
		return CommandOptions{}, fmt.Errorf("缺少子命令，示例: goctl-i18n validate --translator")
	}

	command := strings.TrimSpace(args[0])
	switch command {
	case "validate":
		fs := flag.NewFlagSet(command, flag.ContinueOnError)
		opts := CommandOptions{
			Command:   command,
			LocaleDir: defaultLocaleDir,
		}
		fs.BoolVar(&opts.Custom, "custom", false, "是否生成自定义校验注册文件")
		fs.BoolVar(&opts.Translator, "translator", false, "是否生成多语言翻译辅助文件")
		fs.StringVar(&opts.LocaleDir, "locale-dir", defaultLocaleDir, "语言资源目录")
		fs.BoolVar(&opts.Debug, "debug", false, "是否打印调试日志")
		if err := fs.Parse(args[1:]); err != nil {
			return CommandOptions{}, err
		}

		opts.LocaleDir = filepath.ToSlash(strings.TrimSpace(opts.LocaleDir))
		if opts.LocaleDir == "" {
			opts.LocaleDir = defaultLocaleDir
		}

		return opts, nil
	default:
		return CommandOptions{}, fmt.Errorf("不支持的子命令: %s", command)
	}
}
