package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBlockNumber(t *testing.T) {
	var bn BlockNumber
	err := bn.UnmarshalJSON([]byte("\"0xabcd\""))
	require.NoError(t, err)

	err = bn.UnmarshalJSON([]byte("\"0x1abf87de21\""))
	require.NoError(t, err)
	require.Equal(t, BlockNumber(114882502177), bn)
}
