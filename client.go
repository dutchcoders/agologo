package agologo

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type summarize struct {
	URL string `json:"url"`
}

type Result struct {
	Photos  []string `json:"photos"`
	Summary []struct {
		Metadata struct {
			Icon   interface{} `json:"icon"`
			Origin string      `json:"origin"`
			Source string      `json:"source"`
			Url    string      `json:"url"`
		} `json:"metadata"`
		Quotes    []interface{} `json:"quotes"`
		Ranks     []int64       `json:"ranks"`
		Sentences []string      `json:"sentences"`
	} `json:"summary"`
	Title           string        `json:"title"`
	TitleCandidates []interface{} `json:"title_candidates"`
}

func Title(title string) ArticleOption {
	return func(v map[string]interface{}) {
		v["title"] = title
	}
}

func Text(text string) ArticleOption {
	return func(v map[string]interface{}) {
		v["text"] = text
	}
}

func URL(url string) ArticleOption {
	return func(v map[string]interface{}) {
		v["url"] = url
	}
}

type ArticleOption func(map[string]interface{})

type Summarize struct {
	Articles            []interface{} `json:"articles"`
	Coref               bool          `json:"coref"`
	IncludeAllSentences bool          `json:"include_all_sentences"`
	SortBySalience      bool          `json:"sort_by_salience"`
	SummaryLength       int64         `json:"summary_length"`
}

func Article(options ...ArticleOption) func(*Summarize) {
	return func(s *Summarize) {
		a := map[string]interface{}{}

		for _, fn := range options {
			fn(a)
		}

		s.Articles = append(s.Articles, a)
	}
}

func Coref() func(*Summarize) {
	return func(s *Summarize) {
		s.Coref = true
	}
}

func SortBySalience() func(*Summarize) {
	return func(s *Summarize) {
		s.SortBySalience = true
	}
}

func IncludeAllSentences() func(*Summarize) {
	return func(s *Summarize) {
		s.IncludeAllSentences = true
	}
}

func SummaryLength(v int) func(*Summarize) {
	return func(s *Summarize) {
		s.SummaryLength = int64(v)
	}
}

type SummarizeOption func(*Summarize)

func (c *client) Summarize(options ...SummarizeOption) (*Result, error) {
	input := Summarize{
		SummaryLength:       1,
		Articles:            []interface{}{},
		Coref:               false,
		SortBySalience:      false,
		IncludeAllSentences: false,
	}

	for _, fn := range options {
		fn(&input)
	}

	request, err := c.NewRequest("POST", "/nlp/v0.2/summarize", input)
	if err != nil {
		return nil, err
	}

	output := Result{}
	if err := c.Do(request, &output); err != nil {
		return nil, err
	}

	return &output, nil
}

type client struct {
	*http.Client

	baseURL *url.URL

	token string
}

func (c *client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Ocp-Apim-Subscription-Key", c.token)
	return req, nil
}

func New(token string) (*client, error) {
	baseURL, err := url.Parse("https://api.agolo.com")

	if err != nil {
		return nil, err
	}

	c := &client{
		token:   token,
		baseURL: baseURL,

		Client: http.DefaultClient,
	}

	return c, nil
}

func (wd *client) Do(req *http.Request, v interface{}) error {
	req.Header.Del("Accept-Encoding")

	if dump, err := httputil.DumpRequestOut(req, true); err == nil {
		os.Stdout.Write(dump)
	}

	resp, err := wd.Client.Do(req)
	if err != nil {
		return err
	}

	r := resp.Body
	defer r.Close()

	if resp.StatusCode != http.StatusOK {
		err := Error{}
		json.NewDecoder(r).Decode(&err)
		return &err
	}

	err = json.NewDecoder(r).Decode(&v)
	if err != nil {
		return err
	}

	return nil
}
