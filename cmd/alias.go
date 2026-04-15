package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage device name aliases",
	Long:  `Manage custom short names for your devices. Use subcommands: list, add, rm.`,
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all custom aliases",
	Run: func(cmd *cobra.Command, args []string) {
		if aliasStore == nil || len(aliasStore.Aliases) == 0 {
			fmt.Println("No aliases configured. Use 'micli alias add <name> <device>' to create one.")
			return
		}
		devices, _ := getDeviceListFromLocal()
		deviceMap := make(map[string]string)
		for _, d := range devices {
			deviceMap[d.Did] = d.Name
		}
		fmt.Println("Aliases:")
		for alias, did := range aliasStore.Aliases {
			realName := deviceMap[did]
			if realName != "" {
				fmt.Printf("  %-15s -> %s (%s)\n", alias, did, realName)
			} else {
				fmt.Printf("  %-15s -> %s\n", alias, did)
			}
		}
	},
}

var aliasAddCmd = &cobra.Command{
	Use:   "add <alias> <device_name_or_did>",
	Short: "Add a device alias",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		input := args[1]

		devices, err := getDeviceListFromLocal()
		if err != nil {
			handleResult(nil, fmt.Errorf("failed to load device list: %w", err))
			return
		}

		// Try exact DID match first
		var did string
		for _, d := range devices {
			if d.Did == input {
				did = d.Did
				break
			}
		}

		// If not exact DID match, use resolveDevice for alias/fuzzy matching
		if did == "" {
			did, _, err = resolveDevice(input)
			if err != nil {
				handleResult(nil, fmt.Errorf("failed to resolve device '%s': %w", input, err))
				return
			}
		}

		aliasStore.Add(alias, did)
		if err := aliasStore.Save(); err != nil {
			handleResult(nil, fmt.Errorf("failed to save alias: %w", err))
			return
		}
		fmt.Printf("Alias '%s' -> %s saved successfully\n", alias, did)
	},
}

var aliasRmCmd = &cobra.Command{
	Use:   "rm <alias>",
	Short: "Remove a device alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		if _, ok := aliasStore.Aliases[alias]; !ok {
			fmt.Printf("Alias '%s' not found\n", alias)
			return
		}
		aliasStore.Remove(alias)
		if err := aliasStore.Save(); err != nil {
			handleResult(nil, fmt.Errorf("failed to save: %w", err))
			return
		}
		fmt.Printf("Alias '%s' removed\n", alias)
	},
}

func init() {
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasAddCmd)
	aliasCmd.AddCommand(aliasRmCmd)
	aliasAddCmd.Example = "  alias add lamp \"Bedroom Lamp\"\n  alias add ac 123456789"
	aliasRmCmd.Example = "  alias rm lamp"
}
