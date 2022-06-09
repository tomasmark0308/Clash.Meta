package provider

import (
	"fmt"
	"time"

	"github.com/Dreamacro/clash/adapter/provider"
	"github.com/Dreamacro/clash/common/structure"
	C "github.com/Dreamacro/clash/constant"
	P "github.com/Dreamacro/clash/constant/provider"
	RC "github.com/Dreamacro/clash/rules/common"
	"github.com/Dreamacro/clash/rules/ruleparser"
)

type ruleProviderSchema struct {
	Type     string `provider:"type"`
	Behavior string `provider:"behavior"`
	Path     string `provider:"path"`
	URL      string `provider:"url,omitempty"`
	Interval int    `provider:"interval,omitempty"`
}

func ParseRuleProvider(name string, mapping map[string]interface{}) (P.RuleProvider, error) {
	schema := &ruleProviderSchema{}
	decoder := structure.NewDecoder(structure.Option{TagName: "provider", WeaklyTypedInput: true})
	if err := decoder.Decode(mapping, schema); err != nil {
		return nil, err
	}
	var behavior P.RuleType

	switch schema.Behavior {
	case "domain":
		behavior = P.Domain
	case "ipcidr":
		behavior = P.IPCIDR
	case "classical":
		behavior = P.Classical
	default:
		return nil, fmt.Errorf("unsupported behavior type: %s", schema.Behavior)
	}

	path := C.Path.Resolve(schema.Path)
	var vehicle P.Vehicle
	switch schema.Type {
	case "file":
		vehicle = provider.NewFileVehicle(path)
	case "http":
		vehicle = provider.NewHTTPVehicle(schema.URL, path)
	default:
		return nil, fmt.Errorf("unsupported vehicle type: %s", schema.Type)
	}

	return NewRuleSetProvider(name, behavior, time.Duration(uint(schema.Interval))*time.Second, vehicle), nil
}

func parseRule(tp, payload, target string, params []string) (parsed C.Rule, parseErr error) {
	parsed, parseErr = ruleparser.ParseSameRule(tp, payload, target, params)

	if parseErr != nil {
		return nil, parseErr
	}
	ruleExtra := &C.RuleExtra{
		Network:   RC.FindNetwork(params),
		SourceIPs: RC.FindSourceIPs(params),
	}
	parsed.SetRuleExtra(ruleExtra)
	return parsed, parseErr
}