package axiom

import (
	"context"
	"encoding/json"
	"net/http"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=UserRole -linecomment -output=users_string.go

// UserRole represents the role of an [User].
type UserRole uint8

// All available [User] roles.
const (
	RoleCustom   UserRole = iota // custom
	RoleNone                     // none
	RoleReadOnly                 // read-only
	RoleUser                     // user
	RoleAdmin                    // admin
	RoleOwner                    // owner
)

func userRoleFromString(s string) (ur UserRole) {
	switch s {
	case RoleNone.String():
		ur = RoleNone
	case RoleReadOnly.String():
		ur = RoleReadOnly
	case RoleUser.String():
		ur = RoleUser
	case RoleAdmin.String():
		ur = RoleAdmin
	case RoleOwner.String():
		ur = RoleOwner
	default:
		ur = RoleCustom
	}

	return ur
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the
// UserRole to its string representation because that's what the server expects.
func (ur UserRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(ur.String())
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// UserRole from the string representation the server returns.
func (ur *UserRole) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*ur = userRoleFromString(s)

	return
}

// User represents an user.
type User struct {
	// ID is the unique ID of the user.
	ID string `json:"id"`
	// Name of the user.
	Name string `json:"name"`
	// Emails are the email addresses of the user.
	Emails []string `json:"emails"`
}

// UsersService handles communication with the user related operations of the
// Axiom API.
//
// Axiom API Reference: /v1/users
type UsersService service

// Current retrieves the authenticated user.
func (s *UsersService) Current(ctx context.Context) (*User, error) {
	ctx, span := s.client.trace(ctx, "Users.Current")
	defer span.End()

	var res User
	if err := s.client.Call(ctx, http.MethodGet, "/v1/user", nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}
