package jike

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"
)

type UsersService struct {
	jike   *Jike
	client *http.Client
	debug  bool
}

func NewUsersService(jike *Jike, c *http.Client, debug bool) *UsersService {
	return &UsersService{jike, c, debug}
}

func (u *UsersService) PasswordLogin(ctx context.Context, areaCode, phone, password string) (*LoginOutput, error) {
	input := map[string]interface{}{
		"areaCode":          areaCode,
		"mobilePhoneNumber": phone,
		"password":          password,
	}
	output := &LoginOutput{}
	req := &request{
		Debug:      u.debug,
		HTTPMethod: http.MethodPost,
		HTTPPath:   `/1.0/users/loginWithPhoneAndPassword`,
		Input:      input,
		Output:     &output,
		Context:    ctx,
	}
	return output, req.send(u.jike)
}

func (u *UsersService) Profile(ctx context.Context, username string) (*ProfileOutput, error) {
	params := map[string]string{
		"username": username,
	}
	output := &ProfileOutput{}
	req := &request{
		Debug:      u.debug,
		HTTPMethod: http.MethodGet,
		HTTPPath:   `/1.0/users/profile`,
		Params:     params,
		Output:     &output,
		Context:    ctx,
	}
	return output, req.send(u.jike)
}

func (u *UsersService) FollowingTimeline(ctx context.Context, limit int, loadMoreKey TimelineLoadMoreKey) (*TimelineOutput, error) {
	input := map[string]interface{}{
		"limit":       limit,
		"loadMoreKey": loadMoreKey,
	}
	output := &TimelineOutput{}
	req := &request{
		Debug:      u.debug,
		HTTPMethod: http.MethodPost,
		HTTPPath:   `/1.0/personalUpdate/followingUpdates`,
		Input:      input,
		Output:     &output,
		Context:    ctx,
	}
	return output, req.send(u.jike)
}

type Resp struct {
	Success bool `json:"success"`
	Data    struct {
		Month string `json:"month"`
		Day   string `json:"day"`
		Zh    string `json:"zh"`
		En    string `json:"en"`
		Pic   string `json:"pic"`
	} `json:"data"`
}

func (u *UsersService) Create(ctx context.Context, pictureKeys []string) (*CreateOutput, error) {
	input := map[string]interface{}{
		"type":          "originalPosts",
		"pictureKeys":   pictureKeys,
		"submitToTopic": "59e58bea89ee3f0016b4d2c6",
	}
	res, err := http.Get("https://api.vvhan.com/api/en?type=sj")
	defer res.Body.Close()
	if err != nil {
		log.Println(err)
	} else {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		var resp Resp
		err = json.Unmarshal(body, &resp)
		if err != nil {
			return nil, err
		}
		input["content"] = resp.Data.En + "\n" + resp.Data.Zh
	}
	output := &CreateOutput{}
	req := &request{
		Debug:      false,
		HTTPMethod: http.MethodPost,
		HTTPPath:   `/1.0/originalPosts/create`,
		Input:      input,
		Output:     &output,
		Context:    ctx,
	}
	return output, req.send(u.jike)
}

//func (u *UsersService) UploadPic(ctx context.Context) (*UploadPicOutput, error) {
//	input := map[string]interface{}{
//		"token":"",
//		"fname":"",
//		"file":
//	}
//	req := &request{
//
//	}
//	req.req.FormFile()
//}

type Test struct {
	PictureKeys []string `json:"pictureKeys"`
	TopicId     string   `json:"topicId"`
}

type UploadPicOutput struct {
	FileUrl string `json:"fileUrl"`
	Id      string `json:"id"`
	Key     string `json:"key"`
	Success bool   `json:"success"`
}

// 注意client 本身是连接池，不要每次请求时创建client
var (
	HttpClient = &http.Client{
		//Timeout: 3 * time.Second,
	}
)

type UploadFileInfo struct {
	Token       string
	FName       string
	OriginalUrl string
}

type UploadFileResp struct {
	FileUrl string `json:"fileUrl"`
	Id      string `json:"id"`
	Key     string `json:"key"`
	Success bool   `json:"success"`
}

func UploadFile(info UploadFileInfo) (*UploadFileResp, error) {

	var uploadFileResp UploadFileResp

	params := map[string]string{
		"token": info.Token,
		"fname": info.FName,
	}

	remote, err := getRemote(info.OriginalUrl)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	file := bytes.NewReader(remote)

	body := new(bytes.Buffer)

	writer := multipart.NewWriter(body)

	formFile, err := writer.CreateFormFile("file", info.FName)
	if err != nil {
		return nil, err
	}

	log.Printf("start write %s...\n", info.OriginalUrl)
	_, err = io.Copy(formFile, file)
	if err != nil {
		log.Printf("write %s fail\n", info.OriginalUrl)
		return nil, err
	} else {
		log.Printf("write %s success\n", info.OriginalUrl)
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://upload.qiniup.com", body)
	if err != nil {
		return nil, err
	}
	//req.Header.Set("Content-Type","multipart/form-data")
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(content, &uploadFileResp)
	if err != nil {
		return nil, err
	}
	return &uploadFileResp, nil
}

func getRemote(url string) ([]byte, error) {

	httpClient := http.Client{
		Timeout: 6 * time.Minute,
	}

	log.Printf("start get %s...\n", url)
	res, err := httpClient.Get(url)
	if err != nil {
		log.Printf("get %s fail\n", url)
		// 如果有错误返回错误内容
		return nil, err
	}
	log.Printf("get %s success\n", url)
	// 使用完成后要关闭，不然会占用内存
	defer res.Body.Close()
	log.Printf("start read %s...\n", url)
	time.Now()
	// 读取字节流
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("read %s fail\n", url)
		return nil, err
	} else {
		log.Printf("read %s success\n", url)
	}
	return bytes, nil
}
