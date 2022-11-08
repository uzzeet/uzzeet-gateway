package models

const (
	BodyTypeJSON = "application/json"
	BodyTypeXML  = "application/xml"
	BodyTypeRaw  = "raw"
)

type AuthorizationInfo struct {
	UserID         string      `json:"user_id"`
	Username       string      `json:"username"`
	UserFullname   string      `json:"user_fullname"`
	Email          string      `json:"email"`
	RoleID         string      `json:"role_id"`
	MemberID       string      `json:"member_id"`
	DealerID       interface{} `json:"dealer_id"`
	IDOrganization interface{} `json:"id_organization"`
	GroupID        []GroupID   `json:"group_id"`
	Phone          string      `json:"phone"`
	EpooolToken    string      `json:"epoool_token"`
	Iat            int         `json:"iat"`
	Exp            int         `json:"exp"`
}

/*type AuthorizationInfo struct {
	UserID         string        `json:"id"`
	Username       string        `json:"username"`
	IsOrgAdmin     int           `json:"isorgadmin"`
	IsActive       int           `json:"isactive"`
	OrganizationId string        `json:"organization_id"`
	AppId          string        `json:"app"`
	Exp            int           `json:"exp"`
	UserAccess     []interface{} `json:"user_access"`
}*/

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
