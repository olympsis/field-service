package field

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

/*
Connect to Database

  - Initiates connection to MongoDB

  - Grabs Enviroment Variables
*/
func (f *FieldService) ConnectToDatabase() (bool, error) {
	f.Log.Info("Connecting to Database...")

	hosts := [1]string{os.Getenv("DB_ADDR")}
	cred := options.Credential{
		AuthSource: os.Getenv("DB_NAME"),
		Username:   os.Getenv("DB_USR"),
		Password:   os.Getenv("DB_PASS"),
	}
	opts := options.ClientOptions{
		Auth:  &cred,
		Hosts: hosts[:],
	}
	cl, err := mongo.Connect(context.Background(), &opts)

	// logs connection result and sets client
	if err != nil {
		f.Status = DB_CONN
		panic("Failed to connect to Database: " + err.Error())
	} else {
		f.Client = cl
		f.Status = OK
		f.Log.Info("Database connection successful.")
		f.Collection = f.Client.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("DB_COL"))
		return true, nil
	}
}

/*
  - Pings the database to make sure we have a connection

Returns:

	bool - wether or not we have a response from database
*/
func (f *FieldService) PingDatabase() bool {
	if err := f.Client.Ping(context.TODO(), readpref.Primary()); err != nil {
		f.Status = DB_CONN
		f.Log.Error("failed to ping database")
		return false
	}
	return true
}
