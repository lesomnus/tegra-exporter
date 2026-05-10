package cmd

import (
	"context"

	"github.com/lesomnus/xli"
)

func NewCmdVersion() *xli.Command {
	const Template = `TEGRA_EXPORTER_VERSION=%s
TEGRA_EXPORTER_GIT_REV=%s
TEGRA_EXPORTER_GIT_DIRTY=%v
`
	return &xli.Command{
		Name: "version",
		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			v := build_info
			cmd.Printf(Template,
				v.Version,
				v.GitRev,
				v.GitDirty,
			)
			return nil
		}),
	}
}

type buildInfo struct {
	Version  string
	GitRev   string
	GitDirty bool
}

//go:generate bash -c "../scripts/gen-version-file.sh > /dev/null"
var build_info = buildInfo{
	Version:  "v0.0.0-local",
	GitRev:   "0000000000000000000000000000000000000000",
	GitDirty: false,
}
