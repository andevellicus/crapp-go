// assessment.go
package models

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"gopkg.in/yaml.v3"
)

// Question struct to match the YAML structure
type Question struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	MetricKey   string   `yaml:"metric_key"`
	Type        string   `yaml:"type"`
	MetricsType string   `yaml:"metrics_type"`
	Required    bool     `yaml:"required"`
	Options     []Option `yaml:"options"`
	Placeholder string   `yaml:"placeholder,omitempty"`
	MaxLength   int      `yaml:"max_length,omitempty"`
}

// Option struct for question choices
type Option struct {
	Value       string `yaml:"value"`
	Label       string `yaml:"label"`
	Description string `yaml:"description,omitempty"`
}

// Assessment struct to hold all questions
type Assessment struct {
	Questions []Question `yaml:"questions"`
}

// LoadAssessment reads and parses the questions.yaml file
func LoadAssessment(path string) (*Assessment, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read assessment file: %w", err)
	}

	var assessment Assessment
	err = yaml.Unmarshal(data, &assessment)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal assessment YAML: %w", err)
	}

	return &assessment, nil
}

// ShuffleQuestions randomizes the order of questions
func ShuffleQuestions(questions []Question) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})
}
