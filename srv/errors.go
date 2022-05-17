package srv

type errBadRequest string

// ErrBadRequest represents bad HTTP requests
const ErrBadRequest = errBadRequest("bad request")

func (ebr errBadRequest) Error() string { return string(ebr) }
