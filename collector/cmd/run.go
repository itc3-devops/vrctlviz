// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"encoding/json"
	"fmt"
	"time"

	"strconv"

	"github.com/itc3-devops/procspy"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the Network Operator vizceral collector app",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		resp := EtcdHealthMemberListCheck()
		if resp == true {
			fmt.Println("ETCD cluster is healthy")
			vizAutoRunCollector()
		} else {
			fmt.Println("ETCD cluster unreachable, or unhealthy")
		}
		// Get ETCD lease and start keepalive for auto ETCD cleanup

	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var ipD string
var source string
var errorRate string
var bandwidth string
var norm string
var warn float64
var normalTraffic string
var warningTraffic float64
var metrics VizceralLevels
var deviceName string

func vizAutoRunCollector() {

	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// Do stuff
				// Define wait timer between task cycles
				time.Sleep(time.Second * 5)
				genRegionalServiceLevelData()

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

func genRegionalServiceLevelData() {

	// Get timestamp and convert it to proper format
	ts := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	time := StrintToInt64(ts)

	// Check TCP connections for tincd sessions to determine healthy service
	lookupProcesses := true
	cs, err := procspy.Connections(lookupProcesses)
	if err != nil {
		panic(err)
	}

	for c := cs.Next(); c != nil; c = cs.Next() {
		fmt.Printf(" - %v\n", c)
		Transport := c.Transport
		fmt.Println("Print Transport: ", Transport)
		LocalAddress := c.LocalAddress
		fmt.Println("Print LocalAddress: ", LocalAddress)
		RemoteAddress := c.RemoteAddress
		fmt.Println("Print RemoteAddress: ", RemoteAddress)
		ProcName := c.Proc.Name
		fmt.Println("Print ProcName: ", ProcName)

		// Set vars
		renderer := "region"
		maxvol := float64(50000.100)
		class := "normal"

		// Generate data for structs
		// Generate global level connections between nodes
		genGlobalLevelConnections(cs)

		regionServiceConnections, notices, metadata := genRegionServiceConnections()

		ns := VizceralNode{
			Renderer:    renderer,
			Name:        c.LocalAddress.String(),
			MaxVolume:   maxvol,
			Updated:     time,
			Connections: regionServiceConnections,
			Notices:     notices,
			Class:       class,
			Metadata:    metadata,
		}

		// serialize data for host hierarchy
		j, jErr := json.MarshalIndent(ns, "", " ")
		CheckErr(jErr, "run - genGlobalLevelGraph")
		brjs := fmt.Sprintf("%s", j)
		key := string("viz/vrctlviz::" + "::node::" + c.LocalAddress.String() + "::vrf::global")
		fmt.Println("Print ETCD data: ", key, brjs)
		// upload to etcd and associate with our lease for automatic cleanup
		etcdPutLongLease(key, brjs)
	}
}

// Create service level connections
func genRegionServiceConnections() ([]VizceralConnection, []VizceralNotice, VizceralMetadata) {
	// Create interfaces
	vcg := []VizceralConnection{}
	vng := []VizceralNotice{}
	vm := VizceralMetadata{}

	return vcg, vng, vm
}

// Creates connection information to be loaded into the top level global graph
func genGlobalLevelConnections(cs procspy.ConnIter) {

	fmt.Printf("TCP Connections:\n")
	for c := cs.Next(); c != nil; c = cs.Next() {

		ProcName := c.Proc.Name
		fmt.Println("Print ProcName: ", ProcName)

		//var target string
		var class string

		// Set vars

		class = "normal"
		m := VizceralLevels{
			Normal:  10000,
			Warning: 1,
			Danger:  1,
		}

		cs := VizceralConnection{
			Source:  c.LocalAddress.String(),
			Target:  c.RemoteAddress.String(),
			Metrics: m,
			Class:   class,
		}

		css := fmt.Sprintln(cs)
		fmt.Println("Print json output: ", css)
		// Run some filters to make sure we don't have empty values

		// serialize the data and publish each connection in seperate ETCD key for assembly on a global
		// view by the aggriagte app
		j, jErr := json.MarshalIndent(cs, "", " ")
		CheckErr(jErr, "run - genGlobalLevelConnections")
		brjs := fmt.Sprintf("%s", j)
		key := string("viz/vrctlviz::" + "::connection::" + c.LocalAddress.String() + "::vrf::global")
		fmt.Println("Publishing to ETCD: ", key, lease, brjs)
		// Copy value to etcd and associate with our existing lease for automatic cleanup
		etcdPutShortLease(key, brjs)
	}

}
