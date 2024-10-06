package handlers

import ("Backend/Model")
//hub maintains the set of active clients and broadcasts messages to clietns

type Hub struct{
	clients map[*model.Client]bool;
	register chan *model.Client
	unregister chan *model.Client;
}


//new instance of hub
func NewHub() *Hub{
	return &Hub{
		register: make(chan *model.Client),
		unregister: make(chan *model.Client),
		clients: make(map[*model.Client]bool),
	}
}

//exectues go routines to check incoming socket events
func (hub *Hub) Run(){
	for {
		select {
		case client:= <-hub.register:
			HandleUserRegisterEvent(hub,client)

			break;
		case client := <-hub.unregister:
			HandleUserDisconnectEvent(hub,client)
		}
	}
}
