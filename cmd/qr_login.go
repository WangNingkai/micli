package cmd

import (
	"micli/internal/conf"
	"micli/pkg/miservice"

	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var qrLoginCmd = &cobra.Command{
	Use:   "qr-login",
	Short: "Login using QR code scan with Mi Home app",
	Long: `Login using QR code scan with Mi Home app.
This avoids captcha issues when username/password login requires verification.
The authentication token will be saved for future use.`,
	Run: func(cmd *cobra.Command, args []string) {
		tokenPath := fmt.Sprintf("%s/.mi.token", os.Getenv("HOME"))
		tokenStore := ms.GetTokenStore()
		if tokenStore == nil {
			tokenStore = miservice.NewTokenStore(tokenPath)
		}

		// Clear existing token to force fresh QR login
		_ = tokenStore.SaveToken(nil)

		// Create a fresh service for QR login
		qrService := miservice.New(
			conf.Cfg.Section("account").Key("MI_USER").MustString(""),
			conf.Cfg.Section("account").Key("MI_PASS").MustString(""),
			conf.Cfg.Section("account").Key("REGION").MustString("cn"),
			tokenStore,
		)

		pterm.Info.Println("Starting QR code login...")
		token, err := qrService.QRLogin()
		if err != nil {
			pterm.Error.Printf("QR login failed: %v\n", err)
			os.Exit(1)
		}

		pterm.Success.Println("Login successful!")
		pterm.Info.Printf("User ID: %s\n", token.UserId)
		pterm.Info.Printf("Login mode: %s\n", token.LoginMode)
	},
}

func init() {
	rootCmd.AddCommand(qrLoginCmd)
}
