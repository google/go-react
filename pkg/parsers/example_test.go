package parsers_test

import (
	"fmt"

	"github.com/google/go-react/pkg/parsers"
)

func ExampleJSONParser() {
	type Data struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	p := parsers.NewJSONParser[Data]()
	person, err := p.Parse(`{"name": "John", "age": 30}`)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", person)

	// Output:
	// {Name:John Age:30}
}
