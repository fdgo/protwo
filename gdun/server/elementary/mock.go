package elementary

// import (
// 	"gitlab.hfjy.com/gdun/common"
// 	"gitlab.hfjy.com/gdun/service"
// 	"net/http"
// 	"sort"
// 	"strconv"
// )

//----------------------------------------------------------------------------

// func (es *ElementaryServer) onHttpAnswerFreeMock(w http.ResponseWriter, r *http.Request) {
// 	if es.cache == nil {
// 		es.Send(w, r, -1, "Service unavailable.", "")
// 		return
// 	}

// 	err := r.ParseForm()
// 	if err != nil {
// 		es.Send(w, r, -2, "Invalid form.", "")
// 		return
// 	}

// 	id, err := strconv.Atoi(r.FormValue(common.FIELD_ID))
// 	if err != nil {
// 		es.Send(w, r, -3, "Invalid user ID.", "")
// 		return
// 	}

// 	answer := r.FormValue(common.FIELD_ANSWER)
// 	if len(answer) == 0 {
// 		es.Send(w, r, -4, "Empty answer.", "")
// 		return
// 	}

// 	value := common.GetTimeString() + ":" + answer + ":-1:-1"
// 	err = es.cache.SetField("mock", strconv.Itoa(id), value)
// 	if err != nil {
// 		es.Send(w, r, -5, "Failed to save your answer.", "")
// 		return
// 	}

// 	es.Send(w, r, 0, "", "")
// }

//----------------------------------------------------------------------------

// func (es *ElementaryServer) onHttpGetMyFreeMockResult(w http.ResponseWriter, r *http.Request) {
// 	if es.cache == nil {
// 		es.Send(w, r, -1, "Service unavailable.", "")
// 		return
// 	}

// 	err := r.ParseForm()
// 	if err != nil {
// 		es.Send(w, r, -2, "Invalid form.", "")
// 		return
// 	}

// 	id, err := strconv.Atoi(r.FormValue(common.FIELD_ID))
// 	if err != nil {
// 		es.Send(w, r, -3, "Invalid user ID.", "")
// 		return
// 	}

// 	value, err := es.cache.GetField("mock", strconv.Itoa(id))
// 	if err != nil {
// 		es.Send(w, r, -4, "Failed to get your answer.", "")
// 		return
// 	}

// 	er := service.NewExamResultFromString(value)
// 	if er == nil {
// 		es.Send(w, r, -5, "Failed to get your answer.", "")
// 		return
// 	}

// 	es.Send(w, r, 0, "", er.ToJSON())
// }

//----------------------------------------------------------------------------

// func (es *ElementaryServer) onHttpGetAllFreeMockResults(w http.ResponseWriter, r *http.Request) {
// 	if es.cache == nil {
// 		es.Send(w, r, -1, "Service unavailable.", "")
// 		return
// 	}

// 	result, err := es.cache.GetKey("mock:rank")
// 	if err != nil {
// 		es.Send(w, r, -2, err.Error(), "")
// 		return
// 	}

// 	es.Send(w, r, 0, "", result)
// }

//----------------------------------------------------------------------------

// func (es *ElementaryServer) onHttpReviewFreeMockResults(w http.ResponseWriter, r *http.Request) {
// 	if es.cache == nil {
// 		es.Send(w, r, -1, "Service unavailable.", "")
// 		return
// 	}

// 	result, err := es.toBeDeleted()
// 	if err != nil {
// 		es.Send(w, r, -2, err.Error(), "")
// 		return
// 	}

// 	es.Send(w, r, 0, "", result)
// }

//----------------------------------------------------------------------------

// func (es *ElementaryServer) toBeDeleted() (string, error) {

// 	// Get all answers
// 	m, err := es.cache.GetAllFields("mock")
// 	if err != nil {
// 		return "", err
// 	}

// 	// Get the correct answer.
// 	ans, err := es.cache.GetKey("mock:correct")
// 	if err != nil {
// 		return "", err
// 	}
// 	n := len(ans)
// 	if n == 0 {
// 		return "", common.ERR_NO_EXAM
// 	}

// 	arr := make([]*service.ExamResult, len(m))
// 	i := 0
// 	for k, v := range m {
// 		userID, err := strconv.Atoi(k)
// 		if err != nil {
// 			continue
// 		}
// 		if userID <= 0 {
// 			continue
// 		}

// 		// Construct an exam result.
// 		er := service.NewExamResultFromString(v)
// 		if er == nil {
// 			continue
// 		}
// 		er.UserID = userID

// 		// Compute the number of correct answers.
// 		cnt := 0
// 		for j := 0; (j < n) && (j < len(er.Answer)); j += 2 {
// 			if ans[j] != er.Answer[j] || ans[j+1] != er.Answer[j+1] {
// 				cnt++
// 			}
// 		}
// 		er.Rank = cnt

// 		// Save the result for this student.
// 		arr[i] = er
// 		i++
// 	}

// 	// Sort the results.
// 	sort.Sort(service.ExamResultSlice(arr))

// 	// Write them back to the cache.
// 	n = i
// 	r := `"` + common.FIELD_RANK + `":[`
// 	first := true
// 	for i = 0; i < n; i++ {
// 		arr[i].Rank = i + 1
// 		arr[i].Count = n

// 		err = es.cache.SetField(
// 			"mock",
// 			strconv.Itoa(arr[i].UserID),
// 			strconv.Itoa(arr[i].UpdateTime)+":"+arr[i].Answer+":"+strconv.Itoa(arr[i].Rank)+":"+strconv.Itoa(arr[i].Count))

// 		if err != nil {
// 			return "", err
// 		}

// 		if first {
// 			first = false
// 		} else {
// 			r += `,`
// 		}
// 		r += `{"` + common.FIELD_USER + `":` + strconv.Itoa(arr[i].UserID) + `,"` + common.FIELD_UPDATE_TIME + `":` + strconv.Itoa(arr[i].UpdateTime) + `000,"` + common.FIELD_ANSWER + `":"` + arr[i].Answer + `"}`
// 	}
// 	r += `]`

// 	err = es.cache.SetKey("mock:rank", r)
// 	if err != nil {
// 		return "", err
// 	}

// 	return r, nil
// }

//----------------------------------------------------------------------------
