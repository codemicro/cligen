package main

import (
	"fmt"
	"github.com/codemicro/cligen/internal/gen"
	"github.com/codemicro/cligen/internal/parse"
	"io/ioutil"
)

func main() {
	info, err := parse.Directory("testdata/package")
	fmt.Printf("%#v %v\n", info, err)

	b, err := gen.File("hello.go", info.PackageName, info.Functions)
	fmt.Println(err)
	all, _ := ioutil.ReadAll(b)
	fmt.Println(string(all), err)

	_ = ioutil.WriteFile("testdata/package/runner.cligen.go", all, 0644)
}
