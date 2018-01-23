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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var home string
var err error
var config string
var configFile = "/root/.vrctlvizcfg"
var varsFile = "/root/.vrctlvizcfg"
var cfgFile = "/root/.vrctlvizcfg"
var ConfigString string
var prefixValue string
var prefixKey string
var etcdCertString string
var etcdCaCertString string
var etcdKeyString string
var requestTimeout = 10 * time.Second
var key string
var value string
var lease string
var LeaseHex clientv3.LeaseID
var cli *clientv3.Client
var dialTimeout = 5 * time.Second

// vizceral generates output in the NetflixOSS vizceral format
// https://github.com/Netflix/vizceral/blob/master/DATAFORMATS.md

// Metadata
type VizceralMetadata struct {
	Streaming int `json:"streaming"`
}

// Notice appears in the sidebar
type VizceralNotice struct {
	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	Link     string `json:"link,omitempty"`
	Severity int    `json:"severity,omitempty"`
}

// Levels of trafic in each state
type VizceralLevels struct {
	Danger  float64 `json:"danger,omitempty"`
	Warning float64 `json:"warning,omitempty"`
	Normal  float64 `json:"normal,omitempty"`
}

// One Connection
type VizceralConnection struct {
	Source   string           `json:"source,omitempty"`
	Target   string           `json:"target,omitempty"`
	Metadata VizceralMetadata `json:"metadata,omitempty"`
	Metrics  VizceralLevels   `json:"metrics,omitempty"`
	Status   VizceralLevels   `json:"status,omitempty"`
	Notices  []VizceralNotice `json:"node,omitempty"`
	Class    string           `json:"class,omitempty"`
}

// One node (region/service hierarchy)
type VizceralNode struct {
	Renderer    string               `json:"renderer,omitempty"` // 'region' or omit for service
	Name        string               `json:"name,omitempty"`
	MaxVolume   float64              `json:"maxVolume,omitempty"` // relative base for levels animation
	Updated     int64                `json:"updated,omitempty"`   // Unix timestamp. Only checked on the top-level list of nodes. Last time the data was updated
	Nodes       []VizceralNode       `json:"nodes,omitempty"`
	Connections []VizceralConnection `json:"connections,omitempty"`
	Notices     []VizceralNotice     `json:"notices,omitempty"`
	Class       string               `json:"class,omitempty"` // 'normal', 'warning', or 'danger'
	Metadata    VizceralMetadata     `json:"metadata,omitempty"`
}

// Global level of graph file format
type VizceralGraph struct {
	Renderer    string               `json:"renderer"` // 'global'
	Name        string               `json:"name"`
	MaxVolume   float64              `json:"maxVolume,omitempty"` // relative base for levels animation
	Nodes       []VizceralNode       `json:"nodes,omitempty"`
	Connections []VizceralConnection `json:"connections,omitempty"`
}

// Endpoints : Define ETCD Endpoints
var Endpoints []string

// EtcdApi : Define ETCD Api version
var EtcdApi string

// tlsInfo : Define TLS connection for ETCD
var tlsInfo transport.TLSInfo

// GetEndpointsConfig : Return ETCD Endpoints
func GetEndpointsConfig() []string {
	os.Setenv("ETCDCTL_API", viper.GetString("Etcd.Api"))
	//fmt.Println("ETCD Api version: ", os.Getenv("ETCDCTL_API"))
	os.Setenv("ETCDCTL_ENDPOINTS", viper.GetString("Etcd.Endpoints"))
	//fmt.Println("ETCD ENDPOINTS: ", os.Getenv("ETCDCTL_ENDPOINTS"))
	os.Setenv("ETCDCTL_CERT", viper.GetString("Etcd.Cert"))
	//fmt.Println("ETCD CERT: ", os.Getenv("ETCDCTL_CERT"))
	os.Setenv("ETCDCTL_CACERT", viper.GetString("Etcd.CaCert"))
	//fmt.Println("ETCD CACERT: ", os.Getenv("ETCDCTL_CACERT"))
	os.Setenv("ETCDCTL_KEY", viper.GetString("Etcd.Key"))
	//fmt.Println("ETCD KEY: ", os.Getenv("ETCDCTL_KEY"))
	return []string{(viper.GetViper().GetString("Etcd.Endpoints"))}
}

// GetTlsInfo : Return configured TLS info
func GetTlsInfo() transport.TLSInfo {
	return transport.TLSInfo{
		CertFile:      viper.GetViper().GetString("Etcd.Cert"),
		KeyFile:       viper.GetViper().GetString("Etcd.Key"),
		TrustedCAFile: viper.GetViper().GetString("Etcd.CaCert"),
	}
}

