package client

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"piscine-golang-interact/record"
)

// SubjectNumMap 은 sid에 따른 Subject 이름을 찾는 map 이다.
var SubjectNumMap map[int]string

// SubjectStrMap 은 Subject 이름에 따른 sid를 찾는 map 이다.
var SubjectStrMap map[string]int

// IntervieweeList 는 피평가자의 uid 를 이용하는 Queue 이다.
var IntervieweeList []string

// InterviewerList 는 평가자의 uid 를 이용하는 Queue 이다.
var InterviewerList []string

// IntervieweeMutex 는 피평가자 Queue를 조작하는 Mutex 이다.
var IntervieweeMutex sync.Mutex

// InterviewerMutex 는 평가자 Queue를 조작하는 Mutex 이다.
var InterviewerMutex sync.Mutex

func init() {
	SubjectNumMap = map[int]string{0: "Day00", 1: "Day01", 2: "Day02", 3: "Day03", 4: "Day04", 5: "Day05", 100: "Rush00"}
	SubjectStrMap = map[string]int{"Day00": 0, "Day01": 1, "Day02": 2, "Day03": 3, "Day04": 4, "Day05": 5, "Rush00": 100}
	IntervieweeList = make([]string, 0, 100)
	InterviewerList = make([]string, 0, 100)
	IntervieweeMutex = sync.Mutex{}
	InterviewerMutex = sync.Mutex{}
}

func removeClient(list []string, i int) []string {
	return append(list[:i], list[i+1:]...)
}

// SubjectInfo 구조체는 서브젝트 관련 정보들을 담고 있는 구조체이다.
type SubjectInfo struct {
	// SubjectName 는 Subject 의 이름이다.
	// SubjectURL 는 해당 서브젝트의 공식 문서 url 이다.
	// EvalGuideURL 은 해당 서브젝트 평가표의 url 이다.
	SubjectName  string
	SubjectURL   string
	EvalGuideURL string
}

// MatchInfo 구조체는 평가 매칭이 성공했을 때 전달하는 평가 정보 구조체이다.
type MatchInfo struct {
	// Code 는 매칭 성공시 true, 매칭 취소시 false 이다.
	// InterviewerID 는 평가자의 uid 이다.
	// IntervieweeID 는 피평가자의 uid 이다.
	Code          bool
	InterviewerID string
	IntervieweeID string
	SubjectInfo
}

// Client 구조체는 Piscine Golang 서브젝트의 평가 매칭을 관리하는 오브젝트이다.
type Client struct {
	// MatchMap 은 uid 를 key 로 하여,
	// 해당 유저가 매칭 성공시에 상대의 uid 를 받기 위한 채널을 value 로 한다.
	MatchMap map[string]chan MatchInfo
}

// NewClient 함수는 Client 구조체의 생성자이다.
func NewClient() (ret *Client) {
	ret = &Client{}
	ret.MatchMap = make(map[string]chan MatchInfo)
	return ret
}

// SignUp 함수는 uid(userID) intraID를 받아 DB 에 추가하는 함수이다.
// DB 에 추가하기 전에 기존에 가입된 intraID 라면 가입이 되지 않는다.
func (c *Client) SignUp(uid, name string) (msg string) {
	tx, tErr := record.DB.Begin()
	if tErr != nil {
		return "가입오류: 트랜잭션 초기화"
	}
	defer tx.Rollback()
	if ret, qErr := tx.Query(`SELECT id FROM people WHERE name = $1 ;`, name); ret != nil {
		if qErr == nil {
			return "가입오류: 쿼리 실패"
		}
		if _, eErr := tx.Exec(`INSERT INTO people ( name, password ) VALUES ( ?, ? ) ;`, name, uid); eErr != nil {
			return "가입오류: 생성 실패"
		}
	} else {
		return "가입오류: 기존 사용자"
	}
	tErr = tx.Commit()
	if tErr != nil {
		return "가입오류: 트랜잭션 적용"
	} else {
		return "가입 완료"
	}
}

