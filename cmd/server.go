package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Server struct {
	Active           bool
	port             uint
	fs               fs.ReadFileFS
	configuration    *Configuration
	webserver        *http.Server
	results          map[string]*Result
	resultsChanged   bool
	resultsCacheFile *os.File
	startTime        time.Time
	lastUpdate       time.Time
}

func NewServer(port uint, fs fs.ReadFileFS, cacheFile string, config *Configuration) *Server {
	var err error
	var resultsCacheFile *os.File
	if cacheFile == "" {
		resultsCacheFile = nil
	} else {
		resultsCacheFile, err = os.OpenFile(cacheFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	}
	if err != nil {
		outFatal(EXIT_CACHE_FILE, "Could not open cache file: %s\n", err.Error())
	}

	return &Server{
		Active:           true,
		port:             port,
		fs:               fs,
		configuration:    config,
		resultsCacheFile: resultsCacheFile,
	}
}

func (s *Server) setConfiguration(config *Configuration) {
	s.configuration = config
	s.updateAllGroups()
}

func (s *Server) Run() error {
	s.startTime = time.Now()

	// Initialize web handler
	webHandler := &http.ServeMux{}
	webHandler.HandleFunc("/", s.handleRootRequest)
	webHandler.HandleFunc("/api/", s.handleAPIRequest)
	webHandler.HandleFunc("/status/", s.handleStatusRequest)

	// Serve web application
	s.webserver = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      webHandler,
		// Errorlog:     logger.log.Getlogger(logger.logLevelError),
	}

	s.results = make(map[string]*Result, len(s.configuration.Groups))
	if s.resultsCacheFile != nil {
		data, err := io.ReadAll(s.resultsCacheFile)
		if err != nil {
			s.resultsCacheFile.Truncate(0)
			s.resultsCacheFile.Seek(0, 0)
		} else {
			err = json.Unmarshal(data, &s.results)
			if err != nil {
				outError("Could not parse cache file %s: %s\n", s.resultsCacheFile.Name(), err.Error())
				outError("Starting without cache")
				s.results = make(map[string]*Result, len(s.configuration.Groups))
			}
		}
	}

	// Allow graceful shutdown
	go s.checkForShutdown()

	// Keep Group data up to date
	go s.checkUpdateGroups()

	go s.checkResultsUpdate()

	out("Listening on port %d\n", s.port)
	return s.webserver.ListenAndServe()
}

func (s *Server) checkResultsUpdate() {
	if s.resultsCacheFile == nil {
		return
	}

	for s.Active {
		time.Sleep(2 * time.Second)

		if s.resultsChanged {

			data, err := json.Marshal(s.results)
			if err != nil {
				outError("Cannot save results to cache: %s", err.Error())
			} else {
				err := s.resultsCacheFile.Truncate(0)
				if err != nil {
					outError("Cannot save results to cache: Error truncating cache file")
				}
				written, err := s.resultsCacheFile.Write(data)
				if err != nil {
					outError("Cannot save results to cache: %s", err.Error())
				} else if written != len(data) {
					outError("Cannot save results to cache: Data not written completely: %s", err.Error())
					_ = s.resultsCacheFile.Truncate(0)
				}

				_, err = s.resultsCacheFile.Seek(0, 0)
				if err != nil {
					outError("Cannot save results to cache: Error resetting cache file position")
				}
			}

			outDebug("Results written to %s\n", s.resultsCacheFile.Name())
			s.resultsChanged = false
		}
	}
}

