package msql

import (
	"fmt"

	"github.com/fossmedaddy/dbdaddy/db/msql/msqlq"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/types"
)

func ListTablesInDb() ([]types.Table, error) {
	tables := []types.Table{}
	rows, err := globals.DB.Query(fmt.Sprintf(`select distinct table_name as name, table_schema as 'schema' from information_schema.tables
        where table_schema = '%s'
        order by table_name`, globals.CliConfig.State.CurrentBranch))

	if err != nil {
		return tables, err
	}

	for rows.Next() {
		table := types.Table{}
		if err := rows.Scan(&table.Name, &table.Schema); err != nil {
			return tables, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

func GetTableSchema(schema string, tablename string) (*types.TableSchema, error) {
	table := &types.TableSchema{
		Schema: schema,
		Name:   tablename,
	}

	rows, err := globals.DB.Query(msqlq.QGetTableSchema(schema, tablename))
	if err != nil {
		return table, err
	}

	for rows.Next() {
		column := types.Column{}
		if err := rows.Scan(
			&column.Name,
			&column.Default,
			&column.Nullable,
			&column.DataType,
			// &column.IsPrimaryKey,
			// &column.IsRelation,
			// &column.ForeignTableSchema,
			// &column.ForeignTableName,
			// &column.ForeignColumnName,
		); err != nil {
			return table, err
		}

		table.Columns = append(table.Columns, column)
	}

	return table, nil
}
