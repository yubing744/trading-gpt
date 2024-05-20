package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestToHumanText tests the toHumanText method of the Thoughts struct.
func TestToHumanText(t *testing.T) {
	tests := []struct {
		name     string
		thoughts *Thoughts
		expected string
	}{
		{
			name: "All fields populated",
			thoughts: &Thoughts{
				Plan:       "Complete the project plan by end of the week.",
				Analyze:    map[string]int{"strength": 10, "weakness": 5},
				Detail:     []string{"Task 1: Research", "Task 2: Development", "Task 3: Testing"},
				Reflection: "Reflect on progress weekly.",
				Speak:      "Let's focus on the key objectives.",
			},
			expected: `Plan: Complete the project plan by end of the week.
Analyze: {"strength":10,"weakness":5}
Detail: Task 1: Research, Task 2: Development, Task 3: Testing
Reflection: Reflect on progress weekly.
Speak: Let's focus on the key objectives.
`,
		},
		{
			name: "Nil and empty fields",
			thoughts: &Thoughts{
				Plan:       nil,
				Analyze:    "",
				Detail:     []string{},
				Reflection: nil,
				Speak:      "",
			},
			expected: `Plan: none
Analyze: 
Detail: 
Reflection: none
Speak: 
`,
		},
		{
			name: "Mixed types",
			thoughts: &Thoughts{
				Plan:       12345,
				Analyze:    struct{ Field string }{"Value"},
				Detail:     []string{"Single task"},
				Reflection: nil,
				Speak:      "A statement.",
			},
			expected: `Plan: 12345
Analyze: {"Field":"Value"}
Detail: Single task
Reflection: none
Speak: A statement.
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.thoughts.ToHumanText())
		})
	}
}
