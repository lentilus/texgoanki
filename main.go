package main

import (
	"crypto"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"texgoanki/api"
	"texgoanki/compile"
)

const modelName = "texgoanki"

type Flashcard struct {
	Id    string `json:"id"`
	Front string `json:"front"`
	Back  string `json:"back"`
}

func parse(src string) ([]Flashcard, error) {
	cmd := exec.Command(
		"texflash",
		src,
	)
	stdout, err := cmd.Output()
	// fmt.Println(string(stdout))
	if err != nil {
		return nil, err
	}
	var cards []Flashcard
	err = json.Unmarshal(stdout, &cards)
	if err != nil {
		return nil, err
	}
	return cards, nil
}

func Hash(objs ...interface{}) []byte {
	digester := crypto.MD5.New()
	for _, ob := range objs {
		fmt.Fprint(digester, reflect.TypeOf(ob))
		fmt.Fprint(digester, ob)
	}
	return digester.Sum(nil)
}

func main() {
	if len(os.Args) < 4 {
		log.Fatal("We need at least 4 arguments. Aborting.")
	}

	deck := os.Args[1]
	src := os.Args[2]
	target := os.Args[3]
	croot := os.Args[4]

	var context []string
	for _, a := range os.Args[5:] {
		context = append(context, a)
	}
	fmt.Println(context)

	// extract the relevant source code
	cards, err := parse(src)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(cards)

	// the Anki-Connect API
	connect := api.AnkiConnect{Url: "http://localhost:8765"}

	// if the deck does not exist, create it
	decks, err := api.GetDecks(&connect)
	if err != nil {
		log.Fatal(err)
	}
	present := false
	for _, v := range decks {
		if v == deck {
			present = true
			break
		}
	}
	if !present {
		api.CreateDeck(&connect, deck)
	}

	// if our custom model does not exist, create it
	models, err := api.GetDecks(&connect)
	if err != nil {
		log.Fatal(err)
	}
	modelpresent := false
	for _, v := range models {
		if v == modelName {
			modelpresent = true
			break
		}
	}
	if !modelpresent {
		fields := []string{"front", "back", "id", "hash"}
		template := []map[string]string{
			{
				"Name":  "texgoanki",
				"Front": "{{front}}",
				"Back":  "{{back}}",
			},
		}
		api.CreateModel(
			&connect,
			modelName,
			fields,
			template,
		)
	}

	for _, c := range cards {
		// compute hash
		hash := hex.EncodeToString(Hash(c.Front, c.Back))
		id, err := api.FindCard(&connect, deck, c.Id)

		if err != nil || id == -1 {
			// unable to find flashcard with corresponding id
			// -> create a new card
			front, err := compile.Src2svg(c.Front, croot, context, target, "/dev/shm/texgoanki")
			if err != nil {
				log.Println(err)
				continue
			}

			back, err := compile.Src2svg(c.Back, croot, context, target, "/dev/shm/texgoanki")
			if err != nil {
				log.Println(err)
				continue
			}

			// api.StoreMediaFile(&connect, , data string)
			front_file := hash + "_front.svg"
			back_file := hash + "_back.svg"
			api.StoreMediaFile(&connect, front_file, front)
			api.StoreMediaFile(&connect, back_file, back)

			// create card
			_, err = api.AddCard(
				&connect,
				deck,
				modelName,
				fmt.Sprintf("<img src=%s>", front_file),
				fmt.Sprintf("<img src=%s>", back_file),
				c.Id,
				hash,
			)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("Added new Flashcard!")
			continue
		}

		lasthash, err := api.GetCardField(&connect, id, "hash")
		if err != nil {
			log.Println("Unable to read hash, updating")
		} else if hash == lasthash {
			log.Println("hash did not change, skipping")
			continue
		}

		front, err := compile.Src2svg(c.Front, croot, context, target, "/dev/shm/texgoanki")
		if err != nil {
			log.Println(err)
			continue
		}

		back, err := compile.Src2svg(c.Back, croot, context, target, "/dev/shm/texgoanki")
		if err != nil {
			log.Println(err)
			continue
		}

		front_file := hash + "_front.svg"
		back_file := hash + "_back.svg"
		api.StoreMediaFile(&connect, front_file, front)
		api.StoreMediaFile(&connect, back_file, back)

		// create card
		_, err = api.UpdateCard(
			&connect,
			id,
			fmt.Sprintf("<img src=%s>", front_file),
			fmt.Sprintf("<img src=%s>", back_file),
			c.Id,
			hash,
		)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("Updated Flashcard!")
	}
}
