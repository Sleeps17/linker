package linker

var (
	MsgInternalError = "something went wrong"

	MsgRecordNotFound = "link with this alias was not found"
	MsgAliasNotFound  = "link with this alias was not found"
	MsgUserNotFound   = "unknown username"

	MsgAliasAlreadyExists = "link with such an alias already exists"

	MsgInvalidUsername = "you cannot use a username less than 8 characters long"
	MsgInvalidLink     = "you are trying to post a non-link"
	MsgEmptyAlias      = "it is impossible to find a link using an empty alias"
)
