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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ticker := time.NewTicker(1 * time.Second)
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					// Do stuff
					// Define wait timer between task cycles
					time.Sleep(time.Second * 10)
					genTopLevelView()

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
var ipS string
var errorRate string
var bandwidth string
var norm string
var warn float64
var normalTraffic string
var warningTraffic float64
var metrics tcpMetrics
var deviceName string

type tcpConnections struct {
	Source  string     `json:"source"`
	Target  string     `json:"target"`
	Metrics tcpMetrics `json:"metrics"`
	// Metadata Metadata  `json:"metadata"`
	Class string `json:"class"`
	// Notices  []Notices `json:"notices"`
}

type tcpMetrics struct {
	Normal  string  `json:"normal"`
	Warning float64 `json:"warning"`
}

type tcpNodes struct {
	Name        string `json:"name"`
	Renderer    string `json:"renderer"`
	DisplayName string `json:"displayName"`
	Class       string `json:"class"`
	// MaxVolume string `json:"maxVolume"`
	// Props            Props         `json:"props"`
	// Metadata         Metadata      `json:"metadata"`
	// Notices          []Notices     `json:"notices"`
	// Metrics     Metrics       `json:"metrics"`
	// Nodes       []Nodes       `json:"nodes"`
	Connections []tcpConnections `json:"connections"`
	// ServerUpdateTime string        `json:"serverUpdateTime"`
}

type tcpTopLevelView struct {
	Name        string `json:"name"`
	Renderer    string `json:"renderer"`
	DisplayName string `json:"displayName"`
	Class       string `json:"class"`
	// MaxVolume string `json:"maxVolume"`
	// Props            Props         `json:"props"`
	// Metadata         Metadata      `json:"metadata"`
	// Notices          []Notices     `json:"notices"`
	// Metrics          Metrics       `json:"metrics"`
	Nodes            []tcpNodes       `json:"nodes"`
	Connections      []tcpConnections `json:"connections"`
	ServerUpdateTime string           `json:"serverUpdateTime"`
}

func genConnectionsTcpData(d string, deviceIp string) []tcpConnections {
	// Generate connection data from ETCD using Node map and region filter
	newConnGroup := []tcpConnections{}

	var metrics tcpMetrics

	deviceName = os.Getenv("NodeName")

	ipS = os.Getenv("PrivateIP")
	fmt.Println("Print source IP", ipS)
	scanner := bufio.NewScanner(strings.NewReader(d))
	for scanner.Scan() {
		st := (scanner.Text())

		if strings.Contains(st, deviceIp) {
			// Remove spaces and put each value on its own line for easy grepping
			input := st
			re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
			re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
			final := re_leadclose_whtsp.ReplaceAllString(input, "")
			final = re_inside_whtsp.ReplaceAllString(final, " ")
			// fmt.Println(final)

			// Send data to function to parse out addresses
			if strings.Contains(final, "ESTAB") {
				ipD, metrics = parseDestinationIpAddress(final, ipS)

			}

			cs := tcpConnections{
				Source:  ipS,
				Target:  ipD,
				Metrics: metrics,
				// Metadata: rMetadata,
				Class: "normal",
				// Notices:  rNotices,
			}

			fmt.Println("Print cs: ", cs)

			// Add dataset generated from each ETCD key line item to the connection group so it can be added as an entire set to the regional level dataset
			if ipS == "" {
			} else if ipD == "" {
			} else if normalTraffic == "" {
			} else if warningTraffic == 0 {
			} else {
				newConnGroup = append(newConnGroup, cs)
			}

		}

	}

	// return connection group back to main function as a connection array formatted with the Connection map

	return newConnGroup
}

