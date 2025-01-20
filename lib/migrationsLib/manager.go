package migrationsLib

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/types"
)

// returns: list of migrations, the active migration index (-1 if no active migration found) and "isInit"
// to check if the migrations directory was initialized
func Status(currentState *types.DbSchema, usingRemoteMigDir bool) (MigrationStatus, error) {
	migStat := MigrationStatus{}

	configFilePath, _ := libUtils.FindConfigFilePath()

	var migrationsDirPath string
	if usingRemoteMigDir {
		migrationsDirPath = libUtils.GetRemoteMigrationsDir(path.Dir(configFilePath), currentState.DbName)
	} else {
		migrationsDirPath = libUtils.GetLocalMigrationsDir(path.Dir(configFilePath), currentState.DbName)
	}

	_, migDirErr := libUtils.EnsureDirExists(migrationsDirPath)
	if migDirErr != nil {
		return migStat, migDirErr
	}

	dirs, err := os.ReadDir(migrationsDirPath)
	if err != nil {
		return migStat, err
	}
	slices.SortFunc(dirs, func(a, b fs.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})

	activeI := -1
	for i, migDirEntry := range dirs {
		mig := DbMigration{
			DirPath: path.Join(migrationsDirPath, migDirEntry.Name()),
		}

		state, err := mig.ReadState()
		if err != nil {
			fmt.Println("WARNING: error occured while reading state file in", mig.DirPath)
			fmt.Println(err)
			return migStat, err
		}

		changes := DiffDBSchema(currentState, state)
		if len(changes) == 0 {
			activeI = i
		}

		if i > 0 {
			prevMig := &migStat.Migrations[i-1]
			mig.Down = prevMig
			prevMig.Up = &mig
		}

		migStat.Migrations = append(migStat.Migrations, mig)
	}

	if activeI != -1 {
		activeMig := &migStat.Migrations[activeI]
		activeMig.IsActive = true
		migStat.ActiveMigration = activeMig
	}

	migStat.IsInit = len(migStat.Migrations) == 0

	return migStat, nil
}

func ApplyMigrationSQL(migStat MigrationStatus, isUpMigration bool) error {
	if len(migStat.Migrations) == 0 {
		return fmt.Errorf("no migrations found")
	}

	if migStat.ActiveMigration == nil {
		return fmt.Errorf("no active migration found, please run 'migrations generate' to update migrations")
	}

	var sqlStr string
	if isUpMigration {
		if migStat.ActiveMigration.Up == nil {
			return fmt.Errorf("already on the latest migration, can't go into future")
		}

		upSql, err := migStat.ActiveMigration.GetUpQuery()
		if err != nil {
			return err
		}

		sqlStr = upSql
	} else {
		if migStat.ActiveMigration.Down == nil {
			return fmt.Errorf("can't go down from here, on the initial migration")
		}

		downSql, err := migStat.ActiveMigration.GetDownQuery()
		if err != nil {
			return err
		}

		sqlStr = downSql
	}

	stmts := libUtils.GetSQLStmts(sqlStr)
	if err := db_int.ExecuteStatementsTx(stmts); err != nil {
		return err
	}

	return nil
}

func GetLatestMigrationOrInit(currentState *types.DbSchema, titleIfInit string, usingRemoteDir bool) (*DbMigration, bool, error) {
	latestMig := &DbMigration{}
	isInit := false
	configFilePath, _ := libUtils.FindConfigFilePath()

	var migrationsDirPath string
	if usingRemoteDir {
		migrationsDirPath = libUtils.GetRemoteMigrationsDir(path.Dir(configFilePath), currentState.DbName)
	} else {
		migrationsDirPath = libUtils.GetLocalMigrationsDir(path.Dir(configFilePath), currentState.DbName)
	}

	_, err := libUtils.EnsureDirExists(migrationsDirPath)
	if err != nil {
		return latestMig, isInit, err
	}

	migStat, err := Status(currentState, usingRemoteDir)
	if err != nil {
		return latestMig, isInit, err
	}

	prevState := &types.DbSchema{}
	if !migStat.IsInit {
		latestMig = &migStat.Migrations[len(migStat.Migrations)-1]
		state, err := latestMig.ReadState()
		if err != nil {
			return latestMig, isInit, err
		}

		prevState = state
	} else {
		migDirId, err := libUtils.GenerateMigrationId(migrationsDirPath, titleIfInit)
		if err != nil {
			return latestMig, isInit, err
		}
		migDirPath := path.Join(migrationsDirPath, migDirId)
		initMig, err := NewDbMigration(migDirPath, prevState, "", "", "")
		if err != nil {
			return latestMig, isInit, err
		}

		latestMig = initMig
		isInit = true
	}

	return latestMig, isInit, nil
}

type GenMigOpts struct {
	Title, UpSql, DownSql, InfoStr string
	CurrentState                   *types.DbSchema
	LatestMigration                *DbMigration
	UseRemoteDir                   bool
}

func GenerateMigration(opts GenMigOpts) (*DbMigration, error) {
	var mig *DbMigration
	configDirPath, _ := libUtils.FindConfigDirPath()

	var migrationsDirPath string
	if opts.UseRemoteDir {
		migrationsDirPath = libUtils.GetRemoteMigrationsDir(configDirPath, opts.CurrentState.DbName)
	} else {
		migrationsDirPath = libUtils.GetLocalMigrationsDir(configDirPath, opts.CurrentState.DbName)
	}

	if _, err := libUtils.EnsureDirExists(migrationsDirPath); err != nil {
		return mig, err
	}

	migDirId, err := libUtils.GenerateMigrationId(migrationsDirPath, opts.Title)
	if err != nil {
		return mig, err
	}

	migDirPath := path.Join(migrationsDirPath, migDirId)
	if m, err := NewDbMigration(
		migDirPath,
		opts.CurrentState,
		"",
		opts.DownSql,
		opts.InfoStr,
	); err != nil {
		return m, err
	} else {
		mig = m
	}

	if opts.LatestMigration != nil {
		if err := opts.LatestMigration.WriteUpQuery(opts.UpSql); err != nil {
			return mig, err
		}
	}

	return mig, nil
}
