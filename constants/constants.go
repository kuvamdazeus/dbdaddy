package constants

const (
	// Self
	SelfConfigDirName   = ".dbdaddy"
	SelfConfigFileName  = "config.json"
	SelfStateFileName   = "state.json"
	SelfEnvVarsFileName = ".env.vars"
	SelfDbName          = "__daddys_home"

	//  Config keys
	DbConfigOriginsKey       = "origins"
	DbConfigConnKey          = "connection"
	DbConfigShadowConnKey    = "tmp_db_connection"
	DbConfigCurrentBranchKey = "status.currentBranch"

	// ENV vars file keys
	EnvVarsDatabaseUrlKey       = "DATABASE_URL"
	EnvVarsShadowDatabaseUrlKey = "SHADOW_DATABASE_URL"
	EnvVarsOriginKeyPrefix      = "ORIGIN_"

	// Config possible driver values
	DbDriverPostgres = "postgres"
	DbDriverMySQL    = "mysql"
	DbDriverSqlite   = "sqlite"

	// dumps
	PgDumpDir     = "pg_dumps"
	MySqlDumpDir  = "mysql_dumps"
	SqliteDumpDir = "sqlite_dumps"

	// project
	ScriptsDirName = "scripts"
	SchemaDirName  = "schema"

	// tmp
	TmpDir          = "tmp"
	TextQueryOutput = "query.out"
	CSVQueryOutput  = "query.csv"

	// misc
	UpSqlScriptComment   = "--- UP SQL (APPLY THIS)"
	DownSqlScriptComment = "--- DOWN SQL (APPLY ONLY WHEN NEED TO REVERT)"
	SoftDeleteSuffix     = "__DELETED"
	ShadowDbPrefix       = "__shadow"
)

var DriverDumpDirNames = map[string]string{
	DbDriverPostgres: PgDumpDir,
	DbDriverMySQL:    MySqlDumpDir,
	DbDriverSqlite:   SqliteDumpDir,
}

var SupportedDrivers = []string{DbDriverPostgres, DbDriverMySQL}
