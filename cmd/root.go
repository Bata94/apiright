package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bata94/apiright/pkg/logger"
)

var (
	verbose bool
	debug   bool
	log     logger.Logger
)

var rootCmd = &cobra.Command{
	Use:   "apiright",
	Short: "Apiright CLI Tool",
	Long:  "Apiright CLI Tool",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func init() {
	var err error

	log = logger.NewDefaultLogger()
	if os.Getenv("ENV") == "DEV" {
		log.SetLevel(logger.DebugLevel)
		log.Debug("OS VAR ENV: ", os.Getenv("ENV"), " is set")
	}
	log.Debug("init")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Display more verbose output in console output. (default: false)")
	err = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	if err != nil {
		log.Fatal("Error binding verbose flag: ", err)
		return
	}

	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Display debugging output in the console. (default: false)")
	err = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	if err != nil {
		log.Fatal("Error binding debug flag: ", err)
		return
	}

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	log.Debug("initConfig")

	log.Debug("viper verbose: ", viper.GetBool("verbose"))
	log.Debug("viper debug: ", viper.GetBool("debug"))
	log.Debug("flag verbose: ", verbose)
	log.Debug("flag debug: ", debug)

	if verbose && debug {
		log.Fatal("verbose and debug cannot be set at the same time")
	}

	if verbose {
		log.SetLevel(logger.InfoLevel)
	} else if debug {
		log.SetLevel(logger.DebugLevel)
	}
}
