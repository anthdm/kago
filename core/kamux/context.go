package kamux

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/kamalshkeir/kago/core/utils"
	"github.com/kamalshkeir/kago/core/utils/logger"
)

// Context is a wrapper of responseWriter, request, and params map
type Context struct {
	http.ResponseWriter
	*http.Request
	Params map[string]string
}
// Json return json indented to the client
func (c *Context) Json(code int, body interface{}) {
	c.ResponseWriter.Header().Set("Content-Type","application/json")
	c.WriteHeader(code)
	enc := json.NewEncoder(c.ResponseWriter)
	enc.SetIndent("","\t")
	err := enc.Encode(body)
	if logger.CheckError(err) {return}
}

// Text return text with custom code to the client
func (c *Context) Text(code int, body string) {
	c.ResponseWriter.Header().Set("Content-Type", "text/plain")
	c.WriteHeader(code)
	io.WriteString(c.ResponseWriter, fmt.Sprintf("%s\n", body))
}

// Html return template_name with data to the client
func (c *Context) Html(template_name string, data map[string]interface{}) {
	const key utils.ContextKey = "user"
	if data == nil { data = make(map[string]interface{}) }
	
	user,ok := c.Request.Context().Value(key).(map[string]interface{})
	if ok {		
		data["request"] = c.Request
		data["is_authenticated"] = true
		data["user"] = user
	} else {
		data["request"] = c.Request
		data["is_authenticated"] = false
		data["user"] = nil
	}

	c.ResponseWriter.Header().Set("Content-Type","text/html; charset=utf-8")
	err := allTemplates.ExecuteTemplate(c.ResponseWriter,template_name,data)
	logger.CheckError(err)
}

// GetJson get json body from request and return map
func (c *Context) GetJson() map[string]interface{} {
	// USAGE : data := template.GetJson(r)
	body, err := io.ReadAll(c.Request.Body)
	if logger.CheckError(err) {
		return nil
	}
	defer c.Request.Body.Close()

	request := map[string]interface{}{}
	err = json.Unmarshal(body,&request)
	if logger.CheckError(err) {
		return nil
	}
	return request
}

// Redirect redirect the client to the specified path with a custom code
func (c *Context) Redirect(path string,code int) {
	http.Redirect(c.ResponseWriter,c.Request,path,code)
}

// File serve a file from handler
func (c *Context) File(content_type,path_to_file string) {
	c.ResponseWriter.Header().Set("Content-Type", content_type)
	http.ServeFile(c.ResponseWriter, c.Request, path_to_file)
}

// EmbedFile serve an embeded file from handler
func (c *Context) EmbedFile(content_type string,embed_file []byte) {
	c.ResponseWriter.Header().Set("Content-Type", content_type)
		_,_ = c.ResponseWriter.Write(embed_file)
}

// UploadFileFromFormData upload received_filename into folder_out and return url,fileByte,error
func (c *Context) UploadFileFromFormData(received_filename,folder_out string) (string,[]byte,error) {
	c.Request.ParseMultipartForm(10<<20) //10Mb
	var buff bytes.Buffer
	file, header , err := c.Request.FormFile(received_filename)
	if logger.CheckError(err) {
		return "",nil,err
	}
	defer file.Close()
	// copy the uploaded file to the buffer
	if _, err := io.Copy(&buff, file); err != nil {
		return "",nil,err
	}

	data_string := buff.String()

	// make DIRS if not exist
	err = os.MkdirAll("media/"+folder_out+"/",0664)
	if err != nil {
		return "",nil,err
	}
	// make file
	dst, err := os.Create("media/"+folder_out+"/" + header.Filename)
	if err != nil {
		return "",nil,err
	}
	defer dst.Close()
	dst.Write([]byte(data_string))
	
	url := "media/"+folder_out+"/"+header.Filename
	return url,[]byte(data_string),nil
}

// DELETE FILE
func (c *Context) DeleteFile(path string) error {
	err := os.Remove("."+path)
	if err != nil {
		return err
	} else {
		return nil
	}
}

// Download download data_bytes(content) asFilename(*.json,...) to the client
func (c *Context) Download(data_bytes []byte, asFilename string) {
	bytesReader := bytes.NewReader(data_bytes)
	c.ResponseWriter.Header().Set("Content-Disposition", "attachment; filename=" + strconv.Quote(asFilename))
	c.ResponseWriter.Header().Set("Content-Type", c.Request.Header.Get("Content-Type"))
	io.Copy(c.ResponseWriter,bytesReader)
}