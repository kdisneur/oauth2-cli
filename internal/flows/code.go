package flows

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// OpenBrowser takes an URL and open the client browser
type OpenBrowser func(url string)

// AuthorizationCodeInput represents all the data required to perform an authorization code flow
type AuthorizationCodeInput struct {
	AuthorizeURI string
	TokenURI     string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	Scope        string
	State        string
}

// AuthorizationCodeOptions represents all the possible customization of the authorization code flow
type AuthorizationCodeOptions struct {
	HTTPServerPort int
	OpenBrowser    OpenBrowser
}

// AuthorizationCodeOutput returns back the most useful information (AccessToken and IDToken) alongside the full raw JSON response
type AuthorizationCodeOutput struct {
	AccessToken string
	IDToken     *string
	RawResponse map[string]interface{}
}

// RunAuthorizationCode starts a new authorization code flow:
// - open a browser on the authorization URL
// - spawn a web server to receive the authorization code
// - exchange the authorization code to an access token
// - parse the token response
func RunAuthorizationCode(input AuthorizationCodeInput, opts AuthorizationCodeOptions) (AuthorizationCodeOutput, error) {
	authorizeURI, err := url.Parse(input.AuthorizeURI)
	if err != nil {
		return AuthorizationCodeOutput{}, fmt.Errorf("can't parse authorize uri: %v", err)
	}

	values := authorizeURI.Query()
	values.Set("client_id", input.ClientID)
	values.Set("redirect_uri", input.RedirectURI)
	values.Set("scope", input.Scope)
	values.Set("response_type", "code")
	values.Set("state", input.State)

	authorizeURI.RawQuery = values.Encode()

	opts.OpenBrowser(authorizeURI.String())

	type AuthResult struct {
		Code string
		Err  error
	}

	resultChannel := make(chan AuthResult)
	go func(result chan<- AuthResult, expectedState string) {
		http.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html><body><script type="text/javascript">window.close()</script></body></html>`))
			queryParams := r.URL.Query()

			actualState := queryParams.Get("state")
			if actualState != expectedState {
				err := fmt.Errorf("invalid state. expected:%s; got:%s", expectedState, actualState)
				result <- AuthResult{Err: err}
				return
			}

			errorCode := queryParams.Get("error")
			errorDescription := queryParams.Get("error_description")
			if errorCode != "" || errorDescription != "" {
				err := fmt.Errorf("oauth server returned an error. code:%s; description: %s", errorCode, errorDescription)
				result <- AuthResult{Err: err}
				return
			}

			code := r.URL.Query().Get("code")
			if code == "" {
				err := fmt.Errorf("code is not part of the parameters. url:%s", r.URL.String())
				result <- AuthResult{Err: err}
				return
			}

			result <- AuthResult{Code: code}
		})

		if err := http.ListenAndServe(fmt.Sprintf(":%d", opts.HTTPServerPort), nil); err != nil {
			result <- AuthResult{Err: fmt.Errorf("http server stopped", err.Error())}
		}
	}(resultChannel, input.State)

	result := <-resultChannel
	if result.Err != nil {
		return AuthorizationCodeOutput{}, result.Err
	}

	body := url.Values{}
	body.Set("grant_type", "authorization_code")
	body.Set("code", result.Code)
	body.Set("redirect_uri", input.RedirectURI)
	body.Set("client_id", input.ClientID)
	body.Set("client_secret", input.ClientSecret)

	request, err := http.NewRequest("POST", input.TokenURI, strings.NewReader(body.Encode()))
	if err != nil {
		return AuthorizationCodeOutput{}, fmt.Errorf("can't build request to the '%s' endpoint: %v", input.TokenURI, err)
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.URL.User = url.UserPassword(input.ClientID, input.ClientSecret)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return AuthorizationCodeOutput{}, fmt.Errorf("can't send request to the '%s' endpoint with code '%s': %v", input.TokenURI, result.Code, err)
	}

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return AuthorizationCodeOutput{}, fmt.Errorf("can't read response body from the '%s' endpoint (code was '%s'): %v", input.TokenURI, result.Code, err)
	}

	if response.StatusCode != 200 {
		return AuthorizationCodeOutput{}, fmt.Errorf("unexpected http status from the '%s' endpoint with code '%s'. expected:%d; got:%d; body: %s", input.TokenURI, result.Code, 200, response.StatusCode, string(respBody))
	}

	var tokenResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &tokenResponse); err != nil {
		return AuthorizationCodeOutput{}, fmt.Errorf("cant' decode successful response from the '%s' endpoint with code '%s'. body: %s; err: %v", input.TokenURI, result.Code, string(respBody), err)
	}

	accessToken, ok := tokenResponse["access_token"].(string)
	if !ok {
		return AuthorizationCodeOutput{}, fmt.Errorf("missing 'access_token'' from json response. body: %s; err: %v", string(respBody), err)
	}

	var idToken *string
	maybeIDToken, ok := tokenResponse["id_token"].(string)
	if ok {
		idToken = &maybeIDToken
	}

	return AuthorizationCodeOutput{
		AccessToken: accessToken,
		IDToken:     idToken,
		RawResponse: tokenResponse,
	}, nil
}
