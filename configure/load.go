package configure

import (
	"encoding/json"

	"github.com/google/go-jsonnet"
)

func loadObject(filename string, obj any) (e error) {
	vm := jsonnet.MakeVM()
	str, e := vm.EvaluateFile(filename)
	if e != nil {
		return
	}
	e = json.Unmarshal([]byte(str), obj)
	return
}
