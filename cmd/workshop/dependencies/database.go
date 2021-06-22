package dependencies

import "database/sql"

type store struct {
	db    *sql.DB
	query string
}

func DefaultDatabase(db *sql.DB) Store {
	return &store{db, `SELECT depends_on FROM dependencies WHERE id = $1`}
}

func CustomDatabase(db *sql.DB, query string) Store {
	return &store{db, query}
}

func (store *store) Direct(id string) ([]string, error) {
	rows, err := store.db.Query(store.query, id)
	if err != nil {
		return nil, err
	}

	v := []string{}
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			return nil, err
		}
		v = append(v, s)
	}

	return v, nil
}
