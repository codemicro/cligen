package parse

import (
	"errors"
	"go/ast"
	"regexp"
	"strings"
)

const DirectiveStart = "cligen"

var (
	directiveRegexp = regexp.MustCompile("^(?://)?" + DirectiveStart + ":[A-Za-z]")

	errorNoDirective = errors.New("no directive found")
	errorTooManyDirectives = errors.New("more than one directive found")
)

func getDirective(commentGroup *ast.CommentGroup) (string, error) {
	var matches []string
	for _, comment := range commentGroup.List {
		if directiveRegexp.MatchString(comment.Text) {
			matches = append(matches, comment.Text)
		}
	}

	if len(matches) == 0 {
		return "", errorNoDirective
	}
	if len(matches) > 1 {
		return "", errorTooManyDirectives
	}

	var formatted string
	formatted = strings.TrimPrefix(matches[0], "//")
	formatted = strings.TrimPrefix(formatted, DirectiveStart + ":")

	return formatted, nil
}