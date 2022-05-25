package axiom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	endpoint = "http://axiom.local"
	// apiToken is a placeholder API token.
	apiToken = "xaat-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	// personalToken is a placeholder personal token.
	personalToken = "xapt-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX" //nolint:gosec // Chill, it's just testing.
	// orgID is a placeholder organization id.
	orgID = "awkward-identifier-c3po"
)

var tokenRe = regexp.MustCompile("xa(a|p|)t-[a-zA-z0-9]{8}-[a-zA-z0-9]{4}-[a-zA-z0-9]{4}-[a-zA-z0-9]{4}-[a-zA-z0-9]{12}")

// SetStrictDecoding is a special testing-only client option that - when set to
// 'true' - failes JSON response decoding if fields not present in the
// destination struct are encountered.
func SetStrictDecoding(b bool) Option {
	return func(c *Client) error {
		c.strictDecoding = b
		return nil
	}
}

func TestNewClient(t *testing.T) {
	t.Cleanup(os.Clearenv)

	tests := []struct {
		name        string
		environment map[string]string
		options     []Option
		err         error
	}{
		{
			name: "no environment no options",
			err:  ErrMissingAccessToken,
		},
		{
			name: "no environment accessToken option",
			options: []Option{
				SetAccessToken(personalToken),
			},
			err: ErrMissingOrganizationID,
		},
		{
			name: "no environment accessToken option with API token",
			options: []Option{
				SetAccessToken(apiToken),
			},
		},
		{
			name: "orgID environment no options",
			environment: map[string]string{
				"AXIOM_TOKEN": personalToken,
			},
			err: ErrMissingOrganizationID,
		},
		{
			name: "orgID environment no options with API token",
			environment: map[string]string{
				"AXIOM_TOKEN": apiToken,
			},
		},
		{
			name: "no environment accessToken and orgID option",
			options: []Option{
				SetAccessToken(personalToken),
				SetOrgID(orgID),
			},
		},
		{
			name: "accessToken and orgID environment no options",
			environment: map[string]string{
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": orgID,
			},
		},
		{
			name: "accessToken environment orgID option",
			environment: map[string]string{
				"AXIOM_TOKEN": personalToken,
			},
			options: []Option{
				SetOrgID(orgID),
			},
		},
		{
			name: "orgID environment accessToken option",
			environment: map[string]string{
				"AXIOM_ORG_ID": orgID,
			},
			options: []Option{
				SetAccessToken(personalToken),
			},
		},
		{
			name: "no environment url and accessToken option",
			options: []Option{
				SetURL(endpoint),
				SetAccessToken(personalToken),
			},
		},
		{
			name: "url and accessToken environment no options",
			environment: map[string]string{
				"AXIOM_URL":   endpoint,
				"AXIOM_TOKEN": personalToken,
			},
		},
		{
			name: "accessToken and orgID environment cloudUrl option",
			environment: map[string]string{
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": orgID,
			},
			options: []Option{
				SetURL(CloudURL),
			},
		},
		{
			name: "accessToken and orgID environment enhanced cloudUrl option",
			environment: map[string]string{
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": orgID,
			},
			options: []Option{
				SetURL(CloudURL + "/"),
			},
		},
		{
			name: "cloudUrl accessToken and orgID environment no options",
			environment: map[string]string{
				"AXIOM_URL":    CloudURL,
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": orgID,
			},
		},
		{
			name: "enhanced cloudUrl, accessToken and orgID environment no options",
			environment: map[string]string{
				"AXIOM_URL":    CloudURL + "/",
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": orgID,
			},
		},
		{
			name: "dev url, accessToken and orgID environment no options",
			environment: map[string]string{
				"AXIOM_URL":    "https://dev.axiom.co",
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": orgID,
			},
		},
		{
			name: "cloudUrl and accessToken environment orgID option",
			environment: map[string]string{
				"AXIOM_URL":   CloudURL,
				"AXIOM_TOKEN": personalToken,
			},
			options: []Option{
				SetOrgID(orgID),
			},
		},
		{
			name: "accessToken and orgID environment noEnv option",
			environment: map[string]string{
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": orgID,
			},
			options: []Option{
				SetNoEnv(),
			},
			err: ErrMissingAccessToken,
		},
		{
			name: "no environment noEnv, cloudUrl and accessToken option with API token",
			options: []Option{
				SetNoEnv(),
				SetURL(CloudURL),
				SetAccessToken(apiToken),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			for k, v := range tt.environment {
				os.Setenv(k, v)
			}

			client, err := NewClient(tt.options...)
			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
			} else {
				if assert.NoError(t, err) {
					assert.Regexp(t, tokenRe, client.accessToken)
					assert.NotEmpty(t, client.baseURL)
				}
			}
		})
	}
}

