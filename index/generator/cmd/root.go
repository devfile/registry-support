/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/devfile/registry-support/index/generator/schema"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	shortDesc = "Generate index file"
	longDesc  = "Generate index file based on the index schema and registry devfiles"
	meta      = "meta.yaml"
	devfile   = "devfile.yaml"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "generator <registry directory path> <index file path>",
	Short: shortDesc,
	Long:  longDesc,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		registryDirPath := args[0]
		indexFilePath := args[1]
		err := generateIndex(registryDirPath, indexFilePath)
		if err != nil {
			fmt.Errorf("failed to generate index file: %v", err)
		}
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.generator.yaml)")

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

func generateIndex(registryDirPath string, indexFilePath string) error {
	registryDir, err := ioutil.ReadDir(registryDirPath)
	if err != nil {
		fmt.Errorf("failed to read registry directory %s: %v", registryDirPath, err)
	}

	var index []schema.Schema
	for _, devfileDir := range registryDir {
		if !devfileDir.IsDir() {
			fmt.Errorf("%s is not a directory: %v", filepath.Join(registryDirPath, devfileDir.Name()), err)
		}

		metaFilePath := filepath.Join(registryDirPath, devfileDir.Name(), meta)
		bytes, err := ioutil.ReadFile(metaFilePath)
		if err != nil {
			fmt.Errorf("failed to read %s: %v", metaFilePath, err)
		}
		var indexComponent schema.Schema
		err = yaml.Unmarshal(bytes, &indexComponent)
		if err != nil {
			fmt.Errorf("failed to unmarshal %s data: %v", metaFilePath, err)
		}
		indexComponent.Links = schema.Links{
			Self: fmt.Sprintf("%s/%s:%s", "devfile-catalog", indexComponent.Name, "latest"),
		}
		index = append(index, indexComponent)
	}

	bytes, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		fmt.Errorf("failed to marshal %s data: %v", indexFilePath, err)
	}
	err = ioutil.WriteFile(indexFilePath, bytes, 0644)
	if err != nil {
		fmt.Errorf("failed to write %s: %v", indexFilePath, err)
	}
	return nil
}
