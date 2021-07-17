package parse

import (
	"errors"
	"fmt"
	"go/ast"
	"regexp"
	"strings"
)

const DirectiveStart = "cligen"

var (
	directiveRegexp = regexp.MustCompile("^(?://)?" + DirectiveStart + ":[A-Za-z]")

	errorNoDirective = errors.New("no directive found")
)

func getDirectives(commentGroup *ast.CommentGroup) ([]string, error) {
	var matches []string
	for _, comment := range commentGroup.List {
		if directiveRegexp.MatchString(comment.Text) {
			matches = append(matches, comment.Text)
		}
	}

	if len(matches) == 0 {
		return nil, errorNoDirective
	}

	for i, directive := range matches {
		x := strings.TrimPrefix(directive, "//")
		x = strings.TrimPrefix(x, DirectiveStart + ":")
		matches[i] = x
	}

	return matches, nil
}

func applyDirectives(function *Function) error {
	fmt.Println("---")
	for _, directive := range function.Directives {

		split := strings.Split(directive, " ")
		if len(split) == 0 {
			continue
		}
		opcode, split := split[0], split[1:]
		fmt.Println(opcode, split)

		switch opcode {
		case "cmd":
			if len(split) >= 1 {
				function.UIName = split[0]
			}
		case "rename":
			if len(split) < 2 {
				return errors.New("rename directive missing arguments for old and new argument names")
			}

			searchFor := split[0]
			renameTo := split[1]

			for _, sig := range function.Signature.Argument {
				if strings.EqualFold(searchFor, sig.Name) {
					sig.Name = renameTo
					break
				}
			}
		case "description":
			if len(split) == 0 {
				return errors.New("description directive missing description")
			}
			function.Description = strings.Join(split, " ")
		}

	}
	return nil
}