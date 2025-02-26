package types

import (
	"encoding/json"
	"fmt"
)

func ExampleUnion2() {
	contents := []byte(`"hello"`)
	var u Union2[int, string]
	if err := json.Unmarshal(contents, &u); err != nil {
		panic(err)
	}

	fmt.Printf("%[1]T: %[1]v\n", u.CurrentValue())
	switch u.Selector {
	case 0:
		fmt.Println("It's an integer: ", u.V0)
	case 1:
		fmt.Println("It's a string: ", u.V1)
	}
	// Output:
	// string: hello
	// It's a string:  hello
}

func ExampleUnion3() {
	contents := []byte(`"hello"`)
	var u Union3[int, string, bool]
	if err := json.Unmarshal(contents, &u); err != nil {
		panic(err)
	}

	fmt.Printf("%[1]T: %[1]v\n", u.CurrentValue())
	switch u.Selector {
	case 0:
		fmt.Println("It's an integer: ", u.V0)
	case 1:
		fmt.Println("It's a string: ", u.V1)
	case 2:
		fmt.Println("It's a bool: ", u.V2)
	}
	// Output:
	// string: hello
	// It's a string:  hello
}
