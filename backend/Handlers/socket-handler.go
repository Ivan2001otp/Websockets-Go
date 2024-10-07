package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait*9)/10
	maxMessageSize = 512
)

func CreateNewSocketUser(hub *Hub, connection *websocket.Conn, userID string) {

	client := &Client{
		hub: hub,
		webSocketConnection:connection,
		send:make(chan SocketEventStruct),
		userID: userID,
	}

	go client.writePump()
	go client.readPump()

	client.hub.register <- client;
}

func unRegisterAndCloseConnection(c *Client){
	c.hub.unregister <-c;
	c.webSocketConnection.Close();
}

func setSocketPayloadConfig(c *Client){
	c.webSocketConnection.SetReadLimit(maxMessageSize);
	c.webSocketConnection.SetReadDeadline(time.Now().Add(pongWait))
	c.webSocketConnection.SetPongHandler(func(string) error{
		c.webSocketConnection.SetReadDeadline(time.Now().Add(pongWait));
		return nil;
	})
}

func (c *Client)readPump(){
	var socketEventPayload SocketEventStruct;

	//unregister clients and close the connections
	defer unRegisterAndCloseConnection(c)

	setSocketPayloadConfig(c)

	for{
		_,payload,err := c.webSocketConnection.ReadMessage()

		decoder := json.NewDecoder(bytes.NewReader(payload))

		decoderErr := decoder.Decode(&socketEventPayload)

		if decoderErr!=nil{
			log.Printf("error %v",decoderErr)
			break;
		}

		if err!=nil{
			if websocket.IsUnexpectedCloseError(err,websocket.CloseGoingAway,websocket.CloseAbnormalClosure){
				log.Printf("error -- %v",err);
			}
			break;
		}

		handleSocketPayloadEvents(c,socketEventPayload)
	}


}

func handleSocketPayloadEvents(client *Client,socketEventPayload SocketEventStruct){
	type chatlistResponseStruct struct{
		Type string `json:"type"`
		Chatlist interface{} `json:"chatlist"`
	}

	switch socketEventPayload.EventName{
	case "join":
		userID := (socketEventPayload.EventPayload).(string)
		userDetails := GetUserByUserID(userID)

		if userDetails == (UserDetailsStruct{}){
			log.Println("An invalid user with userID " + userID + " tried to connect to Chat Server.")
		
		}else{
			if userDetails.Online=="N"{
				log.Println("A logged out user with userID " + userID + " tried to connect to Chat Server.")
			}else{
				newUserOnlinePayload := SocketEventStruct{
					EventName:"chatlist-response",
					EventPayload:chatlistResponseStruct{
						Type:"new-user-joined",
						Chatlist:UserDetailsResponsePayloadStruct{
							Online:userDetails.Online,
							UserID:userDetails.ID,
							Username:userDetails.Username,
						},
					},
				}

				BroadcastSocketEventToAllClientExceptMe(client.hub,newUserOnlinePayload,userDetails.ID)

				allOnlineUserPayload := SocketEventStruct{
					EventName: "chatlist-response",
					EventPayload: chatlistResponseStruct{
						Type: "my-chat-list",
						Chatlist: GetAllOnlineUsers(userDetails.ID),
					},
				}

				EmitToSpecificClient(client.hub,allOnlineUserPayload,userDetails.ID)
			}
		}

	case "disconnect":
		if socketEventPayload.EventPayload != nil{
			userID := (socketEventPayload.EventPayload).(string)
			userDetails := GetUserByUserID(userID)
			UpdateUserOnlineStatusByUserID(userID,"N")

			BroadcastSocketEventToAllClient(client.hub,SocketEventStruct{
				EventName: "chatlist-response",
				EventPayload: chatlistResponseStruct{
					Type: "user-disconnected",
					Chatlist: UserDetailsResponsePayloadStruct{
						Online:"N",
						UserID:userDetails.ID,
						Username: userDetails.Username,
					},
				},
			})
		}

	case "message":
		message := (socketEventPayload.EventPayload.(map[string]interface{})["message"]).(string)
		fromUserID  := (socketEventPayload.EventPayload.(map[string]interface{})["fromUserID"]).(string)
		toUserID := (socketEventPayload.EventPayload.(map[string]interface{})["toUserID"]).(string)

		if message!="" && fromUserID!="" && toUserID!=""{
			messagePacket := MessagePayloadStruct{
				FromUserID: fromUserID,
				Message: message,
				ToUserID: toUserID,
			}

			StoreNewChatMessages(messagePacket)

			allOnlineUserPayload := SocketEventStruct{
				EventName: "message-response",
				EventPayload: messagePacket,
			}

			EmitToSpecificClient(client.hub,allOnlineUserPayload,toUserID)
		}
	}
}


func (c *Client) writePump(){
	ticker := time.NewTicker(pingPeriod)

	defer func(){
		ticker.Stop();
		c.webSocketConnection.Close();
	}()

	for {
		select {
		case payload,ok := <-c.send :
			reqBodyBytes := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes).Encode(payload)

			finalPayLoad := reqBodyBytes.Bytes();
			c.webSocketConnection.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				c.webSocketConnection.WriteMessage(websocket.CloseMessage,[]byte{})
				return;
			}

			w,err := c.webSocketConnection.NextWriter(websocket.TextMessage)
			if err!=nil{
				return;
			}

			w.Write(finalPayLoad);

			n:= len(c.send)

			for i:=0 ;i<n;i++{
				json.NewEncoder(reqBodyBytes).Encode(<-c.send)
				w.Write(reqBodyBytes.Bytes())
			}

			if err:=w.Close();err!=nil{
				return;
			}

		case <-ticker.C:
			c.webSocketConnection.SetWriteDeadline(time.Now().Add(writeWait))
			
			if err:=c.webSocketConnection.WriteMessage(websocket.PingMessage,nil);err!=nil{
				return;
			}
		}
	}
}


//will emit the socket event to all socket users
func BroadcastSocketEventToAllClient(hub *Hub,payload SocketEventStruct){
	for client := range hub.clients{
		select{
		case client.send <- payload:
		default:
			close(client.send)
			delete(hub.clients,client);
		}
	}
}


func HandleUserRegisterEvent(hub *Hub,client *Client){
	hub.clients[client]=true;
	handleSocketPayloadEvents(client,SocketEventStruct{
		EventName: "join",
		EventPayload: client.userID,
	})
}

func HandleUserDisconnectEvent(hub *Hub,client *Client){
	_,ok := hub.clients[client];

	if ok {
		delete(hub.clients,client)
		close(client.send)

		handleSocketPayloadEvents(client,SocketEventStruct{
			EventName: "disconnect",
			EventPayload: client.userID,
		})
	}
}

func EmitToSpecificClient(hub *Hub,payload SocketEventStruct,userID string){
	for client := range hub.clients{
		if client.userID == userID{
			select{
			case client.send<-payload:
			default:
				close(client.send);
				delete(hub.clients,client)
			}
		}
	}
}

//emits the socket evetns to all socket users,except for the user who is emitting .
func BroadcastSocketEventToAllClientExceptMe(hub *Hub,payload SocketEventStruct,myUserID string){
	for client := range hub.clients{
		if client.userID != myUserID{
			select{
			case client.send<-payload:
			default:
				close(client.send)
				delete(hub.clients,client);
			}
		}
	}
}

