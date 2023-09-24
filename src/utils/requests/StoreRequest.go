package requests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"spread-data-storage-practice-1/src/utils/adapters"
	"spread-data-storage-practice-1/src/utils/objects"
	"strings"
)

type StoreServer struct {
	adapter *adapters.DatabaseAdapter
	manager *objects.TransactionManager
}

func CreateStoreServer(adapter *adapters.DatabaseAdapter, manager *objects.TransactionManager) *StoreServer {
	return &StoreServer{adapter: adapter, manager: manager}
}

func (server *StoreServer) RequestValue(w http.ResponseWriter, r *http.Request) {
	setSuccessHeader(w)

	path := r.URL.Path[1:]
	dirs := strings.Split(path, "/")

	if len(dirs) < 2 {
		makeErrorResponse(w, "don't have key", http.StatusBadRequest)
		return
	}

	key := dirs[1]

	if len(key) <= 0 {
		makeErrorResponse(w, "empty key", http.StatusBadRequest)
		return
	}

	fmt.Println(fmt.Sprintf("Request Type: %v Key: %v", r.Method, key))

	switch r.Method {
	case http.MethodPost:
		server.PutValue(w, r, key)
	case http.MethodGet:
		server.GetValue(w, key)
	default:
		makeErrorResponse(w, "method isn't allowed", http.StatusMethodNotAllowed)
	}
}

func (server *StoreServer) GetValue(w http.ResponseWriter, key string) {

	request := objects.Request{Command: func() objects.Response {
		value, err := server.adapter.GetValue(key)
		return objects.Response{Body: value, Error: err}
	}}
	response := server.manager.SendRequest(request)

	if response.Error != nil {
		makeErrorResponse(w, "can't find value with key", http.StatusInternalServerError)
		return
	}

	value := response.Body
	w.Write([]byte(fmt.Sprintf("{\"value\" : \"%v\"}", value)))
}

func (server *StoreServer) PutValue(w http.ResponseWriter, r *http.Request, key string) {

	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		makeErrorResponse(w, "can't read body", http.StatusBadRequest)
		return
	}

	value := string(data)

	if len(value) <= 0 {
		makeErrorResponse(w, "empty string", http.StatusBadRequest)
		return
	}

	request := objects.Request{Command: func() objects.Response {
		_, err := server.adapter.SetValue(key, value)
		return objects.Response{Error: err}
	}}
	response := server.manager.SendRequest(request)

	if response.Error != nil {
		makeErrorResponse(w, "can't set value", http.StatusInternalServerError)
		return
	}

	makeErrorResponse(w, "success", http.StatusOK)
}