// Generate top level view data
func genTopLevelView() {
	stats := collectTcpMetrics()

	newNodesGroup := []tcpNodes{}
	tcpConnections := []tcpConnections{}

	renderer := "region"
	deviceName := os.Getenv("NodeName")
	deviceIp := os.Getenv("PrivateIP")
	fmt.Println("Sending device IP to genConnectionsTcpData: ", deviceIp)
	tcpConnections = genConnectionsTcpData(stats, deviceIp)

	// rMetadata := getMetadataData(stats, deviceName)
	// rMetrics := getMetricsData(stats, deviceName)
	// rNotices := getNoticesData(stats, deviceName)
	// rConn := genConnectionsData(stats, keyPrefix, deviceName)

	// rTimestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

	// rNode := genNodesData(stats, rNodePrefix, dL, deviceName)

	// rConn = genConnectionsData(stats, rNodePrefix, dL)

	ns := tcpNodes{
		Renderer:    renderer,
		Name:        deviceIp,
		DisplayName: deviceName,
		Class:       "normal",
		// MaxVolume:        "100000",
		// Metadata:         rMetadata,
		// Notices:          rNotices,
		// Metrics:          rMetrics,
		// Nodes:       rNode,
		Connections: tcpConnections,
	}
	fmt.Println("Adding tcpNodes to array: ", newNodesGroup)
	newNodesGroup = append(newNodesGroup, ns)
	// fmt.Println(newNodesGroup)

	j, jErr := json.MarshalIndent(newNodesGroup, "", " ")
	checkErr(jErr, "Node level view")
	brjs := fmt.Sprintf("%s", j)
	fmt.Println(brjs)
	serialNumber := os.Getenv("SerialNumber")
	key := string("mgmt/vrctlviz-nodes/" + serialNumber)
	etcdPutLeaseForever(key, brjs)
}

func parseDestinationIpAddress(final string, source string) (string, tcpMetrics) {

	var ip string

	// fmt.Println(final)
	// Split on space.
	result := strings.Split(final, " ")

	// Display all elements.
	for i := range result {
		si := fmt.Sprintf("%s", result[i])
		if strings.Contains(si, "rtt:") {
			warningTraffic = (roundTripAverage(si))
		}
		if strings.Contains(si, "bytes_received:") {
			normalTraffic = (bytesSendRec(si))
		}
		metrics = getTcpMetricsData(normalTraffic, warningTraffic)
		fmt.Println("Print si: ", si)
		// Include any string that has .
		if strings.Contains(si, ".") {
			fmt.Println("Print values that passed . filter: ", si)
			// Drop any string with / by invoking a nil action
			if strings.Contains(si, "/") {
				fmt.Println("Dropping this value / filter: ", si)
				//do nothing
				// Drop any strings that have the letter r by invoking a nil action
			} else if strings.Contains(si, "r") {
				fmt.Println("Dropping this value r filter: ", si)
				// do nothing
				// Drop anything that matches our source address
			} else if strings.Contains(si, source) {
				fmt.Println("Dropping this value source filter: ", si)
				// do nothing
				// Of the strings that are left, include any string that has :
			} else if strings.Contains(si, ":") {
				fmt.Println("Print values that passed all filters: ", si)
				// Cut strings before any :
				ip = before(si, ":")

			}
		}
	}

	return ip, metrics
}

func roundTripAverage(data string) float64 {
	// fmt.Println("print data warn", data)
	var warn float64
	var err error
	w := (after(data, "rtt:"))
	// fmt.Println("print w", w)

	w1 := (after(w, "/"))

	if strings.Contains(w1, ".") {
		warn, err = strconv.ParseFloat(w1, 64)
		checkErr(err, "generic - label")

	} else {
		// nothing
	}
	return warn
}

func bytesSendRec(data string) string {

	// fmt.Println("print data norm", data)
	n := (after(data, "bytes_received:"))
	// fmt.Println("print n", n)
	return n
}

func getTcpMetricsData(norm string, warn float64) tcpMetrics {

	rMetrics := tcpMetrics{
		Normal:  norm,
		Warning: warn,
	}

	// return node group back to main function as a node array formatted with the Node map
	return rMetrics
}
