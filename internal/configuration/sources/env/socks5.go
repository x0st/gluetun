package env

import (
	"fmt"

	"github.com/qdm12/gluetun/internal/configuration/settings"
	"github.com/qdm12/gosettings/sources/env"
)

func (s *Source) readSocks5() (socks5 settings.Socks5, err error) {
	socks5.Enabled, err = s.env.BoolPtr("SOCKS5")
	if err != nil {
		return socks5, err
	}

	socks5.Address, err = s.readSocks5Address()
	if err != nil {
		return socks5, err
	}
	socks5.Password = s.env.Get("SOCKS5_PASSWORD", env.ForceLowercase(false))

	return socks5, nil
}

func (s *Source) readSocks5Address() (address *string, err error) {
	const currentKey = "SOCKS5_LISTENING_ADDRESS"
	port, err := s.env.Uint16Ptr("SOCKS5_PORT") // retro-compatibility
	if err != nil {
		return nil, err
	} else if port != nil {
		s.handleDeprecatedKey("SOCKS5_PORT", currentKey)
		return ptrTo(fmt.Sprintf(":%d", *port)), nil
	}

	return s.env.Get(currentKey), nil
}
