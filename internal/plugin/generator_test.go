package plugin

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateFiles(t *testing.T) {
	t.Parallel()

	project := &TargetProject{
		TypesFilePath:    filepath.Join("demo", "apps", "apiuser", "internal", "types", "types.go"),
		SharedDir:        filepath.Join("demo", "tools", "i18n"),
		LocaleDir:        filepath.Join("demo", "tools", "i18n", "locales"),
		SharedImportPath: "example.com/demo/tools/i18n",
	}

	files, err := GenerateFiles(project, CommandOptions{
		Command:    "validate",
		Custom:     true,
		Translator: true,
		LocaleDir:  defaultLocaleDir,
	})
	require.NoError(t, err)
	require.Len(t, files, 6)

	contentByBase := make(map[string]string, len(files))
	for _, file := range files {
		contentByBase[filepath.Base(file.Path)] = string(file.Content)
	}

	require.Contains(t, contentByBase["translator.go"], "package i18n")
	require.Contains(t, contentByBase["validator.go"], "var validate = validator.New()")
	require.Contains(t, contentByBase["validation.go"], "registerValidation")
	require.Contains(t, contentByBase["validation.go"], "package i18n")
	require.Contains(t, contentByBase["i18n.go"], "LoadMessageFileFS")
	require.Contains(t, contentByBase["i18n.go"], "package i18n")
	require.Contains(t, contentByBase["active.zh.yaml"], "validate.required")
	require.Contains(t, contentByBase["active.en.yaml"], "validate.required")
}
