// Copyright 2016 The Chihaya Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

package memory

import (
	"net"
	"testing"

	"github.com/chihaya/chihaya/server/store"

	"github.com/stretchr/testify/require"
)

var (
	v6  = net.ParseIP("0c22:384e:0:0c22:384e::68")
	v4  = net.ParseIP("12.13.14.15")
	v4s = net.ParseIP("12.13.14.15").To4()
)

func TestKey(t *testing.T) {
	var table = []struct {
		input    net.IP
		expected [16]byte
	}{
		{v6, [16]byte{12, 34, 56, 78, 0, 0, 12, 34, 56, 78, 0, 0, 0, 0, 0, 104}},
		{v4, [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 12, 13, 14, 15}},  // IPv4 in IPv6 prefix
		{v4s, [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 12, 13, 14, 15}}, // is equal to the one above, should produce equal output
	}

	for _, tt := range table {
		got := key(tt.input)
		require.Equal(t, got, tt.expected)
	}
}

func TestIPStore(t *testing.T) {
	var d = &ipStoreDriver{}

	s, err := d.New(&store.DriverConfig{})
	require.Nil(t, err)
	require.NotNil(t, s)

	// check default state
	found, err := s.HasIP(v4)
	require.Nil(t, err)
	require.False(t, found)

	// check IPv4
	err = s.AddIP(v4)
	require.Nil(t, err)

	found, err = s.HasIP(v4)
	require.Nil(t, err)
	require.True(t, found)

	found, err = s.HasIP(v4s)
	require.Nil(t, err)
	require.True(t, found)

	found, err = s.HasIP(v6)
	require.Nil(t, err)
	require.False(t, found)

	// check removes
	err = s.RemoveIP(v6)
	require.NotNil(t, err)

	err = s.RemoveIP(v4s)
	require.Nil(t, err)

	found, err = s.HasIP(v4)
	require.Nil(t, err)
	require.False(t, found)

	// check IPv6
	err = s.AddIP(v6)
	require.Nil(t, err)

	found, err = s.HasIP(v6)
	require.Nil(t, err)
	require.True(t, found)

	err = s.RemoveIP(v6)
	require.Nil(t, err)

	found, err = s.HasIP(v6)
	require.Nil(t, err)
	require.False(t, found)

	errChan := s.Stop()
	err = <-errChan
	require.Nil(t, err, "IPStore shutdown must not fail")
}

func TestHasAllHasAny(t *testing.T) {
	var d = &ipStoreDriver{}
	s, err := d.New(&store.DriverConfig{})
	require.Nil(t, err)
	require.NotNil(t, s)

	found, err := s.HasAnyIP(nil)
	require.Nil(t, err)
	require.False(t, found)

	found, err = s.HasAllIPs(nil)
	require.Nil(t, err)
	require.True(t, found)

	found, err = s.HasAllIPs([]net.IP{v4})
	require.Nil(t, err)
	require.False(t, found)

	err = s.AddIP(v4)
	require.Nil(t, err)

	found, err = s.HasAnyIP([]net.IP{v4, v6})
	require.Nil(t, err)
	require.True(t, found)

	found, err = s.HasAllIPs([]net.IP{v4, v6})
	require.Nil(t, err)
	require.False(t, found)

	found, err = s.HasAllIPs([]net.IP{v4})
	require.Nil(t, err)
	require.True(t, found)

	err = s.AddIP(v6)
	require.Nil(t, err)

	found, err = s.HasAnyIP([]net.IP{v4, v6})
	require.Nil(t, err)
	require.True(t, found)

	found, err = s.HasAllIPs([]net.IP{v4, v6})
	require.Nil(t, err)
	require.True(t, found)

	errChan := s.Stop()
	err = <-errChan
	require.Nil(t, err, "IPStore shutdown must not fail")
}

