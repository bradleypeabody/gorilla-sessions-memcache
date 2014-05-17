package gsm

import (
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"testing"
)

// FIXME: need unit tests for the different StoreMethods

func TestMain(t *testing.T) {

	doRun := func(method StoreMethod, listenConfig string) {

		memcacheClient := memcache.New("localhost:11211")
		// fmt.Printf("memcacheClient = %v\n", memcacheClient)
		sessionStore := NewMemcacheStore(memcacheClient, "TestMain_", []byte("example123"))
		sessionStore.StoreMethod = StoreMethod(method)
		sessionStore.Logging = 1

		http.HandleFunc("/test"+string(method), func(w http.ResponseWriter, r *http.Request) {
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
		go http.ListenAndServe(listenConfig, nil)

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

		v := doReq("http://localhost" + listenConfig + "/test" + string(method) + "?store=blah")
		if v != "blah" {
			t.Fatalf("Expected v=blah but got v='%s'\n", v)
		}

		v = doReq("http://localhost" + listenConfig + "/test" + string(method))
		if v != "blah" {
			t.Fatalf("Expected session to give us v=blah but got v='%s'\n", v)
		}

	}

	doRun(StoreMethodSecureCookie, ":18201")
	doRun(StoreMethodGob, ":18202")
	doRun(StoreMethodJson, ":18203")

}
