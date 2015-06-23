package main // import "websecurify/go-http-meta-proxy"

// ---
// ---
// ---

import (
	"os"
	"log"
	"path"
	"strings"
	"net/url"
	"net/http"
	"net/http/httputil"
	
	// ---
	
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	
	// ---
	
	"github.com/abbot/go-http-auth"
)

// ---
// ---
// ---

func main() {
	funcWrapper := func (f http.HandlerFunc) (http.HandlerFunc) {
		return f
	}
	
	// ---
	
	if os.Getenv("HTPASSWD") != "" {
		secrets := auth.HtpasswdFileProvider(os.Getenv("HTPASSWD"))
		
		// ---
		
		authenticator := auth.NewBasicAuthenticator(os.Getenv("REALM"), secrets)
		
		// ---
		
		funcWrapper = func (f http.HandlerFunc) (http.HandlerFunc) {
			return auth.JustCheck(authenticator, f)
		}
	}
	
	// ---
	
	r := mux.NewRouter()
	
	// ---
	
	for _, env := range os.Environ() {
		tokens := strings.SplitN(env, "=", 2)
		
		if len(tokens) != 2 {
			continue
		}
		
		// ---
		
		key := strings.TrimSpace(tokens[0])
		value := strings.TrimSpace(tokens[1])
		
		// ---
		
		if !strings.HasPrefix(key, "BACKEND_") {
			continue
		}
		
		// ---
		
		tokens = strings.SplitN(value, ":::", 2)
		
		if len(tokens) != 2 {
			log.Fatal("cannot parse backend " + key)
		}
		
		// ---
		
		source := strings.TrimSpace(tokens[0])
		backend := strings.TrimSpace(tokens[1])
		
		if source == "" || backend == "" {
			continue
		}
		
		// ---
		
		log.Println("mapping", source, "to", backend)
		
		// ---
		
		sourceURL, sourceURLErr := url.Parse(source)
		
		if sourceURLErr != nil {
			log.Fatal(sourceURLErr)
		}
		
		// ---
		
		backendURL, backendURLErr := url.Parse(backend)
		
		if backendURLErr != nil {
			log.Fatal(backendURLErr)
		}
		
		// ---
		
		proxy := &httputil.ReverseProxy{Director: func (r *http.Request) {
			endsWithSlash := strings.HasSuffix(r.URL.Path, "/")
			
			// ---
			
			r.URL.Scheme = backendURL.Scheme
			r.URL.Host = backendURL.Host
			r.URL.Path = path.Join(backendURL.Path, r.URL.Path[len(sourceURL.Path):])
			
			// ---
			
			if endsWithSlash && !strings.HasSuffix(r.URL.Path, "/") {
				r.URL.Path = r.URL.Path + "/"
			}
		}}
		
		// ---
		
		r.HandleFunc(path.Join(sourceURL.Path, "/{path:.*}"), funcWrapper(func (w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})).Host(sourceURL.Host)
	}
	
	// --
	
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, r))
	
	// ---
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}
