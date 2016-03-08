package install

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"os"
	"time"
	"io/ioutil"
	"path"
	"path/filepath"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindata_file_info struct {
	name string
	size int64
	mode os.FileMode
	modTime time.Time
}

func (fi bindata_file_info) Name() string {
	return fi.name
}
func (fi bindata_file_info) Size() int64 {
	return fi.size
}
func (fi bindata_file_info) Mode() os.FileMode {
	return fi.mode
}
func (fi bindata_file_info) ModTime() time.Time {
	return fi.modTime
}
func (fi bindata_file_info) IsDir() bool {
	return false
}
func (fi bindata_file_info) Sys() interface{} {
	return nil
}

var _templates_bootlocal_sh = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x5c\x8f\xbd\x4e\x04\x31\x0c\x84\x7b\x3f\x85\xef\xa8\x2f\x96\x10\xa2\x80\xf7\xa0\x41\x08\x65\x37\xce\x25\x22\x89\x57\x76\xc2\x22\x4e\xf7\xee\x2c\x3f\xdb\x50\x59\xb6\x35\x33\xdf\xdc\x1c\x68\xca\x8d\x2c\x01\xcc\x01\xe9\xdd\x2b\x95\x3c\xd1\x24\xd2\x6f\x83\xcc\x6f\xac\x04\x90\x23\x3e\xe3\x01\x4f\x11\x8f\x8e\x6a\x56\x15\x3d\xe2\xcb\x23\xf6\xc4\x0d\x10\xd7\x33\x77\x4c\xbd\x2f\xf6\x40\x74\xce\x3d\x8d\xc9\xcd\x52\xa9\x72\xe4\x52\x64\xb5\x3f\x0d\x29\x17\xf6\xc6\x46\x41\xd6\x56\xc4\x07\xba\x5c\xdc\x13\xab\x65\x69\xd7\xeb\x16\xdc\xc6\xc7\xab\xaf\xe1\xfe\xce\x7d\xe6\x65\x73\x1e\x6d\x9b\xf8\xff\x1e\x33\x80\x8d\x20\x38\x2f\xdf\x4c\x3b\x12\xd2\x30\xfd\x69\xf3\xbb\xc3\xfe\x38\x05\xcf\x55\x9a\xb3\x84\xd6\xbd\x76\xf8\x0a\x00\x00\xff\xff\x9f\xce\xf5\x1a\xf5\x00\x00\x00")

func templates_bootlocal_sh_bytes() ([]byte, error) {
	return bindata_read(
		_templates_bootlocal_sh,
		"templates/bootlocal.sh",
	)
}

