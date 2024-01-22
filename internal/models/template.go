package models

import "time"

type Template struct {
	ID         string      `json:"id"`
	FolderID   string      `json:"folder_id"`
	Slug       string      `json:"slug"`
	Name       string      `json:"name"`
	Source     string      `json:"source,omitempty"`
	Author     *Author     `json:"author,omitempty"`
	Submitters []Submitter `json:"submitters"`
	Fields     []Field     `json:"fields"`
	Schema     []Schema    `json:"schema"`
	Documents  []Document  `json:"documents"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	ArchivedAt *time.Time  `json:"archived_at,omitempty"`
}

type Author struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type Submitter struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Field struct {
	ID          string   `json:"id"`
	SubmitterID string   `json:"submitter_id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Areas       []*Areas `json:"areas,omitempty"`
}

type Areas struct {
	AttachmentID string  `json:"attachment_id"`
	Page         int     `json:"page"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	W            float64 `json:"w"`
	H            float64 `json:"z"`
}

type Schema struct {
	AttachmentID string `json:"attachment_id"`
	Name         string `json:"name"`
}

type Document struct {
	ID            string          `json:"id"`
	URL           string          `json:"url"`
	FileName      string          `json:"filename,omitempty"`
	Metadata      DocMetadata     `json:"metadata"`
	PreviewImages []PreviewImages `json:"preview_images"`
	CreatedAt     time.Time       `json:"created_at"`
}

type DocMetadata struct {
	Analyzed bool   `json:"analyzed,omitempty"`
	Pdf      Pdf    `json:"pdf"`
	Sha256   string `json:"sha256,omitempty"`
}

type Pdf struct {
	Annotations   []*Annotations `json:"annotations,omitempty"`
	NumberOfPages int            `json:"number_of_pages"`
}

type Annotations struct {
	Type  string  `json:"type"`
	Value string  `json:"value"`
	Page  int     `json:"page"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	W     float64 `json:"w"`
	H     float64 `json:"z"`
}

type PreviewImages struct {
	ID         string      `json:"id"`
	RecordType string      `json:"record_type,omitempty"`
	RecordID   string      `json:"record_id,omitempty"`
	BlobID     string      `json:"blob_id,omitempty"`
	Metadata   ImgMetadata `json:"metadata"`
	FileName   string      `json:"filename"`
}

type ImgMetadata struct {
	Analyzed   bool `json:"analyzed,omitempty"`
	Identified bool `json:"identified,omitempty"`
	Width      int  `json:"width"`
	Height     int  `json:"height"`
}
