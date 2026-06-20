package phragmaos

import (
	"log"
	"net"
	"net/http"
)

type App struct {
	store *Store
}


func Run(){
	mux := http.NewServeMux()

	app := &App{store: &Store{}}
	mux.HandleFunc("GET /welcome",app.ping)
    
	log.Printf("Started the server :")
    http.ListenAndServe(":8080", mux)

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

	allowed := bucket.Allow()

	if allowed {
	   w.WriteHeader(http.StatusOK)
	   _, _ = w.Write([]byte("Allowed"))
	}else{
		w.WriteHeader(http.StatusTooManyRequests)
	   _, _ = w.Write([]byte(" Too many Requests! Limit Applied"))
	}
	
     
}