func (c *Client) ModifyId(uid, name string) (msg string) {
	tx, tErr := record.DB.Begin()
	if tErr != nil {
		return "가입오류: 트랜잭션 초기화"
	}
	defer tx.Rollback()
	if ret, qErr := tx.Query(`SELECT id FROM people WHERE name = $1 ;`, name); ret != nil {
		if qErr == nil {
			return "가입오류: 쿼리 실패"
		}
		if _, eErr := tx.Exec(`INSERT INTO people ( name, password ) VALUES ( ?, ? ) ;`, name, uid); eErr != nil {
			return "가입오류: 생성 실패"
		}
	} else {
		return "가입오류: 기존 사용자"
	}
	tErr = tx.Commit()
	if tErr != nil {
		return "가입오류: 트랜잭션 적용"
	} else {
		return "가입 완료"
	}
}

// Submit 함수는 sid(subject id) uid(userID) url(github repo link)와
// 매칭된 상대방의 UID 를 공유할 matchedUserId channel 을 인자로 받아
// 서브젝트 제출을 수행하고 작업이 성공적으로 이루어졌는지 여부를 알리는 msg 를 반환하는 함수이다.
// Eval Queue 에 사용자가 있는지 Mutex 를 걸고 확인한 후에 있다면 매칭을 진행해야한다. ** MUTEX 활용 필수!!
func (c *Client) Submit(sid, uid, url string, matchedUserId chan MatchInfo) (msg string) {
	// convertID := SubjectStrMap[sid]
	// 평가 포인트 및 매치 기록
	InterviewerMutex.Lock()
	defer InterviewerMutex.Lock()
	if len(InterviewerList) == 0 {
		IntervieweeMutex.Lock()
		IntervieweeList = append(IntervieweeList, uid)
		IntervieweeMutex.Unlock()
		return "제출완료"
	} else {
		// send to each other
		InterviewerList = removeClient(InterviewerList, 0)
		return "제출완료"
	}
}

// SubmitCancel 함수는 uid 를 인자로 받아 해당 유저의 제출을 취소하는 함수이다.
// 제출 취소의 성공/실패 여부를 msg 로 리턴한다.
func (c *Client) SubmitCancel(uid string) (msg string) {
	// 디비 기록 확인
	IntervieweeMutex.Lock()
	defer IntervieweeMutex.Unlock()
	for i, v := range IntervieweeList {
		if v == uid {
			removeClient(IntervieweeList, i)
			return "취소완료"
		}
	}
	return "취소오류"
}

// Register 함수는 uid 와 매칭된 상대방의 UID 를 공유할 matchedUserId channel 을 인자로 받아
// 평가 등록을 수행하고 작업이 성공적으로 이루어졌는지 여부를 알리는 msg 를 반환하는 함수이다.
// Submit Queue 에 사용자가 있는지 Mutex 를 걸고 확인한 후에 있다면 매칭을 진행해야한다. ** MUTEX 활용 필수!!
func (c *Client) Register(uid string, matchedUid chan MatchInfo) (msg string) {
	// 평가 포인트 및 매치 기록
	IntervieweeMutex.Lock()
	defer IntervieweeMutex.Unlock()
	if len(IntervieweeList) == 0 {
		InterviewerMutex.Lock()
		InterviewerList = append(InterviewerList, uid)
		InterviewerMutex.Unlock()
		return "평가등록완료"
	} else {
		// send to each other
		IntervieweeList = removeClient(IntervieweeList, 0)
		return "평가등록완료"
	}
}

// RegisterCancel 함수는 uid 를 인자로 받아 해당 유저의 평가 등록을 취소하는 함수이다.
// 평가 등록 취소의 성공/실패 여부를 msg 로 리턴한다.
func (c *Client) RegisterCancel(uid string) (msg string) {
	// 디비 기록 확인
	InterviewerMutex.Lock()
	defer InterviewerMutex.Unlock()
	for i, v := range InterviewerList {
		if v == uid {
			removeClient(InterviewerList, i)
			return "평가취소완료"
		}
	}
	return "평가취소오류"
}

