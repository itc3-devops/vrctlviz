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
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"regexp"
	"strconv"
	"strings"

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

		// Start server for exporting expvars
		go func() {
			log.Println(http.ListenAndServe("127.0.0.1:6061", nil))

		}()
		resp := etcdHealthMemberListCheck()
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
	time := strintToInt64(ts)

	// Run script to collect vizceral data from local host
	stats := collectTcpMetrics()
	fmt.Println("Print stats", stats)

	// Set vars
	renderer := "region"
	deviceIp := os.Getenv("PrivateIP")
	maxvol := float64(50000.100)
	class := "normal"

	// Generate data for structs
	// Generate global level connections between nodes
	genGlobalLevelConnections(stats, deviceIp)

	// Generate node level region/service hierarchy
	regionServiceNodes := genRegionServiceNodes(stats)
	regionServiceConnections, notices, metadata := genRegionServiceConnections(stats, deviceIp)

	ns := VizceralNode{
		Renderer:    renderer,
		Name:        deviceIp,
		MaxVolume:   maxvol,
		Updated:     time,
		Nodes:       regionServiceNodes,
		Connections: regionServiceConnections,
		Notices:     notices,
		Class:       class,
		Metadata:    metadata,
	}

	// add some filters to check for empty values
	if renderer == "" {
	} else if deviceIp == "" {
	} else {
		// serialize data for host hierarchy
		j, jErr := json.MarshalIndent(ns, "", " ")
		checkErr(jErr, "run - genGlobalLevelGraph")
		brjs := fmt.Sprintf("%s", j)
		lease := getLeaseNumber()
		key := string("viz/vrctlviz::lease::" + lease + "::node::" + deviceIp + "::vrf::global")
		// upload to etcd and associate with our lease for automatic cleanup
		etcdPutLongLease(key, brjs)
	}
}

// Create node level services
func genRegionServiceNodes(data string) []VizceralNode {
	// Create interfaces
	vsg := []VizceralNode{}

	return vsg
}

// Create service level connections
func genRegionServiceConnections(data string, deviceIp string) ([]VizceralConnection, []VizceralNotice, VizceralMetadata) {
	// Create interfaces
	vcg := []VizceralConnection{}
	vng := []VizceralNotice{}
	vm := VizceralMetadata{}

	return vcg, vng, vm
}

// Creates connection information to be loaded into the top level global graph
func genGlobalLevelConnections(stats string, deviceIp string) {

	// Create vars
	var source string
	var target string
	var metrics VizceralLevels
	var meta VizceralMetadata
	var notices []VizceralNotice
	var class string

	// Set vars
	source = os.Getenv("PrivateIP")
	class = "normal"
	meta = VizceralMetadata{}

	// Read through stats output line by line
	scanner := bufio.NewScanner(strings.NewReader(stats))
	for scanner.Scan() {
		st := (scanner.Text())
		// Each line of the stats output is represented by 'st'

		// Remove spaces and break each value out on its own line for easier filtering
		input := st
		re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
		re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
		final := re_leadclose_whtsp.ReplaceAllString(input, "")
		final = re_inside_whtsp.ReplaceAllString(final, " ")
		fmt.Println("Print final: ", final)

		// Filter out metrics we want to use to build out each data struct
		// Only focus on connections that are in an Established state
		if strings.Contains(final, "ESTAB") {
			// Parse our values out of the filtered Established strings
			target, metrics, _, notices = parseTargetMetricsProtocol(final, source)
			fmt.Println("Print target", target)
		}

		// statically set status to the same as metrics for now
		status := metrics

		cs := VizceralConnection{
			Source:   source,
			Target:   target,
			Metrics:  metrics,
			Metadata: meta,
			Status:   status,
			Notices:  notices,
			Class:    class,
		}

		// Run some filters to make sure we don't have empty values
		if source == "" {
		} else if target == "" {
		} else {

			// serialize the data and publish each connection in seperate ETCD key for assembly on a global
			// view by the aggriagte app
			j, jErr := json.MarshalIndent(cs, "", " ")
			checkErr(jErr, "run - genGlobalLevelConnections")
			brjs := fmt.Sprintf("%s", j)
			lease := getLeaseNumber()
			key := string("viz/vrctlviz::lease::" + lease + "::connection::" + ipD + "::ip::" + deviceIp + "::vrf::global")
			// Copy value to etcd and associate with our existing lease for automatic cleanup
			etcdPutShortLease(key, brjs)
		}

	}

}

