package field

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	ID       primitive.ObjectID `json:"_id" bson:"_id"`
	Owner    string             `json:"owner" bson:"owner"`
	Name     string             `json:"name" bson:"name"`
	Notes    string             `json:"notes" bson:"notes"`
	Sports   []string           `json:"sports" bson:"sports"`
	Images   []string           `json:"images" bson:"images"`
	Location GeoJSON            `json:"location" bson:"location"`
	City     string             `json:"city" bson:"city"`
	State    string             `json:"state" bson:"state"`
	Country  string             `json:"country" bson:"country"`
	IsPublic bool               `json:"isPublic" bson:"isPublic"`
}

type GeoJSON struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

type FieldsResponse struct {
	TotalFields int     `json:"totalFields"`
	Fields      []Field `json:"fields"`
}

/*
Create new field service struct
*/
func NewFieldService(l *logrus.Logger, r *mux.Router) *FieldService {
	return &FieldService{log: l, rtr: r}
}

/*
Connect to Database

  - Initiates connection to MongoDB

  - Grabs Enviroment Variables
*/
func (f *FieldService) ConnectToDatabase() (bool, error) {
	f.log.Info("Connecting to Database...")
	cl, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("DB_URL")))

	// logs connection result and sets client
	if err != nil {
		f.log.Error("Failed to connect to Database!")
		f.log.Error(err.Error())
		return false, err
	} else {
		f.cl = cl // set controller client to client
		f.log.Info("Database connection successful.")
		// set the collection
		f.col = f.cl.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("DB_COL"))
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
func (f *FieldService) CreateField() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req Field

		_, c := context.WithTimeout(context.Background(), 30*time.Second)
		defer c()

		// decode request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "msg": " ` + err.Error() + `" }`))
			return
		}

		field := Field{
			ID:       primitive.NewObjectID(),
			Owner:    req.Owner,
			Name:     req.Name,
			Notes:    req.Notes,
			Sports:   req.Sports,
			Images:   req.Images,
			Location: req.Location,
			City:     req.City,
			State:    req.State,
			Country:  req.Country,
			IsPublic: req.IsPublic,
		}

		// create auth user in database
		_, err = f.col.InsertOne(context.TODO(), field)
		if err != nil {
			f.log.Error(err)
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "msg": " ` + err.Error() + `" }`))
			return
		}

		rw.WriteHeader(http.StatusCreated)
		json.NewEncoder(rw).Encode(field)
	}

}

/*
Get Fields  (Get)

  - Grab params and filter fields

  - Grabs field data from database

Returns:

	Http handler
		- Writes list of fields back to client
*/
func (f *FieldService) GetFields() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		longitude, _ := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
		latitude, _ := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
		radius, _ := strconv.ParseFloat(r.URL.Query().Get("radius"), 64)

		if longitude == 0 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "msg": "you need to have longitude and latitude"}`))
			return

		} else {
			if latitude == 0 {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(`{ "msg": "you need to have longitude and latitude"}`))
				return
			}
		}

		var fields []Field
		loc := GeoJSON{
			Type:        "Point",
			Coordinates: []float64{longitude, latitude},
		}

		filter := bson.D{
			{Key: "location",
				Value: bson.D{
					{Key: "$near", Value: bson.D{
						{Key: "$geometry", Value: loc},
						{Key: "$maxDistance", Value: radius},
					}},
				}},
		}

		cur, err := f.col.Find(context.Background(), filter)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(`{ "msg": "failed to search for fields" }`))
				return
			}
		}

		if cur == nil {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		for cur.Next(context.TODO()) {
			var field Field
			err := cur.Decode(&field)
			if err != nil {
				f.log.Error(err)
			}
			fields = append(fields, field)
		}

		if len(fields) == 0 {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		resp := FieldsResponse{
			TotalFields: len(fields),
			Fields:      fields,
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(resp)
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
func (f *FieldService) GetField() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// grab club id from path
		vars := mux.Vars(r)
		if len(vars["id"]) == 0 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "msg": "no field id found in request." }`))
			return
		}

		if len(vars["id"]) < 24 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "msg": "bad field id found in request." }`))
			return
		}

		id := vars["id"]
		_, c := context.WithTimeout(context.Background(), 30*time.Second)
		defer c()

		// find field data in database
		var field Field
		oid, _ := primitive.ObjectIDFromHex(id)
		filter := bson.D{primitive.E{Key: "_id", Value: oid}}
		err := f.col.FindOne(context.TODO(), filter).Decode(&field)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(`{ "msg": "field does not exist" }`))
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

  - Updates field data

  - Grab parameters and update

