package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rclsilver-org/usg-dns-api/db"
	"github.com/rclsilver-org/usg-dns-api/pkg/pid"
	"github.com/rclsilver-org/usg-dns-api/server"
	"github.com/rclsilver-org/usg-dns-api/unifi"
	"github.com/rclsilver-org/usg-dns-api/version"
)

var (
	pidFile string
	pidLock pid.ProcessLockFile
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the API server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(cmd.Context())

		if pidFile != "" {
			lock, err := pid.AcquireProcessIDLock(pidFile)
			if err != nil {
				logrus.WithContext(ctx).WithError(err).Fatal("unable to write the pid file")
			}
			pidLock = lock
		}

		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

		go func() {
			signal := <-signalCh
			logrus.WithContext(ctx).Warnf("received %v signal", signal)

			cancel()
		}()

		db, err := db.NewDatabase(ctx)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to initialize the database")
		}

		if db.GetMasterToken() == "" {
			logrus.WithContext(ctx).Fatal("no master token generated. please use the 'generate-token' command to generate a new one")
		}

		unifi, err := unifi.NewClient(ctx)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to initialize the unifi client")
		}

		if err := unifi.Login(ctx); err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to login to the unifi-controller")
		}
		logrus.WithContext(ctx).Debug("successfully connected to unifi-controller")

		s, err := server.NewServer(ctx, db, unifi, server.WithVerbose(verbose), server.WithTitle("usg-dns-api"), server.WithVersion(version.VersionFull()))
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to initialize the server")
		}

		s.StartTaskScheduler(ctx)

		if err := s.Serve(ctx); err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to start the HTTP server")
		}
	},
}

func init() {
	serverCmd.PersistentFlags().StringVarP(&pidFile, "pid-file", "p", "", "Write a pid file")
	rootCmd.AddCommand(serverCmd)
}
