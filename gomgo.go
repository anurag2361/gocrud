package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
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

//Data struct
type Data struct {
	Data []Example `json:"data"`
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
	var data Data
	result := Example{}
	iter := Collection.Find(nil).Iter()
	for iter.Next(&result) { //iterating over recieved objects
		example = append(example, result) //appending recieved objects to array
	}
	data.Data = example
	res.Header().Set("Content-Type", "application/json")
	exampleresult, err := json.Marshal(data)
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

//Handler function for index url
func Handler() string {
	return "Index"
}

//UploadHandler is a sample function to upload a file
func UploadHandler(res http.ResponseWriter, req *http.Request) {
	file, handle, err := req.FormFile("file")
	if err != nil {
		fmt.Fprintf(res, "%v", err)
		return
	}
	defer file.Close()
	mimeType := handle.Header.Get("Content-Type")
	switch mimeType {
	case "image/jpeg":
		saveFile(res, file, handle)
	case "image/png":
		saveFile(res, file, handle)
	default:
		jsonResponse(res, http.StatusBadRequest, "The format file is not valid.")
	}
}

func saveFile(res http.ResponseWriter, file multipart.File, handle *multipart.FileHeader) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintf(res, "%v", err)
		return
	}
	err = ioutil.WriteFile("./files/"+handle.Filename, data, 0666)
	if err != nil {
		fmt.Fprintf(res, "%v", err)
		return
	}
	jsonResponse(res, http.StatusCreated, "File Uploaded")
}

func jsonResponse(res http.ResponseWriter, code int, message string) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(code)
	fmt.Fprint(res, message)
}

func main() {
	m := martini.Classic()
	m.Get("/", Handler)
	m.Post("/createname", CreateHandler)
	m.Get("/getnames", GetHandler)
	m.Get("/getname/:id", IDHandler)
	m.Put("/updatenames/:id", UpdateHandler)
	m.Delete("/deletename/:id", DeleteHandler)
	m.Post("/upload", UploadHandler)
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
