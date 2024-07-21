package pokemon

import (
	"errors"
	"fmt"
	"sync"
)

type Pokedex struct {
	mu   *sync.Mutex
	List map[string]PokemonEndpoint
}

var Pokemons = Pokedex{
	mu:   &sync.Mutex{},
	List: make(map[string]PokemonEndpoint),
}

func (c *Pokedex) Add(name string, pokemon PokemonEndpoint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.List[name] = pokemon
}

func (c *Pokedex) Get(name string) (PokemonEndpoint, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	pokemon, ok := c.List[name]
	if !ok {
		return pokemon, errors.New(fmt.Sprintf("You don't have the pokemon: %s", name))
	}
	return pokemon, nil
}

func (c *Pokedex) Print() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.List) == 0 {
		return errors.New("You haven't caught any Pokemons...YET!")
	}
	fmt.Println("==Your Pokedex==")
	for k := range c.List {
		fmt.Println(k)
	}
	return nil
}
