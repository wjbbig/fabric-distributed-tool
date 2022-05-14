package utils

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SplitNameOrgDomain splits url into node name, organization name and domain name.
// the default substr is '.', the first element is node name, the second is org name
// and the last are all domain name.
func SplitNameOrgDomain(url string) (string, string, string) {
	firstDotIndex := strings.Index(url, ".")
	name := url[:firstDotIndex]

	args := strings.Split(url, ".")
	orgName := args[1]
	domain := url[firstDotIndex+1:]
	return name, orgName, domain
}

// Indexes extends strings.Index，it returns all index of sub string.
func Indexes(str string, subStr string) []int {
	var indexes []int
	var lastIndex int
	for {
		if index := strings.Index(str, subStr); index != -1 {
			indexes = append(indexes, index+lastIndex)
			lastIndex += index + 1
			str = str[index+1:]
		} else {
			break
		}
	}
	return indexes
}

// RunLocalCmd runs local cmd
func RunLocalCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	var buf bytes.Buffer
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return err
	}
	if buf.Len() != 0 {
		fmt.Printf("%s\n", buf.String())
	}
	return nil
}

func SplitUrlParam(url string) (hostname string, port string, username string, ip string, sshPort string, password string) {
	args := strings.Split(url, "@")
	hostParam := strings.Split(args[0], ":")
	hostname = hostParam[0]
	port = hostParam[1]
	username = args[1]
	indexes := Indexes(args[2], ":")
	password = args[2][indexes[1]+1:]
	ip = args[2][:indexes[0]]
	sshPort = args[2][indexes[0]+1 : indexes[1]]
	return
}

func CheckLocalIp(ip string) (bool, error) {
	addr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		return false, err
	}
	outBoundIP, err := GetOutBoundIP()
	if err != nil {
		return false, err
	}
	if strings.Contains(ip, outBoundIP) || addr.IP.IsLoopback() || addr.IP.IsUnspecified() {
		return true, nil
	}

	return false, nil
}

func GetRandomPort() int {
	// 7000~60000
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(53000) + 7000
}

// DetectImageNameAndTag finds the docker image with the specified tag
func DetectImageNameAndTag(keyword string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", errors.Wrapf(err, "detect image of %s failed", keyword)
	}
	defer cli.Close()
	options := types.ImageListOptions{
		All:     false,
		Filters: filters.Args{},
	}
	imageList, err := cli.ImageList(context.Background(), options)
	if err != nil {
		return "", errors.Wrapf(err, "detect image of %s failed", keyword)
	}

	for _, summary := range imageList {
		if summary.RepoTags == nil {
			continue
		}
		if strings.Contains(summary.RepoTags[0], keyword) {
			return summary.RepoTags[0], nil
		}
	}

	return "", errors.Errorf("there is no docker image containing keyword %s", keyword)
}

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	dirPath := filepath.Dir(filename)
	exists, err := DirExists(dirPath)
	if err != nil {
		return err
	}
	if !exists {
		err = os.MkdirAll(dirPath, 0750)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(filename, data, perm)
}

func DirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return "", err
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}

func SliceContains(v string, s []string) bool {
	for _, t := range s {
		if strings.Contains(t, v) {
			return true
		}
	}
	return false
}

func DeduplicatedSlice(s []string) []string {
	var deduplicatedSlice []string
	m := make(map[string]struct{})
	for _, s2 := range s {
		m[s2] = struct{}{}
	}
	if len(s) != len(m) {
		for k := range m {
			deduplicatedSlice = append(deduplicatedSlice, k)
		}
	} else {
		deduplicatedSlice = s
	}
	return deduplicatedSlice
}
