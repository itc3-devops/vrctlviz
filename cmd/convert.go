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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Reads data from ETCD and converts to vizceral graph format",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		go func() {
			log.Println(http.ListenAndServe("127.0.0.1:6060", nil))

		}()

		vizAutoRun()

	},
}

func init() {
	RootCmd.AddCommand(convertCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// convertCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// convertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func vizAutoRun() {
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// Do stuff
				// Define wait timer between task cycles
				time.Sleep(time.Second * 1)
				genGlobalLevelGraph()

			case <-quit:
				ticker.Stop()
				fmt.Println("Stopped the ticker!")
				return
			}
		}
	}()
	// Determine how long the scheduled task should run
	// 45000 hours is about 5 years
	time.Sleep(45000 * time.Hour)
	close(quit)
	// Determine delay between running tasks
	time.Sleep(10 * time.Nanosecond)
}

func genGlobalLevelGraph() {
	// Load environment variables
	loadHostEnvironmentVars()
	cli := requestEtcdDialer()

	// Set vars
	renderer := "global"
	name := "edge"
	maxvol := float64(50000.100)

	// Generate node level region/service hierarchy
	regionServiceNodes := regionServiceNodes(cli)
	regionServiceConnections := regionServiceConnections(cli)

	ns := VizceralGraph{
		Renderer:    renderer,
		Name:        name,
		MaxVolume:   maxvol,
		Nodes:       regionServiceNodes,
		Connections: regionServiceConnections,
	}
	n := fmt.Sprintf("%s", ns)
	// serialize and write data to file

	v := vizFileReadata(n)
	vizFileWrite(v)

}

// Creates connection information to be loaded into the top level global graph
func regionServiceConnections(cli *clientv3.Client) []VizceralConnection {
	// create vars
	vcg := []VizceralConnection{}
	vc := VizceralConnection{}

	// set vars
	keyPrefix := "viz"

	// get etcd keys based on connection prefix
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := cli.Get(ctx, keyPrefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	cancel()
	checkErr(err, "vizceral - genTopLevelView - get node keys")

	// iterate through each key for adding to the array
	for _, ev := range resp.Kvs {

		// convert etcd key/values to strings
		cKey := fmt.Sprintf("%s", ev.Key)
		cValue := fmt.Sprintf("%s", ev.Value)

		// filter out anything that is not a connection key
		if strings.Contains(cKey, "connection") {

			// unmarshall value into struct
			err := json.Unmarshal([]byte(cValue), &vc)
			if err != nil {
				log.Fatalf("failed to decode: %s", err)
			}
			// add connection to the interface
			vcg = append(vcg, vc)
		}
	}

	return vcg

}

func regionServiceNodes(cli *clientv3.Client) []VizceralNode {

	// create vars
	vng := []VizceralNode{}
	vn := VizceralNode{}

	keyPrefix := "viz"

	// pull nodes from etcd
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := cli.Get(ctx, keyPrefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	cancel()
	checkErr(err, "vizceral - genTopLevelView - get node keys")

	// iterate through each key for adding to the array
	for _, ev := range resp.Kvs {

		// convert etcd key/values to strings
		cKey := fmt.Sprintf("%s", ev.Key)
		cValue := fmt.Sprintf("%s", ev.Value)

		// filter out anything that is not a node key
		if strings.Contains(cKey, "node") {

			// unmarshall value into struct
			err := json.Unmarshal([]byte(cValue), &vn)
			if err != nil {
				log.Fatalf("failed to decode: %s", err)
			}
			// add node to the interface
			vng = append(vng, vn)
		}
	}

	return vng

}
