package macaroons

import (
	"time"
	"gopkg.in/macaroon-bakery.v2/bakery/checkers"
	macaroon "gopkg.in/macaroon.v2"
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
