package serverstor_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/impr0ver/uploaderGo/internal/serverstor"
	"github.com/stretchr/testify/suite"
)

type DBStorageTestSuite struct {
	suite.Suite
	DB      *serverstor.DBStorage
	TestDSN string
}

func (suite *DBStorageTestSuite) SetupSuite() {
	suite.DB = &serverstor.DBStorage{DB: nil}

	dsn := "postgresql://localhost:5432?user=postgres&password=postgres"
	dbname := "testdb"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return
	}

	db.Exec("DROP DATABASE " + dbname)
	_, err = db.Exec("CREATE DATABASE " + dbname)
	db.Close()
	if err != nil {
		return
	}

	testDSN := "postgresql://localhost:5432/" + dbname + "?user=postgres&password=postgres"
	suite.DB, _ = serverstor.ConnectDB(context.TODO(), testDSN)
}

func (suite *DBStorageTestSuite) TestDBStorageAddCounterAndGetCounter() {
	ctx := context.Background()

	tests := []struct {
		name     string
		fName    string
		fPath    string
		fSize    int64
		createAt time.Time
		want     string
	}{
		{"test#1", "myfile1", "/path1/path2/path3/myfile1", 111111, time.Now(), "/path1/path2/path3/myfile1"},
		{"test#2", "myfile 2.jpg", "/path 1/path 2/path 3/myfile 2.jpg", 2958, time.Now(), "/path 1/path 2/path 3/myfile 2.jpg"},
		{"test#3", "myfile3.pdf", "/path/myfile3.pdf", 101010, time.Now(), "/path/myfile3.pdf"},
		{"test#4", "myfile 4.txt", "/somepath/myfile 4.txt", 10, time.Now(), "/somepath/myfile 4.txt"},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.DB.AddNewFileInfo(ctx, tt.fName, tt.fPath, tt.fSize)
			suite.NoError(err, tt.name+" failed")

			res, err := suite.DB.GetAllFileInfo(ctx)
			suite.NoError(err, tt.name+", GetAllFileInfo failed")

			for _, item := range res {
				suite.Equal(item.FilePath, tt.want)

				suite.DB.DeleteFileInfoByFilePath(ctx, item.FilePath)
			}
		})
	}
}

func (suite *DBStorageTestSuite) TestDBStorageGetAllFileInfo() {
	ctx := context.Background()

	err := suite.DB.AddNewFileInfo(ctx, "myfile1", "/path1/path2/path3/myfile1", 111111)
	suite.NoError(err, "AddNewFileInfo failed")
	err = suite.DB.AddNewFileInfo(ctx, "myfile2", "/path1/path2/path3/myfile2", 222222)
	suite.NoError(err, "AddNewFileInfo failed")
	err = suite.DB.AddNewFileInfo(ctx, "myfile3", "/path1/path2/path3/myfile3", 333333)
	suite.NoError(err, "AddNewFileInfo failed")
	err = suite.DB.AddNewFileInfo(ctx, "myfile4", "/path1/path2/path3/myfile4", 444444)
	suite.NoError(err, "AddNewFileInfo failed")
	err = suite.DB.AddNewFileInfo(ctx, "myfile5", "/path1/path2/path3/myfile5", 555555)
	suite.NoError(err, "AddNewFileInfo failed")

	fileInfos, err := suite.DB.GetAllFileInfo(ctx)
	suite.NoError(err, "GetAllGauges failed")
	len := len(fileInfos)
	suite.Equal(5, len)

}

