package main

import (
	"github.com/beng90/spec2go/generate"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

func main() {
	file, err := ioutil.ReadFile("api.dereferenced.yml")
	if err != nil {
		panic(err)
	}

	validators := []generate.Validator{}
	data := yaml.MapSlice{}
	yaml.Unmarshal(file, &data)

	generate.Generate(&validators, data)

	//fmt.Println("validators", validators)

	templateFiles := []string{
		"validators.tpl",
	}

	t := template.Must(template.New("validators.tpl").ParseFiles(templateFiles...))

	f, err := os.Create("generated/validators.go")
	if err != nil {
		log.Println("create file: ", err)
		return
	}

	err = t.Execute(f, validators)
	if err != nil {
		log.Println("executing template:", err)

		os.Exit(1)
	}

	_ = f.Close()
}
