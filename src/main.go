package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type ToDoItem struct {
	ID          int    `json:"ID"`
	Status      bool   `json:"Status"`
	Description string `json:"Description"`
	Date        int64  `json:"Date"`
}

type ErrorResponse struct {
	ErrorMsg string `json:"ErrorMsg"`
}

var ToDoItemList []ToDoItem

const StartMessage = "Rest API v0.1 - To Do Items"

func main() {
	logger("API working!")
	ToDoItemList = []ToDoItem{
		{ID: 1, Status: false, Description: "Description1", Date: time.Now().Unix()},
		{ID: 2, Status: true, Description: "Description2", Date: time.Now().Unix()},
		{ID: 3, Status: true, Description: "Description3", Date: time.Now().Unix()},
	}
	handleRequests()
}

func handleRequests() {
	mainRouter := mux.NewRouter()
	mainRouter.HandleFunc("/", version).Methods("GET")

	eventRouter := mainRouter.PathPrefix("/event").Subrouter()
	eventRouter.HandleFunc("/all", getAllEvents).Methods("GET")
	eventRouter.HandleFunc("/{id:[0-9]+}", getEventById).Methods("GET")
	eventRouter.HandleFunc("/add", addNewEvent).Methods("POST")
	eventRouter.HandleFunc("/update", updateEvent).Methods("POST")

	http.Handle("/", mainRouter)
	log.Fatal(http.ListenAndServe(":11111", nil))
}

func logger(msg string) {
	fmt.Println("[log] - " + msg)
}

func GetLastId() (lastId int) {
	arrLenght := len(ToDoItemList)
	lastId = 0
	if arrLenght > 0 {
		lastId = ToDoItemList[0].ID
	}
	for i := 1; i < arrLenght; i++ {
		if ToDoItemList[i].ID > lastId {
			lastId = ToDoItemList[i].ID
		}
	}
	lastId += 1
	return
}

func CheckJsonFields(jsonData []byte, checkIdField bool) (isValid bool, missingFields []string) {
	var item *ToDoItem = &ToDoItem{}
	isValid = true
	var jsonField map[string]interface{}
	err := json.Unmarshal(jsonData, &jsonField)

	if err != nil {
		logger(err.Error())
		return
	}
	structFields := reflect.ValueOf(item).Elem().Type()

	for i := 0; i < structFields.NumField(); i++ {
		field := structFields.Field(i)
		if !checkIdField {
			if field.Name == "ID" {
				continue
			}
		}
		_, isExist := jsonField[field.Name]

		if !isExist {
			missingFields = append(missingFields, field.Name)
			isValid = false
			continue
		}
	}
	return
}

func version(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, StartMessage)
	logger(StartMessage)
}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	logger("Get all events.")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ToDoItemList)
}

func getEventById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	variables := mux.Vars(r)
	Id, err := strconv.Atoi(variables["id"])
	if err != nil {
		logger(err.Error())
		fmt.Fprintf(w, "%+v", http.StatusBadRequest)
		return
	}
	logger("Get event by id: " + strconv.Itoa(Id))
	for _, employee := range ToDoItemList {
		if employee.ID == Id {
			json.NewEncoder(w).Encode(employee)
			break
		}
	}
}

func addNewEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger(err.Error())
		json.NewEncoder(w).Encode(ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	isValid, missingFields := CheckJsonFields(requestBody, false)

	if !isValid {
		errorMsg := "Missing fields->"
		for _, value := range missingFields {
			errorMsg += "'" + value + "' "
		}
		logger(errorMsg)
		json.NewEncoder(w).Encode(ErrorResponse{ErrorMsg: errorMsg})
		return
	}

	var item ToDoItem
	unMarshalErr := json.Unmarshal(requestBody, &item)

	if unMarshalErr != nil {
		logger(unMarshalErr.Error())
		json.NewEncoder(w).Encode(ErrorResponse{ErrorMsg: unMarshalErr.Error()})
		return
	}

	item.ID = GetLastId()
	ToDoItemList = append(ToDoItemList, item)
	fmt.Fprintf(w, "%+v", http.StatusOK)
}

func updateEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger(err.Error())
		json.NewEncoder(w).Encode(ErrorResponse{ErrorMsg: err.Error()})
		return
	}

	isValid, missingFields := CheckJsonFields(requestBody, true)

	if !isValid {
		errorMsg := "Missing fields->"
		for _, value := range missingFields {
			errorMsg += "'" + value + "' "
		}
		logger(errorMsg)
		json.NewEncoder(w).Encode(ErrorResponse{ErrorMsg: errorMsg})
		return
	}

	var item ToDoItem
	unMarshalErr := json.Unmarshal(requestBody, &item)

	if unMarshalErr != nil {
		logger(unMarshalErr.Error())
		json.NewEncoder(w).Encode(ErrorResponse{ErrorMsg: unMarshalErr.Error()})
		return
	}
	isExist := false
	for i := 0; i < len(ToDoItemList); i++ {
		value := &ToDoItemList[i]
		if value.ID == item.ID {
			value.Status = item.Status
			value.Description = item.Description
			value.Date = item.Date
			isExist = true
			break
		}
	}

	if !isExist {
		errorMsg := "There is no record has this id!"
		json.NewEncoder(w).Encode(ErrorResponse{ErrorMsg: errorMsg})
		return
	}
	fmt.Fprintf(w, "%+v", http.StatusOK)
}
