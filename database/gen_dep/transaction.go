package gen_dep

import (
	"database/sql"
)

func GetPreparedTransaction(db *sql.DB, query string) (*sql.Stmt, *sql.Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, nil, err
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, nil, err
	}
	return stmt, tx, nil
}