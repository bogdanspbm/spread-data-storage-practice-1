package objects

import (
	"errors"
	"spread-data-storage-practice-1/src/utils/adapters"
	"spread-data-storage-practice-1/src/utils/websocket"
)

type Request struct {
	Name string
	Args []string
}

type Response struct {
	Body  string
	Error error
}

type TransactionManager struct {
	requestQueue  chan Request
	responseQueue chan Response
	database      *adapters.DatabaseAdapter
	socket        *websocket.ClusterSocket
}

func CreateTransactionManager(database *adapters.DatabaseAdapter, socket *websocket.ClusterSocket) *TransactionManager {
	requestQueue := make(chan Request)
	responseQueue := make(chan Response)
	manager := &TransactionManager{requestQueue: requestQueue, responseQueue: responseQueue, database: database, socket: socket}
	go manager.managerTick()
	return manager
}

func (manager *TransactionManager) SendRequest(request Request) Response {
	manager.requestQueue <- request
	return <-manager.responseQueue
}

func (manager *TransactionManager) managerTick() {

	for {
		data := <-manager.requestQueue
		manager.responseQueue <- manager.calculateResponse(data)
	}
}

func (manager *TransactionManager) calculateResponse(request Request) Response {
	command := request.buildCommand(manager.database, manager.socket)

	if command == nil {
		return Response{Error: errors.New("bad command")}
	}

	return command()
}

// TODO: По хорошему тут надо сделать под каждый контроллер надстройку в отдельном файле
func (requset *Request) buildCommand(database *adapters.DatabaseAdapter, socket *websocket.ClusterSocket) func() Response {
	switch requset.Name {
	case "put_value":
		return func() Response {
			resp := Response{}

			if len(requset.Args) < 2 {
				resp.Error = errors.New("bad request arguments")
				return resp
			}

			_, err := database.SetVersionValue(requset.Args[0], requset.Args[1], socket.GetLogicTimeInc())

			if err == nil {
				socket.ReplicateValue(requset.Args[0], requset.Args[1])
			}

			resp.Error = err

			return resp
		}
	case "get_value":
		return func() Response {
			resp := Response{}

			if len(requset.Args) < 1 {
				resp.Error = errors.New("bad request arguments")
				return resp
			}

			value, err := database.GetValue(requset.Args[0])
			resp.Body = value.Value
			resp.Error = err
			return resp
		}
	}
	return nil
}