// MyGrade 함수는 uid 를 인자로 받아 해당 유저의 점수 정보를 리턴하는 함수이다.
func (c *Client) MyGrade(uid string) (grades EmbedInfo) {
	grades.title = "서브젝트 채점 현황"
	tx, tErr := record.DB.Begin()
	if tErr != nil {
		return
	}
	defer tx.Rollback()
	if rows, qErr := tx.Query(`SELECT e.course, e.score, e.pass, e.updated_at FROM evaluation AS e JOIN people AS p ON e.interviewee_id = p.id WHERE p.password = $1 ;`, uid); qErr != nil {
		return
	} else {
		var course int
		var score int
		var pass bool
		var stamp time.Time
		for rows.Next() {
			if sErr := rows.Scan(&course, &score, &pass, &stamp); sErr != nil {
				return
			}
			tempLines := make([]string, 0, 3)
			tempLines = append(tempLines, "Score: "+strconv.Itoa(score))
			if pass {
				tempLines = append(tempLines, "PASS")
			} else {
				tempLines = append(tempLines, "FAIL")
			}
			time := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d\n", stamp.Year(), stamp.Month(), stamp.Day(), stamp.Hour(), stamp.Minute(), stamp.Second())
			tempLines = append(tempLines, "Time: "+time)
			grades.embedRows = append(grades.embedRows, EmbedRow{name: SubjectNumMap[course], lines: tempLines})
		}
		rows.Close()
	}
	_ = tx.Commit()
	return
}

func matchEmbedRow(m *sync.Mutex, s string, p *map[string]string, l *[]string) EmbedRow {
	tempEmbedRow := EmbedRow{name: s}
	tempLines := make([]string, 0, 100)
	m.Lock()
	for _, v := range *l {
		if i, ok := (*p)[v]; ok {
			tempLines = append(tempLines, i)
		}
	}
	m.Unlock()
	if len(tempLines) == 0 {
		tempLines = append(tempLines, "대기열 없음")
	}
	tempEmbedRow.lines = tempLines
	return tempEmbedRow
}

// MatchState 함수는 uid 를 인자로 받아 현재 큐 정보를 리턴하는 함수이다.
func (c *Client) MatchState() (grades EmbedInfo) {
	grades.title = "평가 및 피평가 매칭 현황"
	people := make(map[string]string)
	tx, tErr := record.DB.Begin()
	if tErr != nil {
		return
	}
	defer tx.Rollback()
	if rows, qErr := tx.Query(`SELECT name, password FROM people`); qErr != nil {
		return
	} else {
		var name string
		var pass string
		for rows.Next() {
			if sErr := rows.Scan(&name, &pass); sErr != nil {
				return
			}
			people[pass] = name
		}
		rows.Close()
	}
	tErr = tx.Commit()
	if tErr != nil {
		return
	} else {
		grades.embedRows = append(grades.embedRows, matchEmbedRow(&InterviewerMutex, "평가자", &people, &InterviewerList))
		grades.embedRows = append(grades.embedRows, matchEmbedRow(&IntervieweeMutex, "피평가자", &people, &IntervieweeList))
		return
	}
}

// FindIntraByUID 함수는 uid 를 인자로 받아 intraID 를 반환하는 함수이다.
func (c *Client) FindIntraByUID(uid string) (intraID string) {
	tx, tErr := record.DB.Begin()
	if tErr != nil {
		return "트랜잭션 초기화 오류"
	}
	defer tx.Rollback()
	if rows, qErr := tx.Query(`SELECT name FROM people WHERE password = $1 ;`, uid); qErr != nil {
		return "가입되지 않은 사용자"
	} else {
		for rows.Next() {
			if sErr := rows.Scan(&intraID); sErr != nil {
				return "잘못된 참조"
			}
		}
		rows.Close()
	}
	tErr = tx.Commit()
	if tErr != nil {
		return "트랜잭션 적용 오류"
	} else {
		return
	}
}
