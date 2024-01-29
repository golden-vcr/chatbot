package connection

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golden-vcr/chatbot"
	"github.com/golden-vcr/chatbot/internal/csrf"
	"github.com/golden-vcr/chatbot/internal/state"
	"github.com/golden-vcr/chatbot/internal/tokens"
	"github.com/golden-vcr/server-common/entry"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
)

// Server implements the HTTP endpoints that allow us to verify the status of the
// connection and handle logging in via Twitch in order to enable our bot to connect to
// IRC with a User Access Token
type Server struct {
	agent        state.Agent
	twitchClient TwitchClient
	clientId     string
	redirectUri  string
	botUsername  string

	csrfBuffer csrf.Buffer
	tokenStore tokens.Store
	ircMu      sync.Mutex
}

func NewServer(ctx context.Context, logger *slog.Logger, agent state.Agent, twitchClient TwitchClient, clientId, redirectUri, botUsername string, tokenStore tokens.Store) *Server {
	// Load any previously-stored credentials
	credentials, err := tokenStore.Load()
	if err == nil {
		// If we have stored credentials, use the refresh token to request a new user
		// access token
		credentials, err = twitchClient.RefreshCredentials(ctx, credentials)
		if err == nil {
			// If the refresh was successful, attempt to reinitialize the agent so that
			// we automatically initialize a bot and log it in
			err = agent.Reinitialize(credentials.AccessToken, 3*time.Second)
			if err == nil {
				// If we successfully logged in, store our post-refresh credentials
				logger.Info("Successfully initialized agent from stored credentials")
				err = tokenStore.Save(credentials)
				if err != nil {
					logger.Warn("Failed to store refreshed credentials", "error", err)
				}
			} else {
				logger.Warn("Failed to initialized agent from stored credentials", "error", err)
			}
		} else {
			logger.Warn("Failed to refresh stored credentials", "error", err)
		}
	} else {
		logger.Warn("Failed to load stored credentials", "error", err)
	}

	return &Server{
		agent:        agent,
		twitchClient: twitchClient,
		clientId:     clientId,
		redirectUri:  redirectUri,
		botUsername:  botUsername,

		csrfBuffer: csrf.NewBuffer(ctx, time.Minute, 10),
		tokenStore: tokenStore,
	}
}

func (s *Server) RegisterRoutes(r *mux.Router) {
	r.Path("/status").Methods("GET").HandlerFunc(s.handleGetStatus)
	r.Path("/login").Methods("GET").HandlerFunc(s.handleGetLogin)
	r.Path("/auth").Methods("GET").HandlerFunc(s.handleGetAuth)
}

func (s *Server) handleGetStatus(res http.ResponseWriter, req *http.Request) {
	status := s.agent.GetStatus()
	res.Write([]byte(status))
}

func (s *Server) handleGetLogin(res http.ResponseWriter, req *http.Request) {
	u, err := url.Parse("https://id.twitch.tv/oauth2/authorize")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", s.clientId)
	q.Set("redirect_uri", s.redirectUri)
	q.Set("scope", strings.Join(chatbot.RequiredScopes, " "))
	q.Set("state", s.csrfBuffer.Peek())
	u.RawQuery = q.Encode()

	res.Header().Set("location", u.String())
	res.WriteHeader(http.StatusFound)
}

func (s *Server) handleGetAuth(res http.ResponseWriter, req *http.Request) {
	// Prevent CSRF by ensuring that the 'state' value matches a CSRF token value that
	// we generated within recent memory
	csrfToken := req.URL.Query().Get("state")
	if csrfToken == "" || !s.csrfBuffer.Contains(csrfToken) {
		http.Error(res, "access denied: invalid CSRF token", http.StatusUnauthorized)
		return
	}

	// Check to see if the authorization flow ended in failure, and return 401 if so
	errorValue := req.URL.Query().Get("error")
	errorDescriptionValue := req.URL.Query().Get("error_description")
	if errorValue != "" || errorDescriptionValue != "" {
		http.Error(res, fmt.Sprintf("%s: %s", errorValue, errorDescriptionValue), http.StatusUnauthorized)
		return
	}

	// Otherwise, ensure that we have a code in the URL that we can exchange for a User
	// Access Token
	code := req.URL.Query().Get("code")
	if code == "" {
		http.Error(res, "access denied: missing code", http.StatusUnauthorized)
		return
	}

	// Exchange the code for a User Access Token
	credentials, err := s.twitchClient.GetUserAccessTokenFromCode(code)
	if err != nil {
		http.Error(res, fmt.Sprintf("token exchange failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Verify that the login is coming from the authorized Twitch account for our chat
	// bot: if any other Twitch user attempts to connect to our chat bot app, we want to
	// simply ignore the login
	user, err := s.twitchClient.ResolveUserInfo(credentials.AccessToken)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to resolve user info: %v", err), http.StatusInternalServerError)
		return
	}
	if user.Login != strings.ToLower(s.botUsername) {
		http.Error(res, fmt.Sprintf("user %s is not an authorized chat bot account", user.Login), http.StatusForbidden)
		return
	}

	// Reinitialize the agent with our new user access token, causing any existing bot
	// to be replaced by a new one with a fresh login
	if err := s.agent.Reinitialize(credentials.AccessToken, 3*time.Second); err != nil {
		http.Error(res, fmt.Sprintf("chat bot initialization failed: %v", err), http.StatusInternalServerError)
		return
	}

	// If we successfully reinitialized the agent, our new credentials are good: store
	// them so we can refresh/reuse them in the future, so we can restart the server
	// without having to prompt the user to log in again
	if err := s.tokenStore.Save(credentials); err != nil {
		entry.Log(req).Error("Failed to store Twitch tokens", "error", err)
	}

	res.Write([]byte(fmt.Sprintf("Successfully initialized chat bot %s.", user.DisplayName)))
}