func TestNewClient_Valid(t *testing.T) {
	client := newClient(t)

	// Are endpoints/resources present?
	assert.NotNil(t, client.Dashboards)
	assert.NotNil(t, client.Datasets)
	assert.NotNil(t, client.Monitors)
	assert.NotNil(t, client.Notifiers)
	assert.NotNil(t, client.Organizations.Cloud)
	assert.NotNil(t, client.Organizations.Selfhost)
	assert.NotNil(t, client.StarredQueries)
	assert.NotNil(t, client.Teams)
	assert.NotNil(t, client.Tokens.API)
	assert.NotNil(t, client.Tokens.Personal)
	assert.NotNil(t, client.Users)
	assert.NotNil(t, client.Version)
	assert.NotNil(t, client.VirtualFields)

	// Is default configuration present?
	assert.Equal(t, endpoint, client.baseURL.String())
	assert.Equal(t, personalToken, client.accessToken)
	assert.Empty(t, client.orgID)
	assert.NotNil(t, client.httpClient)
	assert.NotEmpty(t, client.userAgent)
	assert.False(t, client.strictDecoding)
	assert.True(t, client.noEnv) // Disabled for testing.
	assert.False(t, client.noLimiting)
}

func TestClient_Options_SetAccessToken(t *testing.T) {
	client := newClient(t)

	exp := personalToken
	opt := SetAccessToken(exp)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, exp, client.accessToken)
}

func TestClient_Options_SetClient(t *testing.T) {
	client := newClient(t)

	exp := &http.Client{
		Timeout: time.Second,
	}
	opt := SetClient(exp)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, exp, client.httpClient)
}

func TestClient_Options_SetCloudConfig(t *testing.T) {
	client := newClient(t)

	opt := SetCloudConfig(personalToken, orgID)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, personalToken, client.accessToken)
	assert.Equal(t, orgID, client.orgID)
}

func TestClient_Options_SetOrgID(t *testing.T) {
	client := newClient(t)

	exp := orgID
	opt := SetOrgID(exp)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, exp, client.orgID)
}

func TestClient_Options_SetSelfhostConfig(t *testing.T) {
	client := newClient(t)

	opt := SetSelfhostConfig(endpoint, personalToken)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, endpoint, client.baseURL.String())
	assert.Equal(t, personalToken, client.accessToken)
}

func TestClient_Options_SetURL(t *testing.T) {
	client := newClient(t)

	exp := endpoint
	opt := SetURL(exp)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, exp, client.baseURL.String())
}

func TestClient_Options_SetUserAgent(t *testing.T) {
	client := newClient(t)

	exp := "axiom-go/1.0.0"
	opt := SetUserAgent(exp)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, exp, client.userAgent)
}

func TestClient_newRequest_BadURL(t *testing.T) {
	client := newClient(t)

	_, err := client.newRequest(context.Background(), http.MethodGet, ":", nil)
	assert.Error(t, err)

	if assert.IsType(t, new(url.Error), err) {
		urlErr := err.(*url.Error)
		assert.Equal(t, urlErr.Op, "parse")
	}
}

