/*   Copyright 2020-2022 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/devfile/registry-support/registry-library/library"
)

const (
	usage     = "registry"
	shortDesc = "Commands to interact with devfile registry"
	longDesc  = "Commands to interact with devfile registry"
)

var (
	registryList   = os.Getenv("REGISTRY_LIST")
	cfgFile        string
	allResources   bool
	destDir        string
	devfileType    string
	skipTLSVerify  bool
	newIndexSchema bool
	user           string
	architectures  []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   usage,
	Short: shortDesc,
	Long:  longDesc,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		operation := args[0]
		fmt.Printf("%s is not a valid operation\n", operation)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.registry.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	var pullCmd = &cobra.Command{
		Use:   "pull <registry name> <stack name>",
		Short: "Pull stack resources from the registry, by default only pull devfile.yaml from the registry",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			registry := args[0]
			stack := args[1]
			var err error

			options := library.RegistryOptions{
				NewIndexSchema: newIndexSchema,
				Telemetry: library.TelemetryData{
					User: "user",
				},
				SkipTLSVerify: skipTLSVerify,
			}

			if len(architectures) > 0 {
				options.Filter.Architectures = architectures
			}

			if allResources {
				err = library.PullStackFromRegistry(registry, stack, destDir, options)
			} else {
				err = library.PullStackByMediaTypesFromRegistry(registry, stack, library.DevfileMediaTypeList, destDir, options)
			}
			if err != nil {
				fmt.Printf("Failed to pull %s from registry %s: %v\n", stack, registry, err)
			}
		},
	}
	pullCmd.Flags().BoolVarP(&allResources, "all", "a", false, "pull all resources of the given stack")
	pullCmd.Flags().StringArrayVar(&architectures, "arch", []string{}, "architecture filter; example: --arch amd64 --arch arm64")
	pullCmd.Flags().StringVar(&destDir, "context", ".", "destination directory that stores stack resources")
	pullCmd.Flags().BoolVar(&skipTLSVerify, "skip-tls-verify", false, "skip TLS verification")
	pullCmd.Flags().BoolVar(&newIndexSchema, "new-index-schema", false, "pull new index schema")
	pullCmd.Flags().StringVar(&user, "user", "", "consumer name")

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List stacks of the registry",
		Run: func(cmd *cobra.Command, args []string) {
			if devfileType == "" {
				fmt.Printf("Please specify the devfile type by using flag --type\n")
				return
			}

			options := library.RegistryOptions{
				Telemetry: library.TelemetryData{
					User: "user",
				},
				SkipTLSVerify: skipTLSVerify,
			}

			if len(architectures) > 0 {
				options.Filter.Architectures = architectures
			}

			err := library.PrintRegistry(registryList, devfileType, options)
			if err != nil {
				fmt.Printf("Failed to list stacks of registry %s: %v\n", registryList, err)
			}
		},
	}
	listCmd.Flags().StringVar(&devfileType, "type", "", "specify devfile type")
	listCmd.Flags().StringArrayVar(&architectures, "arch", []string{}, "architecture filter; example: --arch amd64 --arch arm64")
	listCmd.Flags().BoolVar(&skipTLSVerify, "skip-tls-verify", false, "skip TLS verification")
	listCmd.Flags().StringVar(&user, "user", "", "consumer name")

	var downloadCmd = &cobra.Command{
		Use:   "download <registry name> <stack name> <starter project name>",
		Short: "Downloads starter project",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			registry, stack, starterProject := args[0], args[1], args[2]
			var err error

			options := library.RegistryOptions{
				NewIndexSchema: newIndexSchema,
				Telemetry: library.TelemetryData{
					User: "user",
				},
				SkipTLSVerify: skipTLSVerify,
			}

			err = library.DownloadStarterProjectAsDir(destDir, registry, stack, starterProject, options)
			if err != nil {
				fmt.Printf("failed to download starter project %s: %v\n", starterProject, err)
			}
		},
	}
	downloadCmd.Flags().StringVar(&destDir, "context", ".", "destination directory that stores stack resources")
	downloadCmd.Flags().BoolVar(&skipTLSVerify, "skip-tls-verify", false, "skip TLS verification")
	downloadCmd.Flags().BoolVar(&newIndexSchema, "new-index-schema", false, "use new index schema")
	downloadCmd.Flags().StringVar(&user, "user", "", "consumer name")

	rootCmd.AddCommand(pullCmd, listCmd, downloadCmd)
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

		// Search config in home directory with name ".registry" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".registry")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
