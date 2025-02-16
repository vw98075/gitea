// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package doctor

import (
	"fmt"
	"sort"
	"strings"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"
)

// Check represents a Doctor check
type Check struct {
	Title                      string
	Name                       string
	IsDefault                  bool
	Run                        func(logger log.Logger, autofix bool) error
	AbortIfFailed              bool
	SkipDatabaseInitialization bool
	Priority                   int
}

type wrappedLevelLogger struct {
	log.LevelLogger
}

func (w *wrappedLevelLogger) Log(skip int, level log.Level, format string, v ...interface{}) error {
	return w.LevelLogger.Log(
		skip+1,
		level,
		" - %s "+format,
		append(
			[]interface{}{
				log.NewColoredValueBytes(
					fmt.Sprintf("[%s]", strings.ToUpper(level.String()[0:1])),
					level.Color()),
			}, v...)...)
}

func initDBDisableConsole(disableConsole bool) error {
	setting.NewContext()
	setting.InitDBConfig()

	setting.NewXORMLogService(disableConsole)
	if err := db.InitEngine(); err != nil {
		return fmt.Errorf("models.SetEngine: %v", err)
	}
	return nil
}

// Checks is the list of available commands
var Checks []*Check

// RunChecks runs the doctor checks for the provided list
func RunChecks(logger log.Logger, autofix bool, checks []*Check) error {
	wrappedLogger := log.LevelLoggerLogger{
		LevelLogger: &wrappedLevelLogger{logger},
	}

	dbIsInit := false
	for i, check := range checks {
		if !dbIsInit && !check.SkipDatabaseInitialization {
			// Only open database after the most basic configuration check
			setting.EnableXORMLog = false
			if err := initDBDisableConsole(true); err != nil {
				logger.Error("Error whilst initializing the database: %v", err)
				logger.Error("Check if you are using the right config file. You can use a --config directive to specify one.")
				return nil
			}
			dbIsInit = true
		}
		logger.Info("[%d] %s", log.NewColoredIDValue(i+1), check.Title)
		logger.Flush()
		if err := check.Run(&wrappedLogger, autofix); err != nil {
			if check.AbortIfFailed {
				logger.Critical("FAIL")
				return err
			}
			logger.Error("ERROR")
		} else {
			logger.Info("OK")
			logger.Flush()
		}
	}
	return nil
}

// Register registers a command with the list
func Register(command *Check) {
	Checks = append(Checks, command)
	sort.SliceStable(Checks, func(i, j int) bool {
		if Checks[i].Priority == Checks[j].Priority {
			return Checks[i].Name < Checks[j].Name
		}
		if Checks[i].Priority == 0 {
			return false
		}
		return Checks[i].Priority < Checks[j].Priority
	})
}
