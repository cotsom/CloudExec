package secretsengine

import (
	_ "embed"
	"fmt"
	"log"
	"regexp"

	"github.com/cotsom/CloudExec/internal/utils"
	"gopkg.in/yaml.v3"
)

//go:embed rules/*.yml
var rulesYAML []byte

var compiledRules []Rule

type Rule struct {
	Name       string
	Re         *regexp.Regexp
	Confidence string
}

func LoadRules() {
	var root struct {
		Patterns []struct {
			Pattern struct {
				Name       string `yaml:"name"`
				Regex      string `yaml:"regex"`
				Confidence string `yaml:"confidence"`
			} `yaml:"pattern"`
		} `yaml:"patterns"`
	}

	if err := yaml.Unmarshal(rulesYAML, &root); err != nil {
		log.Fatalf("Failed to parse rules YAML: %v", err)
	}

	raw := root.Patterns
	compiledRules = make([]Rule, 0, len(raw)) // ← исправлено (было :=)

	for _, r := range raw {
		re := regexp.MustCompile(r.Pattern.Regex)
		compiledRules = append(compiledRules, Rule{
			Name:       r.Pattern.Name,
			Re:         re,
			Confidence: r.Pattern.Confidence,
		})
	}

	utils.Colorize(utils.ColorBlue, fmt.Sprintf("Loaded %d rules\n", len(compiledRules)))
}

func FindSecrets(text string) string {
	for _, rule := range compiledRules {
		matches := rule.Re.FindAllStringSubmatch(text, -1)
		if len(matches) > 0 {
			if len(matches[0]) > 1 {
				return matches[0][1]
			}
			return matches[0][0]
		}
	}
	return ""
}
