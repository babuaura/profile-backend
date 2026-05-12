package web

import "net/http"

func WithCORS(allowedOrigins []string, next http.Handler) http.Handler {
    allowed := make(map[string]bool, len(allowedOrigins))
    for _, origin := range allowedOrigins { allowed[origin] = true }
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        if allowed[origin] { w.Header().Set("Access-Control-Allow-Origin", origin); w.Header().Set("Vary", "Origin") }
        w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
        if r.Method == http.MethodOptions { w.WriteHeader(http.StatusNoContent); return }
        next.ServeHTTP(w, r)
    })
}
