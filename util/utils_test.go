package util

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIndexes(t *testing.T) {
	str := "aa:aa:aab:a"
	indexes := Indexes(str, ":")
	t.Log(indexes)
	require.Equal(t, 3, len(indexes))
	require.Equal(t, 2, indexes[0])
	require.Equal(t, 5, indexes[1])
}

func TestSplitUrlParam(t *testing.T) {
	url := "peer0.org1.example.com:7050@username@127.0.0.1:22:password"

	host, port, username, ip, sshPort, password := SplitUrlParam(url)
	t.Log(host)
	t.Log(port)
	t.Log(username)
	t.Log(ip)
	t.Log(sshPort)
	t.Log(password)
}
