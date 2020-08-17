package ack

const (
	CONTINUE           = 100
	SwitchingProtocols = 101
	PROCESSING         = 102

	OK             = 200
	CREATED        = 201
	ACCEPTED       = 202
	NonAuth        = 203
	NoContent      = 204
	ResetContent   = 205
	PartialContent = 206

	BadRequest      = 400
	UNAUTHORIZED    = 401
	PaymentRequired = 402
	FORBIDDEN       = 403
	NotFound        = 404

	InternalServerError = 500
	NotImplemented      = 501
	BadGateway          = 502
	ServiceUnavailable  = 503
	GatewayTimeout      = 504
)