// If a nil body is passed to NewRequest, make sure that nil is also passed to
// http.NewRequest. In most cases, passing an io.Reader that returns no content
// is fine, since there is no difference between an HTTP request body that is an
// empty string versus one that is not set at all. However in certain cases,
// intermediate systems may treat these differently resulting in subtle errors.
func TestClient_newRequest_EmptyBody(t *testing.T) {
	client := newClient(t)

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	assert.Empty(t, req.Body)
}

func TestClient_do(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, _ = fmt.Fprint(w, `{"A":"a"}`)
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	type foo struct {
		A string
	}

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	var body foo
	_, err = client.do(req, &body)
	require.NoError(t, err)

	assert.Equal(t, foo{"a"}, body)
}

func TestClient_do_ioWriter(t *testing.T) {
	content := `{"A":"a"}`

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, _ = fmt.Fprint(w, content)
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	var buf bytes.Buffer
	_, err = client.do(req, &buf)
	require.NoError(t, err)

	assert.Equal(t, content, buf.String())
}

func TestClient_do_HTTPError(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(http.StatusText(http.StatusBadRequest)))
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	if _, err = client.do(req, nil); assert.ErrorIs(t, err, &Error{
		Status:  http.StatusBadRequest,
		Message: http.StatusText(http.StatusBadRequest),
	}) {
		assert.EqualError(t, err, "API error 400: Bad Request")
	}
}

func TestClient_do_HTTPError_JSON(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", mediaTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		assert.NoError(t, json.NewEncoder(w).Encode(Error{
			Message: "This is a Bad Request error",
		}))
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	if _, err = client.do(req, nil); assert.ErrorIs(t, err, &Error{
		Status:  http.StatusBadRequest,
		Message: "This is a Bad Request error",
	}) {
		assert.EqualError(t, err, "API error 400: This is a Bad Request error")
	}
}

func TestClient_do_HTTPError_Unauthenticated(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", mediaTypeJSON)
		w.WriteHeader(http.StatusUnauthorized)

		assert.NoError(t, json.NewEncoder(w).Encode(Error{
			Message: "You are not allowed here!",
		}))
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	_, err = client.do(req, nil)
	assert.ErrorIs(t, err, ErrUnauthenticated)
}

