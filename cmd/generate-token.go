package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rclsilver-org/usg-dns-api/db"
)

var generateTokenCmd = &cobra.Command{
	Use:   "generate-token",
	Short: "Generate the master token",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		db, err := db.NewDatabase(ctx)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to initialize the database")
		}

		token := db.GenerateMasterToken()

		if err := db.Save(); err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to write the database")
		}

		logrus.WithContext(ctx).Infof("a new master token has been generated: %s", token)
	},
}

func init() {
	rootCmd.AddCommand(generateTokenCmd)
}
