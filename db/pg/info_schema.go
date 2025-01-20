package pg

import (
	"fmt"
	"sync"

	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/db/pg/pgq"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/types"
)

func ListTablesInDb() ([]types.Table, error) {
	tables := []types.Table{}
	rows, err := globals.DB.Query(fmt.Sprintf(`
		select
			table_name as name,
			table_schema as schema,
			table_type as type
		from information_schema.tables

		where
			table_schema not in %s and
			table_type in ('BASE TABLE', 'VIEW')

		order by table_name
    `, constants.PgExcludedDbsSQLStr))
	if err != nil {
		return tables, err
	}

	defer rows.Close()
	for rows.Next() {
		table := types.Table{}
		if err := rows.Scan(&table.Name, &table.Schema, &table.Type); err != nil {
			return tables, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

func GetTableSchema(schema string, tablename string) (*types.TableSchema, error) {
	var tableSchema *types.TableSchema

	dbSchema, err := GetDbSchema(schema, tablename)
	if err != nil {
		return tableSchema, err
	}

	tableid := libUtils.GetTableId(schema, tablename)
	tableSchema = dbSchema.Tables[tableid]
	if tableSchema == nil {
		return tableSchema, fmt.Errorf("table with name '%s' could not be found in db schema", tableid)
	}

	return tableSchema, nil
}

// take schema & tablename as "" if need to fetch the whole db schema
// else specify schema & tablename to get schema for a specific table
func GetDbSchema(schema, tablename string) (*types.DbSchema, error) {
	var wg sync.WaitGroup

	tableid := libUtils.GetTableId(schema, tablename)

	dbSchema := &types.DbSchema{
		DbName: globals.CurrentConnConfig.Database,
	}

	tableSchemaMapping := map[string]*types.TableSchema{}
	viewSchemaMapping := map[string]*types.TableSchema{}
	schemaPresenceMapping := map[string]bool{}
	dbCons := map[string][]*types.DbConstraint{}
	dbTypes := []types.DbType{}
	dbSeqs := []types.DbSequence{}

	wg.Add(3)

	var (
		seqErr error
	)
	go (func() {
		defer wg.Done()

		rows, err := globals.DB.Query(pgq.QGetSequences())
		if err != nil {
			seqErr = err
			return
		}

		defer rows.Close()
		for rows.Next() {
			seq := types.DbSequence{}
			if err := rows.Scan(
				&seq.Schema,
				&seq.Name,
				&seq.DataType,
				&seq.IncrementBy,
				&seq.MinValue,
				&seq.MaxValue,
				&seq.StartValue,
				&seq.CacheSize,
				&seq.Cycle,
			); err != nil {
				seqErr = err
				return
			}

			dbSeqs = append(dbSeqs, seq)
		}
	})()

	var (
		typeErr error
	)
	go (func() {
		defer wg.Done()

		rows, err := globals.DB.Query(`
			select nsp.nspname, typ.typname from pg_type as typ

			inner join pg_namespace as nsp on
				typ.typnamespace = nsp.oid and
				nsp.nspname != 'information_schema' and
				nsp.nspname NOT LIKE 'pg_%'

			where
				typtype not in ('b', 'c')
		`)
		if err != nil {
			typeErr = err
			return
		}

		defer rows.Close()
		for rows.Next() {
			dbType := types.DbType{}

			err := rows.Scan(
				&dbType.Schema,
				&dbType.Name,
			)
			if err != nil {
				typeErr = err
				return
			}

			dbTypes = append(dbTypes, dbType)
		}
	})()

	var (
		conErr error
	)
	go (func() {
		defer wg.Done()

		conRows, _conErr := globals.DB.Query(pgq.QGetConstraints(tableid))
		if _conErr != nil {
			conErr = _conErr
			return
		}
		defer conRows.Close()
		for conRows.Next() {
			con := types.DbConstraint{}
			err := conRows.Scan(
				&con.ConName,
				&con.ConSchema,
				&con.Type,
				&con.UpdateActionType,
				&con.DeleteActionType,
				&con.Syntax,
				&con.TableSchema,
				&con.TableName,
				&con.ColName,
				&con.FTableSchema,
				&con.FTableName,
				&con.FColName,
				&con.MultiColConstraint,
			)
			if err != nil {
				conErr = err
				return
			}

			tableid := libUtils.GetTableId(con.TableSchema, con.TableName)

			if dbCons[tableid] == nil {
				dbCons[tableid] = []*types.DbConstraint{}
			}
			dbCons[tableid] = append(dbCons[tableid], &con)
		}
	})()

	wg.Wait()

	schemaRows, schemaErr := globals.DB.Query(pgq.QGetSchema(tableid))
	if schemaErr != nil {
		return dbSchema, schemaErr
	}
	defer schemaRows.Close()
	for schemaRows.Next() {
		var tabletype, tableschema, tablename string
		column := types.Column{}
		if err := schemaRows.Scan(
			&tabletype,
			&tableschema,
			&tablename,
			&column.Name,
			&column.Default,
			&column.Nullable,
			&column.DataType,
			&column.CharMaxLen,
			&column.NumericPrecision,
			&column.NumericScale,
		); err != nil {
			return dbSchema, schemaErr
		}

		tableid := libUtils.GetTableId(tableschema, tablename)
		column.TableSchema = tableschema
		column.TableName = tablename
		schemaPresenceMapping[tableschema] = true

		var schemaPtr *map[string]*types.TableSchema
		switch tabletype {
		case constants.TableTypeBaseTable:
			schemaPtr = &tableSchemaMapping
		case constants.TableTypeView:
			schemaPtr = &viewSchemaMapping
		}

		schema := *schemaPtr

		tableSchema := schema[tableid]
		tableCons := dbCons[tableid]
		if tableSchema == nil {
			tableSchema = &types.TableSchema{
				Schema:      tableschema,
				Name:        tablename,
				Constraints: tableCons,
			}

			schema[tableid] = tableSchema
		}

		tableSchema.Columns = append(tableSchema.Columns, column)
	}

	wg.Add(2)

	var (
		viewErr error
	)
	go (func() {
		defer wg.Done()

		rows, err := globals.DB.Query(pgq.QGetViews())
		if err != nil {
			viewErr = err
			return
		}
		defer rows.Close()
		for rows.Next() {
			var viewschema, viewname, syntax string
			err := rows.Scan(
				&viewschema,
				&viewname,
				&syntax,
			)
			if err != nil {
				viewErr = err
				return
			}

			viewSchema := viewSchemaMapping[libUtils.GetTableId(viewschema, viewname)]
			viewSchema.ViewDefSyntax = syntax
		}
	})()

	var (
		indErr error
	)
	go (func() {
		defer wg.Done()

		indexQ := pgq.QGetIndexes(tableid)
		rows, err := globals.DB.Query(indexQ)
		if err != nil {
			indErr = err
			return
		}

		defer rows.Close()
		for rows.Next() {
			ind := types.DbIndex{}
			indexrelid, indkey := 0, 0
			if err := rows.Scan(
				&indexrelid,
				&ind.Schema,
				&ind.TableName,
				&ind.Name,
				&ind.NAttributes,
				&ind.IsUnique,
				&ind.NullsNotDistinct,
				&indkey,
				&ind.Syntax,
			); err != nil {
				indErr = err
				return
			}
			_ = indexrelid
			_ = indkey

			tableid := libUtils.GetTableId(ind.Schema, ind.TableName)
			tableSchema := tableSchemaMapping[tableid]
			tableSchema.Indexes = append(tableSchema.Indexes, ind)
		}
	})()

	wg.Wait()

	if typeErr != nil {
		return dbSchema, typeErr
	}
	if indErr != nil {
		return dbSchema, indErr
	}
	if seqErr != nil {
		return dbSchema, seqErr
	}
	if conErr != nil {
		return dbSchema, typeErr
	}
	if viewErr != nil {
		return dbSchema, viewErr
	}

	schemaList := []types.Schema{}
	for schemaname := range schemaPresenceMapping {
		schemaList = append(schemaList, types.Schema{
			Name: schemaname,
		})
	}

	dbSchema.Types = dbTypes
	dbSchema.Tables = tableSchemaMapping
	dbSchema.Views = viewSchemaMapping
	dbSchema.Schemas = schemaList
	dbSchema.Sequences = dbSeqs

	return dbSchema, nil
}