// GetEtcdTlsCli : Return new ETCD TLS client
func GetEtcdTlsCli() (*clientv3.Client, error) {
	Endpoints := GetEndpointsConfig()
	tlsInfo := GetTlsInfo()

	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		//fmt.Println("Error setting TLS config: ", err)
		log.WithFields(log.Fields{"common": "EtcdTlsDialer"}).Error("Failed to create TLS config: ", err)
		os.Exit(1)
	}
	return clientv3.New(clientv3.Config{
		Endpoints:   Endpoints,
		DialTimeout: dialTimeout,
		TLS:         tlsConfig,
	})
}

// GetEtcdCli : Return new ETCD dialer
func GetEtcdCli() (*clientv3.Client, error) {
	Endpoints := GetEndpointsConfig()
	return clientv3.New(clientv3.Config{
		Endpoints: Endpoints,
	})
}

// GetEtcdClient : Return new ETCD dialer based on config TLS settings
func GetEtcdClient() (*clientv3.Client, error) {
	if viper.IsSet("Etcd.Tls") {
		//fmt.Println("Creating new TLS Etcd Dialer")
		return GetEtcdTlsCli()
	}
	//fmt.Println("Creating new ETCD Dialer")
	return GetEtcdCli()

}

// print a Vizceral graph as json
func vizFileWrite(v *VizceralGraph) {
	vJson, _ := json.Marshal(*v)
	sJson := fmt.Sprintf("%s", vJson)

	df := os.Getenv("TRAFFIC_URL")
	dataFile := string("/usr/src/app/dist/" + df)
	createFile(dataFile)
	WriteFile(dataFile, sJson)
}

// Read a Vizceral format file into a graph
func vizFileReadFile(fn string) *VizceralGraph {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.WithFields(log.Fields{"common": "Format vizceral graph"}).Error("NOTIFY - Erro reading file", err)
	}
	v := new(VizceralGraph)
	err = json.Unmarshal(data, v)
	if err != nil {
		log.WithFields(log.Fields{"common": "Format vizceral graph"}).Error("NOTIFY - Error formatting file into a graph", err)
	}
	return v
}

func serializeVizceral(data string) {

	df := os.Getenv("TRAFFIC_URL")
	dataFile := string("/usr/src/app/dist/" + df)
	j, jErr := json.MarshalIndent(data, "", " ")
	CheckErr(jErr, "Viz - Top level global vrf view")
	brjs := fmt.Sprintf("%s", j)
	fmt.Println(brjs)
	deleteFile(dataFile)
	createFile(dataFile)
	WriteFile(dataFile, brjs)

}

// Read a Vizceral format file into a graph
func vizFileReadata(data string) *VizceralGraph {
	v := new(VizceralGraph)
	err := json.Unmarshal([]byte(data), v)
	if err != nil {
		log.WithFields(log.Fields{"common": "Format vizceral graph"}).Error("NOTIFY - Error formatting file into a graph", err)
	}
	return v
}

func CheckErr(err error, label string) {
	if err != nil {
		fmt.Println(err.Error())
		log.WithFields(log.Fields{"common": label}).Error("NOTIFY - General Error Handler", err)

	}
}

func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
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

// Convert string to int64
func StrintToInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
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
		CheckErr(err, "setup - createFile") //okay to call os.exit()
		defer file.Close()
	}
}

func WriteFile1(path string, contents string) {
	err := ioutil.WriteFile(path, []byte(contents), 0666)
	if err != nil {
		log.Fatal(err)
	}
}

var mutex = &sync.Mutex{}

func WriteFile(path string, contents string) {
	mutex.Lock()
	defer mutex.Unlock()
	err := ioutil.WriteFile(path, []byte(contents), 0666)
	if err != nil {
		log.Fatal(err)
	}
}

func deleteFile(path string) error {

	// detect if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		CheckErr(err, "generic - label") //okay to call os.exit()
		defer file.Close()
	}
	// delete file
	var err1 = os.Remove(path)
	CheckErr(err1, "Delete file")
	return err1
}

func mkDir(dir string) {

	if _, mkDirErr := os.Stat(dir); os.IsNotExist(mkDirErr) {
		mkDirErr = os.MkdirAll(dir, 0755)
		CheckErr(mkDirErr, "Make directory")
	}
}

