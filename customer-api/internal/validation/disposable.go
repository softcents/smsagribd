package validation

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

var (
	disposableDomainsOnce sync.Once
	disposableMap         = make(map[string]struct{})
)

// fallbackDomains provides an embedded baseline of common temporary/disposable email providers
// in case storage/framework/disposable_domains.json is not present on disk.
var fallbackDomains = []string{
	"mailinator.com", "guerrillamail.com", "guerrillamailblock.com", "sharklasers.com",
	"tempmail.com", "temp-mail.org", "10minutemail.com", "10minutemail.net",
	"yopmail.com", "trashmail.com", "trashmail.net", "dispostable.com",
	"getairmail.com", "maildrop.cc", "fakeinbox.com", "mohmal.com",
	"crazymailing.com", "throwawaymail.com", "mytemp.email", "nada.ltd",
	"getnada.com", "inboxkitten.com", "burnermail.io", "dropmail.me",
	"yopmail.fr", "yopmail.net", "cool.fr.nf", "jetable.fr.nf", "courriel.fr.nf",
	"moncourrier.fr.nf", "monemail.fr.nf", "monmail.fr.nf", "hideaddress.net",
	"emailondeck.com", "tempail.com", "tempail.net", "disposable.email",
	"0-mail.com", "0box.eu", "0box.net", "0clickemail.com", "0clock.net",
}

// InitDisposableDomains loads domains from storage/framework/disposable_domains.json if available,
// falling back to the bundled list. No external network requests are made.
func InitDisposableDomains(storageJsonPath string) {
	disposableDomainsOnce.Do(func() {
		// Populate fallback baseline first
		for _, d := range fallbackDomains {
			disposableMap[strings.ToLower(strings.TrimSpace(d))] = struct{}{}
		}

		if storageJsonPath == "" {
			storageJsonPath = "./storage/framework/disposable_domains.json"
		}

		data, err := os.ReadFile(storageJsonPath)
		if err == nil {
			var domains []string
			if err := json.Unmarshal(data, &domains); err == nil {
				for _, d := range domains {
					cleaned := strings.ToLower(strings.TrimSpace(d))
					if cleaned != "" {
						disposableMap[cleaned] = struct{}{}
					}
				}
			}
		}
	})
}

// IsDisposable checks if the provided email domain is listed as disposable/temporary.
func IsDisposable(email string) bool {
	// Ensure initialized
	InitDisposableDomains("")

	atIdx := strings.LastIndex(email, "@")
	if atIdx == -1 || atIdx == len(email)-1 {
		return false
	}

	domain := strings.ToLower(strings.TrimSpace(email[atIdx+1:]))
	if domain == "" {
		return false
	}

	// Exact domain match
	if _, ok := disposableMap[domain]; ok {
		return true
	}

	// Subdomain match (e.g. sub.mailinator.com)
	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts)-1; i++ {
		parentDomain := strings.Join(parts[i:], ".")
		if _, ok := disposableMap[parentDomain]; ok {
			return true
		}
	}

	return false
}
