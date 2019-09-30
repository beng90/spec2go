package validate

import (
	"log"
)

var IsDebugMode = false

func debug(v ...interface{}) {
	if IsDebugMode == true {
		log.Println(v...)
	}
}
