package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/srijan-raghavula/pokedex/internal/pokecache"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type command struct {
	name        string
	description string
	callback    func(string, *config) error
}

type config struct {
	next string
	prev string
}

type locList struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type locEndpoint struct {
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int `json:"game_index"`
	ID        int `json:"id"`
	Location  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

var commands map[string]command
var isFirstCall bool = true
var cache pokecache.Cache = pokecache.NewCache(time.Minute * 2)

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
		"explore": {
			name:        "explore",
			description: "takes a location area and lists all the Pokemons in the area",
			callback:    nil,
		},
	}
	for {
		fmt.Printf("Pokedex > ")
		scanned := bufio.NewScanner(os.Stdin)
		if scanned.Scan() {
			stdIn, err := scanned.Text(), scanned.Err()
			if err != nil {
				log.Fatal(err)
			}

			switch stdIn {
			case "help":
				err := commands[stdIn].callback("", &c)
				if err != nil {
					log.Fatal(err)
				}
			case "exit":
				err := commands[stdIn].callback("", &c)
				if err != nil {
					log.Fatal(err)
				}
			case "map":
				err := commands[stdIn].callback("", &c)
				if err != nil {
					log.Fatal(err)
				}
			case "mapb":
				err := commands[stdIn].callback("", &c)
				if err != nil {
					fmt.Println(err)
				}
			case "explore":
				err := commands[stdIn].callback("", &c)
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

func printCommands(s string, c *config) error {
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

func exit(s string, c *config) error {
	os.Exit(1)
	return nil
}

func mapNext(s string, c *config) error {
	val, ok := cache.Get(c.next)
	if ok {
		body := val
		var bodyMap locList
		err := json.Unmarshal(body, &bodyMap)
		if err != nil {
			return err
		}

		for _, result := range bodyMap.Results {
			fmt.Println(result.Name)
		}
		if !isFirstCall {
			c.prev = bodyMap.Previous
		}
		if isFirstCall {
			c.prev = c.next
			isFirstCall = false
		}
		c.next = bodyMap.Next
		return nil
	}
	res, err := http.Get(c.next)
	if err != nil {
		return err
	}

	if sc := res.StatusCode; sc > 399 {
		errorMessage := "response status code: " + fmt.Sprintf("%q", sc)
		return errors.New(errorMessage)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	cache.Add(c.next, body)
	res.Body.Close()

	var bodyMap locList
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return err
	}

	for _, result := range bodyMap.Results {
		fmt.Println(result.Name)
	}
	if !isFirstCall {
		c.prev = bodyMap.Previous
	}
	if isFirstCall {
		c.prev = c.next
		isFirstCall = false
	}
	c.next = bodyMap.Next
	return nil
}

func mapPrev(s string, c *config) error {
	val, ok := cache.Get(c.prev)
	if ok {
		body := val
		var bodyMap locList
		err := json.Unmarshal(body, &bodyMap)
		if err != nil {
			return err
		}

		for _, result := range bodyMap.Results {
			fmt.Println(result.Name)
		}
		return nil
	}
	if isFirstCall {
		return errors.New("no prev locations to show")
	}
	res, err := http.Get(c.prev)
	if err != nil {
		return err
	}

	if sc := res.StatusCode; sc > 399 {
		errorMessage := "response status code:" + fmt.Sprintf("%q", sc)
		return errors.New(errorMessage)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	cache.Add(c.prev, body)
	res.Body.Close()

	var bodyMap locList
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return err
	}

	for _, result := range bodyMap.Results {
		fmt.Println(result.Name)
	}
	return nil
}

func locArea(name string, c *config) error {
	val, ok := cache.Get(c.prev + "/" + name)
	if ok {
		body := val
		var bodyMap locEndpoint
		err := json.Unmarshal(body, &bodyMap)
		if err != nil {
			return err
		}

		for _, result := range bodyMap.PokemonEncounters {
			fmt.Println(result.Pokemon.Name)
		}
		return nil
	}
	res, err := http.Get(c.prev + "/" + name)
	if err != nil {
		return err
	}

	if sc := res.StatusCode; sc > 399 {
		errorMessage := "response status code: " + fmt.Sprintf("%q", sc)
		return errors.New(errorMessage)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	cache.Add(c.next, body)
	res.Body.Close()

	var bodyMap locEndpoint
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return err
	}

	for _, result := range bodyMap.PokemonEncounters {
		fmt.Println(result.Pokemon.Name)
	}
	return nil
}
