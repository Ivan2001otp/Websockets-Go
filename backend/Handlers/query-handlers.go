package handlers

import (
	"Backend/Utils"
	config "Backend/config"
	constants "Backend/constants"
	"context"
	"errors"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//update the user online status by userid
func UpdateUserOnlineStatusByUserID(userID string,status string) error{
	docID,err := primitive.ObjectIDFromHex(userID)

	if err!=nil{
		return nil;
	}

	collection := config.MongoDBClient.Database(os.Getenv("MONGODB_DATABASE")).Collection("users")
	ctx,cancel := context.WithTimeout(context.Background(),10 * time.Second)

	_, queryError := collection.UpdateOne(ctx,bson.M{"_id":docID,},bson.M{"$set":bson.M{"online":status}})

	defer cancel();

	if queryError!=nil{
		return errors.New(constants.ServerFailedResponse)
	}

	return nil;
}

func GetUserByUserID(userID string) UserDetailsStruct{
	var userDetails UserDetailsStruct;

	docID,err := primitive.ObjectIDFromHex(userID)

	if err!=nil{
		return UserDetailsStruct{}
	}

	collection := config.MongoDBClient.Database(os.Getenv("MONGODB_DATABASE")).Collection("users")
	ctx,cancel := context.WithTimeout(context.Background(),10*time.Second)


	_ = collection.FindOne(ctx,bson.M{
		"_id":docID,
	}).Decode(&userDetails)

	defer cancel();

	return userDetails;
}

func GetUserByUsername(username string ) UserDetailsStruct{
	var userDetails UserDetailsStruct;

	collection := config.MongoDBClient.Database(os.Getenv("MONGODB_DATABASE")).Collection("users")


	ctx,cancel := context.WithTimeout(context.Background(),10*time.Second);

	_ = collection.FindOne(ctx,bson.M{
		"username":username,
	}).Decode(&userDetails)

	defer cancel();

	return userDetails;
}

func IsUsernameAvailableQueryHandler(username string)bool{
	userDetails := GetUserByUsername(username)

	if userDetails==(UserDetailsStruct{}){
		return true;
	}
	return false;
}

func LoginQueryHandler(userDetailsRequestPayload UserDetailsRequestPayloadStruct )(UserDetailsResponsePayloadStruct,error){
	if userDetailsRequestPayload.Username==""{
		return UserDetailsResponsePayloadStruct{},errors.New(constants.UsernameCantBeEmpty);

	}else if(userDetailsRequestPayload.Password==""){
		return UserDetailsResponsePayloadStruct{},errors.New(constants.PasswordCantBeEmpty);
	}else{
		userDetails := GetUserByUsername(userDetailsRequestPayload.Username)

		if userDetails == (UserDetailsStruct{}) {
			return UserDetailsResponsePayloadStruct{},errors.New(constants.UserIsNotRegisteredWithUs)
		}

		if isPasswordOkay:= utils.ComparePasswords(userDetailsRequestPayload.Password,userDetails.Password);isPasswordOkay!=nil{
			return UserDetailsResponsePayloadStruct{},errors.New(constants.LoginPasswordIsInCorrect)

		}

			return UserDetailsResponsePayloadStruct{
				UserID: userDetails.ID,
				Username: userDetails.Username,
			},nil;
	}
}

func RegisterQueryHandler(userDetailsRequestPayload UserDetailsRequestPayloadStruct)(string,error){
	if userDetailsRequestPayload.Username==""{
		return "",errors.New(constants.UsernameCantBeEmpty);
	}else if userDetailsRequestPayload.Password==""{
		return "",errors.New(constants.PasswordCantBeEmpty)
	}else{
		newPasswordHash,err := utils.CreatePassword(userDetailsRequestPayload.Password)

		if err!=nil{
			return "",errors.New(constants.ServerFailedResponse);
		}

		collection := config.MongoDBClient.Database(os.Getenv("MONGODB_DATABASE")).Collection("users")
		ctx,cancel := context.WithTimeout(context.Background(),10*time.Second)

		registrationQueryResponse,err := collection.InsertOne(ctx,bson.M{
			"username":userDetailsRequestPayload.Username,
			"password":newPasswordHash,
			"online":"N",
		})

		defer cancel();

		registrationQueryObjID := registrationQueryResponse.InsertedID.(primitive.ObjectID)

		if onlineStatusError:=UpdateUserOnlineStatusByUserID(registrationQueryObjID.Hex(),"Y");onlineStatusError!=nil{
			return "",errors.New(constants.ServerFailedResponse)
		}

		return registrationQueryObjID.Hex(),nil;
	}
}

func GetAllOnlineUsers(userID string)[]UserDetailsResponsePayloadStruct{
	var onlineUsers []UserDetailsResponsePayloadStruct;

	docID,err := primitive.ObjectIDFromHex(userID)

	if err!=nil{
		log.Println("failed to convert hex to object id")
		return onlineUsers;
	}

	collection := config.MongoDBClient.Database(os.Getenv("MONGODB_DATABASE")).Collection("users")
	ctx,cancel := context.WithTimeout(context.Background(),10*time.Second)

	cursor,queryError := collection.Find(ctx,bson.M{
		"online":"Y",
		"_id":bson.M{
			"$ne":docID,
		},
	})

	defer cancel();

	if queryError!=nil{
		log.Println("could not get all online users")
		return onlineUsers;
	}

	for cursor.Next(context.TODO()){
		var singleOnlineUser UserDetailsStruct;
		err := cursor.Decode(&singleOnlineUser)

		if err!=nil{
			onlineUsers = append(onlineUsers, UserDetailsResponsePayloadStruct{
				UserID: singleOnlineUser.ID,
				Online: singleOnlineUser.Online,
				Username: singleOnlineUser.Username,
			})
		}
	}

	return onlineUsers;
}

func StoreNewChatMessages(messagePayload MessagePayloadStruct)bool{
	collection := config.MongoDBClient.Database(os.Getenv("MONGODB_DATABASE")).Collection("messages")
	ctx,cancel := context.WithTimeout(context.Background(),10*time.Second)

	_,registrationError := collection.InsertOne(ctx,bson.M{
		"fromUserID":messagePayload.FromUserID,
		"message":messagePayload.Message,
		"toUserID":messagePayload.ToUserID,
	})

	defer cancel();

	if registrationError!=nil{
		return false;
	}

	return true;
}

//fetch the cconversations messages between the two users.
func GetConversationBetweenTwoUsers(toUserID string,fromUserId string)[]ConversationStruct{
	var conversations []ConversationStruct;

	collection := config.MongoDBClient.Database(os.Getenv("MONGODB_DATABASE")).Collection("messages")
	ctx,cancel := context.WithTimeout(context.Background(),10*time.Second)

	queryCondition := bson.M{
		"$or":[]bson.M{
			{
				"$and":[]bson.M{
					{
						"toUserID":toUserID,
					},{
						"fromUserID":fromUserId,
					},
				},
			},
			{
				"$and":[]bson.M{
					{
						"toUserID":fromUserId,
					},{
						"fromUserID":toUserID,
					},
				},
			},
		},
		
	}


	cursor ,err := collection.Find(ctx,queryCondition)
	defer cancel();

	if err!=nil{
		log.Println("could not get conversations");
		return conversations;
	}

	for cursor.Next(context.TODO()){
		var conversation ConversationStruct;
		err := cursor.Decode(&conversation)

		if err!=nil{
			conversations = append(conversations,ConversationStruct{
				ID: conversation.ID,
				FromUserID: conversation.FromUserID,
				ToUserID: conversation.ToUserID,
				Message: conversation.Message,
			});
		}
	}

	return conversations;
}
