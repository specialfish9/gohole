package dns

import (
	"fmt"
	"net/netip"
	"strings"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
)

func normalizeName(name string) string {
	name = dnsutil.Fqdn(name)
	name = strings.ToLower(name)
	return name
}

func parseCustomDomains(customDomains map[string]any) (map[string]netip.Addr, error) {
	res := make(map[string]netip.Addr)
	for name, addr := range customDomains {
		addrStr, ok := addr.(string)
		if !ok {
			return nil, fmt.Errorf(
				"invalid IP address for entry '%s': expected a string, got '%v' of type %T",
				name,
				addr,
				addr,
			)
		}

		ip, err := netip.ParseAddr(addrStr)
		if err != nil {
			return nil, fmt.Errorf("invalid IP address '%s' in entry '%s': %s", addrStr, name, err)
		}

		name = normalizeName(name)

		res[name] = ip
	}

	return res, nil
}

// checkCustomDomains checks whether the given question contains a name in the custom domains map.
// It will then create a response RR.
func (h *Handler) checkCustomDomains(rc *ReqCtx, question dns.RR) (dns.RR, error) {
	addr, isCustom := h.customDomains[rc.Name]
	if !isCustom {
		rc.Logger.Debug("Name isn't a custom domain", "name", rc.Name)
		return nil, nil
	}

	// Update the request context
	rc.Custom = true
	rc.Logger.Debug("Name is a custom domain", "name", rc.Name)

	return answerFromQuestion(question, addr)
}
