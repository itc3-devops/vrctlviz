package cmd

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _dataCollectSh = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\xd4\x4f\xca\xcc\xd3\x4f\x4a\x2c\xce\xe0\x2a\x2e\x56\xd0\xcd\x2c\x51\xa8\x51\x48\x2f\x4a\x2d\x50\xd0\x2d\x53\x08\x2e\x49\x2c\x49\x55\xa8\x51\x50\x48\x2c\xcf\x56\x50\xf7\x0b\x52\x35\xaa\x2e\x28\xca\xcc\x2b\x49\x53\x50\x52\x2d\x56\x50\xd2\x51\x31\xb0\xce\x4b\xad\x28\xb1\xae\x35\x54\x47\xd2\x65\x68\x64\xae\x67\xa0\x67\xa0\x67\xc8\x05\x08\x00\x00\xff\xff\xf7\x9a\x63\xa2\x5d\x00\x00\x00")

func dataCollectShBytes() ([]byte, error) {
	return bindataRead(
		_dataCollectSh,
		"data/collect.sh",
	)
}

func dataCollectSh() (*asset, error) {
	bytes, err := dataCollectShBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "data/collect.sh", size: 93, mode: os.FileMode(511), modTime: time.Unix(1514577098, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _dataTemplateYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\xcf\xc1\x4a\xf4\x30\x14\x05\xe0\x7d\xa1\xef\x70\xa0\xfb\x29\xfc\xc3\x8f\x90\x9d\xb4\x5d\xc8\xc8\x28\xda\x8d\x2b\x89\xe9\x2d\x13\x4c\x93\x90\xdc\x14\xba\xf1\xd9\x25\xad\x30\x8e\x6e\xc7\x65\x72\x6f\xce\x77\x52\xa1\xa5\x51\x5b\x42\xd7\x37\x2d\x94\x49\x91\x29\x80\x1d\x52\x24\x48\x3b\x20\x12\x43\x51\x60\x3d\x6a\x25\x99\x62\x59\x54\x97\xbb\x53\x8a\x8c\x98\xbc\x77\x81\xb7\xc9\xbc\x2f\x0b\x62\x35\x34\xdb\x86\x28\x0b\x80\x4d\x14\xe0\x90\x28\x1f\xaa\x35\xd5\x4b\x3e\x65\xe9\x5b\x3a\x5c\x58\x47\x64\x67\x1d\x9c\x9d\xc8\x32\x66\x19\xb4\x7c\x33\x84\x49\xda\x24\x8d\x59\x30\xba\xb0\x42\x4d\x7f\xff\xda\x1d\xdb\xc7\x87\xbb\x63\xff\x9c\x73\x81\xcc\x76\x76\xf0\x4e\x5b\x8e\x02\x27\x66\x1f\x45\x5d\xe7\xeb\xdd\x57\xe1\x9d\x76\xe2\xdf\xfe\xe6\xff\xf6\xe0\x8a\x55\x9a\xee\xa9\x3f\xb7\x50\x6c\x1a\x0a\x2c\xf0\x51\xe7\xd4\xb8\x75\xf0\x34\x5d\xdf\xbd\xfd\x2d\xcb\xb3\xab\xe4\x9f\xa8\x87\xee\xe5\x82\x3c\xd0\xf2\xe3\xaf\xef\xb4\x7c\x06\x00\x00\xff\xff\xa9\x7b\xa7\x28\x5e\x02\x00\x00")

func dataTemplateYamlBytes() ([]byte, error) {
	return bindataRead(
		_dataTemplateYaml,
		"data/template.yaml",
	)
}

func dataTemplateYaml() (*asset, error) {
	bytes, err := dataTemplateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "data/template.yaml", size: 606, mode: os.FileMode(511), modTime: time.Unix(1515087611, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"data/collect.sh":    dataCollectSh,
	"data/template.yaml": dataTemplateYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"data": &bintree{nil, map[string]*bintree{
		"collect.sh":    &bintree{dataCollectSh, map[string]*bintree{}},
		"template.yaml": &bintree{dataTemplateYaml, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
