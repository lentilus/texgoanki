package main

import (
	"fmt"
	"texgoanki/api"
)

func main() {
	connect := api.AnkiConnect{Url: "http://localhost:8765"}
	names, _ := api.GetDecks(&connect)
	fmt.Println(names[0])

	fmt.Println(api.CreateDeck(&connect, "bar"))

	// myres := response{result: "foo", error: "foo"}
	// fmt.Print(myres.result.(string))
}
