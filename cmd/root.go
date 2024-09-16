package cmd

import (
	"os"

	"github.com/ovh/configstore"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	verbose bool

	configFile        string
	defaultConfigFile string
)

var rootCmd = &cobra.Command{
	Use:   "usg-dns-api",
	Short: "usg-dns-api server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if configFile != "" {
			logrus.WithContext(cmd.Context()).Infof("loading the configuration from the file %q", configFile)
			configstore.File(configFile)
		} else {
			logrus.WithContext(cmd.Context()).Infof("loading the configuration according the %q environment variable", configstore.ConfigEnvVar)
			configstore.InitFromEnvironment()
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if pidLock != nil {
			if err := pidLock.Unlock(); err != nil {
				logrus.WithContext(cmd.Context()).WithError(err).Warning("unable to delete the pid file")
			}
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable to verbose mode")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config-file", "c", defaultConfigFile, "Configuration file")
}
