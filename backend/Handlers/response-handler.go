package handlers

import (
	"encoding/json"
	"net/http"
	constants "Backend/Constants"
)

type APIResponseStruct struct{
	Code int 	`json:"code"`
	Status string `json:"status"`
	Message string `json:"message"`
	Response interface{} `json:"response"`
}

func ReturnResponse(response http.ResponseWriter,request *http.Request,apiResponse APIResponseStruct){
	var (
		responseMessage,responseStatusText string
		responseHTTPCode 					int
	)

	if apiResponse.Code==0{
		responseHTTPCode = http.StatusOK
	}else{
		responseHTTPCode = apiResponse.Code
	}

	if apiResponse.Status!=""{
		responseStatusText=apiResponse.Status;
	}else{
		responseStatusText = http.StatusText(http.StatusOK)
	}


	if apiResponse.Message!=""{
		responseMessage = apiResponse.Message;
	}else{
		responseMessage = constants.SuccessfulResponse;
	}

	httpResponse := &APIResponseStruct{
		Code: responseHTTPCode,
		Status: responseStatusText,
		Message: responseMessage,
		Response: apiResponse.Response,
	}

	jsonResponse,err := json.Marshal(httpResponse)

	if err!=nil{
		panic(err);
	}

	response.Header().Set("Content-Type","application/json")
	response.WriteHeader(httpResponse.Code);
	response.Write(jsonResponse);
}