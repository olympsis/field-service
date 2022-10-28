package field

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
Field Service Struct
*/
type FieldService struct {
	cl  *mongo.Client
	col *mongo.Collection
	log *logrus.Logger
	rtr *mux.Router
}

type Field struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Name      string             `json:"Name" bson:"Name"`
	Images    []string           `json:"images" bson:"images"`
	Longitude string             `json:"longitude" bson:"longitude"`
	Latitude  string             `json:"latitude" bson:"latitude"`
	City      string             `json:"city" bson:"city"`
	State     string             `json:"state" bson:"state"`
	Parking   string             `json:"parking" bson:"parking"`
}

/*
Create new field service struct
*/
func NewFieldService(l *logrus.Logger, r *mux.Router) *FieldService {
	return &FieldService{log: l, rtr: r}
}

/*
Connect to Database
-	Initiates connection to MongoDB
-	Grabs Enviroment Variables
*/
func (a *FieldService) ConnectToDatabase() (bool, error) {
	a.log.Info("Connecting to Database...")
	cl, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("DATABASE")))

	// logs connection result and sets client
	if err != nil {
		a.log.Error("Failed to connect to Database!")
		a.log.Error(err.Error())
		return false, err
	} else {
		a.cl = cl // set controller client to client
		a.log.Info("Database connection successful.")
		// set the collection
		a.col = a.cl.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("USER_COL"))
		return true, nil
	}
}

/*
Create Field Data (POST)

  - Creates new field for olympsis

  - Grab request body

  - Create field data in user databse

    Returns:
    Http handler

  - Writes object back to client
*/
func (a *FieldService) CreateField() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req Field

		_, c := context.WithTimeout(context.Background(), 30*time.Second)
		defer c()

		// decode request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "Bad HTTP Request": " ` + err.Error() + `" }`))
			return
		}

		field := Field{
			ID:        primitive.NewObjectID(),
			Name:      req.Name,
			Images:    req.Images,
			Longitude: req.Longitude,
			Latitude:  req.Latitude,
			City:      req.City,
			State:     req.State,
			Parking:   req.Parking,
		}

		// create auth user in database
		_, err = a.col.InsertOne(context.TODO(), field)
		if err != nil {
			a.log.WithFields(logrus.Fields{
				"handler": "Create and insert new field",
			}).Error("Failed to insert field into the database")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "Bad HTTP Request": " ` + err.Error() + `" }`))
			return
		}
		rw.WriteHeader(http.StatusCreated)
		json.NewEncoder(rw).Encode(field)
	}

}

/*
Get Field Data (Get)
-	Grab uuid from params
-	Grabs field data from database

Returns:

	Http handler
		- Writes user data back to client
*/
func (a *FieldService) GetField() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// grab uuid from query
		keys, ok := r.URL.Query()["id"]
		if !ok || len(keys[0]) < 1 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "Bad HTTP Request": "No UUID found in request." }`))
			return
		}
		id := keys[0]

		_, c := context.WithTimeout(context.Background(), 30*time.Second)
		defer c()

		// find field data in database
		var field Field
		OID, err := primitive.ObjectIDFromHex(id)
		filter := bson.D{primitive.E{Key: "_id", Value: OID}}
		err = a.col.FindOne(context.TODO(), filter).Decode(&field)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(`{ "Not Found": "Field does not exist" }`))
				return
			}
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(field)
	}
}

/*
Update Field Data (PUT)
-	Updates field data
-	Grab parameters and update

Returns:

	Http handler
		- Writes updated field back to client
*/
func (a *FieldService) UpdateField() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req Field

		_, c := context.WithTimeout(context.Background(), 30*time.Second)
		defer c()

		// decode request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "Bad HTTP Request": " ` + err.Error() + `" }`))
			return
		}

		if req.ID == primitive.NilObjectID {
			rw.WriteHeader(http.StatusBadRequest)
			a.log.Debug(req.ID)
			rw.Write([]byte(`{ "Bad HTTP Request": "Please Provide a Field ID" }`))
			return
		}

		field := Field{
			ID:        req.ID,
			Name:      req.Name,
			Images:    req.Images,
			Longitude: req.Longitude,
			Latitude:  req.Latitude,
			City:      req.City,
			State:     req.State,
			Parking:   req.Parking,
		}

		filter := bson.D{primitive.E{Key: "_id", Value: field.ID}}
		_, err = a.col.ReplaceOne(context.TODO(), filter, field)
		if err != nil {
			a.log.Debug(err.Error())
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(field)
	}
}

/*
Delete Field Data (Delete)
-	Updates field data
-	Grab parameters and update

Returns:

	Http handler
		- Writes token back to client
*/
func (a *FieldService) DeleteField() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// grab uuid from query
		keys, ok := r.URL.Query()["id"]
		if !ok || len(keys[0]) < 1 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "Bad HTTP Request": "No UUID found in request." }`))
			return
		}
		id := keys[0]

		OID, err := primitive.ObjectIDFromHex(id)

		filter := bson.D{primitive.E{Key: "_id", Value: OID}}
		_, err = a.col.DeleteOne(context.TODO(), filter)
		if err != nil {
			a.log.Debug(err.Error())
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`OK`))
	}
}

/*
Validate an Parse JWT Token
-	parse jwt token
- 	return values

Returns:

	uuid - string of the user id token
	createdAt - string of the session token created date
	role - role of user
	error -  if there is an error return error else nil
*/
func (a *FieldService) ValidateAndParseJWTToken(tokenString string) (string, string, string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("KEY")), nil
	})

	if err != nil {
		return "", "", "", err
	} else {
		uuid := claims["uuid"].(string)
		provider := claims["provider"].(string)
		createdAt := claims["createdAt"].(string)
		return uuid, provider, createdAt, nil
	}
}
