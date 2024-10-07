package handlers


//hub maintains the set of active clients and broadcasts messages to clietns

type Hub struct{
	clients map[*Client]bool;
	register chan *Client
	unregister chan *Client;
}


//new instance of hub
func NewHub() *Hub{
	return &Hub{
		register: make(chan *Client),
		unregister: make(chan *Client),
		clients: make(map[*Client]bool),
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
