package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/J4NN0/mycel/internal/logger"
	"github.com/maximhq/bifrost/core/schemas"
	"golang.org/x/net/html"
)

const (
	fetchName = "fetch_url"
	fetchDesc = "Read a web page and return its text. Use it when the user gives you a link to look at, or to read a result web_search returned, and you need what the page actually says rather than a snippet. It only reads: it cannot log in, fill in a form, or reach anything on the user's own machine or private network."

	fetchTimeout  = 20 * time.Second
	dialTimeout   = 10 * time.Second
	fetchMaxBytes = 4 << 20
	fetchMaxChars = 6000

	userAgent = "Mycel (+https://github.com/J4NN0/mycel)"
)

var _ Tool = (*Fetch)(nil)

type fetchArgs struct {
	URL string `json:"url"`
}

type Fetch struct {
	log    logger.Logger
	client *http.Client
}

func NewFetchURL(log logger.Logger) Tool {
	log.Debugf("Tool loaded: %s", fetchName)

	return &Fetch{
		log:    log,
		client: publicHTTPClient(fetchTimeout),
	}
}

func (t *Fetch) Info() (string, string) {
	return fetchName, fetchDesc
}

func (t *Fetch) Definition() schemas.ChatTool {
	const urlDesc = "Absolute http or https URL to read, exactly as the user gave it or exactly as web_search returned it. Never invent a URL and never guess one."

	props := schemas.NewOrderedMapFromPairs(
		schemas.Pair{Key: "url", Value: map[string]string{"type": "string", "description": urlDesc}},
	)

	return schemas.ChatTool{
		Type: schemas.ChatToolTypeFunction,
		Function: &schemas.ChatToolFunction{
			Name:        fetchName,
			Description: schemas.Ptr(fetchDesc),
			Parameters: &schemas.ToolFunctionParameters{
				Type:       "object",
				Properties: props,
				Required:   []string{"url"},
			},
		},
	}
}

func (t *Fetch) Execute(ctx context.Context, raw json.RawMessage) (string, error) {
	var a fetchArgs
	err := json.Unmarshal(raw, &a)
	if err != nil {
		return "", fmt.Errorf("parse fetch args: %w", err)
	}

	target, err := readableURL(a.URL)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", fmt.Errorf("build request for %s: %w", target, err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,text/plain;q=0.9,*/*;q=0.1")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not read %s: %w", target, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("could not read %s: the site answered %s", target, resp.Status)
	}

	mediaType := responseMediaType(resp)
	if !isReadable(mediaType) {
		return "", fmt.Errorf("%s is %s, which is not a document that can be read as text", target, mediaType)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, fetchMaxBytes))
	if err != nil {
		return "", fmt.Errorf("could not read the body of %s: %w", target, err)
	}

	// A redirect chain means the page that answered is not always the one asked for.
	final := resp.Request.URL.String()
	t.log.Debugf("Fetched %s (%d bytes, %s)", final, len(body), mediaType)

	return page(final, mediaType, body)
}

func page(finalURL, mediaType string, body []byte) (string, error) {
	var title, text string

	if strings.Contains(mediaType, "html") {
		var err error
		title, text, err = htmlToText(strings.NewReader(string(body)))
		if err != nil {
			return "", fmt.Errorf("could not read the HTML of %s: %w", finalURL, err)
		}
	} else {
		text = collapseLines(string(body))
	}

	if text == "" {
		return "", fmt.Errorf("%s has no readable text: it may be built entirely in JavaScript", finalURL)
	}

	var sb strings.Builder
	if title != "" {
		fmt.Fprintf(&sb, "Title: %s\n", title)
	}
	fmt.Fprintf(&sb, "URL: %s\n\n%s", finalURL, truncate(text, fetchMaxChars))

	return sb.String(), nil
}

func readableURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("%q is not a valid URL: ask the user for the real link rather than guessing one", raw)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("%q is not an http or https URL, so there is no page to read", raw)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%q has no host: a full URL such as https://example.com/page is needed", raw)
	}

	return u.String(), nil
}

