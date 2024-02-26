package settings

import (
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
	"github.com/qdm12/ss-server/pkg/tcpudp"
)

type Socks5 struct {
	// Enabled is true if the server should be running.
	// It defaults to false, and cannot be nil in the internal state.
	Enabled *bool
	// Settings are settings for the TCP+UDP server.
	tcpudp.Settings
}

func (s Socks5) validate() (err error) {
	return s.Settings.Validate()
}

func (s *Socks5) copy() (copied Socks5) {
	return Socks5{
		Enabled:  gosettings.CopyPointer(s.Enabled),
		Settings: s.Settings.Copy(),
	}
}

// mergeWith merges the other settings into any
// unset field of the receiver settings object.
func (s *Socks5) mergeWith(other Socks5) {
	s.Enabled = gosettings.MergeWithPointer(s.Enabled, other.Enabled)
	s.Settings = s.Settings.MergeWith(other.Settings)
}

// overrideWith overrides fields of the receiver
// settings object with any field set in the other
// settings.
func (s *Socks5) overrideWith(other Socks5) {
	s.Enabled = gosettings.OverrideWithPointer(s.Enabled, other.Enabled)
	s.Settings.OverrideWith(other.Settings)
}

func (s *Socks5) setDefaults() {
	s.Enabled = gosettings.DefaultPointer(s.Enabled, false)
	s.Settings.SetDefaults()
}

func (s Socks5) String() string {
	return s.toLinesNode().String()
}

func (s Socks5) toLinesNode() (node *gotree.Node) {
	node = gotree.New("Socks5 server settings:")

	node.Appendf("Enabled: %s", gosettings.BoolToYesNo(s.Enabled))
	if !*s.Enabled {
		return node
	}

	node.Appendf("Listening address: %s", *s.Address)
	node.Appendf("Cipher: %s", s.CipherName)
	node.Appendf("Password: %s", gosettings.ObfuscateKey(*s.Password))
	node.Appendf("Log addresses: %s", gosettings.BoolToYesNo(s.LogAddresses))

	return node
}
