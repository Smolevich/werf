package get_release

import (
	"fmt"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/werf"
	"github.com/spf13/cobra"
)

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-release",
		DisableFlagsInUseLine: true,
		Short: "Prints Helm Release name that will be used in current configuration with specified params",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runGetRelease(); err != nil {
				return fmt.Errorf("get-release failed: %s", err)
			}
			return nil
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupEnvironment(&CommonCmdData, cmd)

	return cmd
}

func runGetRelease() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("cannot parse werf config: %s", err)
	}

	release, err := common.GetHelmRelease("", *CommonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	fmt.Println(release)

	return nil
}