func responseMediaType(resp *http.Response) string {
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return "text/html"
	}
	return mediaType
}

func isReadable(mediaType string) bool {
	return strings.HasPrefix(mediaType, "text/") ||
		mediaType == "application/json" ||
		mediaType == "application/xml" ||
		strings.HasSuffix(mediaType, "+json") ||
		strings.HasSuffix(mediaType, "+xml")
}

// publicHTTPClient refuses to dial the machine's own networks. Anyone who can message the agent can
// suggest a URL, and so can a page it has just read, so the address is checked at dial time: that
// covers redirects and hostnames that resolve to a private address, and keeps a fetch away from
// Ollama, Redis and anything else listening locally.
func publicHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: dialTimeout}
	dialer.Control = func(_, address string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return fmt.Errorf("unexpected address %q: %w", address, err)
		}

		ip := net.ParseIP(host)
		if ip == nil {
			return fmt.Errorf("unexpected address %q", address)
		}
		if !isPublicIP(ip) {
			return fmt.Errorf("refusing to connect to %s: it is on a local or private network", ip)
		}

		return nil
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{DialContext: dialer.DialContext},
	}
}

func isPublicIP(ip net.IP) bool {
	return !ip.IsLoopback() &&
		!ip.IsPrivate() &&
		!ip.IsUnspecified() &&
		!ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() &&
		!ip.IsInterfaceLocalMulticast() &&
		!ip.IsMulticast()
}

var skippedTags = map[string]bool{
	"script":   true,
	"style":    true,
	"noscript": true,
	"template": true,
	"svg":      true,
	"iframe":   true,
}

var blockTags = map[string]bool{
	"p": true, "div": true, "br": true, "hr": true, "section": true, "article": true,
	"header": true, "footer": true, "nav": true, "aside": true, "main": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"ul": true, "ol": true, "li": true, "table": true, "tr": true, "blockquote": true, "pre": true,
}

// htmlToText is a tokenizer walk, not a readability pass: menus and boilerplate survive it, so
// callers should expect to truncate what comes back.
func htmlToText(r io.Reader) (title, text string, err error) {
	var (
		tokenizer = html.NewTokenizer(r)
		heading   strings.Builder
		body      strings.Builder
		inTitle   bool
		skipping  string
	)

	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			err = tokenizer.Err()
			if errors.Is(err, io.EOF) {
				return strings.TrimSpace(heading.String()), collapseLines(body.String()), nil
			}
			return "", "", err

		case html.StartTagToken, html.SelfClosingTagToken:
			if skipping != "" {
				continue
			}
			name, _ := tokenizer.TagName()
			switch tag := string(name); {
			case skippedTags[tag]:
				skipping = tag
			case tag == "title":
				inTitle = true
			case blockTags[tag]:
				body.WriteString("\n")
			}

		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			tag := string(name)
			if skipping != "" {
				if skipping == tag {
					skipping = ""
				}
				continue
			}
			if tag == "title" {
				inTitle = false
				continue
			}
			if blockTags[tag] {
				body.WriteString("\n")
			}

		case html.TextToken:
			if skipping != "" {
				continue
			}
			text := collapseSpaces(string(tokenizer.Text()))
			if inTitle {
				heading.WriteString(text)
				continue
			}
			body.WriteString(text)
		}
	}
}

// collapseSpaces turns every run of whitespace into one space, keeping the leading and trailing one
// that separates the text of neighbouring inline elements. Line breaks come from block tags instead.
func collapseSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	pending := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			pending = true
			continue
		}
		if pending {
			b.WriteRune(' ')
			pending = false
		}
		b.WriteRune(r)
	}
	if pending {
		b.WriteRune(' ')
	}

	return b.String()
}

func collapseLines(s string) string {
	lines := strings.Split(s, "\n")
	kept := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" && (len(kept) == 0 || kept[len(kept)-1] == "") {
			continue
		}
		kept = append(kept, line)
	}

	return strings.TrimSpace(strings.Join(kept, "\n"))
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "\n\n[…truncated: the page continues past this point]"
}