// Parse out all the node and connection information
func parseTargetMetricsProtocol(final string, source string) (string, VizceralLevels, int, []VizceralNotice) {

	var warningTraffic float64
	var normalTraffic float64
	var dangerTraffic float64

	// create vars
	var ip string
	notices := []VizceralNotice{}
	var protocol int
	// Split on space to filter on continus strings only.
	result := strings.Split(final, " ")

	// Iterate over each attribute in TCP string
	for i := range result {
		si := fmt.Sprintf("%s", result[i])
		if strings.Contains(si, "rtt:") {
			// Set warning level traffic based on RTT which indicates congestion and collisions on the network
			// Send attribute to function to set struct
			warningTraffic = roundTripAverage(si)
		}
		if strings.Contains(si, "Mbps") {
			// Set normal level traffic by data sent
			// Send attribute to function to set struct
			normalTraffic = dataSend(si)
		}

		if strings.Contains(si, "retrans") {
			// Set danger level based on retransmissions
			// Send attribute to function to set struct
			dangerTraffic = retransRate(si)
		}

		// pass levels to function to add to struct
		metrics = formatMetricsStruct(normalTraffic, warningTraffic, dangerTraffic)

		// parse out and format notices
		notices = parseFormatNotices(si)

		// Parse out target IP address
		fmt.Println("Print si: ", si)
		// Include any string that has .
		if strings.Contains(si, ".") {

			// Drop any string with / by invoking a nul action
			if strings.Contains(si, "/") {
				fmt.Println("Dropping this value / filter: ", si)
				// Drop any strings that have the letter r by invoking a nul action
			} else if strings.Contains(si, "r") {
				fmt.Println("Dropping this value r filter: ", si)
				// Drop anything that matches our own source address to filter out loops
			} else if strings.Contains(si, source) {
				fmt.Println("Dropping this value source filter: ", si)
				// Of the strings that are left, include any string that has :
			} else if strings.Contains(si, ":") {
				fmt.Println("Print values that passed all filters: ", si)
				// target ip is listed before the : and the protocol is listed after the :
				ip = before(si, ":")
				protocol = 1
			}
		}
	}

	return ip, metrics, protocol, notices
}

// parse warning traffic from tcp string
func roundTripAverage(data string) float64 {
	// create vars
	var warn float64
	var err error

	// set vars
	// get attribute string that follows rtt
	w := (after(data, "rtt:"))
	// remove trailing / and value
	w1 := (after(w, "/"))

	// make sure value contains a . so we can set it as a float64 value, if not add a .
	if strings.Contains(w1, ".") {
		warn, err = strconv.ParseFloat(w1, 64)
		checkErr(err, "run - roundTripAverage")
	} else {
		w1 := strings.Join([]string{w1, ".100"}, "")
		warn, err = strconv.ParseFloat(w1, 64)
		checkErr(err, "run - roundTripAverage")
	}
	return warn
}

// parse normal traffic from tcp string
func dataSend(data string) float64 {
	// create vars
	var normal float64
	var err error
	// set vars
	// set value to int in front of Mbps
	n1 := before(data, "Mbps")

	// add a 0 in front of the . to add the volume of the traffic graph
	n2 := before(n1, ".")
	n := strings.Join([]string{n2, "0"}, "")

	// make sure string has  a . if not add one so we can convert to float64
	if strings.Contains(n, ".") {
		normal, err = strconv.ParseFloat(n, 64)
		checkErr(err, "run - dataSend")

	} else {
		w := strings.Join([]string{n, ".100"}, "")
		normal, err = strconv.ParseFloat(w, 64)
		checkErr(err, "run - dataSend")
	}
	return normal
}

// parse danger traffic from tcp string
func retransRate(data string) float64 {
	var danger float64
	var err error
	// set vars
	// set danager level on retransmits
	n := (after(data, "retrans:0/"))
	// if value has no . add one so we can set float64
	if strings.Contains(n, ".") {
		danger, err = strconv.ParseFloat(n, 64)
		checkErr(err, "run - retransRate")

	} else {
		w := strings.Join([]string{n, ".100"}, "")
		danger, err = strconv.ParseFloat(w, 64)
		checkErr(err, "run - retransRate")
	}
	return danger
}

// parse notices
// func parseNotices(data string) []VizceralNotice {

// }

// format metrics levels into struct
func formatMetricsStruct(norm float64, warn float64, dang float64) VizceralLevels {

	m := VizceralLevels{
		Normal:  norm,
		Warning: warn,
		Danger:  dang,
	}

	// return node group back to main function as a node array formatted with the Node map
	return m
}

// format metadata into struct
func formatMetadataStruct(proto int) VizceralMetadata {

	rMetadata := VizceralMetadata{
		Streaming: proto,
	}

	// return node group back to main function as a node array formatted with the Node map
	return rMetadata
}

// format notices into struct
func parseFormatNotices(data string) []VizceralNotice {

	// statically set notices for now
	ng := []VizceralNotice{}

	n := VizceralNotice{
		Title:    "Title",
		Subtitle: "Subtitle",
		Link:     "http://netmon.itc3.io",
		Severity: 0,
	}
	ng = append(ng, n)
	// return node group back to main function as a node array formatted with the Node map
	return ng
}
