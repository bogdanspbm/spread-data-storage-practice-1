package objects

type Request struct {
	Command func() Response
}

type Response struct {
	Body  string
	Error error
}

type TransactionManager struct {
	requestQueue  chan Request
	responseQueue chan Response
}

func CreateTransactionManager() *TransactionManager {
	requestQueue := make(chan Request)
	responseQueue := make(chan Response)
	manager := &TransactionManager{requestQueue: requestQueue, responseQueue: responseQueue}
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
	return request.Command()
}
