package main

import (
	"log"
	"net/http"
	"Backend/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func AddApproutes(route *mux.Router){
	log.Println("loading routes...")

	hub := handlers.NewHub();
	go hub.Run()

	route.HandleFunc("/",handlers.RenderHome)
	route.HandleFunc("/isUsernameAvailable/{username}",handlers.IsUsernameAvailable)

	route.HandleFunc("/login",handlers.Login).Methods("POST")

	route.HandleFunc("/registration",handlers.Registration).Methods("POST")

	route.HandleFunc("/userSessionCheck/{userID}",handlers.UserSessionCheck)

	route.HandleFunc("/getConversation/{toUserID}/{fromUserID}",handlers.GetMessagesHandler)


	//websockets
	route.HandleFunc("/ws/{userID}",func(responseWriter http.ResponseWriter,request *http.Request){
		var upgrader = websocket.Upgrader{
			ReadBufferSize:1024,
			WriteBufferSize:1024,
		}

		userID := mux.Vars(request)["userID"]

		upgrader.CheckOrigin = func(r *http.Request)bool{return true;}

		connection,err := upgrader.Upgrade(responseWriter,request,nil)

		if err!=nil{
			log.Println(err);
			return;
		}

		handlers.CreateNewSocketUser(hub,connection,userID)
	})

	log.Println("Routes are loaded.")
}