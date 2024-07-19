package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/srijan-raghavula/internal/pokecache"
	"io"
	"log"
	"net/http"
	"os"
)

type command struct {
	name        string
	description string
	callback    func(*config) error
}

type config struct {
	next string
	prev string
}

var commands map[string]command
var isFirstCall bool = true
var cache pokecache.Cache = pokecache.NewCache()

func main() {
	c := config{
		next: "https://pokeapi.co/api/v2/location-area",
		prev: "",
	}
	commands = map[string]command{
		"help": {
			name:        "help",
			description: "prints a message to help yourself using Pokedex",
			callback:    printCommands,
		},
		"exit": {
			name:        "exit",
			description: "exits Pokedex",
			callback:    exit,
		},
		"map": {
			name:        "map",
			description: "prints the next 20 location areas in the pokemon world",
			callback:    mapNext,
		},
		"mapb": {
			name:        "mapb",
			description: "prints the previous 20 location areas in the pokemon world (returns an error if you haven't started your exploration yet)",
			callback:    mapPrev,
		},
	}
	scanned := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("Pokedex > ")
		if scanned.Scan() {
			stdIn, err := scanned.Text(), scanned.Err()
			if err != nil {
				log.Fatal(err)
			}
			switch stdIn {
			case "help":
				err := commands[stdIn].callback(&c)
				if err != nil {
					log.Fatal(err)
				}
			case "exit":
				err := commands[stdIn].callback(&c)
				if err != nil {
					log.Fatal(err)
				}
			case "map":
				err := commands[stdIn].callback(&c)
				if err != nil {
					log.Fatal(err)
				}
			case "mapb":
				err := commands[stdIn].callback(&c)
				if err != nil {
					fmt.Println(err)
				}
			default:
				fmt.Println(`Available commands to use(use "help" for more details)`)
				for k := range commands {
					fmt.Println(k)
				}
			}
		}
	}
}

func printCommands(c *config) error {
	if len(commands) == 0 {
		return errors.New("no commands yet")
	}
	fmt.Println("=======Pokedex help center=======")
	fmt.Println("\nusage: [command]\n\nCOMMANDS TO USE:")
	for k, v := range commands {
		fmt.Printf("%v\n", k)
		if v.description == "" {
			return errors.New("command not described")
		}
		fmt.Printf("%v\n\n", v.description)
	}
	return nil
}

func exit(c *config) error {
	os.Exit(1)
	return nil
}

func mapNext(c *config) error {
	res, err := http.Get(c.next)
	if err != nil {
		return err
	}

	if sc := res.StatusCode; sc > 299 {
		errorMessage := "response status code:" + string(sc)
		return errors.New(errorMessage)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	var bodyMap map[string]interface{}
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return err
	}

	results := bodyMap["results"].([]interface{})
	for _, result := range results {
		converted := result.(map[string]interface{})
		fmt.Println(converted["name"])
	}
	if !isFirstCall {
		c.prev = bodyMap["previous"].(string)
	}
	if isFirstCall {
		c.prev = c.next
		isFirstCall = false
	}
	c.next = bodyMap["next"].(string)
	return nil
}

func mapPrev(c *config) error {
	if isFirstCall {
		return errors.New("no prev locations to show")
	}
	res, err := http.Get(c.prev)
	if err != nil {
		return err
	}

	if sc := res.StatusCode; sc > 299 {
		errorMessage := "response status code:" + string(sc)
		return errors.New(errorMessage)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	var bodyMap map[string]interface{}
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return err
	}

	results := bodyMap["results"].([]interface{})
	for _, result := range results {
		converted := result.(map[string]interface{})
		fmt.Println(converted["name"])
	}
	return nil
}
