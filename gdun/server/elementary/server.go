package elementary

import (
	"bytes"
	"compress/gzip"
	// "fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"github.com/wangmhgo/go-project/gdun/log"
	"github.com/wangmhgo/go-project/gdun/service"
	"net/http"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

const (
	backlogSize = 1024
)

//----------------------------------------------------------------------------

type Server struct {
	urlPrefix      string
	frontEndDomain string
	uploadDir      string
	lanIPPrefix    []string
	db             *common.Database
	cache          *common.Cache
	studentWeixin  *common.WeixinClient
	teacherWeixin  *common.WeixinClient
	ss             *service.SessionService
	us             *service.UserService
	gs             *service.GroupService
	cs             *service.ClassService
	ms             *service.MeetingService
	sv             *service.ExamService
	ns             *service.NoteService
	is             *service.IssueService
	sbjs           *service.SubjectService
	vs             *service.VideoService
	ts             *service.TagService
	gdp            *service.GdPass
	gda            *service.GdAdapter
	gss            *service.GdSessionService
	accessLog      *log.Logger
	authorizeLog   *log.Logger
	userLog        *log.Logger
}

func NewServer(db *common.Database, cache *common.Cache, gdDB *common.Database, oss *common.ObjectStorage, cfg *Config) (*Server, error) {

	sv := new(Server)

	sv.urlPrefix = cfg.UrlPrefix
	sv.frontEndDomain = cfg.FrontEndDomain
	sv.lanIPPrefix = cfg.LanIPPrefix
	sv.uploadDir = cfg.UploadDir

	//----------------------------------------------------

	sv.db = db
	sv.cache = cache

	sv.studentWeixin = (func() *common.WeixinClient {
		if sv.cache == nil {
			return nil
		}

		key := common.KEY_PREFIX_CONFIG + "weixin"

		appID, err := sv.cache.GetField(key, "studentAppID")
		if err != nil {
			return nil
		}
		appSecret, err := sv.cache.GetField(key, "studentAppSecret")
		if err != nil {
			return nil
		}

		return common.NewWeixinClient(appID, appSecret)
	})()

	//----------------------------------------------------

	sv.gda = service.NewGdAdapter(gdDB)
	sv.gss = service.NewGdSessionService()

	//----------------------------------------------------
	// General

	sv.accessLog = log.GetLogger(cfg.LogDir, "access", log.LEVEL_INFO)
	sv.authorizeLog = log.GetLogger(cfg.LogDir, "authorize", log.LEVEL_INFO)
	sv.userLog = log.GetLogger(cfg.LogDir, "user", log.LEVEL_INFO)

	ts := service.NewTranscodingService(oss, cfg.CoursewareKeyPrefix, cfg.GhostScript, cfg.TmpDir, nil)

	//----------------------------------------------------

	var err error

	sv.vs, err = service.NewVideoService(nil, sv.cache, nil, nil, nil, nil, "")
	if err != nil {
		return nil, err
	}

	sv.ss = service.NewSessionService(sv.cache, sv.accessLog)

	sv.ns, err = service.NewNoteService(sv.db, sv.cache, backlogSize)
	if err != nil {
		return nil, err
	}

	sv.sv, err = service.NewExamService(sv.db, sv.cache, oss, cfg.ExamKeyPrefix, sv.gda)
	if err != nil {
		return nil, err
	}

	sv.us, err = service.NewUserService(sv.db, sv.cache, sv.ss)
	if err != nil {
		return nil, err
	}

	sv.ms, err = service.NewMeetingService(sv.db, sv.cache, cfg.LiveServer, ts, sv.sv)
	if err != nil {
		return nil, err
	}

	sv.cs, err = service.NewClassService(sv.db, sv.cache, oss, sv.ms)
	if err != nil {
		return nil, err
	}

	sv.gs, err = service.NewGroupService(sv.db, sv.cache)
	if err != nil {
		return nil, err
	}

	sv.ts, err = service.NewTagService(sv.db, sv.cache)
	if err != nil {
		return nil, err
	}

	sv.sbjs, err = service.NewSubjectService(db, cache)
	if err != nil {
		return nil, err
	}
	sv.is, err = service.NewIssueService(sv.db, sv.cache, sv.cs, sv.ms)
	if err != nil {
		return nil, err
	}

	sv.gdp = service.NewGdPass(sv.cs, sv.us, cfg.GdEncryptKey)

	//----------------------------------------------------

	sv.registerHttpHandles()

	return sv, nil
}

//----------------------------------------------------------------------------

func (sv *Server) Close() error {
	if sv.db != nil {
		if err := sv.db.Close(); err != nil {
			return err
		}
	}
	if sv.cache != nil {
		if err := sv.cache.Close(); err != nil {
			return err
		}
	}
	return nil
}

//----------------------------------------------------------------------------

func (sv *Server) createLogLine(r *http.Request) string {
	s := r.RemoteAddr + " " + r.Host + r.RequestURI + " (" + r.UserAgent() + ") (" + r.Referer() + ")"
	if r.TLS != nil {
		s += " TLS"
	}

	return s
}

func (sv *Server) isLanIP(ip string) bool {
	if sv.lanIPPrefix == nil {
		return false
	}

	for i := 0; i < len(sv.lanIPPrefix); i++ {
		if strings.HasPrefix(ip, sv.lanIPPrefix[i]) {
			return true
		}
	}

	return false
}

func (sv *Server) Go404(w http.ResponseWriter, r *http.Request, info string) {
	if sv.accessLog != nil {
		go sv.accessLog.Error(r.RemoteAddr + " " + r.Host + r.RequestURI + " (" + info + ") (" + r.UserAgent() + ") (" + r.Referer() + ")")
	}
	w.WriteHeader(404)
}

//----------------------------------------------------------------------------

func (sv *Server) Send(w http.ResponseWriter, r *http.Request, status int, info string, result string) (int, error) {
	if sv.accessLog != nil {
		s := r.RemoteAddr + " " + r.Host + r.RequestURI
		if status == 0 {
			go sv.accessLog.Info(s)
		} else {
			go sv.accessLog.Warning(s + " " + strconv.Itoa(status) + " (" + info + ") (" + r.UserAgent() + ") (" + r.Referer() + ")")
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", sv.frontEndDomain)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Server", "solo/2018.02.23")

	body := common.PlainJSONMessage(status, info, result)

	if len(body) > 10240 {
		ae := r.Header.Get("Accept-Encoding")
		if strings.Index(ae, "gzip") >= 0 {
			var b bytes.Buffer
			gz := gzip.NewWriter(&b)

			if _, err := gz.Write(body); err == nil {
				if err := gz.Flush(); err == nil {
					if err := gz.Close(); err == nil {
						w.Header().Set("Content-Encoding", "gzip")
						return w.Write(b.Bytes())
					}
				}
			}
		}
	}

	return w.Write(body)
}

//----------------------------------------------------------------------------

func (sv *Server) registerHttpHandles() {

	//----------------------------------------------------

	http.HandleFunc(sv.urlPrefix+"/video/query", sv.onHttpQueryVideo)

	http.HandleFunc(sv.urlPrefix+"/player/authorize", sv.onHttpAuthorizeM3U8)

	http.HandleFunc(sv.urlPrefix+"/internal/video/get", sv.onHttpGetInternalVideo)
	http.HandleFunc(sv.urlPrefix+"/internal/video/add", sv.onHttpAddInternalVideo)
	http.HandleFunc(sv.urlPrefix+"/internal/video/delete", sv.onHttpDeleteInternalVideo)

	http.HandleFunc(sv.urlPrefix+"/internal/ip/get", sv.onHttpGetInternalIP)
	http.HandleFunc(sv.urlPrefix+"/internal/ip/add", sv.onHttpAddInternalIP)
	http.HandleFunc(sv.urlPrefix+"/internal/ip/delete", sv.onHttpDeleteInternalIP)

	// Bokecc.
	http.HandleFunc(sv.urlPrefix+"/bokecc/pass", sv.onHttpBokeccLogin)
	http.HandleFunc(sv.urlPrefix+"/bokecc/internal/register", sv.onHttpBokeccInternalRegister)
	http.HandleFunc(sv.urlPrefix+"/bokecc/internal/pass", sv.onHttpBokeccInternalPass)

	//----------------------------------------------------

	http.HandleFunc(sv.urlPrefix+"/app/config/get", sv.onHttpGetAppConfig)

	//----------------------------------------------------

	http.HandleFunc(sv.urlPrefix+"/weixin/callback/login", sv.onHttpWeixinLogin)
	http.HandleFunc(sv.urlPrefix+"/weixin/template/get", sv.onHttpGetWeixinMessageTemplates)

	//----------------------------------------------------

	// Gd Pass.
	http.HandleFunc(sv.urlPrefix+"/gd/login", sv.onHttpLoginAsGdUser)
	http.HandleFunc(sv.urlPrefix+"/gd/3rd/login", sv.onHttpLoginAs3rdUser)
	http.HandleFunc(sv.urlPrefix+"/gd/pass", sv.onHttpLoginAsGdStudent)

	// Gd Adapter.
	http.HandleFunc(sv.urlPrefix+"/gd/student/add", sv.onHttpAddGdStudentToGdCourse)
	http.HandleFunc(sv.urlPrefix+"/gd/student/info/query", sv.onHttpQueryGdStudent)
	http.HandleFunc(sv.urlPrefix+"/gd/student/name/resync", sv.onHttpResyncGdStudentName)
	http.HandleFunc(sv.urlPrefix+"/gd/courseware/get", sv.onHttpGetGdCourseware)
	http.HandleFunc(sv.urlPrefix+"/gd/class/query", sv.onHttpQueryClassBriefs)
	http.HandleFunc(sv.urlPrefix+"/gd/meeting/query", sv.onHttpQueryMeetingViaGdCourseID)

	//----------------------------------------------------

	http.HandleFunc(sv.urlPrefix+"/user/register", sv.onHttpRegister)
	http.HandleFunc(sv.urlPrefix+"/user/login", sv.onHttpLogin)
	http.HandleFunc(sv.urlPrefix+"/user/password/change", sv.onHttpChangePassword)
	http.HandleFunc(sv.urlPrefix+"/user/profile/change", sv.onHttpChangeProfile)

	http.HandleFunc(sv.urlPrefix+"/user/query", sv.onHttpQueryUser)
	http.HandleFunc(sv.urlPrefix+"/user/scan", sv.onHttpScan)

	//----------------------------------------------------

	http.HandleFunc(sv.urlPrefix+"/user/experience/register", sv.onHttpExperienceRegister)
	http.HandleFunc(sv.urlPrefix+"/user/experience/login", sv.onHttpExperienceLogin)
	http.HandleFunc(sv.urlPrefix+"/user/verification/send", sv.onHttpSendVerificationCode)

	http.HandleFunc(sv.urlPrefix+"/user/invitation/generate", sv.onHttpGenerateInvitationToken)
	http.HandleFunc(sv.urlPrefix+"/user/invitation/query", sv.onHttpQueryInvitationToken)
	http.HandleFunc(sv.urlPrefix+"/user/invitation/delete", sv.onHttpDeleteInvitationToken)

	//----------------------------------------------------

	http.HandleFunc(sv.urlPrefix+"/subject/add", sv.onHttpAddSubject)
	http.HandleFunc(sv.urlPrefix+"/subject/change", sv.onHttpChangeSubject)
	http.HandleFunc(sv.urlPrefix+"/subject/query", sv.onHttpQuerySubject)

	http.HandleFunc(sv.urlPrefix+"/tag/add", sv.onHttpAddTag)
	http.HandleFunc(sv.urlPrefix+"/tag/change", sv.onHttpChangeTag)
	http.HandleFunc(sv.urlPrefix+"/tag/query", sv.onHttpQueryTag)

	http.HandleFunc(sv.urlPrefix+"/group/add", sv.onHttpAddGroup)
	http.HandleFunc(sv.urlPrefix+"/group/query", sv.onHttpQueryGroup)
	http.HandleFunc(sv.urlPrefix+"/group/delete", sv.onHttpDeleteGroup)
	// To be deleted:
	http.HandleFunc(sv.urlPrefix+"/user/group/add", sv.onHttpAddGroup)
	http.HandleFunc(sv.urlPrefix+"/user/group/query", sv.onHttpQueryGroup)
	http.HandleFunc(sv.urlPrefix+"/user/group/delete", sv.onHttpDeleteGroup)

	http.HandleFunc(sv.urlPrefix+"/group/subject/add", sv.onHttpAddSubjectToGroup)
	http.HandleFunc(sv.urlPrefix+"/group/subject/delete", sv.onHttpDeleteSubjectFromGroup)
	http.HandleFunc(sv.urlPrefix+"/group/subject/query", sv.onHttpQuerySubjectForGroup)

	//----------------------------------------------------

	http.HandleFunc(sv.urlPrefix+"/class/get", sv.onHttpGetClass)

	http.HandleFunc(sv.urlPrefix+"/class/add", sv.onHttpAddClass)
	http.HandleFunc(sv.urlPrefix+"/class/change", sv.onHttpChangeClass)
	http.HandleFunc(sv.urlPrefix+"/class/import", sv.onHttpImportClass)
	http.HandleFunc(sv.urlPrefix+"/class/clone", sv.onHttpCloneClass)
	http.HandleFunc(sv.urlPrefix+"/class/query", sv.onHttpQueryClass)
	http.HandleFunc(sv.urlPrefix+"/class/end", sv.onHttpEndClass)
	http.HandleFunc(sv.urlPrefix+"/class/delete", sv.onHttpDeleteClass)
	http.HandleFunc(sv.urlPrefix+"/class/publish", sv.onHttpPublishClass)
	http.HandleFunc(sv.urlPrefix+"/class/package", sv.onHttpPackageClass)

	http.HandleFunc(sv.urlPrefix+"/class/subclass/add", sv.onHttpAddSubClass)
	http.HandleFunc(sv.urlPrefix+"/class/subclass/delete", sv.onHttpDeleteSubClass)
	http.HandleFunc(sv.urlPrefix+"/class/subclass/query", sv.onHttpQuerySubClass)

	http.HandleFunc(sv.urlPrefix+"/class/cover/set", sv.onHttpSetClassCover)
	http.HandleFunc(sv.urlPrefix+"/class/ally/set", sv.onHttpSetClassAlly)

	//----------------------------------------------------

	http.HandleFunc(sv.urlPrefix+"/class/meeting/add", sv.onHttpAddMeeting)
	http.HandleFunc(sv.urlPrefix+"/class/meeting/query", sv.onHttpQueryMeeting)
	http.HandleFunc(sv.urlPrefix+"/class/meeting/delete", sv.onHttpDeleteMeetingFromClass)
	http.HandleFunc(sv.urlPrefix+"/class/meeting/end", sv.onHttpEndMeeting)
	http.HandleFunc(sv.urlPrefix+"/class/meeting/copy", sv.onHttpCopyMeeting)
	http.HandleFunc(sv.urlPrefix+"/class/meeting/change", sv.onHttpChangeMeeting)

	http.HandleFunc(sv.urlPrefix+"/class/progress/query", sv.onHttpQueryClassProgress)

	http.HandleFunc(sv.urlPrefix+"/class/invitation/generate", sv.onHttpGenerateClassInvitationToken)
	http.HandleFunc(sv.urlPrefix+"/class/invitation/query", sv.onHttpQueryClassInvitationToken)
	http.HandleFunc(sv.urlPrefix+"/class/experience/query", sv.onHttpQueryExperienceUser)

	//----------------------------------------------------
	// Class-User related.

	http.HandleFunc(sv.urlPrefix+"/class/teacher/add", sv.onHttpClassAddTeacher)
	http.HandleFunc(sv.urlPrefix+"/class/teacher/delete", sv.onHttpClassDeleteTeacher)
	http.HandleFunc(sv.urlPrefix+"/class/teacher/query", sv.onHttpClassQueryTeacher)

	http.HandleFunc(sv.urlPrefix+"/class/student/add", sv.onHttpClassAddStudent)
	http.HandleFunc(sv.urlPrefix+"/class/student/delete", sv.onHttpClassDeleteStudent)
	http.HandleFunc(sv.urlPrefix+"/class/student/query", sv.onHttpClassQueryStudent)
	http.HandleFunc(sv.urlPrefix+"/class/student/remark", sv.onHttpClassRemarkStudent)

	http.HandleFunc(sv.urlPrefix+"/class/gdStudent/add", sv.onHttpClassAddGdStudent)
	http.HandleFunc(sv.urlPrefix+"/class/gdStudent/delete", sv.onHttpClassDeleteGdStudent)

	http.HandleFunc(sv.urlPrefix+"/class/keeper/student/associate", sv.onHttpAssociateKeeperWithStudent)
	http.HandleFunc(sv.urlPrefix+"/class/keeper/student/disassociate", sv.onHttpDisassociateKeeperWithStudent)

	//----------------------------------------------------
	// Issue related.

	http.HandleFunc(sv.urlPrefix+"/issue/get", sv.onHttpGetIssues)
	http.HandleFunc(sv.urlPrefix+"/issue/ask", sv.onHttpAddIssue)
	http.HandleFunc(sv.urlPrefix+"/issue/answer", sv.onHttpAnswerIssue)
	http.HandleFunc(sv.urlPrefix+"/issue/answer/change", sv.onHttpChangeIssueAnswer)
	// To be deleted:
	http.HandleFunc(sv.urlPrefix+"/class/issue/get", sv.onHttpGetIssues)
	http.HandleFunc(sv.urlPrefix+"/class/issue/ask", sv.onHttpAddIssue)
	http.HandleFunc(sv.urlPrefix+"/class/issue/answer", sv.onHttpAnswerIssue)
	http.HandleFunc(sv.urlPrefix+"/class/issue/answer/change", sv.onHttpChangeIssueAnswer)

	http.HandleFunc(sv.urlPrefix+"/issue/resource/get", sv.onHttpGetIssueResource)
	http.HandleFunc(sv.urlPrefix+"/issue/question/get", sv.onHttpGetIssueQuestion)

	//----------------------------------------------------

	// http.HandleFunc(sv.urlPrefix+"/meeting/name/change", sv.onHttpChangeMeetingName)
	// http.HandleFunc(sv.urlPrefix+"/meeting/subject/change", sv.onHttpChangeMeetingSubject)
	// http.HandleFunc(sv.urlPrefix+"/meeting/time/change", sv.onHttpChangeMeetingTime)
	// http.HandleFunc(sv.urlPrefix+"/meeting/type/change", sv.onHttpChangeMeetingType)
	// http.HandleFunc(sv.urlPrefix+"/meeting/section/change", sv.onHttpChangeMeetingSection)

	http.HandleFunc(sv.urlPrefix+"/meeting/join", sv.onHttpJoinMeeting)
	http.HandleFunc(sv.urlPrefix+"/meeting/leave", sv.onHttpLeaveMeeting)
	http.HandleFunc(sv.urlPrefix+"/meeting/score", sv.onHttpScoreMeeting)
	http.HandleFunc(sv.urlPrefix+"/meeting/feedback/get", sv.onHttpGetMeetingFeedback)
	http.HandleFunc(sv.urlPrefix+"/meeting/finish", sv.onHttpFinishMeeting)

	http.HandleFunc(sv.urlPrefix+"/meeting/progress/query", sv.onHttpQueryMeetingProgress)

	// http.HandleFunc(sv.urlPrefix+"/meeting/teacher/add", sv.onHttpMeetingAddTeacher)
	// http.HandleFunc(sv.urlPrefix+"/meeting/teacher/delete", sv.onHttpMeetingDeleteTeacher)

	// http.HandleFunc(sv.urlPrefix+"/meeting/student/notify", sv.onHttpMeetingNotifyStudent)

	//----------------------------------------------------

	http.HandleFunc(sv.urlPrefix+"/meeting/config/set", sv.onHttpSetMeetingConfig)
	http.HandleFunc(sv.urlPrefix+"/meeting/config/get", sv.onHttpGetMeetingConfig)

	http.HandleFunc(sv.urlPrefix+"/meeting/sync", sv.onHttpSyncMeeting)
	http.HandleFunc(sv.urlPrefix+"/meeting/get", sv.onHttpGetMeeting)

	http.HandleFunc(sv.urlPrefix+"/meeting/ally/set", sv.onHttpSetMeetingAlly)

	//----------------------------------------------------
	// Meeting-Resource related.

	http.HandleFunc(sv.urlPrefix+"/meeting/resource/arrange", sv.onHttpArrangeMeetingResources)

	http.HandleFunc(sv.urlPrefix+"/meeting/slides/add", sv.onHttpAddCourseware)
	http.HandleFunc(sv.urlPrefix+"/meeting/slides/delete", sv.onHttpDeleteCourseware)
	http.HandleFunc(sv.urlPrefix+"/meeting/slides/finish", sv.onHttpFinishCourseware)
	// To be deleted:
	http.HandleFunc(sv.urlPrefix+"/meeting/courseware/add", sv.onHttpAddCourseware)
	http.HandleFunc(sv.urlPrefix+"/meeting/courseware/delete", sv.onHttpDeleteCourseware)
	http.HandleFunc(sv.urlPrefix+"/meeting/courseware/finish", sv.onHttpFinishCourseware)

	http.HandleFunc(sv.urlPrefix+"/meeting/video/add", sv.onHttpAddVideo)
	http.HandleFunc(sv.urlPrefix+"/meeting/video/delete", sv.onHttpDeleteVideo)
	http.HandleFunc(sv.urlPrefix+"/meeting/video/finish", sv.onHttpFinishVideo)
	http.HandleFunc(sv.urlPrefix+"/meeting/video/authorize", sv.onHttpAuthorizeVideos)

	http.HandleFunc(sv.urlPrefix+"/meeting/replay/add", sv.onHttpAddReplay)
	http.HandleFunc(sv.urlPrefix+"/meeting/replay/delete", sv.onHttpDeleteReplay)
	http.HandleFunc(sv.urlPrefix+"/meeting/replay/finish", sv.onHttpFinishReplay)
	http.HandleFunc(sv.urlPrefix+"/meeting/replay/authorize", sv.onHttpAuthorizeReplays)

	http.HandleFunc(sv.urlPrefix+"/meeting/exam/add", sv.onHttpAddExamToMeeting)
	http.HandleFunc(sv.urlPrefix+"/meeting/exam/resync", sv.onHttpResyncExam)
	http.HandleFunc(sv.urlPrefix+"/meeting/exam/delete", sv.onHttpDeleteExamFromMeeting)
	http.HandleFunc(sv.urlPrefix+"/meeting/exam/answer", sv.onHttpAnswerMeetingExam)
	http.HandleFunc(sv.urlPrefix+"/meeting/exam/answer/get", sv.onHttpGetMeetingExamAnswer)
	http.HandleFunc(sv.urlPrefix+"/meeting/exam/question/answer", sv.onHttpAnswerMeetingExamQuestion)
	http.HandleFunc(sv.urlPrefix+"/meeting/exam/authorize", sv.onHttpAuhtorizeMeetingExam)

	//----------------------------------------------------

	http.HandleFunc(sv.urlPrefix+"/collection/question/add", sv.onHttpAddQuestionToCollection)
	http.HandleFunc(sv.urlPrefix+"/collection/question/delete", sv.onHttpDeleteQuestionFromCollection)

	//----------------------------------------------------

	http.HandleFunc(sv.urlPrefix+"/note/add", sv.onHttpAddNote)
	http.HandleFunc(sv.urlPrefix+"/note/delete", sv.onHttpAddNote)
	http.HandleFunc(sv.urlPrefix+"/note/get", sv.onHttpGetNote)
}

//----------------------------------------------------------------------------
