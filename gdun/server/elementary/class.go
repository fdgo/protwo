package elementary

import (
	"fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"github.com/wangmhgo/go-project/gdun/service"
	"net/http"
	"strconv"
	// "strings"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	name := common.Prune(r.FormValue(common.FIELD_NAME))
	if len(name) == 0 {
		sv.Send(w, r, -1, "Empty class name.", "")
		return
	}

	subjects := common.StringToIntArray(r.FormValue(common.FIELD_SUBJECT))

	// Get the corresponding Gd course ID.
	gdCourseID := 0
	s := common.Prune(r.FormValue(common.FIELD_GAODUN_COURSE_ID))
	if len(s) > 0 {
		if gdCourseID, err = strconv.Atoi(s); err != nil {
			sv.Send(w, r, -2, "Invalid Gd class ID.", "")
			return
		}
	}

	// Get template of front pages.
	template := 0
	s = common.Prune(r.FormValue(common.FIELD_TEMPLATE))
	if len(s) > 0 {
		if template, err = strconv.Atoi(s); err != nil {
			sv.Send(w, r, -3, "Invalid template.", "")
			return
		}
	}

	// Get configuration of live platform.
	platformID := 0
	s = common.Prune(r.FormValue(common.FIELD_PLATFORM_ID))
	if len(s) > 0 {
		if platformID, err = strconv.Atoi(s); err != nil {
			sv.Send(w, r, -4, "Invalid platform ID.", "")
			return
		}
	}
	platformData := common.Prune(r.FormValue(common.FIELD_PLATFORM_DATA))

	groupID := session.GroupID
	if session.IsSystem() {
		if groupID, err = strconv.Atoi(r.FormValue(common.FIELD_GROUP)); err != nil {
			sv.Send(w, r, -5, "Invalid group ID.", "")
			return
		}
	}

	id, err := sv.cs.AddClass(name, subjects, groupID, gdCourseID, template, platformID, platformData, session)
	if err != nil {
		sv.Send(w, r, -6, err.Error(), "")
		return
	}

	result := `"` + common.FIELD_ID + `":` + strconv.Itoa(id)
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpChangeClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	// Get class ID.
	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, "Invalid class ID.", "")
		return
	}

	// Get class name.
	name := r.FormValue(common.FIELD_NAME)

	subjects := common.StringToIntArray(r.FormValue(common.FIELD_SUBJECT))

	// Get the correspond Gd course ID.
	gdCourseID := -1
	s := common.Prune(r.FormValue(common.FIELD_GAODUN_COURSE_ID))
	if len(s) > 0 {
		if gdCourseID, err = strconv.Atoi(s); err != nil {
			sv.Send(w, r, -2, "Invalid Gd class ID.", "")
			return
		}
	}

	// Get template of front page.
	template := -1
	s = common.Prune(r.FormValue(common.FIELD_TEMPLATE))
	if len(s) > 0 {
		if template, err = strconv.Atoi(s); err != nil {
			sv.Send(w, r, -3, "Invalid template.", "")
			return
		}
	}

	// Get configuration of live platform.
	platformID := -1
	platformData := common.Prune(r.FormValue(common.FIELD_PLATFORM_DATA))
	s = common.Prune(r.FormValue(common.FIELD_PLATFORM_ID))
	if len(s) > 0 {
		if platformID, err = strconv.Atoi(s); err != nil {
			sv.Send(w, r, -4, "Invalid platform ID.", "")
			return
		}
	}

	err = sv.cs.ChangeClass(classID, name, subjects, gdCourseID, template, platformID, platformData, session)
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpImportClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	//----------------------------------------------------
	// Check inputs.

	name := common.Prune(r.FormValue(common.FIELD_NAME))
	if len(name) == 0 {
		sv.Send(w, r, -1, "Empty class name.", "")
		return
	}

	gdCourseID, err := strconv.Atoi(r.FormValue(common.FIELD_GAODUN_COURSE_ID))
	if err != nil {
		sv.Send(w, r, -2, "Invalid Gd course ID.", "")
		return
	}

	groupID := session.GroupID
	if session.IsSystem() {
		groupID, err = strconv.Atoi(r.FormValue(common.FIELD_GROUP))
		if err != nil {
			sv.Send(w, r, -3, "Invalid group ID.", "")
		}
	}

	//----------------------------------------------------
	// Export the Gd course.

	course, err := sv.gda.GetCourse(gdCourseID)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	//----------------------------------------------------
	// Create a class.

	classID, err := sv.cs.AddClass(name, nil, groupID, 0, 0, 0, "", session)
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	//----------------------------------------------------
	// Create meetings.

	for e := course.Coursewares.Front(); e != nil; e = e.Next() {
		cw, okay := e.Value.(*service.GdCourseware)
		if !okay {
			continue
		}
		if len(cw.Name) == 0 {
			continue
		}

		// Create a meeting.
		meetingID, err := sv.cs.AddMeeting(cw.Name, nil, 0, 0, 0, classID, 0, "", session, true)
		if err != nil {
			sv.Send(w, r, -8, err.Error(), "")
			return
		}

		for res := cw.Resources.Front(); res != nil; res = res.Next() {
			v, okay := res.Value.(*service.GdVideoResource)
			if okay {
				// Add this video to the meeting.
				if err = sv.ms.AddVideo(v.ID, v.Name, meetingID, 1, 1, session); err != nil {
					// fmt.Println(cw.ToJSON())
					sv.Send(w, r, -9, err.Error(), "")
					return
				}
				continue
			}

			exam, okay := res.Value.(*service.GdTikuExam)
			if okay {
				// Save this exam.
				id, _, err := sv.sv.ImportGdExam(exam.Name, groupID, 0, exam, session)
				if err != nil {
					sv.Send(w, r, -10, err.Error(), "")
					return
				}

				// Add this exam to the meeting.
				if err = sv.ms.AddExam(meetingID, id, exam.ID, exam.Name, 0, 0, 1, true, groupID, session); err != nil {
					sv.Send(w, r, -11, err.Error(), "")
					return
				}
			}
		}
	}

	result := `"` + common.FIELD_ID + `":` + strconv.Itoa(classID)
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpCloneClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, "Empty original class ID.", "")
		return
	}

	name := r.FormValue(common.FIELD_NAME)
	if len(name) == 0 {
		sv.Send(w, r, -2, "Empty class name.", "")
		return
	}

	// Get configuration of live platform.
	platformID, err := strconv.Atoi(r.FormValue(common.FIELD_PLATFORM_ID))
	if err != nil {
		platformID = 0
	}
	platformData := r.FormValue(common.FIELD_PLATFORM_DATA)

	id, err := sv.cs.CloneClass(classID, name, platformID, platformData, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	result := `"` + common.FIELD_ID + `":` + strconv.Itoa(id)
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpEndClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, "Invalid class ID.", "")
		return
	}

	if err = sv.cs.EndClass(classID, session); err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	if err = sv.cs.DeleteClass(classID, session); err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpPublishClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	if err = sv.cs.PublishClass(classID, session); err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpPackageClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	if err = sv.cs.PackageClass(classID, session); err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpGenerateClassInvitationToken(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, "Invalid class ID.", "")
		return
	}

	teacherID, err := strconv.Atoi(r.FormValue(common.FIELD_TEACHER))
	if err != nil || teacherID < 0 {
		teacherID = 0
	}

	size, err := strconv.Atoi(r.FormValue(common.FIELD_SIZE))
	if (teacherID == 0) && (err != nil || size <= 0 || size > 100) {
		sv.Send(w, r, -2, "Invalid size.", "")
		return
	}

	channel := common.Prune(r.FormValue(common.FIELD_CHANNEL))
	if (teacherID == 0) && (channel == "") {
		sv.Send(w, r, -3, "Invalid teacher ID or channel name.", "")
		return
	}

	endTime, err := strconv.Atoi(r.FormValue(common.FIELD_END_TIME))
	if err != nil || endTime == 0 {
		sv.Send(w, r, -4, "Invalid token end time.", "")
		return
	}

	duration, err := strconv.Atoi(r.FormValue(common.FIELD_DURATION))
	if err != nil {
		duration = 0
		// if err != nil || duration == 0 {
		// sv.Send(w, r, -5, "Invalid token duration.", "")
		// return
	}

	if err = sv.cs.GenerateInvitationToken(classID, endTime, duration, teacherID, channel, size, session); err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryClassInvitationToken(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, "Invalid class ID.", "")
		return
	}

	isToken := false
	s := common.Prune(r.FormValue(common.FIELD_IS_TEACHER))
	if s == "" {
		isToken = true
	}

	result, err := sv.cs.QueryInvitationToken(classID, isToken, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryExperienceUser(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, "Invalid class ID.", "")
		return
	}

	isToken := false
	s := common.Prune(r.FormValue(common.FIELD_IS_TEACHER))
	if s == "" {
		isToken = true
	}

	result, err := sv.cs.QueryExperienceUserLog(classID, isToken, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryClass(w http.ResponseWriter, r *http.Request) {
	// var session *service.Session = nil

	gdStudentID, err := strconv.Atoi(r.FormValue(common.FIELD_GAODUN_STUDENT_ID))
	if err != nil {
		gdStudentID = 0
	}

	if gdStudentID == 0 {
		// Check session.
		session, err := sv.ss.CheckHttpSessionForUser(w, r)
		if err != nil {
			return
		}

		// Check group ID.
		groupID := 0
		if session.IsSystem() {
			if groupID, err = strconv.Atoi(r.FormValue(common.FIELD_GROUP)); err != nil {
				sv.Send(w, r, -1, common.S_INVALID_GROUP, "")
				return
			}
		} else if session.IsAssistant() {
			groupID = session.GroupID
		}

		// Get all classes.
		result, stauts, err := sv.cs.GetClasses(groupID, session)
		if err != nil {
			sv.Send(w, r, stauts-1, err.Error(), "")
			return
		}

		sv.Send(w, r, 0, "", result.ToJSON(session.IsStudent()))

	} else {
		// Check IP address.
		if !sv.isLanIP(r.RemoteAddr) {
			w.WriteHeader(404)
			return
		}

		var gdCourseIDs []int = nil

		// Get class IDs.
		classIDs := common.StringToIntArray(r.FormValue(common.FIELD_CLASS))
		if len(classIDs) == 0 {
			// Get Gd course IDs.
			gdCourseIDs = common.StringToIntArray(r.FormValue(common.FIELD_GAODUN_COURSE_ID))
			if len(gdCourseIDs) == 0 {
				sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
				return
			}

			// Translate them to class IDs.
			classIDs = make([]int, len(gdCourseIDs))
			for i := 0; i < len(gdCourseIDs); i++ {
				classIDs[i], err = sv.cs.GetClassIDViaGdCourseID(gdCourseIDs[i])
				if err != nil {
					sv.Send(w, r, -2, err.Error(), "")
					return
				}
			}
		}

		// Get his user ID.
		userID, err := sv.us.QueryUserID(gdStudentID)
		if err != nil {
			sv.Send(w, r, -3, common.S_INVALID_STUDENT, "")
			return
		}

		// Construct a fake session.
		session := new(service.Session)
		session.UserID = userID
		session.GroupID = common.GROUP_ID_FOR_STUDENT
		session.IP = r.RemoteAddr

		result := ""
		first := true
		for i := 0; i < len(classIDs); i++ {
			classID, _, err := sv.cs.GetClassIDViaUserID(classIDs[i], userID)
			if err != nil {
				// TODO:
				continue
			}

			ci, err := sv.cs.GetClass(classID, session)
			if err != nil {
				// TODO:
				continue
			}

			p, err := sv.cs.GetClassOverallProgress(classID, session.UserID)
			if err != nil {
				fmt.Println(err.Error())
				p = 0
			}

			if first {
				first = false
			} else {
				result += ","
			}

			result += `"`
			if gdCourseIDs == nil {
				result += strconv.Itoa(classIDs[i])
			} else {
				result += strconv.Itoa(gdCourseIDs[i])
			}
			result += `":{` + ci.ToJSON(true) + `,"` + common.FIELD_PROGRESS + `":` + strconv.Itoa(p) + `}`
		}

		sv.Send(w, r, 0, "", result)
	}
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	ci, err := sv.cs.GetClass(classID, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", ci.ToJSON(session.IsStudent()))
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpClassAddTeacher(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	userID, err := strconv.Atoi(r.FormValue(common.FIELD_TEACHER))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	err = sv.cs.ChangeUser(userID, classID, true, true, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpClassDeleteTeacher(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	userID, err := strconv.Atoi(r.FormValue(common.FIELD_TEACHER))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	err = sv.cs.ChangeUser(userID, classID, true, false, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpClassQueryTeacher(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	result, err := sv.cs.QueryUsers(classID, true, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result.ToJSON())
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpClassAddStudent(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	userID, err := strconv.Atoi(r.FormValue(common.FIELD_STUDENT))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	err = sv.cs.ChangeUser(userID, classID, false, true, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpClassAddGdStudent(w http.ResponseWriter, r *http.Request) {
	var session *service.Session = nil
	var err error = nil

	if sv.isLanIP(r.RemoteAddr) {
		session = new(service.Session)
		session.UserID = 0
		session.GroupID = common.GROUP_ID_FOR_SYSTEM
		session.IP = r.RemoteAddr
	} else {
		session, err = sv.ss.CheckHttpSessionForAssitant(w, r)
		if err != nil {
			return
		}
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	gdStudentID, err := strconv.Atoi(r.FormValue(common.FIELD_GAODUN_STUDENT_ID))
	if err != nil {
		gdAccount := common.Prune(r.FormValue(common.FIELD_GAODUN_ACCOUNT))
		if len(gdAccount) == 0 {
			sv.Send(w, r, -2, common.S_INVALID_STUDENT, "")
			return
		}

		gdStudentID = sv.gdp.GetStudentID(gdAccount)
	}
	if gdStudentID <= 0 {
		sv.Send(w, r, -3, common.S_INVALID_STUDENT, "")
		return
	}

	userID, _, _, err := sv.us.GetOrAddGdStudent(gdStudentID, session.IP, true)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	name := common.Prune(r.FormValue(common.FIELD_NAME))
	if len(name) > 0 {
		if err = sv.us.Remark(userID, name, "", session); err != nil {
			sv.Send(w, r, -5, err.Error(), "")
			return
		}
	}

	err = sv.cs.ChangeUser(userID, classID, false, true, session)
	if err != nil {
		sv.Send(w, r, -6, err.Error(), "")
		return
	}

	result := `"` + common.FIELD_USER + `":` + strconv.Itoa(userID)
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddGdStudentToGdCourse(w http.ResponseWriter, r *http.Request) {
	// Check authority.
	okay := sv.isLanIP(r.RemoteAddr)

	// Check authority and construct a fake session.
	var session *service.Session = nil
	var err error = nil
	if okay {
		// Construct a session structure, if this request originates from a back-end server.
		session = new(service.Session)
		session.UserID = 0
		session.GroupID = common.GROUP_ID_FOR_SYSTEM
		session.IP = r.RemoteAddr
	} else {
		session, err = sv.ss.CheckHttpSessionForSystem(w, r)
		if err == nil {
			if session.IsSystem() {
				okay = true
			}
		}
	}
	if !okay {
		w.WriteHeader(404)
		return
	}

	//----------------------------------------------------

	// Prepare HTTP request.
	err = r.ParseForm()
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	// Translate class ID.
	gdCourseID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}
	if sv.cache != nil {
		if okay, err := sv.cache.FieldExist(common.FIELD_GAODUN_COURSE_ID+":"+common.FIELD_DEPRECATED, strconv.Itoa(gdCourseID)); (err == nil) && okay {
			sv.Send(w, r, -7, "Deprecated Gd course ID.", "")
			return
		}
	}
	classID, err := sv.cs.GetClassIDViaGdCourseID(gdCourseID)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	// Translate user ID.
	gdStudentID, err := strconv.Atoi(r.FormValue(common.FIELD_STUDENT))
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}
	userID, _, _, err := sv.us.GetOrAddGdStudent(gdStudentID, r.RemoteAddr, true)
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	//----------------------------------------------------

	ci, err := sv.cs.GetClass(classID, session)
	if err != nil {
		// TODO:
	}
	if ci.Ally > 0 {
		classID = ci.Ally
	}

	// Add this student to the specified class.
	err = sv.cs.ChangeUser(userID, classID, false, true, session)
	if err != nil {
		sv.Send(w, r, -6, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpClassDeleteStudent(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	userID, err := strconv.Atoi(r.FormValue(common.FIELD_STUDENT))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	err = sv.cs.ChangeUser(userID, classID, false, false, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpClassDeleteGdStudent(w http.ResponseWriter, r *http.Request) {
	var session *service.Session = nil
	var err error = nil

	if sv.isLanIP(r.RemoteAddr) {
		session = new(service.Session)
		session.UserID = 0
		session.GroupID = common.GROUP_ID_FOR_SYSTEM
		session.IP = r.RemoteAddr
	} else {
		session, err = sv.ss.CheckHttpSessionForAssitant(w, r)
		if err != nil {
			return
		}
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	gdStudentID, err := strconv.Atoi(r.FormValue(common.FIELD_GAODUN_STUDENT_ID))
	if err != nil {
		gdAccount := common.Prune(r.FormValue(common.FIELD_GAODUN_ACCOUNT))
		if len(gdAccount) == 0 {
			sv.Send(w, r, -2, common.S_INVALID_STUDENT, "")
			return
		}

		gdStudentID = sv.gdp.GetStudentID(gdAccount)
	}
	if gdStudentID <= 0 {
		sv.Send(w, r, -3, common.S_INVALID_STUDENT, "")
		return
	}

	userID, _, _, err := sv.us.GetOrAddGdStudent(gdStudentID, session.IP, false)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	err = sv.cs.ChangeUser(userID, classID, false, false, session)
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	// result := `"` + common.FIELD_USER + `":` + strconv.Itoa(userID)
	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpClassRemarkStudent(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	isTeacher := false
	userID, err := strconv.Atoi(r.FormValue(common.FIELD_STUDENT))
	if err != nil {
		userID, err = strconv.Atoi(r.FormValue(common.FIELD_TEACHER))
		if err != nil {
			sv.Send(w, r, -1, common.S_INVALID_USER, "")
			return
		}
		isTeacher = true
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -2, common.S_INVALID_CLASS_ID, "")
		return
	}

	name := common.Prune(r.FormValue(common.FIELD_NAME))
	remark := common.Prune(r.FormValue(common.FIELD_REMARK))

	if (len(name) == 0) && (len(remark) == 0) {
		sv.Send(w, r, -3, "Empty name and remark.", "")
		return
	}

	// Check authority.

	ci, err := sv.cs.GetClass(classID, session)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	okay := false
	if isTeacher {
		for i := 0; i < len(ci.Teachers); i++ {
			if ci.Teachers[i] == userID {
				okay = true
				break
			}
		}
	} else {
		for i := 0; i < len(ci.Students); i++ {
			if ci.Students[i] == userID {
				okay = true
				break
			}
		}
	}
	if !okay {
		sv.Send(w, r, -5, common.S_NO_AUTHORITY, "")
		return
	}

	err = sv.us.Remark(userID, name, remark, session)
	if err != nil {
		sv.Send(w, r, -6, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAssociateKeeperWithStudent(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	keeperID, err := strconv.Atoi(r.FormValue(common.FIELD_KEEPER))
	if err != nil {
		sv.Send(w, r, -2, common.S_INVALID_KEEPER, "")
		return
	}

	studentID, err := strconv.Atoi(r.FormValue(common.FIELD_STUDENT))
	if err != nil {
		sv.Send(w, r, -3, common.S_INVALID_KEEPER, "")
		return
	}

	err = sv.cs.ChangeKeeperStudentRelation(classID, studentID, keeperID, true, session)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	err = sv.cs.RefreshKeeperList(classID)
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDisassociateKeeperWithStudent(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	keeperID, err := strconv.Atoi(r.FormValue(common.FIELD_KEEPER))
	if err != nil {
		sv.Send(w, r, -2, common.S_INVALID_KEEPER, "")
		return
	}

	studentID, err := strconv.Atoi(r.FormValue(common.FIELD_STUDENT))
	if err != nil {
		sv.Send(w, r, -3, common.S_INVALID_KEEPER, "")
		return
	}

	err = sv.cs.ChangeKeeperStudentRelation(classID, studentID, keeperID, false, session)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	err = sv.cs.RefreshKeeperList(classID)
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpClassQueryStudent(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	ci, err := sv.cs.GetClass(classID, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	if session.IsStudent() {
		result := `"` + common.FIELD_STUDENT + `":` + common.IntArrayToJSON(ci.Students)
		sv.Send(w, r, 0, "", result)
	} else {
		result, err := sv.cs.QueryUsers(classID, false, session)
		if err != nil {
			sv.Send(w, r, -3, err.Error(), "")
			return
		}

		sv.Send(w, r, 0, "", result.ToJSON())
	}
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryClassProgress(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	// ci, err := sv.cs.GetClass(classID, session)
	// if err != nil {
	// 	sv.Send(w, r, -2, err.Error(), "")
	// 	return
	// }

	// For student.
	if session.IsStudent() {
		result, err := sv.cs.QueryUserProgress(classID, session.UserID, session)
		if err != nil {
			sv.Send(w, r, -3, err.Error(), "")
			return
		}

		sv.Send(w, r, 0, "", result.ToJSON(true))
	} else {
		// For other roles.
		userID, err := strconv.Atoi(r.FormValue(common.FIELD_USER))
		if err == nil {
			// For one student.
			result, err := sv.cs.QueryUserProgress(classID, userID, session)
			if err != nil {
				sv.Send(w, r, -4, err.Error(), "")
				return
			}

			sv.Send(w, r, 0, "", result.ToJSON(true))
		} else {
			// For all students.
			result, err := sv.cs.QueryUserProgressesA(classID, session)
			// result, err := sv.cs.GetCachedUserProgresses(ci.Meetings, ci.Students)
			if err != nil {
				sv.Send(w, r, -5, err.Error(), "")
				return
			}

			sv.Send(w, r, 0, "", result)
		}
	}
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryClassBriefs(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(404)
		return
	}

	var session *service.Session = nil
	if len(r.FormValue(common.FIELD_TOKEN)) == 0 {
		userID := (func() int {
			// Try Gd session.
			s := common.Prune(r.FormValue(common.FIELD_SESSION))
			if len(s) == 0 {
				return 0
			}

			// gdStudentID, err := sv.gss.GetStudentID(s)
			// if err != nil {
			// 	return 0
			// }

			gdStudentID, err := sv.gdp.CheckLogin(s)
			if err != nil {
				return 0
			}

			userID, _, _, err := sv.us.GetOrAddGdStudent(gdStudentID, r.RemoteAddr, false)
			if err != nil {
				return 0
			}

			return userID
		})()

		// fmt.Println(userID)
		if userID <= 0 {
			w.WriteHeader(404)
			return
		}

		session = new(service.Session)
		session.UserID = userID
		session.GroupID = common.GROUP_ID_FOR_STUDENT
		session.IP = r.RemoteAddr
	} else {
		session, err = sv.ss.CheckHttpSessionForStudent(w, r)
		if err != nil {
			return
		}
	}

	result, err := sv.cs.QueryClassBriefs(session)
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddSubClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, "Invalid class ID.", "")
		return
	}
	subClassID, err := strconv.Atoi(r.FormValue(common.FIELD_SUBCLASS))
	if err != nil {
		sv.Send(w, r, -2, "Invalid sub-class ID.", "")
		return
	}

	err = sv.cs.AddSubClass(classID, subClassID, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteSubClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, "Invalid class ID.", "")
		return
	}
	subClassID, err := strconv.Atoi(r.FormValue(common.FIELD_SUBCLASS))
	if err != nil {
		sv.Send(w, r, -2, "Invalid sub-class ID.", "")
		return
	}

	err = sv.cs.DeleteSubClass(classID, subClassID, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQuerySubClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	result, err := sv.cs.QuerySubClasses(classID, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result.ToJSON(false))
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpSetClassAlly(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	ally, err := strconv.Atoi(r.FormValue(common.FIELD_ALLY))
	if err != nil || ally < 0 {
		ally = 0
	}

	err = sv.cs.SetAlly(classID, ally, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------
// TODO:

func (sv *Server) onHttpSetClassCover(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	cover, _, err := r.FormFile(common.FIELD_FILE)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	err = sv.cs.SetCover(classID, cover, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------
