package main

import (
	"io/fs"
	"os"
	"strings"
	"time"
)

type Group struct {
	Inactive     bool        `yaml:"inactive" json:"inactive"`
	Name         string      `yaml:"name" json:"name"`
	Category     string      `yaml:"category,omitempty" json:"category,omitempty"`
	URL          string      `yaml:"url" json:"url"`
	Endpoints    []*Endpoint `yaml:"endpoints,omitempty" json:"endpoints,omitempty"`
	ForcedStatus Status      `yaml:"forced_status,omitempty" json:"forced_status,omitempty"`
}

type Endpoint struct {
	Inactive     bool         `yaml:"inactive" json:"inactive"`
	Name         string       `yaml:"name" json:"name"`
	URL          string       `yaml:"url" json:"url"`
	Method       string       `yaml:"method" json:"method"`
	TargetStatus TargetStatus `yaml:"targetStatus" json:"targetStatus"`
}

type TargetStatus struct {
	Code int    `yaml:"code,omitempty" json:"code,omitempty"`
	Body []byte `yaml:"body,omitempty" json:"body,omitempty"`
}

type Error struct {
	Code    int
	Message string
}

type Result struct {
	Status Status `json:"status"`

	Code            int       `json:"code"`
	ContentType     string    `json:"content_type"`
	Body            []byte    `json:"body"`
	RequestDuration float64   `json:"request_duration"`
	Updated         time.Time `json:"updated"`
}

type FrontendFS struct {
	fs fs.FS
}

func NewFrontendFS(frontendDir string) *FrontendFS {
	return &FrontendFS{
		fs: os.DirFS(frontendDir),
	}
}
func (f *FrontendFS) Open(name string) (fs.File, error) {
	return f.fs.Open(strings.TrimPrefix(name, "frontend/"))
}

func (f *FrontendFS) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(f.fs, strings.TrimPrefix(name, "frontend/"))
}
