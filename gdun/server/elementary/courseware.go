package elementary

import (
	"archive/zip"
	"github.com/wangmhgo/go-project/gdun/common"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddCourseware(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	preparation, err := strconv.Atoi(r.FormValue(common.FIELD_PREPARATION))
	if err != nil {
		preparation = 1
	}

	necessary, err := strconv.Atoi(r.FormValue(common.FIELD_NECESSARY))
	if err != nil {
		necessary = 0
	}
	if necessary != 0 {
		necessary = 1
	}

	// Check whether this is a ZIP file.
	isZip, err := strconv.Atoi(r.FormValue(common.FIELD_ZIP))
	if err != nil {
		isZip = 0
	}

	if isZip == 1 {
		if err = (func() error {
			namePrefix := common.Prune(r.FormValue(common.FIELD_NAME))
			if len(namePrefix) > 0 {
				namePrefix += "/"
			}

			fin, _, err := r.FormFile(common.FIELD_FILE)
			if err != nil {
				return err
			}
			defer fin.Close()

			// Get the length of the ZIP file.
			fsize, err := fin.Seek(0, os.SEEK_END)
			if err != nil {
				return err
			}
			_, err = fin.Seek(0, os.SEEK_SET)
			if err != nil {
				return err
			}

			// Load the ZIP file.
			fzip, err := zip.NewReader(fin, fsize)
			if err != nil {
				return err
			}

			// Load each item residing in the ZIP file, respectively.
			for i := 0; i < len(fzip.File); i++ {
				// Ignore directories and empty files.
				info := fzip.File[i].FileInfo()
				if info.IsDir() || info.Size() == 0 {
					continue
				}
				if strings.Index(fzip.File[i].Name, "__MACOSX") >= 0 || strings.Index(fzip.File[i].Name, ".DS_Store") >= 0 {
					continue
				}

				if err := (func() error {
					fzipIn, err := fzip.File[i].Open()
					if err != nil {
						return err
					}
					defer fzipIn.Close()

					filename := sv.uploadDir + "/" + sv.ss.GetUUID()
					fout, err := os.Create(filename)
					if err != nil {
						return err
					}
					defer fout.Close()

					_, err = io.Copy(fout, fzipIn)
					if err != nil {
						return err
					}

					name := namePrefix + common.RemoveFilenameSuffix(fzip.File[i].Name)

					_, err = sv.ms.AddPDF(filename, "", name, meetingID, preparation, necessary, session)
					if err != nil {
						return err
					}

					return nil
				})(); err != nil {
					return err
				}
			}

			return nil
		})(); err != nil {
			sv.Send(w, r, -2, err.Error(), "")
			return
		}
	} else {
		// Get courseware ID, if it exists.
		id := common.Prune(r.FormValue(common.FIELD_COURSEWARE))

		// Get name of this PDF file.
		name := common.Prune(r.FormValue(common.FIELD_NAME))

		// Save the form file to disk.
		fileName, err := (func() (string, error) {
			fin, _, err := r.FormFile(common.FIELD_FILE)
			if err != nil {
				return "", err
			}
			defer fin.Close()

			fileName := sv.uploadDir + "/" + sv.ss.GetUUID()
			fout, err := os.Create(fileName)
			if err != nil {
				return "", err
			}
			defer fout.Close()

			_, err = io.Copy(fout, fin)
			if err != nil {
				return "", err
			}

			return fileName, nil
		})()
		if err != nil {
			fileName = ""
			if len(id) == 0 {
				sv.Send(w, r, -3, err.Error(), "")
				return
			}
		}

		// Add a PDF file.
		_, err = sv.ms.AddPDF(fileName, id, name, meetingID, preparation, necessary, session)
		if err != nil {
			sv.Send(w, r, -4, err.Error(), "")
			return
		}
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteCourseware(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	coursewareID := r.FormValue(common.FIELD_COURSEWARE)
	if len(coursewareID) == 0 {
		sv.Send(w, r, -1, "Invalid courseware ID.", "")
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -2, "Invalid meeting ID.", "")
		return
	}

	err = sv.ms.DeleteCourseware(coursewareID, meetingID, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpFinishCourseware(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	coursewareID := common.Prune(r.FormValue(common.FIELD_COURSEWARE))
	if len(coursewareID) == 0 {
		sv.Send(w, r, -2, common.S_INVALID_COURSEWARE, "")
		return
	}

	err = sv.ms.SetCoursewareProgress(meetingID, coursewareID, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetGdCourseware(w http.ResponseWriter, r *http.Request) {
	_, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	coursewareID, err := strconv.Atoi(r.FormValue(common.FIELD_COURSEWARE))
	if err != nil {
		sv.Send(w, r, -2, "Invalid courseware ID.", "")
		return
	}

	gc, err := sv.gda.GetCourseware(coursewareID)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", gc.ToJSON())
}

//----------------------------------------------------------------------------
