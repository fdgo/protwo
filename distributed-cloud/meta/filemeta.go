package meta

import (
	"github.com/fdgo/distributed-cloud/db"
	"sort"
)

//文件元信息
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string //时间戳
}

var fileMetas map[string]FileMeta

func init()  {
	fileMetas = make(map[string]FileMeta)
}

//新增/更新文件元信息
func UpdateFileMeta(fmeta FileMeta)  {
	fileMetas[fmeta.FileSha1] = fmeta
}
func UpdateFileMetaDB(fmeta FileMeta) bool {
	return db.OnFileUploadFinished(fmeta.FileSha1,fmeta.FileName,fmeta.FileSize,fmeta.Location)
}



//通过sha1获取文件的元信息对象
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}
func GetFileMetaDB(fileSha1 string) (FileMeta,error)  {
	tfile,err := db.GetFileMeta(fileSha1)
	if err!=nil{
		return FileMeta{},err
	}
	fmeta := FileMeta{
		FileSha1:tfile.FileHash,
		FileName:tfile.FileName.String,
		FileSize:tfile.FileSize.Int64,
		Location:tfile.FileAddr.String,
	}
	return fmeta,nil
}




func GetLastFileMetas(count int)[]FileMeta  {
	fMetaArray := make([]FileMeta,len(fileMetas))
	for _, v:= range fileMetas{
		fMetaArray = append(fMetaArray,v)
	}
	sort.Sort(ByUploadTime(fMetaArray))
	return fMetaArray[0:count]
}
func RemoveFileMeta(fileSha1 string)  {
	delete(fileMetas,fileSha1)
}