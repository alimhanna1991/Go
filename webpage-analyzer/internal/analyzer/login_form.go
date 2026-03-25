package analyzer

import (
	"strings"

	"golang.org/x/net/html"
)

// LoginFormDetector handles login form detection
type LoginFormDetector struct{}

// NewLoginFormDetector creates a new login form detector
func NewLoginFormDetector() *LoginFormDetector {
	return &LoginFormDetector{}
}

// Detect detects if the page contains a login form
func (d *LoginFormDetector) Detect(doc *html.Node, pageURL, pageTitle string) bool {
	pageURL = strings.ToLower(pageURL)
	pageTitle = strings.ToLower(pageTitle)
	fullText := strings.ToLower(nodeText(doc))

	if d.hasLoginFormElement(doc) {
		return true
	}

	return d.matchesAuthPageHeuristics(pageURL, pageTitle, fullText, doc)
}

func (d *LoginFormDetector) hasLoginFormElement(doc *html.Node) bool {
	var found bool
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if found {
			return
		}
		if n.Type == html.ElementNode && n.Data == "form" && d.isLoginForm(n) {
			found = true
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return found
}

func (d *LoginFormDetector) matchesAuthPageHeuristics(pageURL, pageTitle, pageText string, doc *html.Node) bool {
	signals := d.collectSignals(doc)
	authHint := looksLikeAuthRoute(pageURL) || looksLikeLoginText(pageTitle) || looksLikeLoginText(pageText)

	switch {
	case signals.hasPassword && (signals.hasIdentity || signals.hasSubmitHint || authHint):
		return true
	case authHint && signals.hasIdentity && signals.hasSubmitHint:
		return true
	case looksLikeAuthRoute(pageURL) && (signals.hasIdentity || signals.hasPassword || signals.hasSubmitHint):
		return true
	default:
		return false
	}
}

func (d *LoginFormDetector) isLoginForm(formNode *html.Node) bool {
	var hasPassword bool
	var hasIdentity bool
	var hasSubmitHint bool

	actionText := strings.ToLower(attributeValue(formNode, "action"))
	formText := strings.ToLower(nodeText(formNode))
	if looksLikePasswordReset(actionText) || looksLikePasswordReset(formText) {
		return false
	}

	if looksLikeLoginText(actionText) {
		hasSubmitHint = true
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type != html.ElementNode {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				traverse(c)
			}
			return
		}

		switch n.Data {
		case "input":
			inputType := strings.ToLower(attributeValue(n, "type"))
			inputName := strings.ToLower(attributeValue(n, "name"))
			inputID := strings.ToLower(attributeValue(n, "id"))
			inputPlaceholder := strings.ToLower(attributeValue(n, "placeholder"))

			if inputType == "password" {
				hasPassword = true
			}

			if inputType == "email" ||
				inputType == "text" ||
				strings.Contains(inputName, "user") ||
				strings.Contains(inputName, "email") ||
				strings.Contains(inputName, "login") ||
				strings.Contains(inputID, "user") ||
				strings.Contains(inputID, "email") ||
				strings.Contains(inputPlaceholder, "email") ||
				strings.Contains(inputPlaceholder, "user") {
				hasIdentity = true
			}

			if inputType == "submit" && looksLikeLoginText(strings.ToLower(attributeValue(n, "value"))) {
				hasSubmitHint = true
			}
		case "button":
			if looksLikeLoginText(strings.ToLower(nodeText(n))) {
				hasSubmitHint = true
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(formNode)
	return hasPassword && (hasIdentity || hasSubmitHint || (!looksLikePasswordReset(actionText) && !looksLikePasswordReset(formText)))
}

func attributeValue(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func nodeText(node *html.Node) string {
	var parts []string
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.TextNode {
			trimmed := strings.TrimSpace(n.Data)
			if trimmed != "" {
				parts = append(parts, trimmed)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(node)
	return strings.Join(parts, " ")
}

func looksLikeLoginText(value string) bool {
	loginHints := []string{
		"log in",
		"login",
		"sign in",
		"signin",
		"sign-in",
		"sign into",
		"log into",
		"authenticate",
		"authentication",
		"continue with email",
		"continue with username",
	}

	for _, hint := range loginHints {
		if strings.Contains(value, hint) {
			return true
		}
	}

	return false
}

func looksLikePasswordReset(value string) bool {
	resetHints := []string{
		"reset password",
		"forgot password",
		"recover password",
		"change password",
	}

	for _, hint := range resetHints {
		if strings.Contains(value, hint) {
			return true
		}
	}

	return false
}

func looksLikeAuthRoute(value string) bool {
	routeHints := []string{
		"/auth",
		"/login",
		"/signin",
		"/sign-in",
		"/account/login",
	}

	for _, hint := range routeHints {
		if strings.Contains(value, hint) {
			return true
		}
	}

	return false
}

type loginSignals struct {
	hasPassword   bool
	hasIdentity   bool
	hasSubmitHint bool
}

func (d *LoginFormDetector) collectSignals(doc *html.Node) loginSignals {
	signals := loginSignals{}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "input":
				inputType := strings.ToLower(attributeValue(n, "type"))
				inputName := strings.ToLower(attributeValue(n, "name"))
				inputID := strings.ToLower(attributeValue(n, "id"))
				inputPlaceholder := strings.ToLower(attributeValue(n, "placeholder"))
				inputValue := strings.ToLower(attributeValue(n, "value"))

				if inputType == "password" {
					signals.hasPassword = true
				}

				if inputType == "email" ||
					inputType == "text" ||
					strings.Contains(inputName, "user") ||
					strings.Contains(inputName, "email") ||
					strings.Contains(inputName, "login") ||
					strings.Contains(inputID, "user") ||
					strings.Contains(inputID, "email") ||
					strings.Contains(inputPlaceholder, "email") ||
					strings.Contains(inputPlaceholder, "user") ||
					strings.Contains(inputPlaceholder, "username") {
					signals.hasIdentity = true
				}

				if inputType == "submit" && looksLikeLoginText(inputValue) {
					signals.hasSubmitHint = true
				}
			case "button", "a":
				if looksLikeLoginText(strings.ToLower(nodeText(n))) {
					signals.hasSubmitHint = true
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return signals
}
