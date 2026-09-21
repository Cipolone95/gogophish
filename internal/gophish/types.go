package gophish

type Campaign struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	CreatedDate   string   `json:"created_date"`
	LaunchDate    string   `json:"launch_date"`
	SendByDate    *string  `json:"send_by_date"`
	CompletedDate string   `json:"completed_date"`
	Template      Template `json:"template"`
	Page          Page     `json:"page"`
	Status        string   `json:"status"`
	Groups        []Group  `json:"groups"`
	SMTP          SMTP     `json:"smtp"`
	URL           string   `json:"url"`
}

type Template struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Subject     string       `json:"subject"`
	Text        string       `json:"text"`
	HTML        string       `json:"html"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Content string `json:"content"`
	Type    string `json:"type"`
	Name    string `json:"name"`
}

type Page struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type SMTP struct {
	ID               int64    `json:"id"`
	Name             string   `json:"name"`
	Username         string   `json:"username,omitempty"`
	Password         string   `json:"password,omitempty"`
	Host             string   `json:"host"`
	InterfaceType    string   `json:"interface_type,omitempty"`
	FromAddress      string   `json:"from_address"`
	IgnoreCertErrors bool     `json:"ignore_cert_errors"`
	ModifiedDate     string   `json:"modified_date,omitempty"`
	Headers          []Header `json:"headers,omitempty"`
}

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Group struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	ModifiedDate string   `json:"modified_date"`
	Targets      []Target `json:"targets"`
}

type Target struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Position  string `json:"position"`
}

type NameRef struct {
	Name string `json:"name"`
}

type CreateCampaignRequest struct {
	Name       string    `json:"name"`
	Template   NameRef   `json:"template"`
	URL        string    `json:"url"`
	Page       *NameRef  `json:"page,omitempty"`
	SMTP       NameRef   `json:"smtp"`
	LaunchDate string    `json:"launch_date,omitempty"`
	SendByDate *string   `json:"send_by_date,omitempty"`
	Groups     []NameRef `json:"groups"`
}
