package docker

import (
	"errors"
	"os"
)

func address() (string, error) {
	address := os.Getenv("DOCKER_API_ADDRESS")
	if address == "" {
		return "", errors.New("请提供DOCKER REMOTE API 接口 IP、端口")
	}
	return address, nil
}
