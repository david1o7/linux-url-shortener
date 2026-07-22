package tests

import (
	"net/http"
)

type MockClient struct {
	Response *http.Response
	Err error
}

func (m *MockClient) Do(req *http.Request)(*http.Response,error){

	if m.Err!=nil{
		return nil,m.Err
	}

	return m.Response,nil
}