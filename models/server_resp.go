package models

type RegisterResp struct {
	Response	string		`json:"response"`
	Username	string		`json:"username"`
}

type GenericResp struct {
	Response	string	`json:"response"`	
}
