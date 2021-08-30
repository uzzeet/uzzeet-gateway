package models

const (
	BodyTypeJSON = "application/json"
	BodyTypeXML  = "application/xml"
	BodyTypeRaw  = "raw"
)

type AuthorizationInfo struct {
	UserID         string `json:"id"`
	Username       string `json:"username"`
	IsOrgAdmin     int    `json:"isorgadmin"`
	IsActive       int    `json:"isactive"`
	OrganizationId string `json:"organizationid"`
	AppId          string `json:"app"`
	Exp            int    `json:"exp"`
}

type ClientInfo struct {
	ClientID string `json:"client_id"`
	Agent    string `json:"agent"`
}

type RequestContext struct {
	path       string
	method     string
	body       []byte
	params     map[string]string
	query      map[string]string
	header     map[string]string
	authInfo   AuthorizationInfo
	clientInfo ClientInfo
}
