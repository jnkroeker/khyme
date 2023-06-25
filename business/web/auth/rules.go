package auth

// These are the current set of rules for Auth
const (
	RuleAuthenticate   = "auth"
	RuleAny            = "ruleAny"
	RuleAdminOnly      = "ruleAdminOnly"
	RuleUserOnly       = "ruleUserOnly"
	RuleAdminOrSubject = "ruleAdminOrSubject"
)

// Package name of out rego code
const (
	opaPackage string = "khyme.rego"
)

// Core OPA policies
var (
	//go:embed rego/authentication.rego
	opaAuthentication string
)
