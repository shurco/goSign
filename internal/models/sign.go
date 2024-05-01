package models

// Sign is ...
type Sign struct {
	Error          string `json:"error,omitempty"`
	FileName       string `json:"file_name,omitempty"`
	FileNameSigned string `json:"file_name_signed,omitempty"`
}
