package validator


func (v *URLValidator) isPrivateOrLoopback(host string) bool {

	ips, err := v.resolver.LookupIP(host)

	if err != nil {
		return true
	}

	for _, ip := range ips {

		if ip.IsLoopback() {
			return true
		}

		if ip.IsPrivate() {
			return true
		}

		if ip.IsMulticast() {
			return true
		}

		if ip.IsUnspecified() {
			return true
		}

	}

	return false
}