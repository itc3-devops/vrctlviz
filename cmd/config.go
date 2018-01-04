// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate a config file template at ~/.vrctlvizcfg.yaml",
	Long: `Command use:
	To generate a config file template located at ~/.vrctlvizcfg.yaml
	vrctlviz config template`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(os.Args) > 2 && os.Args[2] == "template" {
			fmt.Println(configFile)
			createConfigTemplate()
		} else {
			fmt.Println(`Command use:
				To generate a config file template located at ~/.vrctlvizcfg.yaml
				vrctlviz config template`)
		}

	},
}

func init() {
	RootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
