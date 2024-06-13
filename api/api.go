package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// generic type for request body parameters
type P any

// generic type for the response
type R any

type Body struct {
	Action  string `json:"action"`
	Version int    `json:"version"`
	Params  P      `json:"params,omitempty"`
}

// API encapsules the actual api calls in order to
// mock the api in testing
type API interface {
	Request(string, P, R) error
}

// Actual Anki Connect API
type AnkiConnect struct {
	Url string
}

// Make a request to the AnkiConnect REST-API
func (a *AnkiConnect) Request(action string, params P, response R) error {
	body := Body{action, 6, params}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// fmt.Println(bytes.NewBuffer(b))
	req, err := http.NewRequest(http.MethodPost, a.Url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("status code not okay")
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return err
	}

	return nil
}

// returns a list of Anki deck names (and an error)
func GetDecks(api API) ([]string, error) {
	type Response struct {
		Result []string // deck names
		Error  any
	}

	res := new(Response)
	err := api.Request("deckNames", nil, &res)
	if err != nil {
		return nil, err
	}

	if res.Error != nil {
		msg, ok := res.Error.(string)
		if ok {
			return nil, errors.New(msg)
		}
		return nil, errors.New("Cannot read error message.")
	}

	return res.Result, nil
}

// creates a new Deck and returns its id (and an error)
func CreateDeck(api API, name string) (int, error) {
	type Response struct {
		Result int // deck id
		Error  any
	}

	res := new(Response)
	err := api.Request("createDeck", map[string]any{"deck": name}, &res)
	if err != nil {
		return -1, err

	}

	if res.Error != nil {
		msg, ok := res.Error.(string)
		if ok {
			return -1, errors.New(msg)
		}
		return -1, errors.New("Cannot read error message.")
	}

	fmt.Println(res.Error)
	return res.Result, nil
}

// Stores file with base 64 encoded content.
// for some reason the api is unable to process
// absolute paths (the documentation suggests otherwise)
func StoreMediaFile(api API, name string, data string) (string, error) {
	type Response struct {
		Result string // deck id
		Error  any
	}

	res := new(Response)
	params := map[string]any{"filename": name, "data": data}
	err := api.Request("storeMediaFile", params, &res)
	if err != nil {
		return "", err
	}

	if res.Error != nil {
		msg, ok := res.Error.(string)
		if ok {
			return "", errors.New(msg)
		}
		return "", errors.New("Cannot read error message.")
	}

	fmt.Println(res.Error)
	return res.Result, nil
}

func GetModels(api API) ([]string, error) {
	type Response struct {
		Result []string // deck names
		Error  any
	}

	res := new(Response)
	err := api.Request("modelNames", nil, &res)
	if err != nil {
		return nil, err
	}

	if res.Error != nil {
		msg, ok := res.Error.(string)
		if ok {
			return nil, errors.New(msg)
		}
		return nil, errors.New("Cannot read error message.")
	}

	return res.Result, nil
}

func CreateModel(
	api API,
	name string,
	fields []string,
	templates []map[string]string,
) (any, error) {
	params := map[string]any{
		"modelName":     name,
		"inOrderFields": fields,
		"isCloze":       false,
		"cardTemplates": templates,
	}

	type Response struct {
		Result any // deck id
		Error  any
	}

	res := new(Response)
	err := api.Request("createModel", params, &res)
	if err != nil {
		return nil, err
	}

	if res.Error != nil {
		msg, ok := res.Error.(string)
		if ok {
			return "", errors.New(msg)
		}
		return nil, errors.New("Cannot read error message.")
	}

	fmt.Println(res.Error)
	return res.Result, nil

}

// This calls AddNote internally
// func AddCard(api API, name string) (string, error) {
// }

// update Note (update note fields)
// func Sync(api API) (error) {}
