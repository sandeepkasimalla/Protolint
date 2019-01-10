package config_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/yoheimuta/protolint/internal/cmd/subcmds"
	yaml "gopkg.in/yaml.v2"

	"github.com/yoheimuta/protolint/internal/linter/config"
)

func TestExternalConfig_SkipRule(t *testing.T) {
	noDefaultExternalConfig := config.ExternalConfig{
		Lint: struct {
			Ignores []struct {
				ID    string   `yaml:"id"`
				Files []string `yaml:"files"`
			}
			Rules struct {
				NoDefault bool     `yaml:"no_default"`
				Add       []string `yaml:"add"`
				Remove    []string `yaml:"remove"`
			}
			RulesOption config.RulesOption `yaml:"rules_option"`
		}{
			Ignores: []struct {
				ID    string   `yaml:"id"`
				Files []string `yaml:"files"`
			}{
				{
					ID: "ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
					Files: []string{
						"path/to/foo.proto",
						"path/to/bar.proto",
					},
				},
				{
					ID: "ENUM_NAMES_UPPER_CAMEL_CASE",
					Files: []string{
						"path/to/foo.proto",
					},
				},
			},
			Rules: struct {
				NoDefault bool     `yaml:"no_default"`
				Add       []string `yaml:"add"`
				Remove    []string `yaml:"remove"`
			}{
				NoDefault: true,
				Add: []string{
					"FIELD_NAMES_LOWER_SNAKE_CASE",
					"MESSAGE_NAMES_UPPER_CAMEL_CASE",
				},
				Remove: []string{
					"RPC_NAMES_UPPER_CAMEL_CASE",
				},
			},
		},
	}

	defaultExternalConfig := config.ExternalConfig{
		Lint: struct {
			Ignores []struct {
				ID    string   `yaml:"id"`
				Files []string `yaml:"files"`
			}
			Rules struct {
				NoDefault bool     `yaml:"no_default"`
				Add       []string `yaml:"add"`
				Remove    []string `yaml:"remove"`
			}
			RulesOption config.RulesOption `yaml:"rules_option"`
		}{
			Ignores: []struct {
				ID    string   `yaml:"id"`
				Files []string `yaml:"files"`
			}{
				{
					ID: "ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
					Files: []string{
						"path/to/foo.proto",
						"path/to/bar.proto",
					},
				},
				{
					ID: "ENUM_NAMES_UPPER_CAMEL_CASE",
					Files: []string{
						"path/to/foo.proto",
					},
				},
			},
			Rules: struct {
				NoDefault bool     `yaml:"no_default"`
				Add       []string `yaml:"add"`
				Remove    []string `yaml:"remove"`
			}{
				NoDefault: false,
				Add: []string{
					"FIELD_NAMES_LOWER_SNAKE_CASE",
					"MESSAGE_NAMES_UPPER_CAMEL_CASE",
				},
				Remove: []string{
					"RPC_NAMES_UPPER_CAMEL_CASE",
				},
			},
		},
	}

	for _, test := range []struct {
		name                string
		externalConfig      config.ExternalConfig
		inputRuleID         string
		inputDisplayPath    string
		inputDefaultRuleIDs []string
		wantSkipRule        bool
	}{
		{
			name:             "ignore ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
			inputDisplayPath: "path/to/foo.proto",
			wantSkipRule:     true,
		},
		{
			name:             "ignore ENUM_NAMES_UPPER_CAMEL_CASE",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "ENUM_NAMES_UPPER_CAMEL_CASE",
			inputDisplayPath: "path/to/foo.proto",
			wantSkipRule:     true,
		},
		{
			name:             "not ignore FIELD_NAMES_LOWER_SNAKE_CASE",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "path/to/bar.proto",
		},
		{
			name:           "not skip Add rules",
			externalConfig: noDefaultExternalConfig,
			inputRuleID:    "FIELD_NAMES_LOWER_SNAKE_CASE",
		},
		{
			name:           "skip noAdd rules",
			externalConfig: noDefaultExternalConfig,
			inputRuleID:    "RPC_NAMES_UPPER_CAMEL_CASE",
			wantSkipRule:   true,
		},
		{
			name:           "skip Remove rule",
			externalConfig: defaultExternalConfig,
			inputRuleID:    "RPC_NAMES_UPPER_CAMEL_CASE",
			wantSkipRule:   true,
		},
		{
			name:           "not skip noRemove rule",
			externalConfig: defaultExternalConfig,
			inputRuleID:    "FIELD_NAMES_LOWER_SNAKE_CASE",
		},
		{
			name:           "not skip default rules",
			externalConfig: defaultExternalConfig,
			inputRuleID:    "HOGE_RULE",
			inputDefaultRuleIDs: []string{
				"HOGE_RULE",
			},
		},
		{
			name:                "not skip default one",
			externalConfig:      config.ExternalConfig{},
			inputRuleID:         subcmds.DefaultRuleIDs()[0],
			inputDefaultRuleIDs: subcmds.DefaultRuleIDs(),
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got := test.externalConfig.SkipRule(
				test.inputRuleID,
				test.inputDisplayPath,
				test.inputDefaultRuleIDs,
			)
			if got != test.wantSkipRule {
				t.Errorf("got %v, but want %v", got, test.wantSkipRule)
			}
		})
	}
}

func TestIndentOption_UnmarshalYAML(t *testing.T) {
	for _, test := range []struct {
		name             string
		inputConfig      []byte
		wantIndentOption config.IndentOption
		wantExistErr     bool
	}{
		{
			name: "not found supported style",
			inputConfig: []byte(`
style: 4-space
`),
			wantExistErr: true,
		},
		{
			name: "tab",
			inputConfig: []byte(`
style: tab
`),
			wantIndentOption: config.IndentOption{
				Style: "\t",
			},
		},
		{
			name: "4",
			inputConfig: []byte(`
style: 4
`),
			wantIndentOption: config.IndentOption{
				Style: strings.Repeat(" ", 4),
			},
		},
		{
			name: "2",
			inputConfig: []byte(`
style: 2
`),
			wantIndentOption: config.IndentOption{
				Style: strings.Repeat(" ", 2),
			},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var got config.IndentOption

			err := yaml.UnmarshalStrict(test.inputConfig, &got)
			if test.wantExistErr {
				if err == nil {
					t.Errorf("got err nil, but want err")
				}
				return
			}
			if err != nil {
				t.Errorf("got err %v, but want nil", err)
				return
			}

			if !reflect.DeepEqual(got, test.wantIndentOption) {
				t.Errorf("got %v, but want %v", got, test.wantIndentOption)
			}
		})
	}
}
