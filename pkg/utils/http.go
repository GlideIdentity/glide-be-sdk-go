package utils

import (
	"encoding/json"
	"io"
	"net/http"
)

// use json.Unmarshal to parse the response body into the responseStruct
func GetJsonBody(res *http.Response, responseStruct interface{}) error {
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, responseStruct); err != nil {
		return err
	}

	return nil
}
