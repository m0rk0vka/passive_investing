package filepathgetter

type getFileResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		FileID   string `json:"file_id"`
		FilePath string `json:"file_path"`
	} `json:"result"`
}
