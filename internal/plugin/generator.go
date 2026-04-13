package plugin

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

// GeneratedFile 描述一个待写入目标项目的文件。
type GeneratedFile struct {
	Path      string
	Content   []byte
	Overwrite bool
}

// TypesAppendData 是 types.go 追加代码模板的数据。
type TypesAppendData struct {
	RequestTypes []string
}

// GenerateFiles 生成除 types.go 之外的所有模板文件内容。
func GenerateFiles(project *TargetProject, opts CommandOptions) ([]GeneratedFile, error) {
	files := make([]GeneratedFile, 0, 6)

	validatorContent, err := renderTemplate("templates/validator.tmpl", nil)
	if err != nil {
		return nil, err
	}
	files = append(files, GeneratedFile{
		Path:      filepath.Join(project.SharedDir, "validator.go"),
		Content:   validatorContent,
		Overwrite: false,
	})

	if opts.Custom {
		content, err := renderTemplate("templates/validation.tmpl", nil)
		if err != nil {
			return nil, err
		}
		files = append(files, GeneratedFile{
			Path:      filepath.Join(project.SharedDir, "validation.go"),
			Content:   content,
			Overwrite: false,
		})
	}

	if opts.Translator {
		translatorContent, err := renderTemplate("templates/translator.tmpl", nil)
		if err != nil {
			return nil, err
		}
		files = append(files, GeneratedFile{
			Path:      filepath.Join(project.SharedDir, "translator.go"),
			Content:   translatorContent,
			Overwrite: false,
		})

		i18nContent, err := renderTemplate("templates/i18n.tmpl", nil)
		if err != nil {
			return nil, err
		}
		files = append(files, GeneratedFile{
			Path:      filepath.Join(project.SharedDir, "i18n.go"),
			Content:   i18nContent,
			Overwrite: false,
		})

		zhContent, err := renderTemplate("templates/locale_zh.tmpl", nil)
		if err != nil {
			return nil, err
		}
		files = append(files, GeneratedFile{
			Path:      filepath.Join(project.LocaleDir, "active.zh.yaml"),
			Content:   zhContent,
			Overwrite: false,
		})

		enContent, err := renderTemplate("templates/locale_en.tmpl", nil)
		if err != nil {
			return nil, err
		}
		files = append(files, GeneratedFile{
			Path:      filepath.Join(project.LocaleDir, "active.en.yaml"),
			Content:   enContent,
			Overwrite: false,
		})
	}

	return files, nil
}

func renderTemplate(name string, data any) ([]byte, error) {
	raw, err := templateFS.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("读取模板 %s 失败: %w", name, err)
	}

	tpl, err := template.New(filepath.Base(name)).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("解析模板 %s 失败: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("执行模板 %s 失败: %w", name, err)
	}

	return buf.Bytes(), nil
}
