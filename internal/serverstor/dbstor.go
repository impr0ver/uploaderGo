package serverstor

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	DB *sql.DB
}

type DBData struct {
	FileName string `json:"name"`
	FilePath string `json:"path"`
	FileSize int64  `json:"size"`
}

type DBUser struct {
	UserName string `json:"name"`
	Password string `json:"path"`
}

// DB init
func ConnectDB(ctx context.Context, dsn string) (*DBStorage, error) {
	dbs := &DBStorage{}

	if err := checkDSN(dsn); err != nil {
		return dbs, fmt.Errorf("wrong DSN: %w", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return dbs, fmt.Errorf("unable connect to db: %w", err)
	}

	dbs.DB = db

	err = dbs.DB.PingContext(ctx)
	if err != nil {
		return dbs, err
	}

	err = createTables(ctx, dbs)

	return dbs, err
}

func checkDSN(dsn string) (err error) {
	_, err = pgx.ParseDSN(dsn)
	return err
}

func createTables(ctx context.Context, d *DBStorage) (err error) {
	const (
		tableUsers = `CREATE TABLE IF NOT EXISTS Users (id SERIAL PRIMARY KEY, 
			username VARCHAR(32) UNIQUE NOT NULL, 
			password VARCHAR(512) NOT NULL,
			created_at TIMESTAMP NOT NULL);`

		tableData = `CREATE TABLE IF NOT EXISTS Data (id SERIAL PRIMARY KEY, 
				filename VARCHAR(128) NOT NULL, 
				filepath VARCHAR(512) UNIQUE NOT NULL,
				filesize BIGINT NOT NULL,
				created_at TIMESTAMP NOT NULL);`
	)

	if _, err = d.DB.ExecContext(ctx, tableUsers); err != nil {
		return fmt.Errorf("error create table \"Users\": %w", err)
	}
	if _, err = d.DB.ExecContext(ctx, tableData); err != nil {
		return fmt.Errorf("error create table \"Data\": %w", err)
	}
	return nil
}

func (d *DBStorage) AddNewFileInfo(ctx context.Context, fileName string, filePath string, fileSize int64) error {
	_, err := d.DB.ExecContext(ctx, `INSERT INTO Data (filename, filepath, filesize, created_at) VALUES ($1, $2, $3, $4);`, fileName, filePath, fileSize, time.Now())
	return err
}

func (d *DBStorage) DeleteFileInfoByFilePath(ctx context.Context, filePath string) error {
	_, err := d.DB.ExecContext(ctx, `DELETE FROM data WHERE filepath = $1;`, filePath)
	return err
}

func (d *DBStorage) RegisterNewUser(ctx context.Context, userName string, hash string) error {
	_, err := d.DB.ExecContext(ctx, `INSERT INTO users (username, password, created_at) VALUES ($1, $2, $3);`, userName, hash, time.Now())
	return err
}

func (d *DBStorage) GetUserByName(ctx context.Context, userName string) (DBUser, error) {
	var user DBUser
	err := d.DB.QueryRow(`SELECT username, password FROM Users WHERE username = $1;`, userName).Scan(&user.UserName, &user.Password)
	return user, err
}

func (d *DBStorage) GetAllFileInfo(ctx context.Context) ([]DBData, error) {
	var dbDatas []DBData
	var dbData DBData
	selectQuery := `SELECT filename, filepath, filesize FROM Data ORDER BY filename asc;`
	rows, err := d.DB.QueryContext(ctx, selectQuery)
	if err != nil {
		return dbDatas, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&dbData.FileName, &dbData.FilePath, &dbData.FileSize)
		if err != nil {
			return dbDatas, err
		}
		dbDatas = append(dbDatas, dbData)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return dbDatas, nil
}
