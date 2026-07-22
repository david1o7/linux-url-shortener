package validator

func isAllowedScheme(scheme string) bool {

	switch scheme {

	case "http", "https":
		return true

	default:
		return false
	}

}