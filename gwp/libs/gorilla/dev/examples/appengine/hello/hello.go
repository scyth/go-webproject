package hello

import (
	"fmt"
	"net/http"

	"code.google.com/p/gorilla/mux"
	"code.google.com/p/gorilla/sessions"

	gaeSessions "code.google.com/p/gorilla/appengine/sessions"
)

var router = new(mux.Router)
var store = sessions.NewCookieStore([]byte("my-secret-key"),
	[]byte("1234567890123456"))
var dStore = gaeSessions.NewDatastoreStore("", []byte("my-secret-key"),
	[]byte("1234567890123456"))
var mStore = gaeSessions.NewMemcacheStore("", []byte("my-secret-key"),
	[]byte("1234567890123456"))

func init() {
	// Register a couple of routes.
	router.HandleFunc("/", homeHandler).Name("home")
	router.HandleFunc("/{salutation}/{name}", helloHandler).Name("hello")
	router.HandleFunc("/datastore-session", datastoreSessionHandler).Name("datastore-session")
	router.HandleFunc("/cookie-session", cookieSessionHandler).Name("cookie-session")
	router.HandleFunc("/memcache-session", memcacheSessionHandler).Name("memcache-session")
	// Send all incoming requests to router.
	http.Handle("/", router)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	url1, _ := router.GetRoute("hello").URL("salutation", "hello", "name", "world")
	url2, _ := router.GetRoute("cookie-session").URL()
	url3, _ := router.GetRoute("datastore-session").URL()
	url4, _ := router.GetRoute("memcache-session").URL()
	fmt.Fprintf(w, "Try a <a href='%s'>hello</a>. Or a <a href='%s'>cookie</a>, <a href='%s'>datastore</a> or <a href='%s'>memcache</a> session.", url1, url2, url3, url4)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	vars := mux.Vars(r)
	fmt.Fprintf(w, "%s, %s!", vars["salutation"], vars["name"])
}

func cookieSessionHandler(w http.ResponseWriter, r *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := store.Get(r, "cookie-session")
	var msg string
	if value, ok := session.Values["counter"]; ok {
		// Increment value.
		session.Values["counter"] = value.(int) + 1
		msg = fmt.Sprintf(`Value for cookie session "counter" is "%v".`, value)
	} else {
		// Set a value.
		session.Values["counter"] = 1
		msg = fmt.Sprintf(`No value found for cookie session "counter".`)
	}
	// Save it.
	err := session.Save(r, w)
	fmt.Fprintf(w, `%s -- Errors: %v`, msg, err)
}

func datastoreSessionHandler(w http.ResponseWriter, r *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := dStore.Get(r, "datastore-session")
	var msg string
	if value, ok := session.Values["counter"]; ok {
		// Increment value.
		session.Values["counter"] = value.(int) + 1
		msg = fmt.Sprintf(`Value for datastore session "counter" is "%v".`, value)
	} else {
		// Set a value.
		session.Values["counter"] = 1
		msg = fmt.Sprintf(`No value found for datastore session "counter".`)
	}
	// Save it.
	err := session.Save(r, w)
	fmt.Fprintf(w, `%s -- Errors: %v`, msg, err)
}

func memcacheSessionHandler(w http.ResponseWriter, r *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := mStore.Get(r, "memcache-session")
	var msg string
	if value, ok := session.Values["counter"]; ok {
		// Increment value.
		session.Values["counter"] = value.(int) + 1
		msg = fmt.Sprintf(`Value for memcache session "counter" is "%v".`, value)
	} else {
		// Set a value.
		session.Values["counter"] = 1
		msg = fmt.Sprintf(`No value found for memcache session "counter".`)
	}
	// Save it.
	err := session.Save(r, w)
	fmt.Fprintf(w, `%s -- Errors: %v`, msg, err)
}
