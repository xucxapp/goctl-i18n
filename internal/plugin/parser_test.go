package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
)

func TestCollectRequestTypes(t *testing.T) {
	t.Parallel()

	ctx := &PluginContext{
		Api: &spec.ApiSpec{
			Types: []spec.Type{
				spec.DefineStruct{RawName: "CreateUserReq"},
				spec.DefineStruct{RawName: "LoginReq"},
				spec.DefineStruct{RawName: "ProfileResp"},
			},
			Service: spec.Service{
				Groups: []spec.Group{
					{
						Routes: []spec.Route{
							{RequestType: spec.DefineStruct{RawName: "LoginReq"}},
							{RequestType: spec.DefineStruct{RawName: "CreateUserReq"}},
							{RequestType: spec.DefineStruct{RawName: "ProfileResp"}},
						},
					},
				},
			},
		},
	}

	require.Equal(t, []string{"CreateUserReq", "LoginReq"}, CollectRequestTypes(ctx))
}

func TestDiscoverTargetProject(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	serviceDir := filepath.Join(root, "apps", "demo")
	require.NoError(t, os.MkdirAll(filepath.Join(serviceDir, "internal", "types"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/demo\n\ngo 1.26.2\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "internal", "types", "types.go"), []byte("package types\n"), 0o644))

	project, err := DiscoverTargetProject(&PluginContext{Dir: serviceDir}, CommandOptions{
		Command:   "validate",
		LocaleDir: defaultLocaleDir,
	})
	require.NoError(t, err)

	require.Equal(t, "example.com/demo", project.ModulePath)
	require.Equal(t, filepath.Join(serviceDir, "internal", "types", "types.go"), project.TypesFilePath)
	require.Equal(t, filepath.Join(root, "tools", "i18n"), project.SharedDir)
	require.Equal(t, filepath.Join(root, "tools", "i18n", "locales"), project.LocaleDir)
	require.Equal(t, "example.com/demo/tools/i18n", project.SharedImportPath)
}
