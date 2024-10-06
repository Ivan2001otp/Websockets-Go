package handlers

import ("Backend/Model")
//hub maintains the set of active clients and broadcasts messages to clietns

type Hub struct{
	clients map[*model.Client]bool;
	register chan *model.Client
	unregister chan *model.Client;
}

