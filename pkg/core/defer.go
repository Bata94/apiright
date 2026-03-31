package core

import (
	"database/sql"
	"fmt"
	"io"
	"os"
)

func Close(name string, c io.Closer, logger Logger) {
	if err := c.Close(); err != nil {
		logger.Warn(fmt.Sprintf("failed to close %s", name), "error", err)
	}
}

func Rollback(name string, tx *sql.Tx, logger Logger) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		logger.Warn(fmt.Sprintf("failed to rollback %s", name), "error", err)
	}
}

func SyncLogger(logger Logger) {
	if err := logger.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "warn: failed to sync logger: %v\n", err)
	}
}
