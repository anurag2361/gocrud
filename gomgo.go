package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-martini/martini"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var (
	Session    *mgo.Session //initializing mongo session
	Collection *mgo.Collection
	err        error
)

//Example struct
type Example struct {
	Id        bson.ObjectId `bson:"_id" json:id` //initializing mongodb id variable
	Name      string        `json:"name"`
	Surname   string        `json:"surname"`
	CreatedOn time.Time     `json:"createdon"`
}

//CreateHandler function to create document
func CreateHandler(res http.ResponseWriter, req *http.Request) {
	newexample := Example{}                             //create an empty object of type Example struct
	err = json.NewDecoder(req.Body).Decode(&newexample) //decode request body json and store it in above variable
	if err != nil {
		panic(err)
	}
	newexample.Id = bson.NewObjectId()   //assign new objectid
	newexample.CreatedOn = time.Now()    //assign timestamp
	err = Collection.Insert(&newexample) //insert object in mongodb collection
	if err != nil {
		panic(err)
	} else {
		log.Printf("Inserted info")
	}
	examplejson, err := json.Marshal(newexample) //marshal converts objects into json and unmarshal converts json into objects
	if err != nil {
		panic(err)
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(examplejson) //send stored json on response
}

//GetHandler function to retrieve document
func GetHandler(res http.ResponseWriter, req *http.Request) {
	var example []Example //assigning a new empty array to store recieved objects
	result := Example{}
	iter := Collection.Find(nil).Iter()
	for iter.Next(&result) { //iterating over recieved objects
		example = append(example, result) //appending recieved objects to array
	}
	res.Header().Set("Content-Type", "application/json")
	exampleresult, err := json.Marshal(example)
	if err != nil {
		panic(err)
	}
	res.WriteHeader(http.StatusOK)
	res.Write(exampleresult)
}

//IDHandler function to get single doc by id
func IDHandler(res http.ResponseWriter, req *http.Request, params martini.Params) {
	id := bson.ObjectIdHex(params["id"])    //reading id param
	example, object := Example{}, Example{} //assigning 2 vars, one for recieving the object and other for storing the recieved object
	result := Collection.FindId(id).Iter()  //iterating over the found object
	for result.Next(&example) {             //recieved object is stored in example
		object = example //that object is now stored to object variable
	}
	response, err := json.Marshal(object) //marshalling json to send to frontend
	if err != nil {
		panic(err)
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(response)
}

//UpdateHandler function to update document
func UpdateHandler(res http.ResponseWriter, req *http.Request, params martini.Params) {
	id := bson.ObjectIdHex(params["id"]) //storing required ObjectId recieved from url parameter to id variable
	var example Example
	err = json.NewDecoder(req.Body).Decode(&example)
	if err != nil {
		panic(err)
	}
	err = Collection.Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"name": example.Name, "surname": example.Surname}})
	if err != nil {
		panic(err)
	} else {
		log.Printf("Updated")
	}
	res.WriteHeader(http.StatusNoContent)
}

//DeleteHandler function to delete document
func DeleteHandler(res http.ResponseWriter, req *http.Request, params martini.Params) {
	err = Collection.Remove(bson.M{"_id": bson.ObjectIdHex(params["id"])})
	if err != nil {
		panic(err)
	}
	res.WriteHeader(http.StatusNoContent)
}

//MainHandler function for index url
func MainHandler() string {
	return "Index"
}

func main() {
	m := martini.Classic()
	m.Get("/", MainHandler)
	m.Post("/createname", CreateHandler)
	m.Get("/getnames", GetHandler)
	m.Get("/getname/:id", IDHandler)
	m.Put("/updatenames/:id", UpdateHandler)
	m.Delete("/deletename/:id", DeleteHandler)
	log.Println("Connecting to MongoDB")
	Session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer Session.Close()
	Session.SetMode(mgo.Monotonic, true)
	Collection = Session.DB("newgo").C("newcollection")
	m.Run()
}
