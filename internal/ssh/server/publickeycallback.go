package server

import (
	"fmt"
	"io/ioutil"
	"os"
	osUser "os/user"

	"github.com/mimecast/dtail/internal/config"
	"github.com/mimecast/dtail/internal/logger"
	user "github.com/mimecast/dtail/internal/user/server"

	gossh "golang.org/x/crypto/ssh"
)

// PublicKeyCallback is for the server to check whether a public SSH key is authorized ot not.
func PublicKeyCallback(c gossh.ConnMetadata, pubKey gossh.PublicKey) (*gossh.Permissions, error) {
	user := user.New(c.User(), c.RemoteAddr().String())
	logger.Info(user, "Incoming authorization")

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Unable to get current working directory|%s|", err.Error())
	}

	authorizedKeysFile := fmt.Sprintf("%s/%s/%s.authorized_keys", cwd, config.Common.CacheDir, user.Name)
	if _, err := os.Stat(authorizedKeysFile); os.IsNotExist(err) {
		user, err := osUser.Lookup(user.Name)
		if err != nil {
			return nil, fmt.Errorf("Unable to authorize|%s|%s|", user, err.Error())
		}
		// Fallback to ~
		authorizedKeysFile = user.HomeDir + "/.ssh/authorized_keys"
	}

	logger.Info(user, "Reading", authorizedKeysFile)
	authorizedKeysBytes, err := ioutil.ReadFile(authorizedKeysFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to read authorized keys file|%s|%s|%s", authorizedKeysFile, user, err.Error())
	}

	authorizedKeysMap := map[string]bool{}
	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := gossh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse authorized keys bytes|%s|%s", user, err.Error())
		}
		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest
	}

	if authorizedKeysMap[string(pubKey.Marshal())] {
		logger.Debug("Public key fingerprint", gossh.FingerprintSHA256(pubKey), user)
		return &gossh.Permissions{
			Extensions: map[string]string{
				"pubkey-fp": gossh.FingerprintSHA256(pubKey),
			},
		}, nil
	}

	return nil, fmt.Errorf("Unknown public key|%s", user)
}
