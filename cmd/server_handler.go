package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleAPIRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := s.authorized(r)
	if err != nil {
		s.respond(w, r, err)
		return
	}

	parts := strings.Split(r.URL.Path, "/")[2:]

	if len(parts) == 1 {
		switch parts[0] {
		case "config":
			s.respond(w, r, s.respondConfig())

		case "read":
			s.respond(w, r, s.respondRead(r.URL.Query().Get("group"), r.URL.Query().Get("endpoint")))

		case "refresh":
			s.respond(w, r, s.respondRefresh(r.URL.Query().Get("group"), r.URL.Query().Get("endpoint")))

		case "refreshAll":
			s.updateAllGroups()
			s.respond(w, r, s.respondReadAll())

		case "readAll":
			s.respond(w, r, s.respondReadAll())

		default:
			w.WriteHeader(599)
			w.Write([]byte(fmt.Sprintf("{ \"message\": \"Not Yet Implemented\", \"details\": \"%#v\" }", parts)))

		}
	} else {
		w.WriteHeader(599)
		w.Write([]byte(fmt.Sprintf("{ \"message\": \"Not Yet Implemented\", \"details\": \"%#v\" }", parts)))
	}

}

func (s *Server) handleStatusRequest(w http.ResponseWriter, r *http.Request) {
	content := make(map[string]any)

	if r.URL.Query().Get("more") == "true" {
		header := make(map[string]string, len(r.Header))
		for k, v := range r.Header {
			header[k] = strings.Join(v, ", ")
		}
		content["header"] = header

		content["auth-error"] = s.authorized(r)
	}

	content["status"] = "up"
	content["uptime"] = time.Since(s.startTime).String()
	content["lastUpdate"] = s.lastUpdate

	output, err := json.Marshal(content)
	if err != nil {
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Internal Error"))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(output)
	}
}

func (s *Server) handleRootRequest(w http.ResponseWriter, r *http.Request) {
	authErr := s.authorized(r)
	if authErr != nil {
		s.respond(w, r, authErr)
		return
	}

	parts := strings.Split(r.URL.Path[1:], "/")
	if parts[len(parts)-1] == "" {
		parts[len(parts)-1] = "index.html"
	}
	path := strings.Join(parts, "/")

	content, err := s.fs.ReadFile("frontend/" + path)
	if err != nil {
		outDebug("Frontend: Could not find path: %s", path)
		w.WriteHeader(404)
		w.Write([]byte("<!DOCTYPE html>"))
		w.Write([]byte("<h1>404 - Not found</h1>"))
		w.Write([]byte("<p>Path not found: "))
		w.Write([]byte(path))
		w.Write([]byte("</p>"))
		return
	}

	nameParts := strings.Split(path, ".")
	ext := nameParts[len(nameParts)-1]
	tp := mime.TypeByExtension("." + ext)
	if tp == "" {
		tp = http.DetectContentType(content)
	}
	w.Header().Set("Content-Type", tp)
	w.Write(content)

}
