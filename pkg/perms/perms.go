package perms

import (
	"fmt"
	"strings"
)

var ErrCheckFailed = fmt.Errorf("check failed")

func getAfter(str string, after string) string {
	i := strings.Index(str, after)
	if i == -1 {
		return ""
	}
	return str[i+len(after):]
}
func getBefore(str string, before string) string {
	i := strings.Index(str, before)
	if i == -1 {
		return str
	}
	return str[:i]
}

type Manager interface {
	AddRule(actor string, relationship string, object string)
	Check(actor string, relationship string, object string) error
}

type manager struct {
	rules []Rule
}

func (p *manager) AddRule(actor string, relationship string, object string) {
	p.rules = append(p.rules, Rule{
		Actor:        actor,
		Relationship: relationship,
		Object:       object,
	})
}

func (p *manager) Check(actor string, relationship string, object string) error {
	// build parameters from actor.
	parameters := make(map[string]string)

	set := getAfter(actor, ":")
	if len(set) > 0 {
		nameid := getBefore(set, "/")
		split := strings.Split(nameid, "@")
		if len(split) > 0 {
			parameters["set_name"] = split[0]
		}
		if len(split) > 1 {
			parameters["set_id"] = split[1]
		}
	}

	for _, r := range p.rules {
		if r.ActorMatches(parameters, actor) &&
			r.RelationshipMatches(parameters, relationship) &&
			r.ObjectMatches(parameters, object) {
			return nil
		}
	}
	return ErrCheckFailed
}

func NewManager() Manager {
	return &manager{}
}
