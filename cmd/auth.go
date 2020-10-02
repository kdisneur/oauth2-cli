package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kdisneur/oauth2-cli/internal/caching"
	"github.com/kdisneur/oauth2-cli/internal/console"
	"github.com/kdisneur/oauth2-cli/internal/flows"
	"github.com/kdisneur/oauth2-cli/internal/security"
)

var description = `This command handles the whole classical authorization code flow:

- it reads the expected values (client_id, scope, ...) from tty (and keep a cache of the non sensitve values so we don't have to type them everytime)
- it opens the browser with the right authorize parameters
- it spawns an HTTP server to receive the calbback code
- it exchanges the authorization code to a token
- it outputs to STDOUT deither:
	- the full token response as JSON
	- the access token value
	- the id token value

⚠️  To make it works, don't forget to add the callback URL into the allowed callback URLs of your application.
`

// AuthorizeCache are the keys stored in the caching folder to speed up the authorization process
type AuthorizeCache struct {
	AuthorizeURI string `json:"authorize_uri"`
	TokenURI     string `json:"token_uri"`
	ClientID     string `json:"client_id"`
	Scope        string `json:"scope"`
}

// AuthorizeFlags stores all the possible authorization flags
type AuthorizeFlags struct {
	HTTPPort            int
	ClientSecretOnStdin bool
	OnlyAccessToken     bool
	OnlyIDToken         bool
}

var authorizeFlags AuthorizeFlags

var authorizeCmd = &cobra.Command{
	Use:   "authorize",
	Short: "handles a full authorization code flow",
	Long:  description,
	RunE: func(cmd *cobra.Command, args []string) error {
		var cache AuthorizeCache
		caching.Read(rootFlags.ConfigFolder, "authorize.json", &cache)

		if authorizeFlags.OnlyAccessToken && authorizeFlags.OnlyIDToken {
			return fmt.Errorf("you can't ask for only accesstoken and only idtoken simultaneously")
		}

		var clientSecret string
		if authorizeFlags.ClientSecretOnStdin {
			data, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("can't read clientSecret from STDIN: %v", err)
			}

			clientSecret = strings.TrimSpace(string(data))
		}

		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0644)
		if err != nil {
			return fmt.Errorf("can't open /dev/tty: %v", err)
		}
		defer tty.Close()

		reader := console.NewReader(tty)
		readAndSaveValue := func(reader *console.Reader, prompt string, cache *AuthorizeCache, receiver *string) error {
			value, err := reader.ReadStringlnWithOptions(console.ReadStringOptions{
				Writer:       os.Stderr,
				Prompt:       prompt,
				DefaultValue: *receiver,
				Required:     true,
			})
			if err != nil {
				return fmt.Errorf("can't read %s: %v", prompt, err)
			}

			*receiver = value

			caching.Save(rootFlags.ConfigFolder, "authorize.json", &cache)

			return nil
		}

		if err := readAndSaveValue(reader, "Authorize URI", &cache, &cache.AuthorizeURI); err != nil {
			return err
		}

		if err := readAndSaveValue(reader, "Token URI", &cache, &cache.TokenURI); err != nil {
			return err
		}

		redirectURI, err := reader.ReadStringlnWithOptions(console.ReadStringOptions{
			Writer:       os.Stderr,
			Prompt:       "Redirect URI",
			DefaultValue: fmt.Sprintf("http://localhost:%d/oauth/callback", authorizeFlags.HTTPPort),
			Required:     true,
		})
		if err != nil {
			return fmt.Errorf("can't read Redirect URI: %v", err)
		}

		if err := readAndSaveValue(reader, "Client ID", &cache, &cache.ClientID); err != nil {
			return err
		}

		if !authorizeFlags.ClientSecretOnStdin {
			secret, err := reader.ReadStringlnWithOptions(console.ReadStringOptions{
				Writer: os.Stderr,
				Prompt: "Client Secret",
			})
			if err != nil {
				return fmt.Errorf("can't read Client Secret: %v", err)
			}

			clientSecret = secret
		}

		if err := readAndSaveValue(reader, "Scope", &cache, &cache.Scope); err != nil {
			return err
		}

		state, err := reader.ReadStringlnWithOptions(console.ReadStringOptions{
			Writer: os.Stderr,
			Prompt: "State",
		})
		if err != nil {
			return fmt.Errorf("can't read State: %v", err)
		}

		if state == "" {
			random, err := security.GenerateRandomString(10)
			if err != nil {
				return fmt.Errorf("can't generate state: %v", err)
			}

			state = random
		}

		input := flows.AuthorizationCodeInput{
			AuthorizeURI: cache.AuthorizeURI,
			TokenURI:     cache.TokenURI,
			RedirectURI:  redirectURI,
			ClientID:     cache.ClientID,
			ClientSecret: clientSecret,
			Scope:        cache.Scope,
			State:        state,
		}

		opts := flows.AuthorizationCodeOptions{
			HTTPServerPort: authorizeFlags.HTTPPort,
			OpenBrowser:    func(url string) { exec.Command("open", url).Start() },
		}

		output, err := flows.RunAuthorizationCode(input, opts)
		if err != nil {
			return err
		}

		if authorizeFlags.OnlyAccessToken {
			fmt.Println(output.AccessToken)
			return nil
		}

		if authorizeFlags.OnlyIDToken {
			if output.IDToken == nil {
				return fmt.Errorf("can't get the 'id_token' field from response: %s", output.RawResponse)
			}

			fmt.Printf("%s\n", *output.IDToken)
			return nil
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output.RawResponse); err != nil {
			return fmt.Errorf("can't format json response: %s", output.RawResponse)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(authorizeCmd)

	authorizeCmd.Flags().IntVar(&authorizeFlags.HTTPPort, "http-port", 9876, "address used to bind the http server")
	authorizeCmd.Flags().BoolVar(&authorizeFlags.ClientSecretOnStdin, "client-secret-stdin", false, "should read clientSecret from STDIN or ask the question")
	authorizeCmd.Flags().BoolVar(&authorizeFlags.OnlyAccessToken, "only-accesstoken", false, "returns only the accesstoken. can't be used in combination with `--only-idtoken`")
	authorizeCmd.Flags().BoolVar(&authorizeFlags.OnlyIDToken, "only-idtoken", false, "returns only the idtoken. can't be used in combination with `--only-accesstoken`")
}
