package httpserver

import (
	"encoding/json"
	"github.com/function61/eventkit/command"
	"github.com/function61/eventkit/event"
	"github.com/function61/gokit/assert"
	"github.com/function61/gokit/httpauth"
	"github.com/function61/gokit/logex"
	"github.com/function61/passitron/pkg/commands"
	"github.com/function61/passitron/pkg/domain"
	"github.com/function61/passitron/pkg/state"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type testCase struct {
	name   string
	req    *http.Request
	status int
	body   string
}

const (
	testUserId = "99"
)

func TestScenario(t *testing.T) {
	st := state.NewTesting()

	seedDatabase(st)

	handler, err := createHandlerWithWorkdirHack(st)
	if err != nil {
		t.Fatalf("createHandlerWithWorkdirHack: %v", err)
	}
	// srv := httptest.NewServer(handler)

	auther, err := httpauth.NewEcJwtSigner(st.GetJwtSigningKey())
	if err != nil {
		t.Fatalf("NewJwtSigner: %v", err)
	}

	// somewhat expensive operation, so cache this here to do this only once
	jwtToken := auther.Sign(httpauth.UserDetails{
		Id: testUserId,
	}, time.Now())

	auth := func(req *http.Request) {
		cookies := httpauth.ToCookiesWithCsrfProtection(jwtToken)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		req.Header.Set("x-csrf-token", cookies[1].Value)
	}

	jsonHeader := func(req *http.Request) {
		req.Header.Set("Content-Type", "application/json")
	}

	allProperHeaders := func(req *http.Request) {
		auth(req)
		jsonHeader(req)
	}

	tests := []testCase{
		{
			name:   "Without auth details",
			req:    post("/command/account.ChangeUrl", ""),
			status: http.StatusForbidden,
			body:   `{"status":"error","error_code":"not_signed_in","error_description":"auth: cookie login missing"}`,
		},
		{
			name:   "Missing JSON Content-Type header",
			req:    post("/command/account.ChangeUrl", "", auth),
			status: http.StatusBadRequest,
			body:   `{"status":"error","error_code":"expecting_content_type_json","error_description":"expecting Content-Type header with application/json"}`,
		},
		{
			name:   "Missing JSON body",
			req:    post("/command/account.ChangeUrl", "", allProperHeaders),
			status: http.StatusBadRequest,
			body:   `{"status":"error","error_code":"json_parsing_failed","error_description":"EOF"}`,
		},
		{
			name: "Account not found",
			req: post(
				"/command/account.ChangeUrl",
				cmdJson(&commands.AccountChangeUrl{
					Account: "123",
					Url:     "http://example.com/"}),
				allProperHeaders),
			status: http.StatusBadRequest,
			body:   `{"status":"error","error_code":"command_failed","error_description":"Account not found"}`,
		},
		{
			name: "Command succeeds",
			req: post(
				"/command/account.ChangeUrl",
				cmdJson(&commands.AccountChangeUrl{
					Account: "14",
					Url:     "http://example.com/"}),
				allProperHeaders),
			status: http.StatusOK,
			body:   `{"status":"success","error_code":"","error_description":""}`,
		},
	}

	runOne := func(t *testing.T, test testCase) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, test.req)

		res := rec.Result()

		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		res.Body.Close()

		body := string(bodyBytes)

		assert.Assert(t, res.StatusCode == test.status)
		assert.EqualString(t, body, test.body+"\n")
	}

	for _, test := range tests {
		test := test // pin

		t.Run(test.name, func(t *testing.T) {
			runOne(t, test)
		})
	}
}

func seedDatabase(appState *state.AppState) {
	userCreated := domain.NewUserCreated(
		testUserId,
		"admin",
		event.Meta(time.Now(), testUserId))

	accountCreated := domain.NewAccountCreated(
		"14",
		domain.RootFolderId,
		"My test account",
		event.Meta(time.Now(), testUserId))

	// many crypto tokens are derived from master password
	masterPwdChanged := domain.NewDatabaseMasterPasswordChanged(
		"greatpassword",
		event.Meta(time.Now(), testUserId))

	if err := appState.EventLog.Append([]event.Event{
		userCreated,
		accountCreated,
		masterPwdChanged,
	}); err != nil {
		panic(err)
	}
}

type reqMutator func(*http.Request)

func post(path string, body string, muts ...reqMutator) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "http://localhost"+path, strings.NewReader(body))

	for _, mut := range muts {
		mut(req)
	}

	return req
}

func cmdJson(cmd command.Command) string {
	out, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}

	return string(out)
}

func createHandlerWithWorkdirHack(appState *state.AppState) (http.Handler, error) {
	// createHandler() reads a file off of a filesystem, expecting project root as workdir,
	// but during test execution our workdir is at our workdir
	revertWdir, err := chdirTemporarily("../..")
	if err != nil {
		return nil, err
	}
	defer revertWdir()

	return createHandler(appState, nil)
}

func chdirTemporarily(to string) (revert func(), err error) {
	wdBefore, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if err := os.Chdir(to); err != nil {
		return nil, err
	}

	// returns revert func which changes the dir back to what it used to be before
	return func() {
		if err := os.Chdir(wdBefore); err != nil {
			panic(err)
		}
	}, nil
}