func hostCommandWithOutput(command string, arguments []string) (string, error) {
	out, err := exec.Command(command, arguments...).Output()
	if err != nil {
		fmt.Sprintf("Failed to execute command: %s", err)
		CheckErr(err, "common - hostCommandWithOutput")
	}
	outStr := string(out)
	return outStr, err

}

func collectTcpMetrics() string {

	// Set path for collect script
	filePath := "/usr/bin/collect.sh"

	command := "/bin/sh"
	arguments := []string{"-c", filePath}
	stat, _ := hostCommandWithOutput(command, arguments)
	return stat
}

func RoutedInterface(network string, flags net.Flags) *net.Interface {
	switch network {
	case "ip", "ip4", "ip6":
	default:
		return nil
	}
	ift, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, ifi := range ift {
		if ifi.Flags&flags != flags {
			continue
		}
		if _, ok := hasRoutableIP(network, &ifi); !ok {
			continue
		}
		return &ifi

	}
	return nil
}

func hasRoutableIP(network string, ifi *net.Interface) (net.IP, bool) {
	ifat, err := ifi.Addrs()
	if err != nil {
		return nil, false
	}
	for _, ifa := range ifat {
		switch ifa := ifa.(type) {
		case *net.IPAddr:
			if ip := routableIP(network, ifa.IP); ip != nil {
				return ip, true
			}
		case *net.IPNet:
			if ip := routableIP(network, ifa.IP); ip != nil {
				return ip, true
			}
		}
	}
	return nil, false
}

func routableIP(network string, ip net.IP) net.IP {
	if !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsGlobalUnicast() {
		return nil
	}
	switch network {
	case "ip4":
		if ip := ip.To4(); ip != nil {
			return ip
		}
	case "ip6":
		if ip.IsLoopback() { // addressing scope of the loopback address depends on each implementation
			return nil
		}
		if ip := ip.To16(); ip != nil && ip.To4() == nil {
			return ip
		}
	default:
		if ip := ip.To4(); ip != nil {
			return ip
		}
		if ip := ip.To16(); ip != nil {
			return ip
		}
	}
	return nil
}

func EtcdHealthMemberListCheck() bool {
	var r bool

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	// Get EtcD client
	cli, err := GetEtcdClient()

	CheckErr(err, "common - requestEtcdDialer")
	defer cli.Close() // make sure to close the client

	resp, err := cli.MemberList(ctx)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	// Check for healthy ETCD cluster, if unhealthy returns false

	if len(resp.Members) != 0 {
		r = true
	} else {
		r = false
	}
	return r
}

func etcdPutShortLease(key string, value string) {

	CheckErr(err, "common - requestEtcdDialer")
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

	// Get EtcD client
	cli, err := GetEtcdClient()

	CheckErr(err, "common - requestEtcdDialer")
	defer cli.Close() // make sure to close the client
	// request lease from ETCD
	LeaseResp, leaseErr := cli.Grant(ctx, 30)
	if leaseErr != nil {
		log.WithFields(log.Fields{"common": "requestEtcdLease"}).Error("Requesting Lease", leaseErr)
	}

	opts := []clientv3.OpOption{}
	if LeaseResp.ID != 0 {
		opts = append(opts, clientv3.WithLease(clientv3.LeaseID(LeaseResp.ID)))
	}

	log.WithFields(log.Fields{"common": "ETCD putLeaseForever"}).Debug("Print etcdPutOptions:  ", opts)
	resp, err := cli.Put(ctx, key, value, opts...)
	cancel()

	if err != nil {
		log.WithFields(log.Fields{"common": "ETCD putLeaseForever"}).Error("Error putting key in ETCD:  ", err)

	}

	fmt.Println(*resp)

}

func etcdPutLongLease(key string, value string) {

	CheckErr(err, "common - requestEtcdDialer")
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

	// Get EtcD client
	cli, err := GetEtcdClient()

	CheckErr(err, "common - requestEtcdDialer")
	defer cli.Close() // make sure to close the client
	// request lease from ETCD
	LeaseResp, leaseErr := cli.Grant(ctx, 300)
	if leaseErr != nil {
		log.WithFields(log.Fields{"common": "requestEtcdLease"}).Error("Requesting Lease", leaseErr)
	}

	opts := []clientv3.OpOption{}
	if LeaseResp.ID != 0 {
		opts = append(opts, clientv3.WithLease(clientv3.LeaseID(LeaseResp.ID)))
	}

	log.WithFields(log.Fields{"common": "ETCD putLeaseForever"}).Debug("Print etcdPutOptions:  ", opts)
	resp, err := cli.Put(ctx, key, value, opts...)
	cancel()

	if err != nil {
		log.WithFields(log.Fields{"common": "ETCD putLeaseForever"}).Error("Error putting key in ETCD:  ", err)

	}

	fmt.Println(*resp)

}

