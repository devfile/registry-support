//
// Copyright (c) 2020 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

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
	registryList = os.Getenv("REGISTRY_LIST")
	cfgFile      string
	allResources bool
	destDir      string
	devfileType  string
	insecure     bool
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

			if allResources {
				err = library.PullStackFromRegistry(registry, stack, destDir, insecure)
			} else {
				err = library.PullStackByMediaTypesFromRegistry(registry, stack, library.DevfileMediaTypeList, destDir, insecure)
			}
			if err != nil {
				fmt.Printf("Failed to pull %s from registry %s: %v\n", stack, registry, err)
			}
		},
	}
	pullCmd.Flags().BoolVarP(&allResources, "all", "a", false, "pull all resources of the given stack")
	pullCmd.Flags().StringVar(&destDir, "context", ".", "destination directory that stores stack resources")
	pullCmd.Flags().BoolVar(&insecure, "insecure", false, "skip verification for security")

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List stacks of the registry",
		Run: func(cmd *cobra.Command, args []string) {
			if devfileType == "" {
				fmt.Printf("Please specify the devfile type by using flag --type\n")
				return
			}
			err := library.PrintRegistry(registryList, devfileType, insecure)
			if err != nil {
				fmt.Printf("Failed to list stacks of registry %s: %v\n", registryList, err)
			}
		},
	}
	listCmd.Flags().StringVar(&devfileType, "type", "", "specify devfile type")
	listCmd.Flags().BoolVar(&insecure, "insecure", false, "skip verification for security")

	rootCmd.AddCommand(pullCmd, listCmd)
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
