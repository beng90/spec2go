package main

import (
	"fmt"
	"gopkg.in/go-playground/validator.v9"
)

const (
	fieldErrMsg = "Field validation for '%s' failed on the '%s' rule\n"
)

type ValidationErrors map[string]error

func main() {
	v := validator.New()

	params := make(map[string]string)
	params = map[string]string{
		"email":        "joeybloggs",
		"phone":        "123asd",
		"itemsPerPage": "11",
	}

	errs := ValidationErrors{
		"email": v.Var(params["email"], "required,gt=15,lt=20,email"),
		//"phone": v.Var(params["phone"], "required,numeric").(validator.ValidationErrors),
		//"itemsPerPage": v.Var(params["itemsPerPage"], "required,oneof=10 20 30").(validator.ValidationErrors),
	}

	//fmt.Println(errs)

	for fieldName, fieldErrors := range errs {
		switch fieldErrors.(type) {
		case validator.ValidationErrors:
			//fmt.Println("err", fieldName, fieldErrors)
			for _, verr := range fieldErrors.(validator.ValidationErrors) {
				//fmt.Println("err", verr)
				fmt.Printf(fieldErrMsg, fieldName, verr.Tag())
			}
		case *validator.InvalidValidationError:
			fmt.Println("err", fieldName, fieldErrors)
		}

		//fmt.Printf("%T\n", fieldErrors)
		//}
	}
}
