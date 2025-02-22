// SPDX-License-Identifier: Apache-2.0

package migrations_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xataio/pgroll/pkg/migrations"
	"github.com/xataio/pgroll/pkg/roll"
)

func TestDropColumnWithDownSQL(t *testing.T) {
	t.Parallel()

	ExecuteTests(t, TestCases{
		{
			name: "drop column",
			migrations: []migrations.Migration{
				{
					Name: "01_add_table",
					Operations: migrations.Operations{
						&migrations.OpCreateTable{
							Name: "users",
							Columns: []migrations.Column{
								{
									Name: "id",
									Type: "serial",
									Pk:   ptr(true),
								},
								{
									Name:     "name",
									Type:     "varchar(255)",
									Nullable: ptr(false),
								},
								{
									Name:     "email",
									Type:     "varchar(255)",
									Nullable: ptr(false),
								},
							},
						},
					},
				},
				{
					Name: "02_drop_column",
					Operations: migrations.Operations{
						&migrations.OpDropColumn{
							Table:  "users",
							Column: "name",
							Down:   ptr("UPPER(email)"),
						},
					},
				},
			},
			afterStart: func(t *testing.T, db *sql.DB) {
				// The deleted column is not present on the view in the new version schema.
				versionSchema := roll.VersionedSchemaName("public", "02_drop_column")
				ColumnMustNotExist(t, db, versionSchema, "users", "name")

				// But the column is still present on the underlying table.
				ColumnMustExist(t, db, "public", "users", "name")

				// Inserting into the view in the new version schema should succeed.
				MustInsert(t, db, "public", "02_drop_column", "users", map[string]string{
					"email": "foo@example.com",
				})

				// The "down" SQL has populated the removed column ("name")
				results := MustSelect(t, db, "public", "01_add_table", "users")
				assert.Equal(t, []map[string]any{
					{"id": 1, "name": "FOO@EXAMPLE.COM", "email": "foo@example.com"},
				}, results)
			},
			afterRollback: func(t *testing.T, db *sql.DB) {
				// The trigger function has been dropped.
				triggerFnName := migrations.TriggerFunctionName("users", "name")
				FunctionMustNotExist(t, db, "public", triggerFnName)

				// The trigger has been dropped.
				triggerName := migrations.TriggerName("users", "name")
				TriggerMustNotExist(t, db, "public", "users", triggerName)
			},
			afterComplete: func(t *testing.T, db *sql.DB) {
				// The column has been deleted from the underlying table.
				ColumnMustNotExist(t, db, "public", "users", "name")

				// The trigger function has been dropped.
				triggerFnName := migrations.TriggerFunctionName("users", "name")
				FunctionMustNotExist(t, db, "public", triggerFnName)

				// The trigger has been dropped.
				triggerName := migrations.TriggerName("users", "name")
				TriggerMustNotExist(t, db, "public", "users", triggerName)

				// Inserting into the view in the new version schema should succeed.
				MustInsert(t, db, "public", "02_drop_column", "users", map[string]string{
					"email": "bar@example.com",
				})
				results := MustSelect(t, db, "public", "02_drop_column", "users")
				assert.Equal(t, []map[string]any{
					{"id": 1, "email": "foo@example.com"},
					{"id": 2, "email": "bar@example.com"},
				}, results)
			},
		},
		{
			name: "can drop a column with a reserved word as its name",
			migrations: []migrations.Migration{
				{
					Name: "01_add_table",
					Operations: migrations.Operations{
						&migrations.OpCreateTable{
							Name: "users",
							Columns: []migrations.Column{
								{
									Name: "id",
									Type: "serial",
									Pk:   ptr(true),
								},
								{
									Name: "array",
									Type: "int[]",
								},
							},
						},
					},
				},
				{
					Name: "02_drop_column",
					Operations: migrations.Operations{
						&migrations.OpDropColumn{
							Table:  "users",
							Column: "array",
							Down:   ptr("UPPER(email)"),
						},
					},
				},
			},
			afterStart: func(t *testing.T, db *sql.DB) {
			},
			afterRollback: func(t *testing.T, db *sql.DB) {
			},
			afterComplete: func(t *testing.T, db *sql.DB) {
				// The column has been deleted from the underlying table.
				ColumnMustNotExist(t, db, "public", "users", "array")
			},
		},
		{
			name: "can drop a column in a table with a reserved word as its name",
			migrations: []migrations.Migration{
				{
					Name: "01_add_table",
					Operations: migrations.Operations{
						&migrations.OpCreateTable{
							Name: "array",
							Columns: []migrations.Column{
								{
									Name: "id",
									Type: "serial",
									Pk:   ptr(true),
								},
								{
									Name: "name",
									Type: "text",
								},
							},
						},
					},
				},
				{
					Name: "02_drop_column",
					Operations: migrations.Operations{
						&migrations.OpDropColumn{
							Table:  "array",
							Column: "name",
							Down:   ptr("UPPER(email)"),
						},
					},
				},
			},
			afterStart: func(t *testing.T, db *sql.DB) {
			},
			afterRollback: func(t *testing.T, db *sql.DB) {
			},
			afterComplete: func(t *testing.T, db *sql.DB) {
				// The column has been deleted from the underlying table.
				ColumnMustNotExist(t, db, "public", "users", "array")
			},
		},
	})
}
