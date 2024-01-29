package connection

import (
	"context"
	"fmt"

	"github.com/golden-vcr/chatbot"
	"github.com/nicklaw5/helix/v2"
)

// TwitchClient represents the subset of Twitch API functionality required to
// authenticate the chat bot account in order to connect to IRC
type TwitchClient interface {
	GetUserAccessTokenFromCode(code string) (*helix.AccessCredentials, error)
	RefreshCredentials(ctx context.Context, credentials *helix.AccessCredentials) (*helix.AccessCredentials, error)
	ResolveUserInfo(userAccessToken string) (*helix.User, error)
}

// NewTwitchClient initializes a TwitchClient that's prepared to initiate a user
// authentication flow against the configured Twitch client application
func NewTwitchClient(ctx context.Context, clientId, clientSecret, redirectUri string) (TwitchClient, error) {
	c, err := helix.NewClientWithContext(ctx, &helix.Options{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURI:  redirectUri,
	})
	if err != nil {
		return nil, err
	}
	return &twitchClient{
		Client: c,
		makeUserAuthorizedClient: func(userAccessToken string) (*helix.Client, error) {
			return helix.NewClientWithContext(ctx, &helix.Options{
				ClientID:        clientId,
				UserAccessToken: userAccessToken,
			})
		},
	}, nil
}

// twitchClient is a concrete implementation of TwitchClient that uses the helix Client
// library to make calls agains tthe Twitch API
type twitchClient struct {
	*helix.Client
	makeUserAuthorizedClient func(userAccessToken string) (*helix.Client, error)
}

// GetUserAccessTokenFromCode exchanges an access code (presented at the end of an OAuth
// flow completed by the user) for a Twitch User Access Token that can be used to
// connect to IRC as that user
func (c *twitchClient) GetUserAccessTokenFromCode(code string) (*helix.AccessCredentials, error) {
	// Hit the Twitch API to exchange the code for a User Access Token
	res, err := c.RequestUserAccessToken(code)
	if err != nil {
		return nil, fmt.Errorf("failed to request user access token: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("request for user access token failed with status %d: %s", res.StatusCode, res.ErrorMessage)
	}

	// Ensure that the user authorized all required scopes
	for _, requiredScope := range chatbot.RequiredScopes {
		hasScope := false
		for _, scope := range res.Data.Scopes {
			if scope == requiredScope {
				hasScope = true
				break
			}
		}
		if !hasScope {
			return nil, fmt.Errorf("access token is missing required scope %s", requiredScope)
		}
	}

	return &res.Data, nil
}

func (c *twitchClient) RefreshCredentials(ctx context.Context, credentials *helix.AccessCredentials) (*helix.AccessCredentials, error) {
	res, err := c.RefreshUserAccessToken(credentials.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh user access token: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("request to refresh user access token failed with status %d: %s", res.StatusCode, res.ErrorMessage)
	}

	// Ensure that the user authorized all required scopes
	for _, requiredScope := range chatbot.RequiredScopes {
		hasScope := false
		for _, scope := range res.Data.Scopes {
			if scope == requiredScope {
				hasScope = true
				break
			}
		}
		if !hasScope {
			return nil, fmt.Errorf("refreshed access token is missing required scope %s", requiredScope)
		}
	}

	return &res.Data, nil
}

// ResolveUserInfo queries the Twitch API and returns the details of the user to whom
// the given access token has been granted
func (c *twitchClient) ResolveUserInfo(userAccessToken string) (*helix.User, error) {
	client, err := c.makeUserAuthorizedClient(userAccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.GetUsers(&helix.UsersParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get user identity from access token: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("request for user identity failed with status %d: %s", res.StatusCode, res.ErrorMessage)
	}
	if len(res.Data.Users) != 1 {
		return nil, fmt.Errorf("request for user identity got %d user results; expected exactly 1", len(res.Data.Users))
	}
	return &res.Data.Users[0], nil
}