func etcdPutExistingLease(key string, value string) {

	CheckErr(err, "common - requestEtcdDialer")
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

	// Get EtcD client
	cli, err := GetEtcdClient()

	CheckErr(err, "common - requestEtcdDialer")
	defer cli.Close() // make sure to close the client
	opts := getEtcdPutOptions()
	log.WithFields(log.Fields{"common": "ETCD putLeaseForever"}).Debug("Print etcdPutOptions:  ", opts)
	resp, err := cli.Put(ctx, key, value, opts...)
	cancel()

	if err != nil {
		log.WithFields(log.Fields{"common": "ETCD putLeaseForever"}).Error("Error putting key in ETCD:  ", err)

	}

	fmt.Println(*resp)

}

func getEtcdPutOptions() []clientv3.OpOption {

	ls := readFile("/root/lease")
	lease := fmt.Sprintf("%s", ls)
	fmt.Println("Print existing lease: ", lease)
	id, err := strconv.ParseInt(lease, 16, 64)
	if err != nil {
		log.WithFields(log.Fields{"common": "getEtcdPutOptions"}).Error("Error parsing LeaseID:  ", err)
	}

	opts := []clientv3.OpOption{}
	if id != 0 {
		opts = append(opts, clientv3.WithLease(clientv3.LeaseID(id)))
	}

	return opts
}

func readFile(path string) []byte {

	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return content
}

func changeFilePermissions(path string, permission os.FileMode) {
	if err := os.Chmod(path, permission); err != nil {
		log.Fatal(err)
	}
}

func etcdKeyGetPrefix(key string) (string, string) {

	CheckErr(err, "common - requestEtcdDialer")
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

	// Get EtcD client
	cli, err := GetEtcdClient()

	CheckErr(err, "common - requestEtcdDialer")
	defer cli.Close() // make sure to close the client

	for i := range make([]int, 3) {
		_, err = cli.Put(ctx, fmt.Sprintf("key_%d", i), "value")
		cancel()
		CheckErr(err, "common - etcdKeyGetPrefix")
	}

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

	WriteFile(filePath, collectScriptString)
	changeFilePermissions(filePath, 0777)
}

func requestEtcdLease() {

	CheckErr(err, "common - requestEtcdDialer")
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

	// Get EtcD client
	cli, err := GetEtcdClient()

	CheckErr(err, "common - requestEtcdDialer")
	defer cli.Close() // make sure to close the client

	// request lease from ETCD
	LeaseResp, leaseErr := cli.Grant(ctx, 10)
	if leaseErr != nil {
		log.WithFields(log.Fields{"common": "requestEtcdLease"}).Error("Requesting Lease", leaseErr)
	}

	// convert lease to simple hex for use with etcdctl cli tool
	lSimple := fmt.Sprintf("%016x", LeaseResp.ID)

	leaseKey := strings.Join([]string{"vrctlviz::activeLeases", lSimple}, "::")

	_, err = cli.Put(ctx, leaseKey, lSimple, clientv3.WithLease(LeaseResp.ID))
	cancel()
	if err != nil {
		log.WithFields(log.Fields{"common": "requestEtcdLease"}).Error("Adding hex lease to active-devices key in Etcd", err)
	}

	// Write lease to vars file for reference
	createFile("/root/lease")
	WriteFile("/root/lease", lSimple)

	// start lease keepalive
	// leaseKeepAliveCommandFunc(LeaseResp.ID)

}

// leaseKeepAliveCommandFunc executes the "lease keep-alive" command.
func leaseKeepAliveCommandFunc(leaseId clientv3.LeaseID) {
	id := leaseId

	CheckErr(err, "common - requestEtcdDialer")

	// Get EtcD client
	cli, err := GetEtcdClient()

	CheckErr(err, "common - requestEtcdDialer")
	defer cli.Close() // make sure to close the client

	respc, kerr := cli.KeepAlive(context.Background(), id)
	if kerr != nil {
		log.WithFields(log.Fields{"vrctl": "ETCD keepalive"}).Error("Starting Keepalive for lease", kerr)
	}
	for resp := range respc {
		fmt.Println(*resp)
	}

}

func getLeaseNumber() string {
	b := readFile("/root/lease")
	l := string([]byte(b))
	fmt.Println("Print lease: ", l)
	return l
}
