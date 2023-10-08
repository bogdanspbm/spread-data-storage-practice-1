package requests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"spread-data-storage-practice-1/src/utils/adapters"
	"spread-data-storage-practice-1/src/utils/objects"
	"spread-data-storage-practice-1/src/utils/websocket"
	"strings"
)

type StoreServer struct {
	adapter   *adapters.DatabaseAdapter
	websocket *websocket.ClusterSocket
	manager   *objects.TransactionManager
	journal   *[]objects.Request
}

func CreateStoreServer(adapter *adapters.DatabaseAdapter, socket *websocket.ClusterSocket, manager *objects.TransactionManager, journal *[]objects.Request) *StoreServer {
	return &StoreServer{adapter: adapter, manager: manager, websocket: socket, journal: journal}
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

	request := objects.Request{Name: "get_value", Args: []string{key}}
	*server.journal = append(*server.journal, request)
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

	request := objects.Request{Name: "put_value", Args: []string{key, value}}
	*server.journal = append(*server.journal, request)
	response := server.manager.SendRequest(request)

	if response.Error != nil {
		makeErrorResponse(w, "can't set value", http.StatusInternalServerError)
		return
	}

	server.websocket.ReplicateValue(key, value)
	makeErrorResponse(w, "success", http.StatusOK)
}
