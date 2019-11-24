package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	bolt "go.etcd.io/bbolt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)



//save picture as byte in a database
//creating a todo struct
type Todo struct {
	Id 			int `json:"Id"`
	Title		string	`json:"Title"`
	Description	string	`json:"Description"`
}


var todoInstance Todo

func getId(w http.ResponseWriter, r *http.Request) int {
	//get the id and convert it to an integer
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if id < 1 || err != nil {
		http.NotFound(w, r)
		return 0
	}
	return id
}

func openDatabase() *bolt.DB{
	//open database
	db, er := setupDatabase()
	if er != nil{
		_ = fmt.Errorf("could not use database %v", er)
	}
	return db
}



func getAllItem(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")

	//open database
	db:= openDatabase()
	defer db.Close()

	todoListt := []Todo{}

	err := db.View(func(tx *bolt.Tx) error {
		//call the database
		b := tx.Bucket([]byte("TODODB"))
		//main function that extract
		err := b.ForEach(func(k, v []byte) error {

			_ = json.Unmarshal(v, &todoInstance)
			todoListt = append(todoListt, todoInstance)

			return nil
		})

		if err != nil{
			log.Fatal(err)
		}
		return json.NewEncoder(w).Encode(todoListt)
	})

	if err != nil {
		log.Fatal(err)
	}

}

func getTodo(w http.ResponseWriter, r *http.Request){
	fmt.Println("Hit the getTodo route")

	//getId
	id := getId(w, r)

	//open database
	db:= openDatabase()
	defer db.Close()

	//retrieve todo from database
	_ = db.View(func(tx *bolt.Tx) error {
		database := tx.Bucket([]byte("TODODB"))
		getData := database.Get([]byte(string(id)))

		//check for error
		if getData == nil{
			log.Println("No User with that Id")
			return json.NewEncoder(w).Encode("Sorry, No item with that Id")
		}

		err := json.Unmarshal(getData, &todoInstance)
		if err != nil {
			log.Fatal(err)
		}


		return json.NewEncoder(w).Encode(todoInstance)
	})
	}

//Delete Todo
func deleteTodo(w http.ResponseWriter, r *http.Request){
	fmt.Println("Hit the delete route")


	//getId
	id := getId(w, r)

	//open database
	db:= openDatabase()
	defer db.Close()

	//Delete an item
	_ = db.Update(func(tx *bolt.Tx) error {
		//the main process
		database := tx.Bucket([]byte("TODODB"))
		err := database.Delete([]byte(string(id)))
		if err != nil {
			return fmt.Errorf("could not delete db: %v", err)
		}
		_, _ = fmt.Fprintf(w, "Deleted Sucessfully")
		return nil
	})
}


//Update Todo
func update(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hit the update route")
	if r.Method != http.MethodPost || r.Method != http.MethodPut{
		http.Error(w, "Access Denied",http.StatusMethodNotAllowed)
	}

	//getId
	id := getId(w, r)

	//open database
	db:= openDatabase()
	defer db.Close()

	//read or get data from body of request
	dataRaw, err := ioutil.ReadAll(r.Body)
	if err != nil{
		panic(err.Error())
	}

	//create a variable of your struct
	var to Todo
	//convert the incoming data to something of your variable
	_ = json.Unmarshal(dataRaw, &to)

	//open database
	db, er := setupDatabase()
	if er != nil{
		_ = fmt.Errorf("Could not use database %v", er)
	}

	//convert to the best []byte that can be saved
	todo, _ := json.Marshal(to)

	//update database
	err = db.Update(func(tx *bolt.Tx) error {
		//the main process
		database := tx.Bucket([]byte("TODODB"))
		err = database.Put([]byte(string(id)), []byte(todo))
		if err != nil {
			return fmt.Errorf("could not update db: %v", err)
		}
		_, _ = fmt.Fprintf(w, "Updated Sucessfully")
		return nil
	})
}



//createHandler
func create(w http.ResponseWriter, r *http.Request)  {

	fmt.Println("Hit the create route")
	if r.Method != "POST" {
		http.Error(w, "Access Denied", http.StatusMethodNotAllowed)
	}

	//open database
	db := openDatabase()
	defer db.Close()

	//read data from the body
	data, _ := ioutil.ReadAll(r.Body)
	//stringify data and print it if you are interested..return unusable bytes
	//fmt.Println("2st",string(data))

	var todoInstance Todo

	//convert data gotten from the api to an instance of todo
	err := json.Unmarshal(data, &todoInstance)
	if err != nil {
		log.Fatal(err)
	}

	//convert an instance of Todo to bytes
	//todoByte, _ :=  json.Marshal(todoInstance)
	//fmt.Fprintf(w, string(todobyte))

	//save to database
	err = db.Update(func(tx *bolt.Tx) error {

		//the main process
		openDb := tx.Bucket([]byte("TODODB"))
		id, _ := openDb.NextSequence()
		fmt.Println(id)
		todoInstance.Id = int(id)

		//convert an instance of Todo to bytes
		todoByte, _ :=  json.Marshal(todoInstance)

		err = openDb.Put([]byte(string(id)), todoByte)
		if err != nil {
			return fmt.Errorf("could not set config: %v", err)
		}

		fmt.Println("Just created Todo")
		fmt.Println(string(todoByte))
		_, _ = fmt.Fprintf(w, string(todoByte))

		return nil
	})

}

//set up database
func setupDatabase()(*bolt.DB, error){
	db, err := bolt.Open("todo.db", 0600, nil)
	if err!= nil{
		return nil, fmt.Errorf("could not open db %v", err)
	}

	//the update function is use whenever we want to create database
	err = db.Update(func(tx *bolt.Tx) error {
		//create TODODB database
		_, er := tx.CreateBucketIfNotExists([]byte("TODODB"))
		if er != nil{
			return fmt.Errorf("could not create TODODB %v", er)
		}

		return nil
	})

	if err != nil{
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}
	fmt.Println("Set Database Sucessfully")
	return db, nil
}

//creating routeHandlers and web server
func RouteHandlers(){
	//creating a new serverMux
	mux := http.NewServeMux()
	mux.HandleFunc("/all", getAllItem)
	mux.HandleFunc("/create", create)
	mux.HandleFunc("/update", update)
	mux.HandleFunc("/get", getTodo)
	mux.HandleFunc("/delete", deleteTodo)

	//creating the server
	log.Println("Starting server on :10000")
	log.Fatal(http.ListenAndServe(":7000", mux))
}



func main(){

	RouteHandlers()
}


