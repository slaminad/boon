package db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

var createTableStatements = []string{
	`CREATE DATABASE IF NOT EXISTS boon DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`,
	`USE boon;`,
	`CREATE TABLE IF NOT EXISTS reports (
		id INT UNSIGNED NOT NULL AUTO_INCREMENT,
		header VARCHAR(255) NULL,
		description TEXT NULL,
		author VARCHAR(255) NULL,
		lat INT NULL,
		lon INT NULL,
		community VARCHAR(255) NULL,
		PRIMARY KEY (id)
	)`,
	`CREATE TABLE IF NOT EXISTS communities (
		name VARCHAR(255) NULL,
		description VARCHAR(255) NULL,
		tips TEXT NULL,
		PRIMARY KEY(name)
	)`,
}

// Report type
type Report struct {
	ID          int64
	Header      string
	Description string
	Author      string
	Lat         float64
	Lon         float64
	Community   string
}

// MysqlDB persists reports to a MySQL instance.
type MysqlDB struct {
	conn *sql.DB

	list   *sql.Stmt
	insert *sql.Stmt
	get    *sql.Stmt
	update *sql.Stmt
	delete *sql.Stmt
}

// NewMySQLDB creates a new ReportDatabase backed by a given MySQL server.
func NewMySQLDB() (*MysqlDB, error) {
	// Check database and table exists. If not, create it.
	if err := ensureTableExists(); err != nil {
		return nil, err
	}

	conn, err := sql.Open("mysql", "root:password@/boon")
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("mysql: could not establish a good connection: %v", err)
	}

	db := &MysqlDB{
		conn: conn,
	}

	// Prepared statements. The actual SQL queries are in the code near the
	// relevant method (e.g. addReport).
	if db.list, err = conn.Prepare(listStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare get: %v", err)
	}
	if db.insert, err = conn.Prepare(insertStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare insert: %v", err)
	}
	if db.update, err = conn.Prepare(updateStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare update: %v", err)
	}
	if db.delete, err = conn.Prepare(deleteStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare delete: %v", err)
	}

	return db, nil
}

// Close closes the database, freeing up any resources.
func (db *MysqlDB) Close() {
	db.conn.Close()
}

// rowScanner is implemented by sql.Row and sql.Rows
type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanReport reads a report from a sql.Row or sql.Rows
func scanReport(s rowScanner) (*Report, error) {
	var (
		id          int64
		header      sql.NullString
		description sql.NullString
		author      sql.NullString
		lat         float64
		lon         float64
		community   sql.NullString
	)
	if err := s.Scan(&id, &header, &description, &author, &lat, &lon, &community); err != nil {
		return nil, err
	}

	r := &Report{
		ID:          id,
		Header:      header.String,
		Description: description.String,
		Author:      author.String,
		Lat:         lat,
		Lon:         lon,
		Community:   community.String,
	}
	return r, nil
}

const listStatement = `SELECT * FROM reports ORDER BY id`

// ListReports returns a list of reports, ordered by id.
func (db *MysqlDB) ListReports() ([]*Report, error) {
	rows, err := db.list.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*Report
	for rows.Next() {
		r, err := scanReport(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		reports = append(reports, r)
	}

	return reports, nil
}

const insertStatement = `
  INSERT INTO reports (
    header, description, author, lat, lon, community
  ) VALUES (?, ?, ?, ?, ?)`

// AddReport saves a given report, assigning it a new ID.
func (db *MysqlDB) AddReport(rep *Report) (id int64, err error) {
	r, err := execAffectingOneRow(db.insert, rep.Header, rep.Description, rep.Author,
		rep.Lat, rep.Lon, rep.Community)
	if err != nil {
		return 0, err
	}

	lastInsertID, err := r.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("mysql: could not get last insert ID: %v", err)
	}
	return lastInsertID, nil
}

const deleteStatement = `DELETE FROM reports WHERE id = ?`

// DeleteReport removes a given report by its ID.
func (db *MysqlDB) DeleteReport(id int64) error {
	if id == 0 {
		return errors.New("mysql: report with unassigned ID passed into deleteReport")
	}
	_, err := execAffectingOneRow(db.delete, id)
	return err
}

const updateStatement = `
  UPDATE reports
  SET id=?, header=?, description=?, author=?, lat=?, lon=?, community=?
  WHERE id = ?`

// UpdateReport updates the entry for a given report.
func (db *MysqlDB) UpdateReport(r *Report) error {
	if r.ID == 0 {
		return errors.New("mysql: report with unassigned ID passed into updateReport")
	}

	_, err := execAffectingOneRow(db.insert, r.Header, r.Description, r.Author,
		r.Lat, r.Lon, r.Community, r.ID)
	return err
}

// ensureTableExists checks the table exists. If not, it creates it.
func ensureTableExists() error {
	conn, err := sql.Open("mysql", "root:password@/boon")
	if err != nil {
		return fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	defer conn.Close()

	// Check the connection.
	if conn.Ping() == driver.ErrBadConn {
		return fmt.Errorf("mysql: could not connect to the database. " +
			"could be bad address, or this address is not whitelisted for access.")
	}

	if _, err := conn.Exec("USE boon"); err != nil {
		// MySQL error 1049 is "database does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1049 {
			return createTable(conn)
		}
	}

	if _, err := conn.Exec("DESCRIBE reports"); err != nil {
		// MySQL error 1146 is "table does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1146 {
			return createTable(conn)
		}
		// Unknown error.
		return fmt.Errorf("mysql: could not connect to the database: %v", err)
	}
	return nil
}

// createTable creates the table, and if necessary, the database.
func createTable(conn *sql.DB) error {
	for _, stmt := range createTableStatements {
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// execAffectingOneRow executes a given statement, expecting one row to be affected.
func execAffectingOneRow(stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	r, err := stmt.Exec(args...)
	if err != nil {
		return r, fmt.Errorf("mysql: could not execute statement: %v", err)
	}
	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return r, fmt.Errorf("mysql: could not get rows affected: %v", err)
	} else if rowsAffected != 1 {
		return r, fmt.Errorf("mysql: expected 1 row affected, got %d", rowsAffected)
	}
	return r, nil
}
