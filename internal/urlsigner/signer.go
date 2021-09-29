package urlsigner

import (
	"fmt"
	"log"
	"strings"
	"time"

	goalone "github.com/bwmarrin/go-alone"
)

type Signer struct {
	Secret []byte
}

// GenerateTokenFromString generates a new token string and returns it
func (s *Signer) GenerateTokenFromString(data string) string {
	var urlToSign string

	crypt := goalone.New(s.Secret, goalone.Timestamp)
	if strings.Contains(data, "?") {
		// append hash token to url
		urlToSign = fmt.Sprintf("%s&hash=", data)
	} else {
		// append hash token to url after the url queries
		urlToSign = fmt.Sprintf("%s?hash=", data)
	}

	// sign email
	tokenBytes := crypt.Sign([]byte(urlToSign))
	token := string(tokenBytes)

	// return fully signed url
	return token
}

// VerifyToken verifies whether the link has been changed
func (s *Signer) VerifyToken(token string) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)

	// unsign token to verify it
	_, err := crypt.Unsign([]byte(token))
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func (s *Signer) Expired(token string, minutesUntilExpire int) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	ts := crypt.Parse([]byte(token))

	return time.Since(ts.Timestamp) > time.Duration(minutesUntilExpire)*time.Minute
}
