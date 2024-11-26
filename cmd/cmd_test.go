package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/rusher2004/canis-lupus-arctos/store"
	"github.com/stretchr/testify/assert"
)

type risk struct {
	State       string `json:"state"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// some risks to seed the store to test against
var risks = []risk{
	{
		// risk with all fields
		State:       "open",
		Title:       "Risk 1",
		Description: "Risk 1 description",
	},
	{
		// risk excluding optional fields
		State: "closed",
	},
}

type testCase struct {
	name          string
	id            string
	method        string
	reqBody       string
	skipCheckBody bool

	wantStatus int
	wantBody   string
}

func setup() (*store.MemoryStore, []string, error) {
	rs := store.NewMemoryStore()

	// we'll use these IDs to fetch via the GET endpoint
	ids := make([]string, 0, len(risks))
	for _, r := range risks {
		res, err := rs.CreateRisk(r.State, r.Title, r.Description)
		if err != nil {
			return nil, nil, fmt.Errorf("create risk: %w", err)
		}

		ids = append(ids, res.ID.String())
	}

	return rs, ids, nil
}

func runTest(t *testing.T, baseURL string, tc testCase, a *assert.Assertions, cl http.Client) {
	t.Run(tc.name, func(t *testing.T) {
		reqPath, err := url.JoinPath(baseURL, tc.id)
		if err != nil {
			t.Fatalf("join path: %v", err)
		}

		var buf io.Reader
		// create a request body if needed
		if tc.reqBody != "" {
			buf = strings.NewReader(tc.reqBody)
		}

		req, err := http.NewRequest(tc.method, reqPath, buf)
		if err != nil {
			t.Fatalf("error creating request: %v", err)
		}

		// make http request
		res, err := cl.Do(req)
		if err != nil {
			t.Fatalf("error making request: %v", err)
		}
		defer res.Body.Close()

		// check http status
		a.Equal(tc.wantStatus, res.StatusCode)

		// check contents of body
		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("error reading body: %v", err)
		}

		var out any
		if err := json.Unmarshal(b, &out); err != nil {
			t.Fatalf("error unmarshalling body: %v", err)
		}

		var want any
		if err := json.Unmarshal([]byte(tc.wantBody), &want); err != nil {
			t.Fatalf("error unmarshalling want body: %v", err)
		}

		// we can't know the ID of created risks ahead of time, so we can't compare the entire response
		if !tc.skipCheckBody {
			a.Equal(want, out)
		}
	})
}

func TestRun(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	t.Cleanup(cancel)

	rs, ids, err := setup()
	if err != nil {
		t.Fatalf("error setting up test: %v", err)
	}

	// start the server
	port := ":8181"
	go run(ctx, rs, port)

	cl := http.Client{Timeout: 5 * time.Second}
	a := assert.New(t)
	baseURL := "http://localhost" + port
	fullPath, err := url.JoinPath(baseURL, "v1", "risk")
	if err != nil {
		t.Fatalf("join path: %v", err)
	}

	// wait for server to start. realistically, we'd poll on a health endpoint. but there is no heavy
	// setup of remote dependencies in this app, so waiting should be fine.
	time.Sleep(1 * time.Second)

	tests := []testCase{
		// fetch some existing risks
		{
			name:       "get risk with all properties",
			id:         ids[0],
			method:     "GET",
			wantStatus: http.StatusOK,
			wantBody:   fmt.Sprintf("{\"id\": \"%s\",\"state\":\"%s\",\"title\":\"%s\",\"description\":\"%s\"}", ids[0], risks[0].State, risks[0].Title, risks[0].Description),
		},
		{
			name:       "get risk without optional properties",
			id:         ids[1],
			method:     "GET",
			wantStatus: http.StatusOK,
			wantBody:   fmt.Sprintf("{\"id\": \"%s\",\"state\":\"%s\"}", ids[1], risks[1].State),
		},
		{
			name:       "get all risks",
			method:     "GET",
			wantStatus: http.StatusOK,
			wantBody: fmt.Sprintf("[{\"id\": \"%s\",\"state\":\"%s\",\"title\":\"%s\",\"description\":\"%s\"},{\"id\": \"%s\",\"state\":\"%s\"}]",
				ids[0], risks[0].State, risks[0].Title, risks[0].Description,
				ids[1], risks[1].State,
			),
		},
		// create risks
		{
			name:          "create risk",
			method:        "POST",
			reqBody:       `{"state":"accepted","title":"Risk 3","description":"Risk 3 description"}`,
			skipCheckBody: true,
			wantStatus:    http.StatusCreated,
			wantBody:      "{\"id\": \"3\",\"state\":\"accepted\",\"title\":\"Risk 3\",\"description\":\"Risk 3 description\"}",
		},
		// input validation
		{
			name:       "missing state",
			method:     "POST",
			reqBody:    `{"title":"Risk 4","description":"Risk 4 description"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "{\"error\":\"state required\"}",
		},
		{
			name:       "invalid state",
			method:     "POST",
			reqBody:    `{"state":"invalid"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "{\"error\":\"state must be one of [open, closed, accepted, investigating]\"}",
		},
	}

	for _, tt := range tests {
		runTest(t, fullPath, tt, a, cl)
	}
}
