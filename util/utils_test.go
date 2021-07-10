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