func TestNetworks(t *testing.T) {
	var (
		d          = &ipStoreDriver{}
		net1       = "192.168.22.255/24"
		net2       = "192.168.23.255/24"
		includedIP = net.ParseIP("192.168.22.23")
		excludedIP = net.ParseIP("192.168.23.22")
	)

	s, err := d.New(&store.DriverConfig{})
	require.Nil(t, err)
	require.NotNil(t, s)

	match, err := s.HasIP(includedIP)
	require.Nil(t, err)
	require.False(t, match)

	match, err = s.HasIP(excludedIP)
	require.Nil(t, err)
	require.False(t, match)

	err = s.AddNetwork("")
	require.NotNil(t, err)

	err = s.RemoveNetwork("")
	require.NotNil(t, err)

	err = s.AddNetwork(net1)
	require.Nil(t, err)

	match, err = s.HasIP(includedIP)
	require.Nil(t, err)
	require.True(t, match)

	match, err = s.HasIP(excludedIP)
	require.Nil(t, err)
	require.False(t, match)

	err = s.RemoveNetwork(net2)
	require.NotNil(t, err)

	err = s.RemoveNetwork(net1)
	require.Nil(t, err)

	match, err = s.HasIP(includedIP)
	require.Nil(t, err)
	require.False(t, match)

	match, err = s.HasIP(excludedIP)
	require.Nil(t, err)
	require.False(t, match)

	errChan := s.Stop()
	err = <-errChan
	require.Nil(t, err, "IPStore shutdown must not fail")
}

func TestHasAllHasAnyNetworks(t *testing.T) {
	var (
		d        = &ipStoreDriver{}
		net1     = "192.168.22.255/24"
		net2     = "192.168.23.255/24"
		inNet1   = net.ParseIP("192.168.22.234")
		inNet2   = net.ParseIP("192.168.23.123")
		excluded = net.ParseIP("10.154.243.22")
	)
	s, err := d.New(&store.DriverConfig{})
	require.Nil(t, err)
	require.NotNil(t, s)

	match, err := s.HasAnyIP([]net.IP{inNet1, inNet2, excluded})
	require.Nil(t, err)
	require.False(t, match)

	match, err = s.HasAllIPs([]net.IP{inNet1, inNet2, excluded})
	require.Nil(t, err)
	require.False(t, match)

	err = s.AddNetwork(net1)
	require.Nil(t, err)

	match, err = s.HasAnyIP([]net.IP{inNet1, inNet2})
	require.Nil(t, err)
	require.True(t, match)

	match, err = s.HasAllIPs([]net.IP{inNet1, inNet2})
	require.Nil(t, err)
	require.False(t, match)

	err = s.AddNetwork(net2)
	require.Nil(t, err)

	match, err = s.HasAnyIP([]net.IP{inNet1, inNet2, excluded})
	require.Nil(t, err)
	require.True(t, match)

	match, err = s.HasAllIPs([]net.IP{inNet1, inNet2})
	require.Nil(t, err)
	require.True(t, match)

	match, err = s.HasAllIPs([]net.IP{inNet1, inNet2, excluded})
	require.Nil(t, err)
	require.False(t, match)

	err = s.RemoveNetwork(net1)
	require.Nil(t, err)

	match, err = s.HasAnyIP([]net.IP{inNet1, inNet2})
	require.Nil(t, err)
	require.True(t, match)

	match, err = s.HasAllIPs([]net.IP{inNet1, inNet2})
	require.Nil(t, err)
	require.False(t, match)

	err = s.RemoveNetwork(net2)
	require.Nil(t, err)

	match, err = s.HasAnyIP([]net.IP{inNet1, inNet2})
	require.Nil(t, err)
	require.False(t, match)

	match, err = s.HasAllIPs([]net.IP{inNet1, inNet2})
	require.Nil(t, err)
	require.False(t, match)

	errChan := s.Stop()
	err = <-errChan
	require.Nil(t, err, "IPStore shutdown must not fail")
}
