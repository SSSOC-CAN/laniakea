/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10

Copyright (C) 2015-2018 Lightning Labs and The Lightning Network Developers

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package macaroons

import (
	"strings"
	"time"

	"github.com/SSSOC-CAN/laniakea/errors"
	"github.com/SSSOC-CAN/laniakea/utils"
	bg "github.com/SSSOCPaulCote/blunderguard"
	"gopkg.in/macaroon-bakery.v2/bakery/checkers"
	macaroon "gopkg.in/macaroon.v2"
)

const (
	ErrCantGetPeerFromContext   = bg.Error("unable to get peer info from context")
	ErrInvalidListOfPlugins     = bg.Error("invalid list of plugins")
	ErrUnauthorizedPluginAction = bg.Error("unauthorized plugin action")
	ErrNoMacaroonsFromContext   = bg.Error("no macaroons received from context")
)

type Constraint func(*macaroon.Macaroon) error

type Checker func() (string, checkers.Func)

// AddConstraints returns new derived macaroon by applying every passed
// constraint and tightening its restrictions.
func AddConstraints(mac *macaroon.Macaroon, cs ...Constraint) (*macaroon.Macaroon, error) {
	newMac := mac.Clone()
	for _, constraint := range cs {
		if err := constraint(newMac); err != nil {
			return nil, err
		}
	}
	return newMac, nil
}

// Each *Constraint function is a functional option, which takes a pointer
// to the macaroon and adds another restriction to it. For each *Constraint,
// the corresponding *Checker is provided if not provided by default.

// TimeoutConstraint restricts the lifetime of the macaroon
// to the amount of seconds given.
func TimeoutConstraint(seconds int64) func(*macaroon.Macaroon) error {
	return func(mac *macaroon.Macaroon) error {
		caveat := TimeoutCaveat(seconds)
		return mac.AddFirstPartyCaveat([]byte(caveat.Condition))
	}
}

// TimeoutCaveat is a wrapper function which returns a checkers.Caveat struct
func TimeoutCaveat(seconds int64) checkers.Caveat {
	macaroonTimeout := time.Duration(seconds)
	requestTimeout := time.Now().Add(time.Second * macaroonTimeout)
	return checkers.TimeBeforeCaveat(requestTimeout)
}

// PluginConstraint locks a macaroon to a given set of plugins. The plugin names are validated but not checked against currently registered list of plugins
func PluginConstraint(pluginNames []string) func(*macaroon.Macaroon) error {
	return func(mac *macaroon.Macaroon) error {
		if len(pluginNames) != 0 {
			for _, plug := range pluginNames {
				if ok := utils.ValidatePluginName(plug); !ok {
					return errors.ErrInvalidPluginName
				}
			}
			caveat := checkers.Condition("plugins", strings.Join(pluginNames, ":"))
			return mac.AddFirstPartyCaveat([]byte(caveat))
		}
		return nil
	}
}

// PluginCaveat is a wrapper function which returns a checkers.Caveat struct
func PluginCaveat(pluginNames []string) checkers.Caveat {
	if len(pluginNames) == 1 {
		return checkers.DeclaredCaveat("plugins", pluginNames[0])
	}
	return checkers.DeclaredCaveat("plugins", strings.Join(pluginNames, ":"))
}
