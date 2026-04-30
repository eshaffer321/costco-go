package costco

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ImportTokenResponse converts a raw TokenResponse from the Costco token endpoint
// into StoredTokens ready to be persisted with SaveTokens.
//
// The expected source is the JSON response body from:
//
//	POST https://signin.costco.com/.../oauth2/v2.0/token
//
// Users can obtain this by logging into costco.com, opening DevTools → Network,
// filtering by Fetch/XHR, searching "token", and copying the response body.
func ImportTokenResponse(resp *TokenResponse) (*StoredTokens, error) {
	if resp.IDToken == "" {
		return nil, fmt.Errorf("id_token is missing from token response")
	}
	if resp.RefreshToken == "" {
		return nil, fmt.Errorf("refresh_token is missing from token response")
	}

	return &StoredTokens{
		IDToken:               resp.IDToken,
		RefreshToken:          resp.RefreshToken,
		TokenExpiry:           parseTokenExpiry(resp.IDToken),
		RefreshTokenExpiresAt: time.Now().Add(time.Duration(resp.RefreshTokenExpiresIn) * time.Second),
	}, nil
}

func parseTokenExpiry(tokenString string) time.Time {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return time.Now().Add(15 * time.Minute)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			return time.Unix(int64(exp), 0)
		}
	}

	return time.Now().Add(15 * time.Minute)
}
