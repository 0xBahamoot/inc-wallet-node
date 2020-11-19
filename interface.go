package main

type httpHandler func(*API_Server, interface{}, <-chan struct{}) (interface{}, error)

// Commands valid for normal user
var HttpHandler = map[string]httpHandler{
	api_retrieveoutputcoins: (*API_Server).RetrieveOutputCoins,
}
