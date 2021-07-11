package gen

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/codemicro/cligen/internal/parse"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"unicode"
)

var (
	usedIds  = make(map[string]struct{})
	replacer = strings.NewReplacer(
		"0", "a",
		"1", "b",
		"2", "c",
		"3", "d",
		"4", "e",
		"5", "f",
		"6", "g",
		"7", "h",
		"8", "i",
		"9", "j",
		"A", "k",
		"B", "l",
		"C", "m",
		"D", "n",
		"E", "o",
		"F", "p",
	)
	rng = rand.New(rand.NewSource(45986749679038456)) // so the same IDs always appear in the same sequence
)

func nextIdentifier(exported ...bool) string {
	alreadyInUse := true
	var x string
	for alreadyInUse {

		x = replacer.Replace(strconv.FormatInt(rng.Int63n(100000), 16)) // ensure string has no digits in it

		transformFunc := unicode.ToLower
		if len(exported) > 0 && exported[0] {
			transformFunc = unicode.ToUpper
		}

		x = string(transformFunc(rune(x[0]))) + x[1:]
		_, alreadyInUse = usedIds[x]
	}
	return x
}

type generator struct {
	b *bytes.Buffer
}

func newGenerator() *generator {
	return &generator{
		b: new(bytes.Buffer),
	}
}

func (g *generator) w(x string, args ...interface{}) {
	g.b.WriteString(fmt.Sprintf(x, args...) + "\n")
}

func (g *generator) ifErrNotNil(x string) {
	if x == "" {
		x = "err"
	}
	g.b.WriteString("if " + x + " != nil {\nreturn " + x + "\n}\n")
}

func (g *generator) returnError(x string, args ...interface{}) {
	g.b.WriteString("return errors.New(\"" + fmt.Sprintf(x, args...) + "\")")
}

func File(filename string, packageName string, functions map[string]*parse.Function) (io.Reader, error) {

	g := newGenerator()

	g.w("package %s", packageName)

	imports := []string{"errors", "strings", "github.com/codemicro/cligen/parsecli", "math/bits", "strconv"}
	g.w("import (")
	for _, i := range imports {
		g.w(`"%s"`, i)
	}
	g.w(")")

	g.w("var intSize = bits.UintSize")
	g.w("var funcNames = map[string]string{")
	for funcName := range functions {
		g.w(`"%s": "%s",`, strings.ToLower(funcName), funcName)
	}
	g.w("}")

	g.w("func Start(input []string) error {")

	g.w("if len(input) == 0 {")
	g.returnError("not enough arguments")
	g.w("}")

	g.w("var runFunc string")

	g.w(`if fname, ok := funcNames[strings.ToLower(input[0])]; !ok {`)
	g.returnError("no matching targets found")
	g.w("} else {")
	g.w("runFunc = fname")
	g.w("}")

	g.w("parsedFlags, parsedArgs, err := parsecli.Slice(input[1:])")
	g.ifErrNotNil("")

	g.w("switch runFunc {")

	for f, finfo := range functions {

		g.w(`case "%s":`, f)

		if err := checkArgs(g, finfo.Signature); err != nil {
			return nil, err
		}
		if err := callFunc(g, finfo); err != nil {
			return nil, err
		}

		g.w("return nil")
	}

	g.w("}")

	g.w("return nil")
	g.w("}")

	return g.b, nil
}

func checkArgs(g *generator, sig *parse.Signature) error {

	var numArgs int
	for _, arg := range sig.Argument {
		if !arg.IsPointer {
			numArgs += 1
			continue
		}
	}

	g.w(`if len(parsedArgs) < %d {`, numArgs)
	g.returnError("not enough arguments")
	g.w("}")

	return nil
}

func callFunc(g *generator, f *parse.Function) error {

	var varIDs []string
	var currentArgIndex int

	getSource := func(param *parse.Param) string {
		if param.IsPointer {
			return `parsedFlags["` + param.Name + `"]`
		} else {
			source := fmt.Sprintf("parsedArgs[%d]", currentArgIndex)
			currentArgIndex += 1
			return source
		}
	}

	checkProvided := func(g *generator, param *parse.Param, id string, block []byte) {
		if !param.IsPointer {
			g.b.Write(block)
			return
		}
		g.w(`if _, ok := parsedFlags["%s"]; !ok {`, param.Name)
		g.w("%s = nil", id)
		g.w("} else {")
		g.b.Write(block)
		g.w("}")

	}

	writeVar := func(g *generator, arg *parse.Param, id, t string) {
		var pChar string
		if arg.IsPointer {
			pChar = "*"
		}
		g.w("var %s %s%s", id, pChar, t)
	}

	numConv := func(g *generator, arg *parse.Param, source, id, t string, f string) {
		writeVar(g, arg, id, t)
		x := newGenerator()
		tempID := nextIdentifier()
		x.w("%s, err := strconv.%s(%s, 10, intSize)", tempID, f, source)
		x.ifErrNotNil("")

		if arg.IsPointer {
			y := nextIdentifier()
			x.w("%s := %s(%s)", y, t, tempID)
			x.w("%s = &%s", id, y)
		} else {
			x.w("%s = %s(%s)", id, t, tempID)
		}
		checkProvided(g, arg, id, x.b.Bytes())
	}

	for _, arg := range f.Signature.Argument {
		id := nextIdentifier()
		varIDs = append(varIDs, id)

		source := getSource(arg)

		switch arg.Type {
		case "int":
			numConv(g, arg, source, id, "int", "ParseInt")

		case "uint":
			numConv(g, arg, source, id, "uint", "ParseUint")

		case "float32":
			writeVar(g, arg, id, "float32")

			x := newGenerator()

			tempID := nextIdentifier()
			x.w("%s, err := strconv.ParseFloat(%s, 32)", tempID, source)
			x.ifErrNotNil("")

			if arg.IsPointer {
				y := nextIdentifier()
				x.w("%s := float32(%s)", y, tempID)
				x.w("%s = &%s", id, y)
			} else {
				x.w("%s = float32(%s)", id, tempID)
			}

			checkProvided(g, arg, id, x.b.Bytes())

		case "bool":
			writeVar(g, arg, id, "bool")

			x := newGenerator()

			var t string
			if arg.IsPointer {
				t = nextIdentifier()
			} else {
				t = id
			}

			x.w("%s, err := strconv.ParseBool(%s)", t, source)
			x.ifErrNotNil("")

			if arg.IsPointer {
				y := nextIdentifier()
				x.w("%s := %s", y, t)
				x.w("%s = &%s", id, y)
			}

			checkProvided(g, arg, id, x.b.Bytes())

		case "string":
			writeVar(g, arg, id, "string")

			x := newGenerator()

			if arg.IsPointer {
				y := nextIdentifier()
				x.w("%s := %s", y, source)
				x.w("%s = &%s", id, y)
			} else {
				x.w("%s = %s", id, source)
			}

			checkProvided(g, arg, id, x.b.Bytes())
		}

	}

	var returns []string
	var errID string

	for _, arg := range f.Signature.Return {
		var x string
		if arg.Type == "error" {
			if errID != "" {
				return errors.New("cannot have more than one error return")
			}
			x = nextIdentifier()
			errID = x
		} else {
			x = "_"
		}
		returns = append(returns, x)
	}

	var returnBlock string
	if len(returns) != 0 {
		returnBlock = strings.Join(returns, ", ") + " := "
	}

	g.w("%s%s(%s)", returnBlock, f.Name, strings.Join(varIDs, ", "))

	if len(returns) != 0 {
		g.ifErrNotNil(errID)
	}

	return nil
}