Returns:

	Http handler
		- Writes updated field back to client
*/
func (f *FieldService) UpdateField() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req Field

		_, c := context.WithTimeout(context.Background(), 30*time.Second)
		defer c()

		// decode request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "msg": " ` + err.Error() + `" }`))
			return
		}

		// grab club id from path
		vars := mux.Vars(r)
		if len(vars["id"]) == 0 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "msg": "no field id found in request." }`))
			return
		}

		if len(vars["id"]) < 24 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "msg": "bad field id found in request." }`))
			return
		}

		id := vars["id"]
		oid, _ := primitive.ObjectIDFromHex(id)
		filter := bson.D{primitive.E{Key: "_id", Value: oid}}
		changes := bson.M{"$set": bson.M{
			"owner":    req.Owner,
			"name":     req.Name,
			"notes":    req.Notes,
			"sports":   req.Sports,
			"images":   req.Images,
			"location": req.Location,
			"city":     req.City,
			"state":    req.State,
			"country":  req.Country,
			"isPublic": req.IsPublic,
		}}

		_, err = f.col.UpdateOne(context.TODO(), filter, changes)
		if err != nil {
			f.log.Debug(err.Error())
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`OK`))
	}
}

/*
Delete Field Data (Delete)

  - Updates field data

  - Grab parameters and update

Returns:

	Http handler
		- Writes token back to client
*/
func (f *FieldService) DeleteField() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// grab club id from path
		vars := mux.Vars(r)
		if len(vars["id"]) == 0 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "msg": "no field id found in request." }`))
			return
		}

		if len(vars["id"]) < 24 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "msg": "bad field id found in request." }`))
			return
		}

		id := vars["id"]
		oid, _ := primitive.ObjectIDFromHex(id)

		filter := bson.D{primitive.E{Key: "_id", Value: oid}}
		_, err := f.col.DeleteOne(context.TODO(), filter)
		if err != nil {
			f.log.Debug(err.Error())
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`OK`))
	}
}

/*
Validate an Parse JWT Token

  - parse jwt token

  - return values

Returns:

	uuid - string of the user id token
	createdAt - string of the session token created date
	role - role of user
	error -  if there is an error return error else nil
*/
func (a *FieldService) ValidateAndParseJWTToken(tokenString string) (string, string, float64, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("KEY")), nil
	})

	if err != nil {
		return "", "", 0, err
	} else {
		uuid := claims["uuid"].(string)
		provider := claims["provider"].(string)
		createdAt := claims["createdAt"].(float64)
		return uuid, provider, createdAt, nil
	}
}

/*
Middleware

  - Makes sure user is authenticated before taking requests

  - If there is no token or a bad token it returns the request with a unauthorized or forbidden error

Returns:

	Http handler
	- Passes the request to the next handler
*/
func (f *FieldService) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		bearerToken := r.Header.Get("Authorization")
		tokenSplit := strings.Split(bearerToken, "Bearer ")

		if bearerToken == "" {
			f.log.WithFields(logrus.Fields{
				"Middleware": "ValidateAndParseJWTToken",
			}).Error("Failed to validate token")
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := tokenSplit[1]
		if token == "" {
			f.log.WithFields(logrus.Fields{
				"Middleware": "ValidateAndParseJWTToken",
			}).Error("Failed to validate token")
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}

		_, _, _, err := f.ValidateAndParseJWTToken(token)

		if err != nil {
			f.log.WithFields(logrus.Fields{
				"Middleware": "ValidateAndParseJWTToken",
			}).Error("Failed to validate token")
			http.Error(rw, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(rw, r)
	})
}

// Later we want to ping the db and if the db goes down or something is wrong with this service we want to restart it.
func (f *FieldService) Healthz() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("ok"))
	}
}

func (f *FieldService) WhoAmi() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`
		{
			"version": "0.1",
			"service": "field"
		}
		`))
	}
}
