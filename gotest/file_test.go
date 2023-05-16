package gotest

import (
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"testing"
)

func Test_DownloadFile(t *testing.T) {
	url := "https://oaidalleapiprodscus.blob.core.windows.net/private/org-rc2vQqQ0k9YiBzjOoDQ2qMUV/user-A805NZ20yjQhscbVrmudnsmU/img-mbrIsYHY1RBpokBR6CCbFO94.png?st=2023-05-16T06%3A08%3A16Z&se=2023-05-16T08%3A08%3A16Z&sp=r&sv=2021-08-06&sr=b&rscd=inline&rsct=image/png&skoid=6aaadede-4fb3-4698-a8f6-684d7786b067&sktid=a48cca56-e6da-484e-a814-9c849652bcb3&skt=2023-05-15T22%3A11%3A15Z&ske=2023-05-16T22%3A11%3A15Z&sks=b&skv=2021-08-06&sig=td0D5K6OcaHvcn7zb7VHY4eaw/6yoKr8tSrpBX5GZjI%3D"

	_, _, err := file_util.DownloadImgFromUrl(url)
	if err != nil {
		panic(err)
	}

}
