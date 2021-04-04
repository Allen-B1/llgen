package main

import (
	"flag"
	"fmt"
	"github.com/allen-b1/llgen/parser"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

var showTree bool

func init() {
	flag.BoolVar(&showTree, "tree", false, "whether to print tree or not")
}

func print(n interface{}) string {
	if tok, ok := n.(parser.Token); ok {
		return tok.Type + "<" + tok.Data + ">"
	}

	val := reflect.ValueOf(n)
	str := val.Type().Name()
	if val.Type().Kind() == reflect.Slice {
		str = "[]" + val.Type().Elem().Name()
		for i := 0; i < val.Len(); i++ {
			str += "\n\t" + fmt.Sprint(i) + ": " + strings.Replace(print(val.Index(i).Interface()), "\n", "\n\t", -1)
		}
	} else {
		for i := 0; i < val.Type().NumField(); i++ {
			str += "\n\t" + val.Type().Field(i).Name + ": " + strings.Replace(print(val.Field(i).Interface()), "\n", "\n\t", -1)
		}
	}
	return str
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: llgen [FLAGS] FILE\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.Arg(0) == "" {
		flag.Usage()
		os.Exit(1)
	}

	body, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	tokens, err := tokenize(string(body))
	if err != nil {
		panic(err)
	}

	a, n, err := parser.ParseStatements(tokens)
	if err != nil {
		panic(err)
	}
	if n != len(tokens) {
		panic("invalid document")
	}

	if showTree {
		fmt.Println(print(a))
	} else {
		res, err := generateAll(a)
		if err != nil {
			panic(err)
		}

		fmt.Println(res)
	}
}
