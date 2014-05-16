package gsm

import (
	"encoding/base32"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
	"strings"
)

// Sessions implemented with a dumb in-memory
// map and no expiration.  Good for local
// development so you don't have to run
// memcached on your laptop just to fire up
// your app and hack away.
func NewDumbMemorySessionStore() *DumbMemoryStore {

	keyPair := []byte("stub")

	return &DumbMemoryStore{
		Codecs: securecookie.CodecsFromPairs(keyPair),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		Data: make(map[string]string),
	}
}

// DumbMemoryStore stores sessions in memcache
//
type DumbMemoryStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options // default configuration
	Data    map[string]string // session data goes here
}

// MaxLength restricts the maximum length of new sessions to l.
// If l is 0 there is no limit to the size of a session, use with caution.
// The default for a new DumbMemoryStore is 4096.
func (s *DumbMemoryStore) MaxLength(l int) {
	for _, c := range s.Codecs {
		if codec, ok := c.(*securecookie.SecureCookie); ok {
			codec.MaxLength(l)
		}
	}
}

// Get returns a session for the given name after adding it to the registry.
//
// See CookieStore.Get().
func (s *DumbMemoryStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// See CookieStore.New().
func (s *DumbMemoryStore) New(r *http.Request, name string) (*sessions.Session, error) {
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
func (s *DumbMemoryStore) Save(r *http.Request, w http.ResponseWriter,
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

// save writes encoded session.Values using the memcache client
func (s *DumbMemoryStore) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
		s.Codecs...)
	if err != nil {
		return err
	}

	s.Data[session.ID] = encoded

	return nil
}

// load reads a file and decodes its content into session.Values.
func (s *DumbMemoryStore) load(session *sessions.Session) error {

	if err := securecookie.DecodeMulti(session.Name(), string(s.Data[session.ID]),
		&session.Values, s.Codecs...); err != nil {
		return err
	}
	return nil
}
