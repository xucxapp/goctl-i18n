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

	if !astutil.UsesImport(file, project.SharedImportPath) {
		astutil.AddNamedImport(fset, file, "goctli18n", project.SharedImportPath)
	}
	removeImport(file, "github.com/go-playground/validator/v10")
	removeValidateVar(file)
	rewriteValidateMethods(file, requestTypes)

	formatted, err := formatNode(fset, file)
	if err != nil {
		return fmt.Errorf("格式化 types.go 失败: %w", err)
	}
	formatted = stripLegacyValidateBlock(formatted)

	existingMethods := collectValidateMethods(file)
	missingMethods := make([]string, 0, len(requestTypes))
	for _, requestType := range requestTypes {
		if _, ok := existingMethods[requestType]; !ok {
			missingMethods = append(missingMethods, requestType)
		}
	}
	slices.Sort(missingMethods)

	appendData := TypesAppendData{
		RequestTypes: missingMethods,
	}
	if len(appendData.RequestTypes) == 0 {
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

func removeValidateVar(file *ast.File) {
	decls := make([]ast.Decl, 0, len(file.Decls))
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			decls = append(decls, decl)
			continue
		}

		keepSpecs := make([]ast.Spec, 0, len(genDecl.Specs))
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				keepSpecs = append(keepSpecs, spec)
				continue
			}

			shouldRemove := false
			for _, name := range valueSpec.Names {
				if name.Name == "validate" {
					shouldRemove = true
					break
				}
			}
			if !shouldRemove {
				keepSpecs = append(keepSpecs, spec)
			}
		}

		if len(keepSpecs) == 0 {
			continue
		}
		genDecl.Specs = keepSpecs
		decls = append(decls, genDecl)
	}
	file.Decls = decls
}

func removeImport(file *ast.File, importPath string) {
	decls := make([]ast.Decl, 0, len(file.Decls))
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			decls = append(decls, decl)
			continue
		}

		keepSpecs := make([]ast.Spec, 0, len(genDecl.Specs))
		for _, spec := range genDecl.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if !ok {
				keepSpecs = append(keepSpecs, spec)
				continue
			}
			if strings.Trim(importSpec.Path.Value, `"`) == importPath {
				continue
			}
			keepSpecs = append(keepSpecs, spec)
		}

		if len(keepSpecs) == 0 {
			continue
		}
		genDecl.Specs = keepSpecs
		decls = append(decls, genDecl)
	}
	file.Decls = decls
}

func rewriteValidateMethods(file *ast.File, requestTypes []string) {
	requestTypeSet := make(map[string]struct{}, len(requestTypes))
	for _, requestType := range requestTypes {
		requestTypeSet[requestType] = struct{}{}
	}

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name == nil || funcDecl.Name.Name != "Validate" || funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}

		receiverName, receiverType := getReceiverInfo(funcDecl)
		if receiverType == "" {
			continue
		}
		if _, ok := requestTypeSet[receiverType]; !ok {
			continue
		}
		if receiverName == "" {
			receiverName = "r"
		}

		funcDecl.Body = &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("goctli18n"),
								Sel: ast.NewIdent("ValidateStruct"),
							},
							Args: []ast.Expr{ast.NewIdent(receiverName)},
						},
					},
				},
			},
		}
	}
}

func getReceiverInfo(funcDecl *ast.FuncDecl) (string, string) {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return "", ""
	}

	var receiverName string
	if len(funcDecl.Recv.List[0].Names) > 0 {
		receiverName = funcDecl.Recv.List[0].Names[0].Name
	}

	receiver := funcDecl.Recv.List[0].Type
	starExpr, ok := receiver.(*ast.StarExpr)
	if !ok {
		return receiverName, ""
	}
	ident, ok := starExpr.X.(*ast.Ident)
	if !ok {
		return receiverName, ""
	}

	return receiverName, ident.Name
}

func stripLegacyValidateBlock(content []byte) []byte {
	text := string(content)
	legacyBlocks := []string{
		"\n// validate 是请求结构体验证器的共享实例。\nvar validate = validator.New()\n",
		"\n// validate 是请求结构体验证器的共享实例。\n",
		"\nvar validate = validator.New()\n",
	}
	for _, block := range legacyBlocks {
		text = strings.ReplaceAll(text, block, "\n")
	}
	return []byte(text)
}
