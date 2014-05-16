package gsm

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"testing"
)

func TestStubMain(t *testing.T) {

	sessionStore := NewDumbMemorySessionStore()

	http.HandleFunc("/testdumb", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))

		session, _ := sessionStore.Get(r, "example")

		storeval := r.FormValue("store")
		if len(storeval) > 0 {
			session.Values["thevalue"] = storeval
		} else {
			storeval, _ = session.Values["thevalue"].(string)
		}

		err := session.Save(r, w)
		if err != nil {
			fmt.Printf("Error while saving session: %v\n", err)
		}

		fmt.Fprintf(w, "%s", storeval)

	})

	// run the server
	go http.ListenAndServe(":18210", nil)

	// now do some tests as a client make sure things work as expected

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	httpClient := &http.Client{
		Jar: jar,
	}

	doReq := func(u string) string {
		resp, err := httpClient.Get(u)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)

		}

		fmt.Printf("Got Set-Cookie: %s\n", resp.Header.Get("Set-Cookie"))

		return string(b)
	}

	v := doReq("http://localhost:18210/testdumb?store=blah")
	if v != "blah" {
		t.Fatalf("Expected v=blah but got v='%s'\n", v)
	}

	v = doReq("http://localhost:18210/testdumb")
	if v != "blah" {
		t.Fatalf("Expected session to give us v=blah but got v='%s'\n", v)
	}

}
