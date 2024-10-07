package handlers

import ("github.com/gorilla/websocket"	
)

type UserDetailsStruct struct {
	ID       string `bson:"_id,omit_empty"`
	Username string `json:"user_name"`
	Password string `json:"password"`
	Online   string `json:"online"`
	SocketID string `json:"socket_id"`
}

// map cconversations
type ConversationStruct struct {
	ID         string `json:"id" bson:"_id,omit_empty"`
	Message    string `json:"message"`
	ToUserID   string `json:"toUserID"`
	FromUserID string `json:"fromUserID"`
}

// payload for login and registration request
type UserDetailsRequestPayloadStruct struct {
	Username string
	Password string
}

// payload for login and registration response
type UserDetailsResponsePayloadStruct struct {
	Username string `json:"username"`
	UserID   string `json:"userID"`
	Online   string `json:"online"`
}

// socket events
type SocketEventStruct struct {
	EventName    string      `json:"eventName"`
	EventPayload interface{} `json:"eventPayload"`
}

// client is a middleman between websocket conn & the hub.
type Client struct {
	hub                 *Hub
	webSocketConnection *websocket.Conn
	send 	chan SocketEventStruct //<- imp
	userID string
}


//message payload
type MessagePayloadStruct struct{
	FromUserID 	string 		`json:"fromUserID"`
	ToUserID 	string  	`json:"toUserID"`
	Message 	string 	 `json:"message"`
}