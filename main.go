package main

import (
	"log"
)

func main() {

	typeFile := "types/example.go"
	logicFile := "logic/logic.go"

	typeInfo := parseTypes(typeFile, "/login")

	logicFunc := genLogicFunc(logicFile, typeInfo)

	handlerFunc := genHandlerFunc("handler/handler.go", typeInfo, logicFunc)

	err := addRouter1("router/router.go", "UserRouter", typeInfo, handlerFunc)
	if err != nil {
		log.Fatal(err)
	}
}
