package models

type ApiError struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type PostTopicRequest struct {
	Username string `json:"username"`
	Topic    string `json:"topic"`
}

type PostTopicResponse struct {
	TopicID uint32 `json:"topic_id"`
}

type DeleteTopicRequest struct {
	Username string `json:"username"`
	Topic    string `json:"topic"`
}

type DeleteTopicResponse struct {
	TopicID uint32 `json:"topic_id"`
}

type ListTopicsRequest struct {
	Username string `json:"username"`
}

type ListTopicsResponse struct {
	Topics []string `json:"topics"`
}

type PostLinkRequest struct {
	Username string `json:"username"`
	Topic    string `json:"topic"`
	Link     string `json:"link"`
	Alias    string `json:"alias"`
}

type PostLinkResponse struct {
	Alias string `json:"alias"`
}

type PickLinkRequest struct {
	Username string `form:"username"`
	Topic    string `form:"topic"`
	Alias    string `form:"alias"`
}

type PickLinkResponse struct {
	Link string `json:"link"`
}

type DeleteLinkRequest struct {
	Username string `json:"username"`
	Topic    string `json:"topic"`
	Alias    string `json:"alias"`
}

type DeleteLinkResponse struct {
	Alias string `json:"alias"`
}

type ListLinksRequest struct {
	Username string `form:"username"`
	Topic    string `form:"topic"`
}

type ListLinksResponse struct {
	Links   []string `json:"links"`
	Aliases []string `json:"aliases"`
}
