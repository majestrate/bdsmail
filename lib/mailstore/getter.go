package mailstore

type MailRouter interface {
	FindStoreFor(username string) (Store, bool)
}
