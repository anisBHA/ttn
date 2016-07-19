package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// devicesInfoCmd represents the `device info` command
var devicesInfoCmd = &cobra.Command{
	Use:   "info [Device ID]",
	Short: "Get information about a device",
	Long:  `ttnctl devices info can be used to get information about a device.`,
	Run: func(cmd *cobra.Command, args []string) {

		auth, err := util.LoadAuth(viper.GetString("ttn-account-server"))
		if err != nil {
			ctx.WithError(err).Fatal("Failed to load authentication")
		}
		if auth == nil {
			ctx.Fatal("No authentication found, please login")
		}

		if len(args) == 0 {
			cmd.Usage()
			return
		}

		devID := args[0]
		if !api.ValidID(devID) {
			ctx.Fatalf("Invalid Device ID") // TODO: Add link to wiki explaining device IDs
		}

		appID := viper.GetString("app-id")
		if appID == "" {
			ctx.Fatal("Missing AppID. You should run ttnctl applications use [AppID] [AppEUI]")
		}

		manager, err := handler.NewManagerClient(viper.GetString("ttn-handler"), auth.AccessToken)
		if err != nil {
			ctx.WithError(err).Fatal("Could not create Handler client")
		}

		dev, err := manager.GetDevice(appID, devID)
		if err != nil {
			ctx.WithError(err).Fatal("Could not get existing device.")
		}

		byteFormat, _ := cmd.Flags().GetString("format")

		ctx.Info("Found device")

		fmt.Println()

		fmt.Printf("  Application ID: %s\n", dev.AppId)
		fmt.Printf("       Device ID: %s\n", dev.DevId)
		if lorawan := dev.GetLorawanDevice(); lorawan != nil {
			lastSeen := "never"
			if lorawan.LastSeen > 0 {
				lastSeen = fmt.Sprintf("%s", time.Unix(0, 0).Add(time.Duration(lorawan.LastSeen)))
			}

			fmt.Printf("       Last Seen: %s\n", lastSeen)
			fmt.Println()
			fmt.Println("    LoRaWAN Info:")
			fmt.Println()
			fmt.Printf("     AppEUI: %s\n", formatBytes(lorawan.AppEui, byteFormat))
			fmt.Printf("     DevEUI: %s\n", formatBytes(lorawan.DevEui, byteFormat))
			fmt.Printf("    DevAddr: %s\n", formatBytes(lorawan.DevAddr, byteFormat))
			fmt.Printf("     AppKey: %s\n", formatBytes(lorawan.AppKey, byteFormat))
			fmt.Printf("    AppSKey: %s\n", formatBytes(lorawan.AppSKey, byteFormat))
			fmt.Printf("    NwkSKey: %s\n", formatBytes(lorawan.NwkSKey, byteFormat))

			fmt.Printf("     FCntUp: %d\n", lorawan.FCntUp)
			fmt.Printf("   FCntDown: %d\n", lorawan.FCntDown)
			options := []string{}
			if lorawan.DisableFCntCheck {
				options = append(options, "DisableFCntCheck")
			}
			if lorawan.Uses32BitFCnt {
				options = append(options, "Uses32BitFCnt")
			}
			fmt.Printf("    Options: %s\n", strings.Join(options, ", "))
		}

	},
}

type formattableBytes interface {
	IsEmpty() bool
	Bytes() []byte
}

func formatBytes(toPrint interface{}, format string) string {
	if i, ok := toPrint.(formattableBytes); ok {
		if i.IsEmpty() {
			return "<nil>"
		}
		switch format {
		case "msb":
			return cStyle(i.Bytes(), true) + " (msb first)"
		case "lsb":
			return cStyle(i.Bytes(), false) + " (lsb first)"
		case "hex":
			return fmt.Sprintf("%X", i.Bytes())
		}
	}
	return fmt.Sprintf("%s", toPrint)
}

// cStyle prints the byte slice in C-Style
func cStyle(bytes []byte, msbf bool) string {
	output := "{"
	if !msbf {
		bytes = reverse(bytes)
	}
	for i, b := range bytes {
		if i != 0 {
			output += ", "
		}
		output += fmt.Sprintf("0x%02X", b)
	}
	output += "}"
	return output
}

// reverse is used to convert between MSB-first and LSB-first
func reverse(in []byte) (out []byte) {
	for i := len(in) - 1; i >= 0; i-- {
		out = append(out, in[i])
	}
	return
}

func init() {
	devicesCmd.AddCommand(devicesInfoCmd)
	devicesInfoCmd.Flags().String("format", "hex", "Formatting: hex/msb/lsb")
}