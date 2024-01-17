package utils

import (
	"os"
	"testing"

	"github.com/prebid/prebid-cache/constant"
)

func TestGetServerNameDefaultValues(t *testing.T) {
	expectedServerName := constant.DefaultNodeName + ":" + constant.DefaultPodName
	actualServerName := GetServerName()
	if actualServerName != expectedServerName {
		t.Errorf("Expected server name was: %v, but actual is: %v", expectedServerName, actualServerName)
	}
}

func TestGetServerNameWithSpecialCharacters(t *testing.T) {
	nodeName := "$@^$839.sfo1hy*&2p265.sfo1.pubmatic.local.9(&*@!$"
	podName := "wtrackerserver-&^@!-5cfcdc97fc-j5dlw-&^#"
	os.Setenv(constant.EnvVarNodeName, nodeName)
	os.Setenv(constant.EnvVarPodName, podName)
	expectedServerName := "$@^$839" + ":" + "&^@!-5cfcdc97fc-j5dlw-&^#"
	actualServerName := GetServerName()
	if actualServerName != expectedServerName {
		t.Errorf("Expected server name was: %v, but actual is: %v", expectedServerName, actualServerName)
	}
}