func templates_bootlocal_sh() (*asset, error) {
	bytes, err := templates_bootlocal_sh_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "templates/bootlocal.sh", size: 245, mode: os.FileMode(493), modTime: time.Unix(1457396388, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

var _templates_mirror_daemon_sh = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x7c\x54\x5d\x6f\xe2\x3a\x10\x7d\x8e\x7f\xc5\xdc\x90\xcb\x47\x25\x1a\x68\x6f\x1f\x2e\x15\x48\x48\xa5\x57\x95\x4a\xa9\x68\xef\xbe\x6c\xbb\xab\xd4\x38\x60\x11\xec\xc8\x0e\xec\x76\x81\xff\xbe\xe3\x8f\xd0\x14\xb5\xcb\x0b\xce\xcc\x9c\x33\x33\x67\xc6\xae\xfd\x15\xbf\x70\x11\xeb\x05\xa9\x91\x1a\xc4\xac\xa0\x31\x17\xbc\x38\x9d\xc5\xab\x57\xfd\xaa\x0b\xb6\x42\xf3\xc3\xfa\xc5\x9d\x21\xe5\x19\x83\x54\x2a\x08\x57\x5c\x29\xa9\x42\xd0\x4c\x6d\x98\xb2\x68\xba\x58\x52\x29\x52\x3e\xef\xc1\xd9\xf9\x3f\x17\xf0\xef\x05\x74\x2e\x82\x66\xb7\x85\xbe\x19\xd3\x54\xf1\xbc\xe0\x52\xf4\xc0\x61\x3d\x14\x66\x09\x5b\x49\x61\x19\x72\x25\x29\xd3\x5a\x24\x2b\xd6\x73\x41\x86\xd6\x73\xda\xe2\x9c\xd5\xff\x9d\x1a\xd7\x71\x04\x56\xea\xbe\x7d\xa8\xa1\xe5\x33\x53\x38\x06\x6c\x12\x15\xab\xb5\x28\xf1\xe8\x20\xe4\x6e\x38\x1e\xf5\xcb\x7e\x48\x0f\xa2\xed\xf8\x66\x3a\x9d\x4c\xbf\xdf\x4e\xfe\xbb\xbe\xb9\x1d\xf5\xfa\x16\x96\xc9\x79\x09\xc3\xe3\xde\x06\x5e\x0d\x47\xe3\xc9\xdd\xe4\xfe\xf1\xa1\xd7\x0f\xdb\x6d\x2e\x34\xa3\x6b\xc5\xc2\x3d\xb9\xbf\xb9\x32\xd8\xfe\x87\x19\x6b\xa0\xe5\x5a\x51\x54\x72\x2d\xa8\x91\x04\x32\xfe\xa2\x12\xf5\x4a\xbe\x42\x3b\x75\x6d\x28\x8a\x33\xf0\xa3\x28\xc3\x34\x3c\x43\xbd\x0e\xa7\x7f\x88\x30\xe4\xf9\x3a\xcb\x80\x0b\x38\x28\x81\x4a\x17\x05\x17\x73\x5d\xe1\x3f\x38\x7d\x65\xef\xa8\x8f\x9d\x84\x4c\x47\x8f\x5f\x86\xb7\xfd\x0e\x21\xba\x48\x54\xd1\x6c\xc1\x96\x04\x8c\x2e\x24\xb4\x05\x84\x0f\xc6\x86\x09\x20\x32\x62\xf6\x20\x24\x81\x67\x75\xc3\x7d\x27\xd5\x1e\x06\x03\x08\x8f\x55\xde\x87\x70\x36\xa8\x77\xa1\x4e\x02\xd4\xae\x1f\x35\x73\x0d\x6d\x96\xc2\x0e\xe6\x8a\xe5\xe5\xca\xf8\x2f\xcf\xba\x83\xe4\xc7\x12\xda\xd7\x21\x84\xd0\xd8\xe6\x8a\x8b\x02\xa2\xb3\x7d\xa3\x45\x00\x78\x0a\xd8\xec\x2f\x4c\x84\x74\x21\x3c\x5f\x42\xb1\x60\x02\x1d\xe6\x67\x43\x53\x08\xff\xd6\x4f\x02\xc1\xd7\x09\xcf\x42\x74\xb1\x4c\x33\x1f\x61\x5b\xf3\xd8\x01\x44\x7e\x9c\x1f\xc3\x27\x4b\x03\x4e\x39\xd9\x13\xa2\x58\x26\x93\xd9\x91\x3c\x53\x6b\x7c\xd3\x07\xe5\x51\xcc\xea\x68\x20\xba\x90\xb9\x03\x2c\x39\x0e\x2e\x6a\xd2\xa4\x38\x64\x6c\x39\xd2\x8a\xe8\xd8\x18\x5d\x30\xba\xf4\x0d\x05\xa6\x20\x43\x61\xc6\xa7\x33\x86\xea\x74\xed\xd1\xd2\x07\xb6\x25\x1f\x63\x0d\xae\x4c\xcb\xe0\xf8\xec\x4a\x94\xe9\xdc\x16\x18\xe9\x87\xd0\x96\xe6\xe2\x94\x92\x87\xdf\x9e\xf4\xc9\x51\x6d\x91\xd1\x26\x9e\xb1\x4d\x2c\xcc\xca\x99\xf9\xb9\x86\x92\x62\xad\x2d\xbb\x49\x7c\x54\x30\x40\x45\xe1\xc6\xf8\xdd\x9a\x70\x0d\x78\x57\x04\x2a\xd5\x78\x8b\xfb\xc9\x0b\xe8\xd8\xcf\xca\x7c\x3e\xc5\x0b\x59\x7c\xc2\xd1\xb5\x9f\xbe\xff\x44\x33\x1c\x6f\x37\xc4\x7b\x42\x02\x2b\x4d\x8b\x04\x81\xd7\x28\xb8\xbc\x34\x46\x99\x3b\x9b\xcc\xbd\xc9\xcf\xa1\x62\xad\x02\xdc\xe8\x8d\xd3\x9d\xbc\x19\x2f\xd2\xac\x02\x74\x7b\x99\x82\x7f\x53\xe8\x32\xd6\xf6\x75\x8d\xed\x6a\xa0\xfe\xe5\x5c\x0f\x39\x82\x1a\x24\x1b\x89\x93\x50\x09\x65\xd6\x6e\xa7\x7c\xee\x42\x5c\x7e\xec\xaa\xac\xda\x68\xef\x7b\xc1\x93\x37\x9f\x18\x8b\xdb\xe9\xff\x75\x32\xc7\xa7\x30\xea\xc0\xd6\xa2\x77\x26\xcd\xce\x57\xb8\x73\xa5\xef\x2a\x45\xef\x1c\xd1\x1e\x97\x36\xf0\xaf\x40\x97\x30\x9d\x50\x62\x55\x8d\x9c\x8d\xfc\x0e\x00\x00\xff\xff\x5b\x0e\x1b\x9d\x49\x06\x00\x00")

func templates_mirror_daemon_sh_bytes() ([]byte, error) {
	return bindata_read(
		_templates_mirror_daemon_sh,
		"templates/mirror-daemon.sh",
	)
}

func templates_mirror_daemon_sh() (*asset, error) {
	bytes, err := templates_mirror_daemon_sh_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "templates/mirror-daemon.sh", size: 1609, mode: os.FileMode(420), modTime: time.Unix(1457396388, 0)}
	a := &asset{bytes: bytes, info:  info}
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
	"templates/bootlocal.sh": templates_bootlocal_sh,
	"templates/mirror-daemon.sh": templates_mirror_daemon_sh,
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
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() (*asset, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"templates": &_bintree_t{nil, map[string]*_bintree_t{
		"bootlocal.sh": &_bintree_t{templates_bootlocal_sh, map[string]*_bintree_t{
		}},
		"mirror-daemon.sh": &_bintree_t{templates_mirror_daemon_sh, map[string]*_bintree_t{
		}},
	}},
}}

// Restore an asset under the given directory
func RestoreAsset(dir, name string) error {
        data, err := Asset(name)
        if err != nil {
                return err
        }
        info, err := AssetInfo(name)
        if err != nil {
                return err
        }
        err = os.MkdirAll(_filePath(dir, path.Dir(name)), os.FileMode(0755))
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

// Restore assets under the given directory recursively
func RestoreAssets(dir, name string) error {
        children, err := AssetDir(name)
        if err != nil { // File
                return RestoreAsset(dir, name)
        } else { // Dir
                for _, child := range children {
                        err = RestoreAssets(dir, path.Join(name, child))
                        if err != nil {
                                return err
                        }
                }
        }
        return nil
}

func _filePath(dir, name string) string {
        cannonicalName := strings.Replace(name, "\\", "/", -1)
        return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

