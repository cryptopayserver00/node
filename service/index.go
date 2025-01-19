package service

import (
	NODE_Client "node/utils/http"
)

type NService struct{}

var (
	NodeService = new(NService)
	client      NODE_Client.Client
)
