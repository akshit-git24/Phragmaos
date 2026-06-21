package phragmaos

import (
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type App struct {
	store *Store
}


func Run(){
	mux := http.NewServeMux()

	app := &App{store: &Store{}}
	mux.HandleFunc("GET /welcome",app.ping)
    
	log.Printf("| Started the server :")
     
	addr := getEnvOrDefault("LISTEN_PORT",":8080")
    if addr == ""{
		addr = ":8080"
	}

	host := addr
	if strings.HasPrefix(addr, ":") {
		host = "localhost" + addr
	}
	log.Printf("| API listening on http://%s\n", host)

    http.ListenAndServe(addr, mux)

}

func getUserIp(r *http.Request) string{
    ip,_,err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (a *App)ping(w http.ResponseWriter, r *http.Request){

    ip :=  getUserIp(r)
	bucket := a.store.GetOrCreate(ip)

	result := bucket.Allow()
    //Set these manuall headers before  WriteHeader call, because it cause direct flush of the response . The writes called after this functions aren't considered.
	w.Header().Set("X-RateLimit-Limit", "15")
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))

	if result.Allowed {
	   w.WriteHeader(http.StatusOK)
	   _, _ = w.Write([]byte("Allowed"))
	}else{
	   w.Header().Set("Retry-After",strconv.Itoa(result.RetryAfter))
	   w.WriteHeader(http.StatusTooManyRequests)
	   _, _ = w.Write([]byte(" Too many Requests! Limit Applied"))
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return defaultValue
}