func TestClient_do_RateLimit(t *testing.T) {
	// Truncated time for testing as the `Error()` method for the `LimitError`
	// uses `time.Until()` which will yield different milliseconds when
	// comparing the time values on `errors.Is()`.
	reset := time.Now().Add(time.Hour).Truncate(time.Second)

	expErr := &LimitError{
		Limit: Limit{
			Scope:     LimitScopeAnonymous,
			Limit:     1000,
			Remaining: 0,
			Reset:     reset,

			limitType: limitRate,
		},
		Message: "rate limit exceeded",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set(headerRateScope, "anonymous")
		w.Header().Set(headerRateLimit, "1000")
		w.Header().Set(headerRateRemaining, "0")
		w.Header().Set(headerRateReset, strconv.FormatInt(reset.Unix(), 10))
		w.WriteHeader(http.StatusTooManyRequests)
		assert.NoError(t, json.NewEncoder(w).Encode(Error{
			Message: "rate limit exceeded",
		}))
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	// Request should fail with a `*LimitError`.
	resp, err := client.do(req, nil)

	if assert.ErrorIs(t, err, expErr) {
		assert.EqualError(t, err, "rate limit exceeded: try again in 59m59s")
	}
	assert.Equal(t, expErr.Limit, resp.Limit)
}

func TestClient_do_RateLimit_ShortCircuit(t *testing.T) {
	// Truncated time for testing as the `Error()` method for the `LimitError`
	// uses `time.Until()` which will yield different milliseconds when
	// comparing the time values on `errors.Is()`.
	reset := time.Now().Add(time.Hour).Truncate(time.Second)

	expErr := &LimitError{
		Limit: Limit{
			Scope:     LimitScopeAnonymous,
			Limit:     1000,
			Remaining: 0,
			Reset:     reset,

			limitType: limitRate,
		},
		Message: "anonymous rate limit exceeded, not making remote request",
	}

	var exceeded bool
	hf := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set(headerRateScope, "anonymous")
		w.Header().Set(headerRateLimit, "1000")
		w.Header().Set(headerRateRemaining, "0")
		w.Header().Set(headerRateReset, strconv.FormatInt(reset.Unix(), 10))

		if !exceeded {
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		} else {
			w.WriteHeader(http.StatusTooManyRequests)
			assert.NoError(t, json.NewEncoder(w).Encode(Error{
				Message: "rate limit exceeded",
			}))
		}
		exceeded = true
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	// First request should succeed with no rate limit remaining.
	resp, err := client.do(req, nil)
	require.NoError(t, err)

	assert.EqualValues(t, "anonymous", resp.Limit.Scope.String())
	assert.EqualValues(t, 1000, resp.Limit.Limit)
	assert.EqualValues(t, 0, resp.Limit.Remaining)
	assert.EqualValues(t, reset, resp.Limit.Reset)
	assert.Equal(t, limitRate, resp.Limit.limitType)

	// Second request should short circuit as the client is aware that there is
	// no rate remaining.
	resp, err = client.do(req, nil)
	if assert.ErrorIs(t, err, expErr) {
		assert.EqualError(t, err, "anonymous rate limit exceeded, not making remote request: try again in 59m59s")
	}
	assert.Equal(t, expErr.Limit, resp.Limit)
}

func TestClient_do_RateLimit_NoLimiting(t *testing.T) {
	// Truncated time for testing as the `Error()` method for the `LimitError`
	// uses `time.Until()` which will yield different milliseconds when
	// comparing the time values on `errors.Is()`.
	reset := time.Now().Add(time.Hour).Truncate(time.Second)

	hf := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set(headerRateScope, "anonymous")
		w.Header().Set(headerRateLimit, "1000")
		w.Header().Set(headerRateRemaining, "0")
		w.Header().Set(headerRateReset, strconv.FormatInt(reset.Unix(), 10))

		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}

	client, teardown := setup(t, "/", hf, SetNoLimiting())
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	// First request should succeed with no rate limit remaining.
	resp, err := client.do(req, nil)
	require.NoError(t, err)

	assert.EqualValues(t, "anonymous", resp.Limit.Scope.String())
	assert.EqualValues(t, 1000, resp.Limit.Limit)
	assert.EqualValues(t, 0, resp.Limit.Remaining)
	assert.EqualValues(t, reset, resp.Limit.Reset)
	assert.Equal(t, limitRate, resp.Limit.limitType)

	// Second request should succeed as well as there is no client side limiting
	// happening.
	resp, err = client.do(req, nil)
	require.NoError(t, err)

	assert.EqualValues(t, "anonymous", resp.Limit.Scope.String())
	assert.EqualValues(t, 1000, resp.Limit.Limit)
	assert.EqualValues(t, 0, resp.Limit.Remaining)
	assert.EqualValues(t, reset, resp.Limit.Reset)
	assert.Equal(t, limitRate, resp.Limit.limitType)
}

func TestClient_do_UnprivilegedToken(t *testing.T) {
	client, teardown := setup(t, "/", nil)
	defer teardown()

	err := client.Options(SetAccessToken("xaat-123"))
	require.NoError(t, err)

	_, err = client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.ErrorIs(t, err, ErrUnprivilegedToken)
}

func TestClient_do_RedirectLoop(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	_, err = client.do(req, nil)
	assert.IsType(t, new(url.Error), err)
}

func TestClient_do_ValidOnlyAPITokenPaths(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {}

	tests := []string{
		"/api/v1/datasets/test/query",
		"/api/v1/datasets/_apl",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			client, teardown := setup(t, tt, hf)
			defer teardown()

			err := client.Options(SetAccessToken("xaat-123"))
			require.NoError(t, err)

			req, err := client.newRequest(context.Background(), http.MethodGet, tt, nil)
			require.Nil(t, err)

			_, err = client.do(req, nil)
			require.NoError(t, err)
		})
	}
}