func (suite *DBStorageTestSuite) TestDBStorageRegisterNewUser() {
	ctx := context.Background()

	tests := []struct {
		name     string
		userName string
		hash     string
		want     string
	}{
		{"test#1", "Alice", "$2a$18$.yoyWQKqvmkrkmb5akkgoROqsv8Iic8pktR0x7o/5jfDjN/JYGoZbS", "Alice"},
		{"test#2", "Bob", "$2a$13$.yoyWQKq3gwwmb5akkgekmrgmkqsv8Iic8pktR0x7o/5jfDjN/JYGojJ", "Bob"},
		{"test#3", "Gopher", "$2a$16$.yoyWQKq3gwwmb5akkgoROqsv8Iic8pktR0x7o/5jfDjN/JYGooek", "Gopher"},
		{"test#4", "Noname", "$2a$14$.yoyWQKq3glrk4jkakkgoROqsv8Iic8pktR0x7o/5jfDjN/lElrIjc", "Noname"},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.DB.RegisterNewUser(ctx, tt.userName, tt.hash)
			suite.NoError(err, tt.name+" failed")

			res, err := suite.DB.GetUserByName(ctx, tt.userName)
			suite.NoError(err, tt.name+", GetUserByName failed")

			suite.Equal(res.UserName, tt.want)
		})
	}
}

func (suite *DBStorageTestSuite) TestDBStorageGetUserHash() {
	ctx := context.Background()

	tests := []struct {
		name     string
		userName string
		password string
		want     string
	}{
		{"test#1", "Alice", "$2a$18$.yoyWQKqvmkrkmb5akkgoROqsv8Iic8pktR0x7o/5jfDjN/JYGoZbS", "$2a$18$.yoyWQKqvmkrkmb5akkgoROqsv8Iic8pktR0x7o/5jfDjN/JYGoZbS"},
		{"test#2", "Bob", "$2a$13$.yoyWQKq3gwwmb5akkgekmrgmkqsv8Iic8pktR0x7o/5jfDjN/JYGojJ", "$2a$13$.yoyWQKq3gwwmb5akkgekmrgmkqsv8Iic8pktR0x7o/5jfDjN/JYGojJ"},
		{"test#3", "Gopher", "$2a$16$.yoyWQKq3gwwmb5akkgoROqsv8Iic8pktR0x7o/5jfDjN/JYGooek", "$2a$16$.yoyWQKq3gwwmb5akkgoROqsv8Iic8pktR0x7o/5jfDjN/JYGooek"},
		{"test#4", "Noname", "$2a$14$.yoyWQKq3glrk4jkakkgoROqsv8Iic8pktR0x7o/5jfDjN/lElrIjc", "$2a$14$.yoyWQKq3glrk4jkakkgoROqsv8Iic8pktR0x7o/5jfDjN/lElrIjc"},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.DB.RegisterNewUser(ctx, tt.userName, tt.password)
			suite.NoError(err, tt.name+" RegisterNewUser failed")

			res, err := suite.DB.GetUserByName(ctx, tt.userName)
			suite.NoError(err, tt.name+", GetUserByName failed")

			suite.Equal(res.Password, tt.want)
		})
	}
}

func (suite *DBStorageTestSuite) TestDBStorageDeleteFileInfoByFilePath() {
	ctx := context.Background()
	err := suite.DB.AddNewFileInfo(ctx, "myfile1", "/path1/path2/path3/myfile1", 111111)
	suite.NoError(err, "AddNewFileInfo failed")

	err = suite.DB.DeleteFileInfoByFilePath(ctx, "/path1/path2/path3/myfile1")
	suite.NoError(err, "DeleteFileInfoByFilePath failed")

	_, err = suite.DB.GetUserByName(ctx, "myfile1")
	suite.Equal(err.Error(), "sql: no rows in result set")
}

func (suite *DBStorageTestSuite) SetupTest() {
	suite.DB.DB.Exec("TRUNCATE Users, Data CASCADE;")
}

func TestDBStorageTestSuite(t *testing.T) {
	dsn := "postgresql://localhost:5432?user=postgres&password=postgres"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return
	}
	err = db.PingContext(context.TODO())
	if err != nil {
		return
	}
	db.Close()
	suite.Run(t, new(DBStorageTestSuite))
}
