package xtemplate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRender tests the Render function with various input scenarios.
func TestRender(t *testing.T) {
	tests := []struct {
		name    string
		tmplStr string
		ctx     map[string]interface{}
		want    string
		wantErr bool
	}{
		{
			name: "Valid template with condition true",
			tmplStr: `
				{{if .Condition}}
					Condition is true!
				{{else}}
					Condition is false!
				{{end}}
				<ul>
					{{range .Items}}
						<li>{{.}}</li>
					{{end}}
				</ul>
			`,
			ctx: map[string]interface{}{
				"Condition": true,
				"Items":     []string{"Item1", "Item2", "Item3"},
			},
			want: `
					Condition is true!
				
				<ul>
					
						<li>Item1</li>
					
						<li>Item2</li>
					
						<li>Item3</li>
					
				</ul>
			`,
			wantErr: false,
		},
		{
			name: "Valid template with condition false",
			tmplStr: `
				{{if .Condition}}
					Condition is true!
				{{else}}
					Condition is false!
				{{end}}
				<ul>
					{{range .Items}}
						<li>{{.}}</li>
					{{end}}
				</ul>
			`,
			ctx: map[string]interface{}{
				"Condition": false,
				"Items":     []string{"Item1", "Item2", "Item3"},
			},
			want: `
					Condition is false!
				
				<ul>
					
						<li>Item1</li>
					
						<li>Item2</li>
					
						<li>Item3</li>
					
				</ul>
			`,
			wantErr: false,
		},
		{
			name: "Invalid template",
			tmplStr: `
				{{if .Condition}
					Condition is true!
				{{else}}
					Condition is false!
				{{end}}
				<ul>
					{{range .Items}}
						<li>{{.}}</li>
					{{end}}
				</ul>
			`,
			ctx: map[string]interface{}{
				"Condition": true,
				"Items":     []string{"Item1", "Item2", "Item3"},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmplStr, tt.ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			trimmedGot := strings.TrimSpace(got)
			trimmedWant := strings.TrimSpace(tt.want)

			assert.Equal(t, trimmedWant, trimmedGot)
		})
	}
}
