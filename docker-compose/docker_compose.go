package docker_compose

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"strings"
)

type DockerCompose struct {
	Version  string             `yaml:"version"`
	Networks map[string]string  `yaml:"networks"`
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	ContainerName string   `yaml:"container_name,omitempty"`
	Image         string   `yaml:"image,omitempty"`
	WorkDir       string   `yaml:"work_dir,omitempty"`
	Environment   []string `yaml:"environment,omitempty"`
	Command       string   `yaml:"command,omitempty"`
	Volumes       []string `yaml:"volumes,omitempty"`
	Ports         []string `yaml:"ports,omitempty"`
	dependsOn     []string `yaml:"depends_on,omitempty"`
	Networks      []string `yaml:"networks,omitempty"`
	ExtraHosts    []string `yaml:"extra_hosts,omitempty"`
}

// detectImageNameAndTag 找到包含某个关键字的docker image的tag
func detectImageNameAndTag(keyword string) (string, error) {
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
		if strings.Contains(summary.RepoTags[0], keyword) {
			return summary.RepoTags[0], nil
		}
	}

	return "", nil
}

func GenerateOrdererDockerComposeFile(filePath string, ordererUrl string, otherUrls []string) error {
	return nil
}

func GeneratePeerDockerComposeFile(filePath string, peerUrl string, otherUrls []string) error {
	imageName, err := detectImageNameAndTag("peer")
	if err != nil {
		return err
	}
	peerService := Service{
		ContainerName: "",
		Image:         imageName,
		WorkDir:       "/opt/gopath/src/github.com/hyperledger/fabric/peer",
		Environment:   nil,
		Command:       "peer node start",
		Volumes:       nil,
		Ports:         nil,
		Networks:      nil,
		ExtraHosts:    nil,
	}
	_ = peerService
	return nil
}
