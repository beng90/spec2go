package validators

import (
	"fmt"
)

const isDebugMode = true

func debug(v ...interface{}) {
	if isDebugMode == true {
		fmt.Println("LOG: ", v)
	}
}