func (s *Server) checkForShutdown() {
	for s.Active {
		time.Sleep(2 * time.Second)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.webserver.Shutdown(ctx)
}

func (s *Server) checkUpdateGroups() {
	var nextUpdate = time.Now()
	for s.Active {
		if nextUpdate.Before(time.Now()) {
			s.updateAllGroups()
			nextUpdate = time.Now().Add(time.Duration(int64(s.configuration.RefreshInterval) * int64(time.Second)))
		}
		time.Sleep(10 * time.Second)
	}
}

func (s *Server) updateAllGroups() {
	for _, group := range s.configuration.Groups {
		for _, endpoint := range group.Endpoints {
			s.updateEndpoint(group, endpoint)
		}
	}
}

func (s *Server) updateEndpoint(group *Group, endpoint *Endpoint) {
	uri := s.getEndpointUrl(group, endpoint)

	result, ok := s.results[uri.String()]
	if !ok {
		result = &Result{}
	}

	s.lastUpdate = time.Now()

	if uri == nil || group.Inactive || endpoint.Inactive {
		result.Status = STATUS_INACTIVE
		result.Updated = time.Now()
		s.resultsChanged = result.Status != STATUS_INACTIVE
	} else {
		// Do not refresh more than once per minute
		if time.Since(result.Updated).Seconds() < s.configuration.RefreshInterval {
			return
		}

		result.Status = STATUS_GREEN

		startTime := time.Now()
		response, err := http.Get(uri.String())
		result.RequestDuration = time.Since(startTime).Seconds()
		result.Updated = time.Now()
		if err != nil {
			result.Body = []byte(err.Error())
			result.Code = 999
			result.Status = STATUS_RED
		} else {
			result.Code = response.StatusCode

			body, err := io.ReadAll(response.Body)
			if err != nil {
				result.Body = []byte(err.Error())
				result.Code = 998
				result.Status = STATUS_RED
			} else {
				result.ContentType = response.Header.Get("Content-Type")
				result.Body = body
			}

			if endpoint.TargetStatus.Code == 0 {
				// Check for Code in 200 range
				if response.StatusCode < 200 || response.StatusCode >= 300 {
					result.Status = STATUS_RED
				}
			}

			if result.Status == STATUS_GREEN && endpoint.TargetStatus.Code > 0 {
				// Check for exact code
				if response.StatusCode != endpoint.TargetStatus.Code {
					result.Status = STATUS_RED
				}
			}

			if result.Status == STATUS_GREEN && len(endpoint.TargetStatus.Body) > 0 {
				// Copare response body
				if !bytes.Equal(body, endpoint.TargetStatus.Body) {
					result.Status = STATUS_RED
				}
			}
		}

		// TODO: Save result for statistical analysis

		if endpoint.TargetStatus.Code != 0 && result.Code != endpoint.TargetStatus.Code {
			// Check response status
			result.Status = STATUS_RED
		}
		if len(endpoint.TargetStatus.Body) != 0 && !bytes.Equal(result.Body, endpoint.TargetStatus.Body) {
			// Check response status
			result.Status = STATUS_RED
		}

		outDebug("GET %s%s --> %d (%f)\n", group.URL, endpoint.URL, result.Code, result.RequestDuration)
		s.resultsChanged = true
	}

	s.results[uri.String()] = result
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, response any) {
	data, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("{ \"message\": \"Internal Server Error\", \"details\": \"%#v\" }", err.Error())))
		return
	}
	w.Write(data)
}

func (s *Server) respondConfig() any {
	return s.configuration
}

func (s *Server) respondReadAll() any {
	return s.results
}

func (s *Server) groupByName(groupName string) *Group {
	for _, l := range s.configuration.Groups {
		if l.Name == groupName {
			return l
		}
	}
	return nil
}

func (s *Server) endpointByName(group *Group, endpointName string) *Endpoint {
	for _, e := range group.Endpoints {
		if e.Name == endpointName {
			return e
		}
	}
	return nil
}

func (s *Server) getEndpointUrl(group *Group, endpoint *Endpoint) *url.URL {
	baseUri, err := url.Parse(group.URL)
	if err != nil {
		outError("Invalid base URL for group %s: %s. Error: %s", group.Name, group.URL, err.Error())
		return nil
	}

	uri, err := baseUri.Parse(endpoint.URL)
	if err != nil {
		outError("Invalid URL for endpoint %s in group %s: %s. Error: %s", endpoint.Name, group.Name, endpoint.URL, err.Error())
		return nil
	}

	return uri
}

func (s *Server) respondRead(groupName string, endpointName string) any {
	group := s.groupByName(groupName)
	endpoint := s.endpointByName(group, endpointName)

	uri := s.getEndpointUrl(group, endpoint)

	res, ok := s.results[uri.String()]
	if !ok {
		return Error{
			Code:    400,
			Message: "Invalid group/endpoint selection",
		}
	}
	return res
}

func (s *Server) respondRefresh(groupName string, endpointName string) any {
	group := s.groupByName(groupName)
	endpoint := s.endpointByName(group, endpointName)

	if group == nil || endpoint == nil {
		return Error{
			Code:    400,
			Message: "Invalid group/endpoint selection",
		}
	}

	s.updateEndpoint(group, endpoint)
	return s.respondRead(group.Name, endpoint.Name)
}
