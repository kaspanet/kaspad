package network

import (
	"net"
)

// NormalizeAddresses returns a new slice with all the passed peer addresses
// normalized with the given default port, and all duplicates removed.
func NormalizeAddresses(addrs []string, defaultPort string) ([]string, error) {
	for i, addr := range addrs {
		var err error
		addrs[i], err = NormalizeAddress(addr, defaultPort)
		if err != nil {
			return nil, err
		}
	}

	return removeDuplicateAddresses(addrs), nil
}

// NormalizeAddress returns addr with the passed default port appended if
// there is not already a port specified.
func NormalizeAddress(addr, defaultPort string) (string, error) {
	_, _, err := net.SplitHostPort(addr)
	// net.SplitHostPort returns an error if the given host is missing a
	// port, but theoretically it can return an error for other reasons,
	// and this is why we check addrWithPort for validity.
	if err != nil {
		addrWithPort := net.JoinHostPort(addr, defaultPort)
		_, _, err := net.SplitHostPort(addrWithPort)
		if err != nil {
			return "", err
		}

		return addrWithPort, nil
	}
	return addr, nil
}

// removeDuplicateAddresses returns a new slice with all duplicate entries in
// addrs removed.
func removeDuplicateAddresses(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	seen := map[string]struct{}{}
	for _, val := range addrs {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}
