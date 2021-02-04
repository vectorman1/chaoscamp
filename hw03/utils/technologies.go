package utils

type ScannerData struct {
	Categories   map[string]Category   `json:"categories,omitempty"`
	Technologies map[string]Technology `json:"technologies,omitempty"`
}

type Category struct {
	Name     string `json:"name,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

type Technology struct {
	Cats        []int             `json:"cats,omitempty"`
	Description string            `json:"description,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Html        []string          `json:"html,omitempty"` //problematic
	Icon        string            `json:"icon,omitempty"`
	Implies     []string          `json:"implies,omitempty"` //problematic
	Scripts     []string          `json:"scripts,omitempty"` //problematic - replace for import "scripts":\s"(.+)" with "scripts": [ "\1" ]
	Cookies     map[string]string `json:"cookies,omitempty"`
	Website     string            `json:"website,omitempty"`
	CertIssuer  string            `json:"certIssuer,omitempty"`
	// TODO: Add remaining fields
}
