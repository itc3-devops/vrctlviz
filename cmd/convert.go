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

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
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
		resp := etcdHealthMemberListCheck()
		if resp == true {
			fmt.Println("ETCD cluster is healthy")
			vizAutoRun()
		} else {
			fmt.Println("ETCD cluster unreachable, or unhealthy")
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
	n := fmt.Sprintf("%v", ns)
	// serialize and write data to file

	v := vizFileReadata(n)
	vizFileWrite(v)

	serializeVizceral(n)

	fmt.Println("Print v: \n", v)
}

// Creates connection information to be loaded into the top level global graph
func regionServiceConnections() []VizceralConnection {
	dialTimeout := 5 * time.Second
	requestTimeout := 10 * time.Second
	endpoints := []string{(os.Getenv("ETCDCTL_ENDPOINTS"))}
	tlsInfo := transport.TLSInfo{

		CertFile:      os.Getenv("ETCDCTL_CERT"),
		KeyFile:       os.Getenv("ETCDCTL_KEY"),
		TrustedCAFile: os.Getenv("ETCDCTL_CACERT"),
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	checkErr(err, "common - requestEtcdDialer")
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
		TLS:         tlsConfig,
	})
	checkErr(err, "common - requestEtcdDialer")
	defer cli.Close() // make sure to close the client
	// create vars
	vcg := []VizceralConnection{}
	vc := VizceralConnection{}

	// set vars
	keyPrefix := "viz/vrctlviz::lease::"

	// get etcd keys based on connection prefix
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

func regionServiceNodes() []VizceralNode {
	dialTimeout := 5 * time.Second
	requestTimeout := 10 * time.Second
	endpoints := []string{(os.Getenv("ETCDCTL_ENDPOINTS"))}
	tlsInfo := transport.TLSInfo{

		CertFile:      os.Getenv("ETCDCTL_CERT"),
		KeyFile:       os.Getenv("ETCDCTL_KEY"),
		TrustedCAFile: os.Getenv("ETCDCTL_CACERT"),
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	checkErr(err, "common - requestEtcdDialer")
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
		TLS:         tlsConfig,
	})
	checkErr(err, "common - requestEtcdDialer")
	defer cli.Close() // make sure to close the client
	// create vars
	vng := []VizceralNode{}
	vn := VizceralNode{}
	vc := []VizceralConnection{}

	keyPrefix := "viz/vrctlviz::lease::"

	// pull nodes from etcd
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

			// fmt.Println("Print output: ", cValue)
			// unmarshall value into struct
			err := json.Unmarshal([]byte(cValue), &vn)
			if err != nil {
				log.Fatalf("failed to decode: %s", err)
			}
			// add node to the interface
			vng = append(vng, vn)
		}
	}

	// Get timestamp and convert it to proper format
	ts := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	time := strintToInt64(ts)

	vni := VizceralNode{
		Renderer:    "region",
		Name:        "INTERNET",
		Connections: vc,
		Nodes:       vng,
		Updated:     time,
	}
	vng = append(vng, vni)

	return vng

}
