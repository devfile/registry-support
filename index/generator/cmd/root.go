//
// Copyright Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/devfile/registry-support/index/generator/library"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	shortDesc = "Generate index file"
	longDesc  = "Generate index file based on the index schema and registry devfiles"
)

var cfgFile string
var force bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "generator <registry directory path> <index file path>",
	Short:        shortDesc,
	Long:         longDesc,
	Args:         cobra.ExactArgs(2),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		registryDirPath := args[0]
		indexFilePath := args[1]

		index, err := library.GenerateIndexStruct(registryDirPath, force)
		if err != nil {
			return fmt.Errorf("failed to generate index struct: %v", err)
		}

		err = library.CreateIndexFile(index, indexFilePath)
		if err != nil {
			return fmt.Errorf("failed to create index file: %v", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.generator.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "force to generate index file, ignore validation errors")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".generator" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".generator")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
