package utils

import (
	"os"
	"strings"

	"git.pubmatic.com/PubMatic/go-common/logger"
	"github.com/gofrs/uuid"
	"github.com/prebid/prebid-cache/constant"
)

// GenerateRandomID generates a "github.com/gofrs/uuid" UUID
func GenerateRandomID() (string, error) {
	u2, err := uuid.NewV4()
	return u2.String(), err
}

// GetServerName Generates server name from node and pod name in K8S environment
func GetServerName() string {
	var (
		nodeName string
		podName  string
	)

	if nodeName, _ = os.LookupEnv(constant.EnvVarNodeName); nodeName == "" {
		nodeName = constant.DefaultNodeName
		logger.Info("Node name not set. Using default name: '%s'", nodeName)
	} else {
		nodeName = strings.Split(nodeName, ".")[0]
	}

	if podName, _ = os.LookupEnv(constant.EnvVarPodName); podName == "" {
		podName = constant.DefaultPodName
		logger.Info("Pod name not set. Using default name: '%s'", podName)
	} else {
		podName = strings.TrimPrefix(podName, "creativecache-")
	}

	serverName := nodeName + ":" + podName
	logger.Info("Server name: '%s'", serverName)

	return serverName
}
