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
	UserID         string `json:"id"`
	Username       string `json:"username"`
	IsOrgAdmin     int    `json:"isorgadmin"`
	IsActive       int    `json:"isactive"`
	OrganizationId string `json:"organizationid"`
	AppId          string `json:"app"`
	Exp            int    `json:"exp"`
}

type Token struct {
	Type  string
	Value string
}
