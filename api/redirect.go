package api

import (
	"net/http"
)

func redirect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := redirectResponseWriter{
			WrappedWriter: w,
		}
		next.ServeHTTP(&rw, r)
		if rw.Redirect && r.URL.RequestURI() != "/" {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}
	})
}

type redirectResponseWriter struct {
	WrappedWriter http.ResponseWriter
	Redirect      bool
}

func (r *redirectResponseWriter) Header() http.Header {
	return r.WrappedWriter.Header()
}

func (r *redirectResponseWriter) Write(buffer []byte) (int, error) {
	if r.Redirect {
		return len(buffer), nil
	}
	return r.WrappedWriter.Write(buffer)
}

func (r *redirectResponseWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusNotFound {
		r.Redirect = true
		return
	}
	r.WrappedWriter.WriteHeader(statusCode)
}
