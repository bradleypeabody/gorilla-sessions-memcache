gorilla-sessions-memcache
=========================

Memcache session support for Gorilla Web Toolkit.  

Dependencies
------------

The usual gorilla stuff:

    go get github.com/gorilla/sessions

Plus Brad Fitz' memcache client:

    go get github.com/bradfitz/gomemcache/memcache

Usage
-----

    import (
      "github.com/bradfitz/gomemcache/memcache"
      gsm "github.com/bradleypeabody/gorilla-sessions-memcache"
    )

    ...

    // set up your memcache client
    memcacheClient := memcache.New("localhost:11211")
    
    // set up your session store
    store := gsm.NewMemcacheStore(memcacheClient, "session_prefix_", []byte("secret-key-goes-here"))
    
    // and the rest of it is the same as any other gorilla session handling:
    func MyHandler(w http.ResponseWriter, r *http.Request) {
      session, _ := store.Get(r, "session-name")
      session.Values["foo"] = "bar"
      session.Values[42] = 43
      session.Save(r, w)
    }

Things to Know
--------------

* This is still experimental as of May 2014.

* You can also call NewDumbMemorySessionStore() for local development without a memcache server (it's a stub that just stuffs your session data in a map - definitely do not use this for anything but local dev and testing).
