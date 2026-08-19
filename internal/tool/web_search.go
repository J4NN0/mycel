package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/J4NN0/mycel/internal/logger"
	"github.com/maximhq/bifrost/core/schemas"
)

const (
	searchName = "web_search"
	searchDesc = "Search the web and get back a short list of results, each with a title, a link and a snippet. Use it when an answer depends on something current, changing, or outside what you already know — news, prices, releases, documentation, anything that may have moved on since you were trained. The snippets are not the page: call fetch_url on a result when you need what it actually says. Do not call it for something you can already answer, and never call it just to test it."

	searchTimeout       = 20 * time.Second
	searchHealthTimeout = 3 * time.Second
	searchHealthPath    = "/healthz"

	searchMaxResults  = 5
	searchMaxSnippet  = 300
	searchMaxBodySize = 1 << 20
)

var _ Tool = (*Search)(nil)

type searchArgs struct {
	Query string `json:"query"`
}

type searxngResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

type Search struct {
	log     logger.Logger
	client  *http.Client
	baseURL string
}

func NewWebSearch(log logger.Logger, searxngURL string) Tool {
	if searxngURL == "" {
		log.Debugf("Tool skipped: %s (SEARXNG_URL not set)", searchName)
		return nil
	}

	u, err := url.Parse(strings.TrimSpace(searxngURL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		log.Errorf("Tool skipped: %s (SEARXNG_URL %q is not an http or https URL)", searchName, searxngURL)
		return nil
	}

	s := &Search{
		log:     log,
		client:  &http.Client{Timeout: searchTimeout},
		baseURL: strings.TrimSuffix(u.String(), "/"),
	}

	err = s.reachable()
	if err != nil {
		log.Warningf("Tool skipped: %s (nothing answering at %s: %v). Start it with: docker compose up -d searxng", searchName, s.baseURL, err)
		return nil
	}
	s.log.Debugf("Tool loaded: %s (via %s)", searchName, s.baseURL)

	return s
}

// reachable keeps an unreachable instance from being advertised to the model as a working tool,
// which matters because SEARXNG_URL has a default: the container may simply not be running.
func (t *Search) reachable() error {
	ctx, cancel := context.WithTimeout(context.Background(), searchHealthTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.baseURL+searchHealthPath, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s answered %s", searchHealthPath, resp.Status)
	}

	return nil
}

func (t *Search) Info() (string, string) {
	return searchName, searchDesc
}

func (t *Search) Definition() schemas.ChatTool {
	const queryDesc = "What to search for, in plain words. Keep it short and keyword-like, the way a search box expects — not a sentence addressed to a person."

	props := schemas.NewOrderedMapFromPairs(
		schemas.Pair{Key: "query", Value: map[string]string{"type": "string", "description": queryDesc}},
	)

	return schemas.ChatTool{
		Type: schemas.ChatToolTypeFunction,
		Function: &schemas.ChatToolFunction{
			Name:        searchName,
			Description: schemas.Ptr(searchDesc),
			Parameters: &schemas.ToolFunctionParameters{
				Type:       "object",
				Properties: props,
				Required:   []string{"query"},
			},
		},
	}
}

func (t *Search) Execute(ctx context.Context, raw json.RawMessage) (string, error) {
	var a searchArgs
	err := json.Unmarshal(raw, &a)
	if err != nil {
		return "", fmt.Errorf("parse search args: %w", err)
	}

	query := strings.TrimSpace(a.Query)
	if query == "" {
		return "", fmt.Errorf("a search needs something to search for: ask the user what they want looked up")
	}

	ctx, cancel := context.WithTimeout(ctx, searchTimeout)
	defer cancel()

	body, err := t.request(ctx, query)
	if err != nil {
		return "", err
	}

	var results searxngResponse
	err = json.Unmarshal(body, &results)
	if err != nil {
		return "", fmt.Errorf("search backend did not answer with JSON: check that the SearXNG instance at %s has json in its search.formats setting", t.baseURL)
	}

	t.log.Debugf("Search %q returned %d result(s)", query, len(results.Results))

	return formatResults(query, results), nil
}

func (t *Search) request(ctx context.Context, query string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/search?%s", t.baseURL, url.Values{
		"q":      {query},
		"format": {"json"},
	}.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build search request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach the search backend at %s: %w", t.baseURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// SearXNG answers a disallowed output format, and a request its rate limiter rejects, with 403.
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("the search backend at %s refused the request: its json format or its limiter is blocking it", t.baseURL)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("the search backend at %s answered %s", t.baseURL, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, searchMaxBodySize))
	if err != nil {
		return nil, fmt.Errorf("could not read the search response: %w", err)
	}

	return body, nil
}

func formatResults(query string, results searxngResponse) string {
	if len(results.Results) == 0 {
		return fmt.Sprintf("No results for %q. Try different words, or tell the user nothing came back.", query)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Results for %q:\n\n", query)

	for i, r := range results.Results {
		if i == searchMaxResults {
			break
		}
		fmt.Fprintf(&sb, "%d. %s\n   %s\n", i+1, strings.TrimSpace(r.Title), r.URL)
		if snippet := clip(r.Content, searchMaxSnippet); snippet != "" {
			fmt.Fprintf(&sb, "   %s\n", snippet)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("These are snippets, not the pages themselves. Call fetch_url on whichever result looks right if you need more than this.")

	return sb.String()
}

func clip(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")

	runes := []rune(s)
	if len(runes) <= max {
		return s
	}

	return string(runes[:max]) + "…"
}
