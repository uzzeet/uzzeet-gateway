package models

import (
	"github.com/dgrijalva/jwt-go"
	"net/url"
	"time"
)

const (
	TimestampFormat string = "2006-01-02T15:04:05-0700"
	TokenTypeBearer string = "bearer"
)

type Request struct {
	Method      string
	URL         *url.URL
	Body        []byte
	CompositeID CompositeID
	Timestamp   time.Time
	Signature   string
	Token       string
}

type CompositeID string

type Composite struct {
	ID     CompositeID `json:"id"`
	Secret string      `json:"secret"`
}

type TokenClaims struct {
	jwt.StandardClaims
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
	SoleType       string      `json:"role_type"`
}

type GroupID struct {
	GroupID string `json:"group_id"`
}

/*type TokenClaims struct {
	jwt.StandardClaims
	UserID         string        `json:"id"`
	Username       string        `json:"username"`
	IsOrgAdmin     int           `json:"isorgadmin"`
	IsActive       int           `json:"isactive"`
	OrganizationId string        `json:"organization_id"`
	AppId          string        `json:"app"`
	Exp            int           `json:"exp"`
	UserAccess     []interface{} `json:"user_access"`
}*/

type Token struct {
	Type  string
	Value string
}
