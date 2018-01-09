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
	"fmt"
	"os"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(os.Args) > 1 {
			dialTimeout := 5 * time.Second
			requestTimeout := 10 * time.Second
			endpoints := []string{(os.Getenv("ETCDCTL_ENDPOINTS"))}
			tlsInfo := transport.TLSInfo{

				CertFile:      os.Getenv("ETCDCTL_CERT"),
				KeyFile:       os.Getenv("ETCDCTL_KEY"),
				TrustedCAFile: os.Getenv("ETCDCTL_CACERT"),
			}
			tlsConfig, err := tlsInfo.ClientConfig()
			CheckErr(err, "common - requestEtcdDialer")
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			cli, err := clientv3.New(clientv3.Config{
				Endpoints:   endpoints,
				DialTimeout: dialTimeout,
				TLS:         tlsConfig,
			})
			CheckErr(err, "common - requestEtcdDialer")
			defer cli.Close() // make sure to close the client

			prefix := os.Args[2]
			// pull nodes from etcd
			resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
			cancel()
			CheckErr(err, "vizceral - genTopLevelView - get node keys")

			// iterate through each key for adding to the array
			for _, ev := range resp.Kvs {

				// convert etcd key/values to strings
				cKey := fmt.Sprintf("%s", ev.Key)
				cValue := fmt.Sprintf("%s", ev.Value)

				fmt.Println("Print response: ", cKey, cValue)
			}
		} else {
			fmt.Println("Need a key")
		}
	},
}

func init() {
	RootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
