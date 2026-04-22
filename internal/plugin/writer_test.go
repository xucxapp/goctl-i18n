package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateTypesFileIdempotent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	typesDir := filepath.Join(root, "internal", "types")
	require.NoError(t, os.MkdirAll(typesDir, 0o755))

	source, err := os.ReadFile(filepath.Join("..", "..", "testdata", "types_input.go"))
	require.NoError(t, err)

	typesFile := filepath.Join(typesDir, "types.go")
	require.NoError(t, os.WriteFile(typesFile, source, 0o644))

	project := &TargetProject{
		TypesFilePath:    typesFile,
		SharedImportPath: "example.com/demo/tools/i18n",
	}
	requestTypes := []string{"CreateUserReq", "LoginReq"}

	require.NoError(t, UpdateTypesFile(project, requestTypes))

	firstRun, err := os.ReadFile(typesFile)
	require.NoError(t, err)
	require.Contains(t, string(firstRun), `"example.com/demo/tools/i18n"`)
	require.NotContains(t, string(firstRun), `"github.com/go-playground/validator/v10"`)
	require.NotContains(t, string(firstRun), "var validate = validator.New()")
	require.Contains(t, string(firstRun), "func (r *CreateUserReq) Validate() error")
	require.Contains(t, string(firstRun), "func (r *LoginReq) Validate() error")
	require.Contains(t, string(firstRun), "return i18n.ValidateStruct(r)")

	require.NoError(t, UpdateTypesFile(project, requestTypes))

	secondRun, err := os.ReadFile(typesFile)
	require.NoError(t, err)
	require.Equal(t, string(firstRun), string(secondRun))
}

func TestUpdateTypesFileRewriteLegacyValidate(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	typesDir := filepath.Join(root, "internal", "types")
	require.NoError(t, os.MkdirAll(typesDir, 0o755))

	legacy := `package types

import "github.com/go-playground/validator/v10"

type LoginReq struct {
	Email string
}

// validate 是请求结构体验证器的共享实例。
var validate = validator.New()

func (r *LoginReq) Validate() error {
	return validate.Struct(r)
}
`

	typesFile := filepath.Join(typesDir, "types.go")
	require.NoError(t, os.WriteFile(typesFile, []byte(legacy), 0o644))

	project := &TargetProject{
		TypesFilePath:    typesFile,
		SharedImportPath: "example.com/demo/tools/i18n",
	}

	require.NoError(t, UpdateTypesFile(project, []string{"LoginReq"}))

	content, err := os.ReadFile(typesFile)
	require.NoError(t, err)
	require.Contains(t, string(content), `"example.com/demo/tools/i18n"`)
	require.NotContains(t, string(content), `"github.com/go-playground/validator/v10"`)
	require.NotContains(t, string(content), "var validate = validator.New()")
	require.Contains(t, string(content), "return i18n.ValidateStruct(r)")
}

func TestWriteGeneratedFilesSkipExisting(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "translator.go")
	require.NoError(t, os.WriteFile(target, []byte("existing"), 0o644))

	err := WriteGeneratedFiles([]GeneratedFile{
		{
			Path:      target,
			Content:   []byte("new"),
			Overwrite: false,
		},
	}, CommandOptions{Debug: true})
	require.NoError(t, err)

	content, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, "existing", string(content))
}
