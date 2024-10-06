package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func CreatePassword(passwordString string)(string,error){
	hashedPassword,err := bcrypt.GenerateFromPassword([]byte(passwordString),8)

	if err!=nil{
		return "",errors.New("error occured while creating hash")
	}

	return string(hashedPassword),nil;
}

func ComparePasswords(password string,hashedPassword string)error{
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword),[]byte(password))

	if err!=nil{
		return errors.New("The password and hashpassword does not match");
	}

	return nil;
}