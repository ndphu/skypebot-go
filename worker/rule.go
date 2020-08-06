package worker

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
)

type Rule struct {
	Threads []string `json:"threads"`
	Users   []string `json:"users"`
	Actions []Action `json:"actions"`
	Exclude Exclude `json:"exclude"`
}

type Exclude struct {
	Users []string `json:"users"`
	Threads []string `json:"threads"`
}


type ActionType string

const ActionTypeReact = "react"

type Action struct {
	Type ActionType             `json:"type"`
	Data map[string]interface{} `json:"data"`
}

var rules []Rule
var lock sync.RWMutex

func init() {
	data, err := ioutil.ReadFile("rules.json")
	if err != nil {
		log.Println("Fail to load rules from disk")
		initSampleRules()
		return
	}
	json.Unmarshal(data, &rules)
}

func initSampleRules() {
	AddRules([]Rule{
		{
			Threads: []string{"ALL"},
			Users:   []string{"ngdacphu"},
			Actions: []Action{
				{
					Type: ActionTypeReact,
					Data: map[string]interface{}{"emotions": []string{"heart"}},
				},
			},
		},
		{
			Threads: []string{"ALL"},
			Users:   []string{"letuankhang", "huyvo1301", "tamnv3011", "nguyenngoctuan17", "nxthanhk09", "tyha.tran", "live:tienphat14", "bathach1995"},
			Actions: []Action{
				{
					Type: ActionTypeReact,
					Data: map[string]interface{}{"emotions": []string{"poop"}},
				},
			},
		},
		{
			Threads: []string{"ALL"},
			Users:   []string{"letanminhquan"},
			Actions: []Action{
				{
					Type: ActionTypeReact,
					Data: map[string]interface{}{"emotions": []string{"poop"}},
				},
			},
		},
	})
}

func IsRuleMatched(threadId, userId string) ([]Action) {
	actions := make([]Action, 0)
	for _, rule := range rules {
		if ((contains(rule.Users, "ALL") || contains(rule.Users, userId)) &&
			!contains(rule.Exclude.Users, userId)) &&
			(contains(rule.Threads, "ALL") || contains(rule.Threads, threadId)) {
			log.Println("Rule matched.")
			actions = append(actions, rule.Actions...)
		}
	}
	return actions
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func AddRule(rule Rule) (error) {
	lock.Lock()
	defer lock.Unlock()
	rules = append(rules, rule)
	return writeRulesToFile()
}

func AddRules(newRules []Rule) (error) {
	lock.Lock()
	defer lock.Unlock()
	rules = append(rules, newRules...)
	return writeRulesToFile()
}

func writeRulesToFile() error {
	if data, err := json.Marshal(rules); err != nil {
		return err
	} else {
		return ioutil.WriteFile("rules.json", data, 0755)
	}
}
