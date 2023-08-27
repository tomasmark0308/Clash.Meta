package common

import (
	"errors"
	C "github.com/Dreamacro/clash/constant"
)

var (
	errPayload = errors.New("payloadRule error")
	noResolve  = "no-resolve"
)

type Base struct {
}

func (b *Base) ShouldFindProcess() bool {
	return false
}

func (b *Base) ShouldResolveIP() bool {
	return false
}

func HasNoResolve(params []string) bool {
	for _, p := range params {
		if p == noResolve {
			return true
		}
	}
	return false
}


// function signature for easier reference
// see rules/parser.go
type ParseRuleFunc func(
	tp, payload,
	target string,
	params []string,
	subRules map[string][]C.Rule,
) (parsed C.Rule, parseErr error)