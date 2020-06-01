package service

import (
	"errors"
	"github.com/wangmhgo/go-project/gdun/common"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

func (ts *TranscodingService) TrascodePDF(t *TranscodingTask) (int, int, error) {
	// Create a temporary directory for storing the output PNG files.
	destDir := ts.tmpDir + "/" + t.StorageKey

	err := os.MkdirAll(destDir+"/thumbnail", 0644)
	if err != nil {
		return -1, 0, err
	}
	defer os.RemoveAll(destDir)

	err = ts.ConvertPDF2PNG(destDir, t.FileName)
	if err != nil {
		return -2, 0, err
	}

	size, err := ts.uploadPNG(t.StorageKey, destDir, t.FileName, t.UserIP, t.UserID)
	if err != nil {
		return -3, 0, err
	}

	err = os.Remove(t.FileName)
	if err != nil {
		return -4, 0, err
	}

	return 0, size, nil
}

//----------------------------------------------------------------------------
// Step 1.

func (ts *TranscodingService) ConvertPDF2PNG(destDir string, srcFile string) error {
	// Generate normal PNG files.
	cmd := exec.Command(ts.ghostScript,
		"-q",
		"-sDEVICE=pngalpha",
		"-dBATCH",
		"-dNOPAUSE",
		"-dNOPROMPT",
		"-dDOINTERPOLATE",
		"-r96",
		"-sOutputFile="+destDir+"/%d.png",
		srcFile)

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	// Generate small PNG files.

	cmd = exec.Command(ts.ghostScript,
		"-q",
		"-sDEVICE=pngalpha",
		"-dBATCH",
		"-dNOPAUSE",
		"-dNOPROMPT",
		"-dDOINTERPOLATE",
		"-r16",
		"-sOutputFile="+destDir+"/thumbnail/%d.png",
		srcFile)

	err = cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}

//----------------------------------------------------------------------------
// Step 2.

func (ts *TranscodingService) uploadPNG(key string, dirName string, fileName string, ip string, userID int) (int, error) {
	// Get the output PNG files.
	fis, err := ioutil.ReadDir(dirName)
	if err != nil {
		return 0, err
	}
	size := 0
	for i := 0; i < len(fis); i++ {
		if fis[i].IsDir() {
			continue
		}
		if strings.HasSuffix(fis[i].Name(), ".png") {
			size++
		}
	}
	if size == 0 {
		return 0, errors.New("No PNG file has been generated.")
	}

	// Upload the PDF file.
	// err = ts.cwOss.UploadFile(ts.cwKeyPrefix+key, fileName)
	// if err != nil {
	// 	return 0, err
	// }

	buf, err := (func(fileName string) ([]byte, error) {
		f, err := os.Open(fileName)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		buf, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		return buf, nil

	})(fileName)

	if err != nil {
		return 0, err
	}

	// Upload the PDF file.
	ts.cwOss.UploadBuffer(ts.cwKeyPrefix+key+".pdf", buf)

	// Upload these PNG files.
	for i := 1; i <= size; i++ {
		s := "/" + strconv.Itoa(i) + ".png"
		err = ts.cwOss.UploadFile(ts.cwKeyPrefix+key+s, dirName+s)
		if err != nil {
			return 0, err
		}

		s = "/thumbnail/" + strconv.Itoa(i) + ".png"
		err = ts.cwOss.UploadFile(ts.cwKeyPrefix+key+s, dirName+s)
		if err != nil {
			return 0, err
		}
	}

	// Upload the meta-data file.
	//meta := `metaInfo["` + key + `"]={` +
	meta := `var metaInfo={` +
		`"` + common.FIELD_TYPE + `":"PDF",` +
		`"` + common.FIELD_PAGE + `":` + strconv.Itoa(size) + `,` +
		`"` + common.FIELD_IP + `":"` + ip + `",` +
		`"` + common.FIELD_TIMESTAMP + `":` + common.GetTimeString() + `000,` +
		`"` + common.FIELD_USER + `":` + strconv.Itoa(userID) +
		`};`
	err = ts.cwOss.UploadString(ts.cwKeyPrefix+key+"/meta.js", meta)
	if err != nil {
		return 0, err
	}

	return size, nil
}

//----------------------------------------------------------------------------
