package cmd

import (
	"micli/internal/conf"
	"micli/pkg/util"
	"time"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats <did_or_name>",
	Short: "View device statistics (power, usage, etc.)",
	Long:  `View device statistics like power consumption. Requires --key flag in siid.piid format.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		// Resolve device: config DID -> alias/fuzzy -> interactive
		if did == "" {
			did = conf.Cfg.Section("account").Key("MI_DID").MustString("")
		}
		if did != "" && util.IsDigit(did) {
			// Config DID is a valid numeric DID, use it directly
		} else if did != "" {
			// Config DID might be an alias or name, resolve it
			var err error
			did, _, err = resolveDevice(did)
			if err != nil {
				did = ""
			}
		}

		if did == "" {
			var err error
			did, _, err = resolveDevice(input)
			if err != nil {
				// If resolution fails, fall back to interactive selection
				did, err = chooseDevice()
				if err != nil {
					handleResult(nil, err)
					return
				}
			}
		}

		now := time.Now()
		var timeStart int64
		switch statsType {
		case "hour":
			timeStart = now.Add(-24 * time.Hour).Unix()
		case "day":
			timeStart = now.Add(-7 * 24 * time.Hour).Unix()
		case "week":
			timeStart = now.Add(-4 * 7 * 24 * time.Hour).Unix()
		case "month":
			timeStart = now.Add(-6 * 30 * 24 * time.Hour).Unix()
		default:
			timeStart = now.Add(-7 * 24 * time.Hour).Unix()
		}

		dataType := "stat_" + statsType + "_v3"
		res, err := ioSrv.GetStatistics(did, statsKey, dataType, timeStart, now.Unix())
		handleResult(res, err)
	},
}

var statsKey string
var statsType string

func init() {
	statsCmd.Flags().StringVarP(&statsKey, "key", "k", "", "Property key in siid.piid format (e.g., 7.1)")
	statsCmd.Flags().StringVarP(&statsType, "type", "t", "day", "Statistics type: hour, day, week, month")
	statsCmd.Flags().StringVarP(&did, "did", "d", "", "Device ID")
	statsCmd.MarkFlagRequired("key")
	statsCmd.Example = "  stats \"Bedroom Lamp\" -k 7.1 -t day\n  stats 123456789 -k 2.1 -t week"
}
