package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/uzzeet/uzzeet-gateway/controller/redis/repo"
	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"github.com/uzzeet/uzzeet-gateway/models"
	"net/url"
	"strings"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/uzzeet/uzzeet-gateway/libs/utils/uttime"
)

type Service interface {
	Authorize(models.Request) (*models.AuthorizationInfo, error)
}

type authService struct {
	method        string
	signingKey    []byte
	compositeRepo repo.CompositeRepository
	useSignature  bool
}

func NewService(method, signingKey string, compositeRepo repo.CompositeRepository, useSignature bool) Service {
	return authService{method, []byte(signingKey), compositeRepo, useSignature}
}

func (svc authService) Authorize(request models.Request) (*models.AuthorizationInfo, error) {
	var (
		userID          string
		username        string
		isOrgAdmin      int
		isActive        int
		organization_id string
		app_id          string
		exp             int
	)

	token, err := svc.parseToken(request.Token)
	if err != nil {
		return nil, err
	}

	switch token.Type {
	default:
		return nil, AuthorizationError{
			errors.New("unsupported authorization"),
			map[string]string{
				"id": "Otorisasi tidak didukung",
			},
		}
	case models.TokenTypeBearer:
		claims, err := svc.claimToken(token.Value)
		if err != nil {
			return nil, err
		}

		switch svc.method {
		case "strict":
			userID = claims.UserID
			isOrgAdmin = claims.IsOrgAdmin
			isActive = claims.IsActive
			organization_id = claims.OrganizationId
			app_id = claims.AppId

		case "private":
			if helper.MD5(fmt.Sprintf("private:[%s:%s:%s:%s:%s]", claims.UserID, claims.Username, claims.IsOrgAdmin, claims.IsActive, claims.OrganizationId, claims.AppId)) != claims.Id {
				return nil, fmt.Errorf("invalid authorization token, 0x10001")
			}

		case "protect":
			userID = claims.UserID
			isOrgAdmin = claims.IsOrgAdmin
			isActive = claims.IsActive
			organization_id = claims.OrganizationId
			app_id = claims.AppId
		}

		if svc.useSignature {
			composite, err := svc.compositeRepo.GetRedisByID(request.CompositeID)
			if err != nil {
				if err == errors.New("composite not found") {
					return nil, AuthorizationError{
						errors.New("composite authorization failed"),
						map[string]string{
							"id": "Otorisasi klien gagal",
						},
					}
				}

				return nil, err
			}

			signature, err := svc.createSignature(composite, request.Method, request.URL, token.Value, request.Timestamp, request.Body)
			if err != nil {
				return nil, err
			}

			err = svc.verifySignature([]byte(request.Signature), signature)
			if err != nil {
				return nil, err
			}
		}
	}

	return &models.AuthorizationInfo{
		UserID:         userID,
		Username:       username,
		IsOrgAdmin:     isOrgAdmin,
		IsActive:       isActive,
		OrganizationId: organization_id,
		AppId:          app_id,
		Exp:            exp,
	}, nil
}

func (svc authService) parseToken(source string) (models.Token, error) {
	var t models.Token

	separator := " "
	typeSection := 0
	valueSection := 1
	expectedTokenLength := 2

	if source == "" {
		return t, AuthorizationError{
			errors.New("token not found"),
			map[string]string{
				"id": "Token tidak ditemukan",
			}}
	}

	tokens := strings.Split(source, separator)
	if len(tokens) != expectedTokenLength {
		return t, AuthorizationError{
			errors.New("invalid token"),
			map[string]string{
				"id": "Token tidak valid",
			}}
	}

	t.Type = strings.ToLower(tokens[typeSection])
	t.Value = tokens[valueSection]

	return t, nil
}

