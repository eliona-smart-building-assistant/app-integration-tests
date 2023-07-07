//  This file is part of the eliona project.
//  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package assert

import (
	"fmt"
	"github.com/eliona-smart-building-assistant/app-integration-tests/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func AssetTypeExists(t *testing.T, expected string, expectedAttributes []string, msgAndArgs ...any) bool {
	db, err := app.GetDB()
	require.NoError(t, err, "Connect to database")

	rows, err := db.Query(`
		SELECT *
		FROM public.asset_type
		WHERE asset_type = $1;`, expected)
	require.NoError(t, err, msgAndArgs)
	if !rows.Next() {
		return assert.Fail(t, fmt.Sprintf("Asset type %s not found", expected), msgAndArgs)
	}

	if expectedAttributes != nil {
		for _, expectedAttribute := range expectedAttributes {
			rows, err := db.Query(`
				SELECT *
				FROM public.attribute_schema
				WHERE asset_type = $1 and attribute = $2;`, expected, expectedAttribute)
			require.NoError(t, err, msgAndArgs)
			if !rows.Next() {
				return assert.Fail(t, fmt.Sprintf("Attribute %s for asset type %s not found", expectedAttribute, expected), msgAndArgs)
			}
		}
	}

	return true
}

func WidgetTypeExists(t *testing.T, expected string, msgAndArgs ...any) bool {
	db, err := app.GetDB()
	require.NoError(t, err, "Connect to database")

	rows, err := db.Query(`
		SELECT *
		FROM public.widget_type
		WHERE widget_type.name = $1;`, expected)
	require.NoError(t, err, msgAndArgs)
	if !rows.Next() {
		return assert.Fail(t, fmt.Sprintf("Widget type %s not found", expected), msgAndArgs)
	}

	return true
}

func SchemaExists(t *testing.T, expected string, expectedTables []string, msgAndArgs ...any) bool {
	db, err := app.GetDB()
	require.NoError(t, err, "Connect to database")

	rows, err := db.Query(`
		SELECT *
		FROM information_schema.schemata
		WHERE schema_name = $1;`, expected)
	require.NoError(t, err, msgAndArgs)
	if !rows.Next() {
		return assert.Fail(t, fmt.Sprintf("Schema %s not found", expected), msgAndArgs)
	}

	if expectedTables != nil {
		for _, expectedTable := range expectedTables {
			rows, err := db.Query(`
				SELECT *
				FROM information_schema.tables
				WHERE table_schema = $1 and table_name = $2;`, expected, expectedTable)
			require.NoError(t, err, msgAndArgs)
			if !rows.Next() {
				return assert.Fail(t, fmt.Sprintf("Table %s for schema %s not found", expectedTable, expected), msgAndArgs)
			}
		}
	}

	return true
}
