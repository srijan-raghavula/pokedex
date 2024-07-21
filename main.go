package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/srijan-raghavula/pokedex/internal/pokecache"
	"github.com/srijan-raghavula/pokedex/internal/pokemon"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	cfg := config{
		next: "https://pokeapi.co/api/v2/location-area",
		prev: "https://pokeapi.co/api/v2/location-area",
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
			callback:    pokemonList,
		},
		"catch": {
			name:        "catch",
			description: "catches pokemon and adds to Pokedex",
			callback:    catchPokemon,
		},
		"inspect": {
			name:        "inspect",
			description: "inspects a Pokemon in your Pokedex and shows the details of the Pokemon",
			callback:    inspectPokemon,
		},
		"pokedex": {
			name:        "pokedex",
			description: "lists all the Pokemons caught",
			callback:    pokedex,
		},
	}
	for {
		scanned := bufio.NewScanner(os.Stdin)
		fmt.Printf("Pokedex > ")
		scanned.Scan()
		stdIn, err := scanned.Text(), scanned.Err()
		if err != nil {
			fmt.Println(err)
		}

		words := strings.Fields(stdIn)
		for i, word := range words {
			words[i] = strings.ToLower(word)
		}
		cmd := words[0]
		noOfWords := len(words)

		switch cmd {
		case "help":
			err := commands[cmd].callback(&cfg)
			if err != nil {
				log.Fatal(err)
			}
		case "exit":
			err := commands[cmd].callback(&cfg)
			if err != nil {
				log.Fatal(err)
			}
		case "map":
			err := commands[cmd].callback(&cfg)
			if err != nil {
				log.Fatal(err)
			}
		case "mapb":
			err := commands[cmd].callback(&cfg)
			if err != nil {
				fmt.Println(err)
			}
		case "explore":
			if noOfWords < 2 {
				fmt.Println("usage: explore <location-area-name>")
				break
			}
			locationAreaEndpoint := (words[1])
			err := commands[cmd].callback(&cfg, locationAreaEndpoint)
			if err != nil {
				fmt.Println(err)
			}
		case "catch":
			if noOfWords < 2 {
				fmt.Println("usage: catch <pokemon-name>")
				break
			}
			pokemonName := strings.ToLower(words[1])
			err := commands[cmd].callback(&cfg, pokemonName)
			if err != nil {
				fmt.Println(err)
			}
		case "inspect":
			if noOfWords < 2 {
				fmt.Println("usage: catch <pokemon-name>")
				break
			}
			pokemonName := strings.ToLower(words[1])
			err := commands[cmd].callback(&cfg, pokemonName)
			if err != nil {
				fmt.Println(err)
			}
		case "pokedex":
			err := commands[cmd].callback(&cfg)
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

type command struct {
	name        string
	description string
	callback    func(*config, ...string) error
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

var commands map[string]command
var isFirstCall bool = true
var cache pokecache.Cache = pokecache.NewCache(time.Minute * 2)

func printCommands(c *config, s ...string) error {
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

func exit(c *config, s ...string) error {
	os.Exit(1)
	return nil
}

func mapNext(c *config, s ...string) error {
	val, ok := cache.Get(c.next)
	if ok {
		var unmarshaled locList
		err := json.Unmarshal(val, &unmarshaled)
		if err != nil {
			return err
		}

		for _, result := range unmarshaled.Results {
			fmt.Println(result.Name)
		}
		c.prev = unmarshaled.Previous
		c.next = unmarshaled.Next
		return nil
	}
	res, err := http.Get(c.next)
	if err != nil {
		return err
	}

	if sc := res.StatusCode; sc > 299 {
		errorMessage := "response status code: " + fmt.Sprintf("%d", sc)
		return errors.New(errorMessage)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	cache.Add(c.next, body)
	res.Body.Close()

	var unmarshaled locList
	err = json.Unmarshal(body, &unmarshaled)
	if err != nil {
		return err
	}

	for _, result := range unmarshaled.Results {
		fmt.Println(result.Name)
	}
	if !isFirstCall {
		c.prev = unmarshaled.Previous
	}
	c.next = unmarshaled.Next
	isFirstCall = false
	return nil
}

func mapPrev(c *config, s ...string) error {
	val, ok := cache.Get(c.prev)
	if ok {
		var unmarshaled locList
		err := json.Unmarshal(val, &unmarshaled)
		if err != nil {
			return err
		}

		for _, result := range unmarshaled.Results {
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

	if sc := res.StatusCode; sc > 299 {
		errorMessage := "response status code:" + fmt.Sprintf("%d", sc)
		return errors.New(errorMessage)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	cache.Add(c.prev, body)
	res.Body.Close()

	var unmarshaled locList
	err = json.Unmarshal(body, &unmarshaled)
	if err != nil {
		return err
	}

	for _, result := range unmarshaled.Results {
		fmt.Println(result.Name)
	}
	return nil
}

func pokemonList(c *config, names ...string) error {
	if len(names) < 1 {
		return errors.New("check the string passed into the function")
	}
	endpoint := names[0]
	unmarshaled, err := pokemonsInArea(endpoint, c.next, c.prev)
	if err != nil {
		return err
	}
	for _, pokemon := range unmarshaled.PokemonEncounters {
		fmt.Println(pokemon.Pokemon.Name)
	}
	return nil
}

func pokemonsInArea(endpoint, next, prev string) (locEndpoint, error) {
	val, ok := cache.Get(endpoint)
	var unmarshaled locEndpoint
	if ok {
		var unmarshaled locEndpoint
		err := json.Unmarshal(val, &unmarshaled)
		if err != nil {
			return unmarshaled, err
		}

		return unmarshaled, nil
	}

	// players most likely will explore an area
	// they know the name of and they know the name
	// from using map previously, if they used
	// map before, config.prev will have the appropriate url
	// if they magically just knew the area name from next,
	// we will try that url in an if case
	var url string
	querySeparated := strings.Split(prev, "?")
	if len(querySeparated) == 1 {
		url = fmt.Sprintf("%s/%s", prev, endpoint)
	} else {
		url = fmt.Sprintf("%s/%s?%s", querySeparated[0], endpoint, querySeparated[1])
	}

	res, err := http.Get(url)
	if err != nil {
		return unmarshaled, err
	}
	if sc := res.StatusCode; sc > 299 {
		if sc == 404 {
			return unmarshaled, errors.New("invalid location-area-name (possible spelling mistakes)")
		}
		errorMessage := "response status code: " + fmt.Sprintf("%d", sc)
		return unmarshaled, errors.New(errorMessage)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return unmarshaled, err
	}

	err = json.Unmarshal(body, &unmarshaled)
	if err != nil {
		return unmarshaled, err
	}

	if len(unmarshaled.PokemonEncounters) == 0 {
		var url string
		querySeparated := strings.Split(prev, "?")
		if len(querySeparated) == 1 {
			url = fmt.Sprintf("%s/%s", next, endpoint)
		} else {
			url = fmt.Sprintf("%s/%s?%s", querySeparated[0], endpoint, querySeparated[1])
		}
		res, err := http.Get(url)
		if err != nil {
			return unmarshaled, err
		}
		if sc := res.StatusCode; sc > 299 {
			if sc == 404 {
				return unmarshaled, errors.New("invalid location-area-name (possible spelling mistakes)")
			}
			errorMessage := "response status code: " + fmt.Sprintf("%d", sc)
			return unmarshaled, errors.New(errorMessage)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return unmarshaled, err
		}
		cache.Add(endpoint, body)
		res.Body.Close()

		var unmarshaled locEndpoint
		err = json.Unmarshal(body, &unmarshaled)
		if err != nil {
			return unmarshaled, err
		}

		return unmarshaled, nil
	}

	cache.Add(endpoint, body)
	res.Body.Close()

	return unmarshaled, nil
}

func catchPokemon(c *config, name ...string) error {
	if len(name) < 1 {
		return errors.New("check the string passed into the function")
	}
	isCaught, err := pokemon.IsCaught(name[0])
	if err != nil {
		return err
	}
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣤⣶⣶⣿⣿⣿⣿⣿⣶⣶⣤⣄⡀⠀⠀⠀⠀⠀⠀⠀\n⠀⠀⠀⠀⠀⠀⣠⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣄⠀⠀⠀⠀⠀\n⠀⠀⠀⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⡄⠀⠀⠀\n⠀⠀⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡏⠀⠀⠙⣿⣿⣿⣿⣿⣆⠀⠀\n⠀⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠿⠿⢿⣧⡀⠀⢠⣿⠟⠛⠛⠿⣿⡆⠀\n⠀⢰⣿⣿⣿⣿⣿⣿⠿⠟⠋⠉⠁⠀⠀⠀⠀⠀⠙⠿⠿⠟⠋⠀⠀⠀⣠⣿⠇⠀\n⠀⢸⣿⣿⡿⠟⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣤⣾⠟⠋⠀⠀\n⠀⢸⣿⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⣤⣴⣾⠿⠛⠉⠀⠀⠀⠀⠀\n⠀⠈⢿⣷⣤⣤⣄⣠⣤⣤⣤⣤⣶⣶⣾⠿⠿⠛⠛⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀\n⠀⢠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣦⣤⣀⠀⠀⠀⠀⠀⠀⠀⠀\n⠀⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣦⣄⠀⠀⠀⠀\n⠀⢸⣿⡛⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⡀⠀\n⠀⠀⢻⣧⠀⠈⠙⠛⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀\n⠀⠀⠈⢿⣧⠀⠀⠀⠀⠀⠀⠉⠙⠛⠻⠿⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠁⠀\n⠀⠀⠀⠀⠻⣷⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠹⣿⣿⣿⣿⠟⠀⣠⣾⠟⠀⠀⠀\n⠀⠀⠀⠀⠀⠈⠻⣷⣦⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠉⢀⣤⣾⠟⠁⠀⠀⠀⠀\n⠀⠀⠀⠀⠀⠀⠀⠀⠙⠻⠿⣶⣦⣤⣤⣤⣤⣤⣤⣶⡿⠟⠋⠁⠀⠀⠀⠀⠀⠀\n⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠉⠉⠉⠉⠉⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀\n\n\n")
	time.Sleep(time.Second * 1)
	fmt.Printf("Catching %s ", name[0])
	time.Sleep(time.Millisecond * 750)
	fmt.Printf(". ")
	time.Sleep(time.Millisecond * 750)
	fmt.Printf(". ")
	time.Sleep(time.Millisecond * 750)
	fmt.Printf(".\n")
	time.Sleep(time.Second * 1)
	if isCaught {
		fmt.Printf("%s was caught and added to your Pokedex\n", name[0])
		return nil
	}
	fmt.Printf("%s managed to not get caught\n", name[0])
	return nil
}

func inspectPokemon(c *config, name ...string) error {
	pokemon, err := pokemon.Pokemons.Get(name[0])
	if err != nil {
		return err
	}
	fmt.Printf("Pokemon: %s\n", pokemon.Name)
	fmt.Printf("Height: %d | Weight: %d\n", pokemon.Height, pokemon.Weight)
	fmt.Println("==TYPES==")
	for _, pokemonType := range pokemon.Types {
		fmt.Println(pokemonType.Type.Name)
	}
	fmt.Println("==STATS==")
	for _, stat := range pokemon.Stats {
		fmt.Printf("%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	return err
}

func pokedex(c *config, name ...string) error {
	return pokemon.Pokemons.Print()
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
