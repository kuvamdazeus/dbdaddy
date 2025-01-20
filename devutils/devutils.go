package devutils

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "    ")
	fmt.Println(err, string(b))
}
