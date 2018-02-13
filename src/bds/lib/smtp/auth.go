package smtp

type Auth interface {
	PermitSend(from, username string) bool
	Plain(username, password string) bool
}
