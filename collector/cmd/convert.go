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
	"os"
	"strconv"
	"strings"
	"time"
	"github.com/pkg/errors"

	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/cobra"
)

var viz []VizceralGraph

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "server",
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
		resp := EtcdHealthMemberListCheck()
		if resp == true {
			// fmt.Println("ETCD cluster is healthy")
			vizAutoRun()
		} else {
			// fmt.Println("ETCD cluster unreachable, or unhealthy")
		}

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

	// start scheduler to run the app at a certain interval
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// Do stuff
				// Define wait timer between task cycles
				time.Sleep(time.Second * 5)
				genGlobalLevelGraph()

			case <-quit:
				ticker.Stop()
				// fmt.Println("Stopped the ticker!")
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

// create top level graph
func genGlobalLevelGraph() {

	// Set vars
	renderer := "global"
	name := "edge"
	maxvol := float64(50000.100)

	// Generate node level region/service hierarchy
	regionServiceNodes := regionServiceNodes()
	regionServiceConnections := regionServiceConnections()

	ns := VizceralGraph{
		Renderer:    renderer,
		Name:        name,
		MaxVolume:   maxvol,
		Nodes:       regionServiceNodes,
		Connections: regionServiceConnections,
	}
	// n := fmt.Sprintf("%v", ns)
	// serialize and write data to file

	df := os.Getenv("TRAFFIC_URL")
	dataFile := string("/usr/src/app/dist/" + df)
	j, jErr := json.MarshalIndent(ns, "", " ")
	CheckErr(jErr, "Viz - Top level global vrf view")
	brjs := fmt.Sprintf("%s", j)
	fmt.Println(brjs)
	WriteFile(dataFile, brjs)
}

// create top level graph for api calls
func genApiGlobalLevelGraph(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept Content-Type")
	w.Header().Set("content-type", "application/json")

	// Set vars
	renderer := "global"
	name := "edge"
	maxvol := float64(50000.100)

	// Generate node level region/service hierarchy
	regionServiceNodes, regionServiceConnections := fetchDataFromEtcD()

	ns := VizceralGraph{
		Renderer:    renderer,
		Name:        name,
		MaxVolume:   maxvol,
		Nodes:       regionServiceNodes,
		Connections: regionServiceConnections,
	}
	// n := fmt.Sprintf("%v", ns)
	// serialize and write data to file

	fmt.Println("serializing data")
	//viz = append(viz, ns)
	json.NewEncoder(w).Encode(ns)
}

// Creates connection information to be loaded into the top level global graph
func regionServiceConnections() []VizceralConnection {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"etcd:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		// handle error!
	}
	defer cli.Close()
	// create vars
	vcg := []VizceralConnection{}
	vc := []VizceralConnection{}

	// set vars
	keyPrefix := "viz/vrctlviz::"

	// get etcd keys based on connection prefix
	resp, err := cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))

	CheckErr(err, "vizceral - genTopLevelView - get node keys")

	// iterate through each key for adding to the array
	for _, ev := range resp.Kvs {

		// convert etcd key/values to strings
		cKey := fmt.Sprintf("%s", ev.Key)
		cValue := fmt.Sprintf("%s", ev.Value)
		fmt.Println("Print all keys: ", cKey)

		if strings.Contains(cKey, "connection") {
			fmt.Println("Print all keys that pass connection filter: ", cKey)
			fmt.Println("Print lease: ", lease)
			fmt.Println("Print etcd connections key: ", cValue)
			//unmarshall value into struct
			err := json.Unmarshal([]byte(cValue), &vc)
			if err != nil {
				log.Fatalf("failed to decode: %s", err)
				log.Fatalf("%+v", errors.WithStack(err))
			}
			// add connection to the interface
			fmt.Println("Print unmarshalled connections: ", vc)
			vcg = append(vcg, vc...)

		}
	}
	return vcg

}

func regionServiceNodes() []VizceralNode {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"etcd:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		// handle error!
	}
	defer cli.Close()
	// create vars
	vng := []VizceralNode{}
	vig := []VizceralNode{}
	vn := VizceralNode{}
	vc := []VizceralConnection{}

	keyPrefix := "viz/vrctlviz::"

	// pull nodes from etcd
	resp, err := cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))

	CheckErr(err, "vizceral - genTopLevelView - get node keys")

	// iterate through each key for adding to the array
	for _, ev := range resp.Kvs {

		// convert etcd key/values to strings
		cKey := fmt.Sprintf("%s", ev.Key)
		cValue := fmt.Sprintf("%s", ev.Value)
		// fmt.Println("Print all keys: ", cKey)
		// filter out anything that is not a node key
		if strings.Contains(cKey, "node") {
			// fmt.Println("Print all keys that pass the node filter: ", cKey)
			// fmt.Println("Print lease: ", lease)
			// fmt.Println("Print etcd node keys: ", cValue)
			// unmarshall value into struct
			err := json.Unmarshal([]byte(cValue), &vn)
			if err != nil {
				log.Fatalf("failed to decode: %s", err)
			}
			// fmt.Println("Print unmarshalled ndoes: ", vn)
			// add node to the interface
			vng = append(vng, vn)
		}
	}
	// Get timestamp and convert it to proper format
	ts := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	time := StrintToInt64(ts)

	vni := VizceralNode{
		Renderer:    "region",
		Name:        "INTERNET",
		Connections: vc,
		Nodes:       vig,
		Updated:     time,
	}
	vng = append(vng, vni)

	return vng

}

