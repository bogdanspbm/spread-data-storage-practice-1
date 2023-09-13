package requests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"spread-data-storage-practice-1/src/utils/adapters"
	"strings"
)

type StoreServer struct {
	adapter *adapters.DatabaseAdapter
}

func CreateStoreServer(adapter *adapters.DatabaseAdapter) *StoreServer {
	return &StoreServer{adapter: adapter}
}

func (server *StoreServer) RequestValue(w http.ResponseWriter, r *http.Request) {
	setSuccessHeader(w)

	path := r.URL.Path[1:]
	dirs := strings.Split(path, "/")

	if len(dirs) < 2 {
		makeErrorResponse(w, "{}", http.StatusBadRequest)
		return
	}

	key := dirs[1]

	switch r.Method {
	case http.MethodPost:
		server.PutValue(w, r, key)
	case http.MethodGet:
		server.GetValue(w, key)
	default:
		makeErrorResponse(w, "{}", http.StatusMethodNotAllowed)
	}
}

func (server *StoreServer) GetValue(w http.ResponseWriter, key string) {
	value, err := server.adapter.GetValue(key)

	if err != nil {
		makeErrorResponse(w, "{}", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("{\"value\" : %v}", value)))
}

func (server *StoreServer) PutValue(w http.ResponseWriter, r *http.Request, key string) {

	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		makeErrorResponse(w, "{}", http.StatusBadRequest)
		return
	}

	value := string(data)

	_, err = server.adapter.SetValue(key, value)

	if err != nil {
		makeErrorResponse(w, "{}", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("{\"status\" : \"success\"}", value)))
}
