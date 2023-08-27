package route
import (
	"fmt"
	"net/http"
	"strings"
) 

import (
	"github.com/Dreamacro/clash/constant"

	"github.com/Dreamacro/clash/config"
	"github.com/Dreamacro/clash/tunnel"
	"github.com/Dreamacro/clash/rules"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func ruleRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", getRules)
	r.Put("/", insertRules)
	// r.Delete("/", deleteRules)
	// r.Patch("/", changeRules)
	return r
}

type Rule struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Proxy   string `json:"proxy"`
	Size    int    `json:"size"`
}

func getRules(w http.ResponseWriter, r *http.Request) {
	rawRules := tunnel.Rules()
	rules := []Rule{}
	for _, rule := range rawRules {
		r := Rule{
			Type:    rule.RuleType().String(),
			Payload: rule.Payload(),
			Proxy:   rule.Adapter(),
			Size:    -1,
		}
		if rule.RuleType() == constant.GEOIP || rule.RuleType() == constant.GEOSITE {
			r.Size = rule.(constant.RuleGroup).GetRecodeSize()
		}
		rules = append(rules, r)

	}

	render.JSON(w, r, render.M{
		"rules": rules,
	})
}

// NOTE: no subrule support
// mostly borrowed from config.parseRules
func parseSingleRule(line string) (constant.Rule, error){
	rule := config.TrimArr(strings.Split(line, ","))
	
	var (
		payload  string
		target   string
		params   []string
		ruleName = strings.ToUpper(rule[0])
	)
	l := len(rule)

	if l < 2 {
		return nil, fmt.Errorf("error: format invalid: %s", line)
	}
	if l < 4 {
		rule = append(rule, make([]string, 4-l)...)
	}
	if ruleName == "MATCH" {
		l = 2
	}
	if l >= 3 {
		l = 3
		payload = rule[1]
	}
	target = rule[l-1]
	params = rule[l:]

	if _, ok := tunnel.Proxies()[target]; !ok {
		return nil, fmt.Errorf("[%s] error: proxy [%s] not found", line, target)
	}

	params = config.TrimArr(params)
	parsed, parseErr := rules.ParseRule(ruleName, payload, target, params, nil)
	if parseErr != nil {
		return nil, fmt.Errorf("[%s] error: %s", line, parseErr.Error())
	}

	return parsed, nil
}

func insertRules(w http.ResponseWriter, r *http.Request) {
	// Plan
	// 1. Retrieve (the list of) rules from request r
	// 2. Parse rules with config.parseRules
	// 3. Insert the `parsed` rules into tunnel.rules
	req := struct {
		Rule string `json:"rule"`
	}{}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrBadRequest)
		return
	}

	if req.Rule == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError("Rule body is empty"))
		return
	}

	parsed, parse_err := parseSingleRule(req.Rule);
	if parse_err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError(parse_err.Error()))
		return
	}

	// Insert at the front
	tunnel.InsertRule(parsed)

	render.NoContent(w, r)
}

// func deleteRules(w http.ResponseWriter, r *http.Request) {

// }