package servicestemp

import (
	
	"github.com/ButterHost69/PKr-Base/config"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

func SendUpdateMeRequest(server_ip, username, password string) (bool, error) {
	workspace_name := "WorkSpace1"

	url := "http://"+server_ip+"/register/workspace"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	err := writer.WriteField("username", username)
	if err != nil {
		return false, fmt.Errorf("error writing field: %v", err)
	}

	err = writer.WriteField("password", password)
	if err != nil {
		return false, fmt.Errorf("error writing field: %v", err)
	}

	err = writer.WriteField("workspace_name", workspace_name)
	if err != nil {
		return false, fmt.Errorf("error writing field: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return false, fmt.Errorf("error in closing writer: %v", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return false, fmt.Errorf("error failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error failed to make send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("error failed to read from the response: %v", err)
	}

	// t.Logf("Response status: %v", resp.Status)
	// t.Logf("Response body: %s", body)

	var repsonse models.GenericResp
	err = json.Unmarshal(body, &repsonse)
	if err != nil {
		return false, fmt.Errorf("error failed to umarshall repsonse: %v", err)
	}
	if resp.Status != "200 OK" && repsonse.Response == "success" {
		return false, fmt.Errorf("error Expected Status: 200 OK  ||  Body: 'response':'success,\nreceived: Status: %s, Body: %s", resp.Status, string(body))
	}

	return true, nil
}