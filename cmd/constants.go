package main

const (
	EXIT_OK           = 0
	EXIT_CLI_ARGS     = 1
	EXIT_PARSE_CONFIG = 2
	EXIT_CACHE_FILE   = 4
)

type Status string

const (
	STATUS_GREEN    Status = "green"
	STATUS_YELLOW   Status = "yellow"
	STATUS_RED      Status = "red"
	STATUS_INACTIVE Status = "grey"
)

const (
	AUTH_TYPE_NONE      = "none"
	AUTH_TYPE_CERT      = "client-cert"
	AUTH_TYPE_CERT_INFO = "client-cert-info"
)