func TestAPITokenPathRegex(t *testing.T) {
	tests := []struct {
		input string
		match bool
	}{
		{
			input: "/api/v1/datasets/test/ingest",
			match: true,
		},
		{
			input: "/api/v1/datasets/test/ingest?timestamp-format=unix",
			match: true,
		},
		{
			input: "/api/v1/datasets/test/query",
			match: true,
		},
		{
			input: "/api/v1/datasets/_apl",
			match: true,
		},
		{
			input: "/api/v1/datasets/test/query?nocache=true",
			match: true,
		},
		{
			input: "/api/v1/datasets/_apl?nocache=true",
			match: true,
		},
		{
			input: "/api/v1/datasets//query",
			match: false,
		},
		{
			input: "/api/v1/datasets/query",
			match: false,
		},
		{
			input: "/api/v1/datasets/test/elastic",
			match: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.match, validOnlyAPITokenPaths.MatchString(tt.input))
		})
	}
}

func TestClient_do_Backoff(t *testing.T) {
	var currentCalls int
	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		currentCalls++
		switch currentCalls {
		case 1:
			w.WriteHeader(http.StatusInternalServerError)
		case 2:
			w.WriteHeader(http.StatusBadGateway)
		case 3:
			w.WriteHeader(http.StatusGatewayTimeout)
		default:
			w.WriteHeader(http.StatusOK)
		}
		return
	})
	srv := httptest.NewServer(r)
	defer srv.Close()

	clientOptions := []Option{
		SetURL(srv.URL),
		SetAccessToken(personalToken),
		SetOrgID(orgID),
		SetClient(srv.Client()),
		SetStrictDecoding(true),
		SetNoEnv(),
	}

	client, err := NewClient(clientOptions...)
	require.NoError(t, err)

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	resp, err := client.do(req, nil)
	require.NoError(t, err)
	if currentCalls > 4 {
		t.Fatal("expected to attempt 4 times")
	}
	assert.Equal(t, 4, currentCalls)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// setup sets up a test HTTP server along with a client that is configured to
// talk to that test server. Tests should pass a handler function which provides
// the response for the API method being tested.
func setup(t *testing.T, path string, handler http.HandlerFunc, options ...Option) (*Client, func()) {
	t.Helper()

	r := http.NewServeMux()
	r.HandleFunc(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("Authorization"), "no authorization header present on the request")
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Accept"), "bad accept header present on the request")
		assert.Equal(t, "axiom-go", r.Header.Get("User-Agent"), "bad user-agent header present on the request")
		if orgIDHeader := r.Header.Get("X-Axiom-Org-Id"); orgIDHeader != "" {
			assert.Equal(t, orgID, orgIDHeader, "bad x-axiom-org-id header present on the request")
		}

		if r.ContentLength > 0 {
			assert.NotEmpty(t, r.Header.Get("Content-Type"), "no Content-Type header present on the request")
		}

		handler.ServeHTTP(w, r)
	}))
	srv := httptest.NewServer(r)

	clientOptions := []Option{
		SetURL(srv.URL),
		SetAccessToken(personalToken),
		SetOrgID(orgID),
		SetClient(srv.Client()),
		SetStrictDecoding(true),
		SetNoEnv(),
	}
	clientOptions = append(clientOptions, options...)

	client, err := NewClient(clientOptions...)
	require.NoError(t, err)

	return client, func() { srv.Close() }
}

// newClient returns a new client with stub properties for testing methods that
// don't actually make a http call.
func newClient(t *testing.T) *Client {
	t.Helper()

	client, err := NewClient(
		SetURL(endpoint),
		SetAccessToken(personalToken),
		SetNoEnv(),
	)
	require.NoError(t, err)

	return client
}

func mustTimeParse(tb testing.TB, layout, value string) time.Time {
	ts, err := time.Parse(layout, value)
	require.NoError(tb, err)
	return ts
}
