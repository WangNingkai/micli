package cmd

import (
	"github.com/spf13/cobra"
)

var sceneCmd = &cobra.Command{
	Use:   "scene",
	Short: "Manage and execute smart scenes",
}

var sceneListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all scenes",
	Run: func(cmd *cobra.Command, args []string) {
		scenes, err := ioSrv.SceneList()
		handleResult(scenes, err)
	},
}

var sceneRunCmd = &cobra.Command{
	Use:   "run <scene_id_or_name>",
	Short: "Execute a scene by ID or name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sceneID, homeID, err := ioSrv.ResolveScene(args[0])
		if err != nil {
			handleResult(nil, err)
			return
		}
		res, err := ioSrv.RunScene(sceneID, homeID)
		handleResult(res, err)
	},
}

func init() {
	sceneCmd.AddCommand(sceneListCmd)
	sceneCmd.AddCommand(sceneRunCmd)
	sceneRunCmd.Example = "  scene run \"Good Morning\"\n  scene run abc123"
}
