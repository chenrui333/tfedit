package migration

import (
	"bytes"
	"fmt"
	"text/template"
)

// StateMigration is a type which is equivalent to tfmigrate.StateMigratorConfig of
// minamijoyo/tfmigrate.
// The current implementation doesn't encode migration actions to a file
// directly with gohcl, so we define only what we need here.
type StateMigration struct {
	// Dir is a working directory for executing terraform command.
	Dir string
	// Actions is a list of state action.
	Actions []StateAction
}

var migrationTemplate = `migration "state" "awsv4upgrade" {
  actions = [
  {{- range .Actions }}
    "{{ .MigrationAction }}",
  {{- end }}
  ]
}
`

var compiledMigrationTemplate = template.Must(template.New("migration").Parse(migrationTemplate))

// AppendActions appends a list of actions to migration.
func (m *StateMigration) AppendActions(actions ...StateAction) {
	m.Actions = append(m.Actions, actions...)
}

// Render converts a state migration config to bytes
// Encoding StateMigratorConfig directly with gohcl has some problems.
// An array contains multiple elements is output as one line. It's not readable
// for multiple actions. In additon, the default value is set explicitly, it's
// not only redundant but also increases cognitive load for user who isn't
// familiar with tfmigrate.
// So we use text/template to render a migration file.
func (m *StateMigration) Render() ([]byte, error) {
	var output bytes.Buffer
	if err := compiledMigrationTemplate.Execute(&output, m); err != nil {
		return nil, fmt.Errorf("failed to render migration file: %s", err)
	}

	return output.Bytes(), nil
}
