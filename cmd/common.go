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
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/joho/godotenv"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
)

var prefixValue string
var prefixKey string
var etcdCertString string
var etcdCaCertString string
var etcdKeyString string
var dialTimeout = 5 * time.Second
var requestTimeout = 10 * time.Second
var endpoints = []string{(os.Getenv("ETCDCTL_ENDPOINTS"))}
var tlsInfo = transport.TLSInfo{

	CertFile:      os.Getenv("ETCDCTL_CERT"),
	KeyFile:       os.Getenv("ETCDCTL_KEY"),
	TrustedCAFile: os.Getenv("ETCDCTL_CACERT"),
}

func checkErr(err error, label string) {
	if err != nil {
		fmt.Println(err.Error())
		log.WithFields(log.Fields{"vrctl": label}).Error("NOTIFY - General Error Handler", err)

	}
}

// Loads environment variables by sourcing /.shared/status/vars file
func loadHostEnvironmentVars() {
	// Load environment variables
	createFile("/.shared/status/vars")
	envErr := godotenv.Load("/.shared/status/vars")
	if envErr != nil {

		log.WithFields(log.Fields{"run": "Load Environment"}).Error("Dang! Error loading Environment Variables", envErr)
	}
}

// sleep timer in seconds
func sleep(length time.Duration) {

	time.Sleep(time.Second * length)

}

func before(value string, a string) string {
	// Get substring before a string.
	// Used to parse out the device name
	pos := strings.Index(value, a)
	if pos == -1 {
		return ""
	}
	return value[0:pos]
}

func between(value string, a string, b string) string {
	// Get substring between two strings.
	// Used to parse out the full key string
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func after(value string, a string) string {
	// Get substring after a string.
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:len(value)]
}

func createFile(path string) {
	// Create folder structure for file if not already exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir, _ := filepath.Split(path)
		mkDir(dir)
	}
	// create file if not exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		var file, err = os.Create(path)
		checkErr(err, "setup - createFile") //okay to call os.exit()
		defer file.Close()
	}
}

func writeFile(path string, contents string) {
	// If file does not exist, create directory structure then create file
	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir, file := filepath.Split(path)
		mkDir(dir)
		createFile(file)
	}
	// open file using READ & WRITE permission
	var file, err = os.OpenFile(path, os.O_RDWR, 0644)
	checkErr(err, "setup - writeFile")
	defer file.Close()

	// write some text to file
	_, err = file.WriteString(contents)
	if err != nil {
		fmt.Println(err.Error())
		return //must return here for defer statements to be called
	}

	// save changes
	err = file.Sync()
	if err != nil {
		fmt.Println(err.Error())
		return //same as above
	}
}

func mkDir(dir string) {

	if _, mkDirErr := os.Stat(dir); os.IsNotExist(mkDirErr) {
		mkDirErr = os.MkdirAll(dir, 0755)
		checkErr(mkDirErr, "Make directory")
	}
}

func hostCommandWithOutput(command string, arguments []string) (string, error) {
	out, err := exec.Command(command, arguments...).Output()
	if err != nil {
		fmt.Sprintf("Failed to execute command: %s", err)
		checkErr(err, "hostCommandWithOutput")
	}
	outStr := string(out)
	return outStr, err

}

func collectTcpMetrics() string {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Set path for collect script
	filePath := string(home + "collect.sh")

	// See if collect.sh exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// If script does not exist, create it
		writeCollectorScript()
	}

	command := "/bin/sh"
	arguments := []string{"-c", filePath}
	stat, _ := hostCommandWithOutput(command, arguments)
	return stat
}

func etcdPutLeaseForever(key string, value string) {

	loadHostEnvironmentVars()

	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.WithFields(log.Fields{"vrctl": "ETCD putLeaseForever"}).Error("Error exporting TLS config:  ", err)
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
		TLS:         tlsConfig,
	})
	if err != nil {
		log.WithFields(log.Fields{"vrctl": "ETCD putLeaseForever"}).Error("NOTIFY - Creating new ETCD client listener", err)
	}
	defer cli.Close() // make sure to close the client

	opts := getEtcdPutOptions()
	log.WithFields(log.Fields{"vrctl": "ETCD putLeaseForever"}).Debug("Print etcdPutOptions:  ", opts)
	resp, err := cli.Put(context.TODO(), key, value, opts...)

	if err != nil {
		log.WithFields(log.Fields{"vrctl": "ETCD putLeaseForever"}).Error("Error putting key in ETCD:  ", err)

	}

	fmt.Println(*resp)

}

func getEtcdPutOptions() []clientv3.OpOption {
	loadHostEnvironmentVars()
	serialNumber := os.Getenv("SerialNumber")
	var err error

	// Get lease ID from active device key in ETCD, the most accurate source for the LeaseID
	key1 := strings.Join([]string{"mgmt/active-devices", serialNumber}, "/")
	_, leaseStr := etcdKeyGetPrefix(key1)

	// leaseInt, err := strconv.ParseInt(leaseStr, 64)
	log.WithFields(log.Fields{"vrctl": "ETCD putLeaseForever"}).Debug("Current Lease:  ", leaseStr)
	id, err := strconv.ParseInt(leaseStr, 16, 64)
	if err != nil {
		log.WithFields(log.Fields{"vrctl": "ETCD putLeaseForever"}).Error("Error parsing LeaseID:  ", err)

	}

	opts := []clientv3.OpOption{}
	if id != 0 {
		opts = append(opts, clientv3.WithLease(clientv3.LeaseID(id)))
	}

	return opts
}

func changeFilePermissions(path string, permission os.FileMode) {
	if err := os.Chmod(path, permission); err != nil {
		log.Fatal(err)
	}
}

func etcdKeyGetPrefix(key string) (string, string) {

	tlsConfig, err := tlsInfo.ClientConfig()
	checkErr(err, "generic - label")
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
		TLS:         tlsConfig,
	})
	checkErr(err, "generic - label")
	defer cli.Close() // make sure to close the client

	for i := range make([]int, 3) {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		_, err = cli.Put(ctx, fmt.Sprintf("key_%d", i), "value")
		cancel()
		checkErr(err, "generic - label")
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := cli.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)

		prefixValue = fmt.Sprintf("%s", ev.Value)
		prefixKey = fmt.Sprintf("%s", ev.Key)

	}
	return prefixKey, prefixValue
}

func writeCollectorScript() {
	log.WithFields(log.Fields{"run": "writeCollectorScript"}).Debug("Writing collector script")
	collectScript, err := Asset("data/collect.sh")
	collectScriptString := fmt.Sprintf("%s", collectScript)

	if err != nil {
		log.WithFields(log.Fields{"run": "writeCollectorScript"}).Error("Can not recover collect.sh")
	}

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	filePath := string(home + "collect.sh")
	createFile(filePath)

	writeFile(filePath, collectScriptString)
	changeFilePermissions(filePath, 0777)
}