func fetchDataFromEtcD() ([]VizceralNode, []VizceralConnection) {
	// Connect to EtcD
	cli, _ := clientv3.New(clientv3.Config{
		Endpoints:   []string{"etcd:2379"},
		DialTimeout: 5 * time.Second,
	})

	// Close the connection when complete
	defer cli.Close()

	// Get data
	resp, _ := cli.Get(context.Background(),
						 "viz/vrctlviz::",
						 clientv3.WithPrefix(),
						 clientv3.WithSort(clientv3.SortByCreateRevision,
						 clientv3.SortDescend))

	// Setup containers
	nodes := make(map[string]VizceralNode)
	raw_connections := make(map[string][]VizceralConnection)
	connections := make(map[string]VizceralConnection)

	// Iterate over the data and extract the nodes and connections
	for _, item := range resp.Kvs {
		switch {
		// Extract the nodes
		case strings.Contains(string(item.Key), "node"):
			var n VizceralNode

			// Convert JSON data
			err := json.Unmarshal([]byte(item.Value), &n)

			// Handle errors
			if err != nil {
				log.Fatalf("failed to decode: %s", err)
			}

			// Add node to the map if it's not already there

			if _, ok := nodes[n.Name]; ! ok {
				nodes[n.Name] = n
			}

		// Extract the connection
		case strings.Contains(string(item.Key), "connection"):
			var c []VizceralConnection

			// Convert JSON data
			err := json.Unmarshal([]byte(item.Value), &c)

			// Handle errors
			if err != nil {
				log.Fatalf("failed to decode: %s", err)
			}

			for _, con := range c {
				// Add node to the map if it's not already there
				if _, ok := raw_connections[con.Source]; ! ok {
					raw_connections[con.Source] = c
				}
				break
			}
		}
	}

	// Parse the connections
	for _, c_group := range raw_connections {
		for _, c := range c_group {
			key := c.Source + "->" + c.Target
			connections[key] = c
		}
	}

	// Convert maps to arrays
	ret_nodes := []VizceralNode{}
	ret_connections := []VizceralConnection{}
	for _, n := range nodes {
		ret_nodes = append(ret_nodes, n)
	}
	for _, c := range connections {
		ret_connections = append(ret_connections, c)
	}

	// Add internet node
	vni := VizceralNode{
		Renderer:    "region",
		Name:        "INTERNET",
		Connections: []VizceralConnection{},
		Nodes:       []VizceralNode{},
		Updated:     time.Now().UTC().Unix(),
	}

	ret_nodes = append(ret_nodes, vni)

	parsed_nodes := mergeNodes(ret_nodes, extractNodesFromConnections(ret_connections))

	return parsed_nodes, ret_connections
}

func extractNodesFromConnections(connections []VizceralConnection) []VizceralNode {
	nodes_map := make(map[string]VizceralNode)
	nodes := []VizceralNode{}

	// Go through the connections and extract all the nodes
	for _, con := range connections {
		// Extract the Source
		if _, ok := nodes_map[con.Source]; ! ok {
			nodes_map[con.Source] = VizceralNode{
				Renderer:    "region",
				Name:        con.Source,
				Connections: []VizceralConnection{},
				Nodes:       []VizceralNode{},
				Updated:     time.Now().UTC().Unix(),
			}
		}

		// Extract the Target
		if _, ok := nodes_map[con.Target]; ! ok {
			nodes_map[con.Target] = VizceralNode{
				Renderer:    "region",
				Name:        con.Target,
				Connections: []VizceralConnection{},
				Nodes:       []VizceralNode{},
				Updated:     time.Now().UTC().Unix(),
			}
		}
	}

	// Convert the map into an array
	for _, n := range nodes_map {
		nodes = append(nodes, n)
	}

	return nodes
}

func mergeNodes(a []VizceralNode, b []VizceralNode) []VizceralNode {
	nodes_map := make(map[string]VizceralNode)
	nodes := []VizceralNode{}

	// Start with b as it wins
	for _, node := range b {
		if _, ok := nodes_map[node.Name]; ! ok {
			nodes_map[node.Name] = node
		}
	}

	// Now do the same for a
	for _, node := range a {
		if _, ok := nodes_map[node.Name]; ! ok {
			nodes_map[node.Name] = node
		}
	}

	// At this point, we're merged and deduped so lets convert back to an array
	for _, n := range nodes_map {
		nodes = append(nodes, n)
	}
	return nodes
}
