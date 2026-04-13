package plugin

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// UpdateTypesFile 更新目标项目中的 types.go。
func UpdateTypesFile(project *TargetProject, requestTypes []string) error {
	original, err := os.ReadFile(project.TypesFilePath)
	if err != nil {
		return fmt.Errorf("读取 types.go 失败: %w", err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, project.TypesFilePath, original, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("解析 types.go 失败: %w", err)
	}

	if !astutil.UsesImport(file, "github.com/go-playground/validator/v10") {
		astutil.AddImport(fset, file, "github.com/go-playground/validator/v10")
	}

	formatted, err := formatNode(fset, file)
	if err != nil {
		return fmt.Errorf("格式化 types.go 失败: %w", err)
	}

	existingMethods := collectValidateMethods(file)
	missingMethods := make([]string, 0, len(requestTypes))
	for _, requestType := range requestTypes {
		if _, ok := existingMethods[requestType]; !ok {
			missingMethods = append(missingMethods, requestType)
		}
	}
	slices.Sort(missingMethods)

	appendData := TypesAppendData{
		NeedsValidateVar: !hasValidateVar(file),
		RequestTypes:     missingMethods,
	}
	if !appendData.NeedsValidateVar && len(appendData.RequestTypes) == 0 {
		if bytes.Equal(original, formatted) {
			return nil
		}
		return os.WriteFile(project.TypesFilePath, formatted, 0o644)
	}

	appendBlock, err := renderTemplate("templates/types_append.tmpl", appendData)
	if err != nil {
		return err
	}

	finalContent := strings.TrimRight(string(formatted), "\r\n") + "\n\n" + strings.TrimSpace(string(appendBlock)) + "\n"
	if bytes.Equal(original, []byte(finalContent)) {
		return nil
	}

	return os.WriteFile(project.TypesFilePath, []byte(finalContent), 0o644)
}

// WriteGeneratedFiles 写入模板文件。
func WriteGeneratedFiles(files []GeneratedFile, opts CommandOptions) error {
	for _, file := range files {
		if err := os.MkdirAll(filepath.Dir(file.Path), 0o755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %w", filepath.Dir(file.Path), err)
		}

		if !file.Overwrite && fileExists(file.Path) {
			debugf(opts, "跳过已存在文件: %s", file.Path)
			continue
		}

		if err := os.WriteFile(file.Path, file.Content, 0o644); err != nil {
			return fmt.Errorf("写入文件失败 %s: %w", file.Path, err)
		}

		debugf(opts, "已生成文件: %s", file.Path)
	}

	return nil
}

func formatNode(fset *token.FileSet, node *ast.File) ([]byte, error) {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func hasValidateVar(file *ast.File) bool {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range valueSpec.Names {
				if name.Name == "validate" {
					return true
				}
			}
		}
	}

	return false
}

func collectValidateMethods(file *ast.File) map[string]struct{} {
	methods := make(map[string]struct{})
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name == nil || funcDecl.Name.Name != "Validate" || funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}

		receiver := funcDecl.Recv.List[0].Type
		starExpr, ok := receiver.(*ast.StarExpr)
		if !ok {
			continue
		}
		ident, ok := starExpr.X.(*ast.Ident)
		if !ok {
			continue
		}
		methods[ident.Name] = struct{}{}
	}

	return methods
}
