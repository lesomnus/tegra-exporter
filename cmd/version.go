package cmd

import (
	"context"

	"github.com/lesomnus/tegra-exporter/cmd/version"
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
			v := version.Get()
			cmd.Printf(Template,
				v.Version,
				v.GitRev,
				v.GitDirty,
			)
			return nil
		}),
	}
}
