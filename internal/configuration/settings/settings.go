package settings

import (
	"fmt"

	"github.com/qdm12/gluetun/internal/configuration/settings/helpers"
	"github.com/qdm12/gluetun/internal/constants/providers"
	"github.com/qdm12/gluetun/internal/constants/vpn"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/pprof"
	"github.com/qdm12/gotree"
)

type Settings struct {
	ControlServer ControlServer
	DNS           DNS
	Firewall      Firewall
	Health        Health
	HTTPProxy     HTTPProxy
	Log           Log
	PublicIP      PublicIP
	Shadowsocks   Shadowsocks
	Socks5        Socks5
	System        System
	Updater       Updater
	Version       Version
	VPN           VPN
	Pprof         pprof.Settings
}

type Storage interface {
	GetFilterChoices(provider string) models.FilterChoices
}

// Validate validates all the settings and returns an error
// if one of them is not valid.
// TODO v4 remove pointer for receiver (because of Surfshark).
func (s *Settings) Validate(storage Storage, ipv6Supported bool) (err error) {
	nameToValidation := map[string]func() error{
		"control server":  s.ControlServer.validate,
		"dns":             s.DNS.validate,
		"firewall":        s.Firewall.validate,
		"health":          s.Health.Validate,
		"http proxy":      s.HTTPProxy.validate,
		"log":             s.Log.validate,
		"public ip check": s.PublicIP.validate,
		"shadowsocks":     s.Shadowsocks.validate,
		"socks5":          s.Socks5.validate,
		"system":          s.System.validate,
		"updater":         s.Updater.Validate,
		"version":         s.Version.validate,
		// Pprof validation done in pprof constructor
		"VPN": func() error {
			return s.VPN.Validate(storage, ipv6Supported)
		},
	}

	for name, validation := range nameToValidation {
		err = validation()
		if err != nil {
			return fmt.Errorf("%s settings: %w", name, err)
		}
	}

	return nil
}

func (s *Settings) copy() (copied Settings) {
	return Settings{
		ControlServer: s.ControlServer.copy(),
		DNS:           s.DNS.Copy(),
		Firewall:      s.Firewall.copy(),
		Health:        s.Health.copy(),
		HTTPProxy:     s.HTTPProxy.copy(),
		Log:           s.Log.copy(),
		PublicIP:      s.PublicIP.copy(),
		Shadowsocks:   s.Shadowsocks.copy(),
		Socks5:        s.Socks5.copy(),
		System:        s.System.copy(),
		Updater:       s.Updater.copy(),
		Version:       s.Version.copy(),
		VPN:           s.VPN.Copy(),
		Pprof:         s.Pprof.Copy(),
	}
}

func (s *Settings) MergeWith(other Settings) {
	s.ControlServer.mergeWith(other.ControlServer)
	s.DNS.mergeWith(other.DNS)
	s.Firewall.mergeWith(other.Firewall)
	s.Health.MergeWith(other.Health)
	s.HTTPProxy.mergeWith(other.HTTPProxy)
	s.Log.mergeWith(other.Log)
	s.PublicIP.mergeWith(other.PublicIP)
	s.Shadowsocks.mergeWith(other.Shadowsocks)
	s.Socks5.mergeWith(other.Socks5)
	s.System.mergeWith(other.System)
	s.Updater.mergeWith(other.Updater)
	s.Version.mergeWith(other.Version)
	s.VPN.mergeWith(other.VPN)
	s.Pprof.MergeWith(other.Pprof)
}

func (s *Settings) OverrideWith(other Settings,
	storage Storage, ipv6Supported bool) (err error) {
	patchedSettings := s.copy()
	patchedSettings.ControlServer.overrideWith(other.ControlServer)
	patchedSettings.DNS.overrideWith(other.DNS)
	patchedSettings.Firewall.overrideWith(other.Firewall)
	patchedSettings.Health.OverrideWith(other.Health)
	patchedSettings.HTTPProxy.overrideWith(other.HTTPProxy)
	patchedSettings.Log.overrideWith(other.Log)
	patchedSettings.PublicIP.overrideWith(other.PublicIP)
	patchedSettings.Shadowsocks.overrideWith(other.Shadowsocks)
	patchedSettings.Socks5.overrideWith(other.Socks5)
	patchedSettings.System.overrideWith(other.System)
	patchedSettings.Updater.overrideWith(other.Updater)
	patchedSettings.Version.overrideWith(other.Version)
	patchedSettings.VPN.OverrideWith(other.VPN)
	patchedSettings.Pprof.OverrideWith(other.Pprof)
	err = patchedSettings.Validate(storage, ipv6Supported)
	if err != nil {
		return err
	}
	*s = patchedSettings
	return nil
}

func (s *Settings) SetDefaults() {
	s.ControlServer.setDefaults()
	s.DNS.setDefaults()
	s.Firewall.setDefaults()
	s.Health.SetDefaults()
	s.HTTPProxy.setDefaults()
	s.Log.setDefaults()
	s.PublicIP.setDefaults()
	s.Shadowsocks.setDefaults()
	s.Socks5.setDefaults()
	s.System.setDefaults()
	s.Version.setDefaults()
	s.VPN.setDefaults()
	s.Updater.SetDefaults(*s.VPN.Provider.Name)
	s.Pprof.SetDefaults()
}

func (s Settings) String() string {
	return s.toLinesNode().String()
}

func (s Settings) toLinesNode() (node *gotree.Node) {
	node = gotree.New("Settings summary:")

	node.AppendNode(s.VPN.toLinesNode())
	node.AppendNode(s.DNS.toLinesNode())
	node.AppendNode(s.Firewall.toLinesNode())
	node.AppendNode(s.Log.toLinesNode())
	node.AppendNode(s.Health.toLinesNode())
	node.AppendNode(s.Shadowsocks.toLinesNode())
	node.AppendNode(s.Socks5.toLinesNode())
	node.AppendNode(s.HTTPProxy.toLinesNode())
	node.AppendNode(s.ControlServer.toLinesNode())
	node.AppendNode(s.System.toLinesNode())
	node.AppendNode(s.PublicIP.toLinesNode())
	node.AppendNode(s.Updater.toLinesNode())
	node.AppendNode(s.Version.toLinesNode())
	node.AppendNode(s.Pprof.ToLinesNode())

	return node
}

func (s Settings) Warnings() (warnings []string) {
	if *s.VPN.Provider.Name == providers.HideMyAss {
		warnings = append(warnings, "HideMyAss dropped support for Linux OpenVPN "+
			" so this will likely not work anymore. See https://github.com/qdm12/gluetun/issues/1498.")
	}

	if helpers.IsOneOf(*s.VPN.Provider.Name, providers.SlickVPN) &&
		s.VPN.Type == vpn.OpenVPN {
		warnings = append(warnings, "OpenVPN 2.5 uses OpenSSL 3 "+
			"which prohibits the usage of weak security in today's standards. "+
			*s.VPN.Provider.Name+" uses weak security which is out "+
			"of Gluetun's control so the only workaround is to allow such weaknesses "+
			`using the OpenVPN option tls-cipher "DEFAULT:@SECLEVEL=0". `+
			"You might want to reach to your provider so they upgrade their certificates. "+
			"Once this is done, you will have to let the Gluetun maintainers know "+
			"by creating an issue, attaching the new certificate and we will update Gluetun.")
	}

	return warnings
}
