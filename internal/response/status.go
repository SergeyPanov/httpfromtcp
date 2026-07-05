package response

type StatusCode int

const (
	Success     StatusCode = 200
	BadRequest  StatusCode = 400
	ServerError StatusCode = 500
)
