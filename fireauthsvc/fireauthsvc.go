// Package fireauthsvc provides services for Firebase authentication.
// It includes functionality to initialize a Firebase authentication client
// using a Firebase app instance. The package utilizes Google Wire for dependency
// injection, allowing for easy integration and testing of the authentication service.

package fireauthsvc

import (
	"context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/google/wire"
	"github.com/pkg/errors"
)

var DefaultWireset = wire.NewSet(
	NewFirebaseAuthSvc,
)

// NewFirebaseAuthSvc initializes and returns a Firebase authentication client.
// It takes a Firebase app instance as input and returns the authentication client.

func NewFirebaseAuthSvc(
	firebaseApp *firebase.App,
) (*auth.Client, error) {
	client, err := firebaseApp.Auth(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize firebase auth client")
	}
	return client, nil
}
