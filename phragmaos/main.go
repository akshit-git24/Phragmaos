package phragmaos

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type App struct {
	store *Store
	extractor IdentifierExtractor
	config   *Config
}

type IdentifierExtractor interface {
    Extract(r *http.Request) string
}

// Rate Limiting params
type IPExtractor struct{}
type APIKeyExtractor struct{}

func (e *IPExtractor) Extract(r *http.Request) (string){
	ip,_,err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (e *APIKeyExtractor)Extract(r *http.Request) (string){
	 api_key := r.Header.Get("X-API-KEY")
     return api_key
}

func Run(){

	data,err := os.ReadFile("sample.json")

	var cfg Config
    if err == nil {
       json.Unmarshal(data, &cfg)
	   log.Printf("Config :",&cfg)
	}else{
		log.Printf("No config setted")
		log.Fatal(err)
	}
	
	mux := http.NewServeMux()
    extractorType := os.Getenv("EXTRACTOR")
	app := &App{
		store: &Store{},
		extractor: getExtractor(os.Getenv("EXTRACTOR")),
		config: &cfg,
    }
	mux.HandleFunc("GET /welcome",app.ping) 
    
	if extractorType == ""{
		extractorType = "IP"
	}
    log.Printf("| Rate Limit Model  : %s ",extractorType)

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

func getExtractor(t string) (IdentifierExtractor){
	 if t =="apikey" || t =="APIKEY" || t=="API_KEY" || t=="API-KEY"{
		return &APIKeyExtractor{}
	 }
    return &IPExtractor{}
}

func (a *App) ping(w http.ResponseWriter, r *http.Request){

    ip :=  a.extractor.Extract(r)

	endpoints := a.config.findEndpoint(r.URL.Path)
    key := ip + ":" + r.URL.Path
	bucket := a.store.GetOrCreate(key,&endpoints)

	result := bucket.Allow()
    //Set these manual headers before WriteHeader call, because it cause direct flush of the response . The writes called after this functions aren't considered.
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

// func getEnvOrDefaultInt(key string, defaultValue int) int {
//     if val := strings.TrimSpace(os.Getenv(key)); val != "" {
//         if parsed, err := strconv.Atoi(val); err == nil {
//             return parsed
//         }
//     }
//     return defaultValue
// }

func (c *Config) findEndpoint(path string) EndpointConfig {
    for _, e := range c.Endpoints {
        if e.Path == path {
            log.Printf("e :",e)
            return e
        }
    }
    // default if path not in config
    return EndpointConfig{
        Path:       path,
        Limit:      15,
        RefillRate: 1,
    }
}
// func getExtractor(t string) IdentifierExtractor {
//     if t == "apikey" {
//         return &APIKeyExtractor{}
//     }
//     return &IPExtractor{} // default
// }

