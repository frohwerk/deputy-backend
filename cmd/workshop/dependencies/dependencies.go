package dependencies

import (
	"fmt"

	"github.com/frohwerk/deputy-backend/internal/util"
)

type Store interface {
	Direct(id string) ([]string, error)
}

type Lookup struct {
	Store
}

type collector struct {
	*Lookup
	depth int
	found util.Set
}

func (r *Lookup) Transitive(id string) ([]string, error) {
	c := &collector{Lookup: r, found: make(util.Set)}
	err := c.all(id)
	if err != nil {
		return nil, err
	}
	return c.result(), nil
}

func (c *collector) result() []string {
	values := []string{}
	for key := range c.found {
		values = append(values, key)
	}
	return values
}

func (c *collector) all(id string) error {
	c.depth++
	defer func() { c.depth-- }()

	if c.depth > 20 {
		return fmt.Errorf("too many recursions, please check if there are circular relationships in your model")
	}

	deps, err := c.Direct(id)
	if err != nil {
		return err
	}

	for _, dep := range deps {
		c.found.Put(dep)
		c.all(dep)
	}

	return nil
}

func (c *collector) direct(id string) ([]string, error) {
	if c.Store == nil {
		return []string{}, nil
	}
	return c.Direct(id)
}