func (svc authService) claimToken(tokenString string) (*models.TokenClaims, error) {
	switch tokenString {
	case helper.Env("X_TOKEN_UNTIL_20210101", "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJleHAiOjE1OTM4MzgyNTksImlhdCI6MTU5MTI0NjI1OSwiaXNzIjoiYmFja29mZmljZS0yLjAiLCJqdGkiOiI0NDRmNzAzZTJkYTI0ZGYxMzZlOWM5YTk2OTc3N2E3OCIsInVhdCI6IjFkYzg2MGFmMjBkNGM3Yzg2ZDdkZWNmYzAxOGZmMzU4WkRRMk9HTXdZekF0Wm1KaFlTMHhNV1U1TFdJME16Y3RNREF4TmpObE1ERTJaRFJqIiwidWVtIjoiMWRjODYwYWYyMGQ0YzdjODZkN2RlY2ZjMDE4ZmYzNThZVzVrZDJsQWEyOXBibmR2Y210ekxtTnZiUT09IiwidWlkIjoiMWRjODYwYWYyMGQ0YzdjODZkN2RlY2ZjMDE4ZmYzNThNUT09IiwidW5xIjoiMWRjODYwYWYyMGQ0YzdjODZkN2RlY2ZjMDE4ZmYzNThOVFkzT1RNek5tTXRZekkxTXkwMFpUWXlMVGxtWm1FdFlqY3paR0l4TVRCak4yTTAiLCJ1YWQiOiIxZGM4NjBhZjIwZDRjN2M4NmQ3ZGVjZmMwMThmZjM1OFlYTnpaWE56YjNJPSJ9.XbS48y2XHkSqf9WMRFqDCzSBJSXtTDhHYz-imbU8mos"):
		if svc.method != "strict" {
			return nil, AuthorizationError{
				errors.New("Invalid token"),
				map[string]string{
					"id": "Token tidak valid",
				},
			}
		}

		// max lifetime of token
		maxTime, _ := uttime.Parse("2021-01-15")
		if time.Now().After(maxTime) {
			return nil, AuthorizationError{
				errors.New("Token has been blocked, due moving to gRPC."),
				map[string]string{
					"id": "Token tidak valid",
				},
			}
		}

		return &models.TokenClaims{
			UserID:         "1",
			Username:       fmt.Sprintf("admin+%s@cleva.com", uttime.ToString("Ymd", time.Now())),
			IsOrgAdmin:     0,
			IsActive:       0,
			OrganizationId: "0",
			AppId:          "0",
			Exp:            0,
		}, nil
	}

	//key := string(svc.signingKey)

	token, err := jwt.ParseWithClaims(tokenString, &models.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, AuthorizationError{
				errors.New("invalid token"),
				map[string]string{
					"id": "Token tidak valid",
				},
			}
		}

		return []byte("um_phrase"), nil
	})
	if err != nil {
		return nil, AuthorizationError{
			err,
			map[string]string{
				"id": "Token tidak valid",
			},
		}
	}

	return token.Claims.(*models.TokenClaims), nil
}

func (svc authService) createSignature(composite models.Composite, method string, uri *url.URL, token string, timestamp time.Time, body []byte) ([]byte, error) {
	emptyString := ""
	relativePath := uri.EscapedPath()
	encodedQuery := strings.Replace(uri.Query().Encode(), "+", "%20", -1)
	if encodedQuery != emptyString {
		relativePath = fmt.Sprintf("%s?%s", relativePath, encodedQuery)
	}

	stringBody := string(body)
	trimmedCharacters := []string{
		" ", "\t", "\n", "\r",
	}
	for _, trimmedCharacter := range trimmedCharacters {
		stringBody = strings.ReplaceAll(stringBody, trimmedCharacter, emptyString)
	}

	sha256Body := sha256.Sum256([]byte(stringBody))
	hexEncodedBody := strings.ToLower(hex.EncodeToString(sha256Body[:]))
	stringToSign := fmt.Sprintf(
		"%s;%s;%s;%s;%s",
		method,
		relativePath,
		token,
		timestamp.Format(models.TimestampFormat),
		hexEncodedBody,
	)

	mac := hmac.New(sha256.New, []byte(composite.Secret))
	_, err := mac.Write([]byte(stringToSign))
	if err != nil {
		return nil, fmt.Errorf("while signing signature: %v", err)
	}

	return []byte(hex.EncodeToString(mac.Sum(nil))), nil
}

func (svc authService) verifySignature(signature, expectedSignature []byte) error {
	if !hmac.Equal([]byte(signature), expectedSignature) {
		return AuthorizationError{
			errors.New("signature mismatch"),
			map[string]string{
				"id": "Signature tidak cocok",
			},
		}
	}

	return nil
}
