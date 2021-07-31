package utils

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

// SplitNameOrgDomain 将url拆分成节点名称,组织名称和域名
// 默认以'.'为分割符,分割后第1个元素是节点名称,第二个是组织名,
// 第二个到之后所有的内容组为域名
func SplitNameOrgDomain(url string) (string, string, string) {
	firstDotIndex := strings.Index(url, ".")
	name := url[:firstDotIndex]

	args := strings.Split(url, ".")
	orgName := args[1]
	domain := url[firstDotIndex+1:]
	return name, orgName, domain
}

// Indexes 拓展了strings.Index方法，返回子字符串的全部位置
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

// RunLocalCmd 执行本地命令
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
	if addr.IP.IsLoopback() || addr.IP.IsUnspecified() {
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
