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

	project := &TargetProject{TypesFilePath: typesFile}
	requestTypes := []string{"CreateUserReq", "LoginReq"}

	require.NoError(t, UpdateTypesFile(project, requestTypes))

	firstRun, err := os.ReadFile(typesFile)
	require.NoError(t, err)
	require.Contains(t, string(firstRun), `"github.com/go-playground/validator/v10"`)
	require.Contains(t, string(firstRun), "var validate = validator.New()")
	require.Contains(t, string(firstRun), "func (r *CreateUserReq) Validate() error")
	require.Contains(t, string(firstRun), "func (r *LoginReq) Validate() error")

	require.NoError(t, UpdateTypesFile(project, requestTypes))

	secondRun, err := os.ReadFile(typesFile)
	require.NoError(t, err)
	require.Equal(t, string(firstRun), string(secondRun))
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
