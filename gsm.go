// Memcache session support for Gorilla Web Toolkit,
// without Google App Engine dependency.
package gsm

import (
	"encoding/base32"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
	"strings"
)

// type MemcacheSessionStore struct {
// 	client *memcache.Client
// }

// // Create a new memcache session store
// func NewMemcacheSessionStore(client *memcache.Client) *MemcacheSessionStore {
// 	return &MemcacheSessionStore{client: client}
// }

// MemcacheStore ------------------------------------------------------------

// var fileMutex sync.RWMutex

// NewMemcacheStore returns a new MemcacheStore.
//
// The path argument is the directory where sessions will be saved. If empty
// it will use os.TempDir().
//
// See NewCookieStore() for a description of the other parameters.
// func NewMemcacheStore(path string, keyPairs ...[]byte) *MemcacheStore {
func NewMemcacheStore(client *memcache.Client, keyPrefix string, keyPairs ...[]byte) *MemcacheStore {
	// if path == "" {
	// 	path = os.TempDir()
	// }
	// if path[len(path)-1] != '/' {
	// 	path += "/"
	// }

	if client == nil {
		panic("Cannot have nil memcache client")
	}

	return &MemcacheStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		KeyPrefix: keyPrefix,
		Client:    client,
		// path: path,
	}
}

// MemcacheStore stores sessions in the filesystem.
//
// It also serves as a referece for custom stores.
//
// This store is still experimental and not well tested. Feedback is welcome.
type MemcacheStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options // default configuration
	// path    string
	Client    *memcache.Client
	KeyPrefix string
}

// MaxLength restricts the maximum length of new sessions to l.
// If l is 0 there is no limit to the size of a session, use with caution.
// The default for a new MemcacheStore is 4096.
func (s *MemcacheStore) MaxLength(l int) {
	for _, c := range s.Codecs {
		if codec, ok := c.(*securecookie.SecureCookie); ok {
			codec.MaxLength(l)
		}
	}
}

// Get returns a session for the given name after adding it to the registry.
//
// See CookieStore.Get().
func (s *MemcacheStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// See CookieStore.New().
func (s *MemcacheStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			err = s.load(session)
			if err == nil {
				session.IsNew = false
			}
		}
	}
	return session, err
}

// Save adds a single session to the response.
func (s *MemcacheStore) Save(r *http.Request, w http.ResponseWriter,
	session *sessions.Session) error {
	if session.ID == "" {
		// Because the ID is used in the filename, encode it to
		// use alphanumeric characters only.
		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32)), "=")
	}
	if err := s.save(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID,
		s.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

// save writes encoded session.Values to a file.
func (s *MemcacheStore) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
		s.Codecs...)
	if err != nil {
		return err
	}

	key := s.KeyPrefix + session.ID

	fmt.Printf("key = %v, encoded = %v\n", key, encoded)

	err = s.Client.Set(&memcache.Item{Key: key, Value: []byte(encoded)})
	if err != nil {
		return err
	}

	// // fileMutex.Lock()
	// // defer fileMutex.Unlock()
	// fp, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	// if err != nil {
	// 	return err
	// }
	// if _, err = fp.Write([]byte(encoded)); err != nil {
	// 	return err
	// }
	// fp.Close()
	return nil
}

// load reads a file and decodes its content into session.Values.
func (s *MemcacheStore) load(session *sessions.Session) error {

	key := s.KeyPrefix + session.ID

	it, err := s.Client.Get(key)
	// it.Value

	// filename := s.path + "session_" + session.ID
	// fp, err := os.OpenFile(filename, os.O_RDONLY, 0400)
	// if err != nil {
	// 	return err
	// }
	// defer fp.Close()
	// var fdata []byte
	// buf := make([]byte, 128)
	// for {
	// 	var n int
	// 	n, err = fp.Read(buf[0:])
	// 	fdata = append(fdata, buf[0:n]...)
	// 	if err != nil {
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		return err
	// 	}
	// }

	if err = securecookie.DecodeMulti(session.Name(), string(it.Value),
		&session.Values, s.Codecs...); err != nil {
		return err
	}
	return nil
}

// Sessions implemented with a dumb in-memory
// map and no expiration.  Good for local
// development so you don't have to run
// memcached on your laptop just to fire up
// your app and hack away.
// func NewDumbMemorySessionStore() {

// }
