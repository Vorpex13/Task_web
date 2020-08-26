package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

type Group struct {
	GroupName        string `json:"group_name"`
	GroupDescription string `json:"group_description"`
	Id               int    `json:"group_id"`
	ParentId         int    `json:"parent_id"`
}
type ArrStruct struct {
	AllGroup []Group `json:"all_group"`
	AllTask  []Task  `json:"all_task"`
}

type Task struct {
	TaskId      string `json:"task_id"`
	GroupId     int    `json:"group_id"`
	TaskName    string `json:"task_name"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
	CompletedAt string `json:"completed_at"`
}

const filename = "task.json"

func main() {
	addgroup()
	addtask()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/groups", GetGroups).Methods(http.MethodGet)
	router.HandleFunc("/group/{id}", GetGroupId).Methods(http.MethodGet)
	router.HandleFunc("/group/child/{id}", GetChildId).Methods(http.MethodGet)
	router.HandleFunc("/group/top_parents", GetGroupParents).Methods(http.MethodGet)
	router.HandleFunc("/group/new", PostNewGroup).Methods(http.MethodPost)
	router.HandleFunc("/group/{id}", PutId).Methods(http.MethodPut)
	router.HandleFunc("/group/{id}", Delete).Methods(http.MethodDelete)
	router.HandleFunc("/tasks", Tasks).Methods(http.MethodGet)
	router.HandleFunc("/tasks/new", PostNewTasks).Methods(http.MethodPost)
	router.HandleFunc("/tasks/{id}", PutIdTask).Methods(http.MethodPut)
	router.HandleFunc("/tasks/group/{id}", GetCompleted).Methods(http.MethodGet)
	router.HandleFunc("/tasks/{id}", GetCompletedEdit).Methods(http.MethodPost)
	router.HandleFunc("/tasks/time", GetTime).Methods(http.MethodGet)
	log.Fatal(http.ListenAndServe(":8080", router))
}
func addgroup() {
	var CopyStruct ArrStruct
	CopyStruct.read()
	CopyStruct.AllGroup = append(CopyStruct.AllGroup, Group{
		GroupName:        "Groups",
		GroupDescription: "Groups created task",
		Id:               10,
		ParentId:         0,
	})
	write(CopyStruct)
}

func addtask() {
	var CopyStruct ArrStruct
	CopyStruct.read()
	var TaskTest Task
	TaskTest.TaskName = "Home work"
	TaskTest.GroupId = 5
	TaskTest.CreatedAt = Time(TaskTest.CreatedAt)
	var indefer int
	for i := 0; i < len(CopyStruct.AllGroup); i++ {
		if CopyStruct.AllGroup[i].Id == TaskTest.GroupId {
			indefer = i
			break
		}
	}
	TaskTest.TaskId = TaskIdIndefer(CopyStruct.AllGroup[indefer].GroupDescription, TaskTest.GroupId)
	CopyStruct.read()
	CopyStruct.AllTask = append(CopyStruct.AllTask, TaskTest)
	write(CopyStruct)
}

//TODO:Сортировка по имени.
func SortName(ArrCopy ArrStruct) ArrStruct {

	sort.SliceStable(ArrCopy.AllGroup, func(i, j int) bool {
		return ArrCopy.AllGroup[i].GroupName < ArrCopy.AllGroup[j].GroupName
	},
	)
	return ArrCopy
}

//TODO:Обрезка Массива.
func Limit(limit int, ArrCopy ArrStruct) ArrStruct {
	var ArrLimit ArrStruct
	if limit > len(ArrCopy.AllGroup) {
		limit = len(ArrCopy.AllGroup)
		ArrLimit.AllGroup = ArrCopy.AllGroup[0:limit]
	} else {
		ArrLimit.AllGroup = ArrCopy.AllGroup[0:limit]
	}
	return ArrLimit
}

//TODO:Чтение с файла.
func (cp *ArrStruct) read() {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(data, &cp); err != nil {
	}
}

//TODO:Запись в файл.
func write(CopyStruct ArrStruct) {
	data, err := json.Marshal(CopyStruct)
	ioutil.WriteFile(filename, data, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

//TODO:Размещение для вывода Отец-Сыновья.
func TopFamily(CopyStruct ArrStruct) ArrStruct {
	var ParentsArr, ChildrenArr []int
	for i := 0; i < len(CopyStruct.AllGroup); i++ {
		if CopyStruct.AllGroup[i].ParentId == 0 {
			ParentsArr = append(ParentsArr, CopyStruct.AllGroup[i].Id-1)
		} else {
			ChildrenArr = append(ChildrenArr, CopyStruct.AllGroup[i].Id-1)
		}
	}
	var MediumArr ArrStruct
	for i := 0; i < len(ParentsArr); i++ {
		MediumArr.AllGroup = append(MediumArr.AllGroup, CopyStruct.AllGroup[ParentsArr[i]])
		for j := 0; j < len(ChildrenArr); j++ {
			if CopyStruct.AllGroup[ChildrenArr[j]].ParentId == CopyStruct.AllGroup[ParentsArr[i]].Id {
				MediumArr.AllGroup = append(MediumArr.AllGroup, CopyStruct.AllGroup[ChildrenArr[j]])
			}
		}
	}
	return MediumArr
}

//TODO:Размещение для вывода Топ отцы и их сыновья.
func TopParents(CopyStruct ArrStruct) ArrStruct {
	var ParentsArr, ChildrenArr ArrStruct
	for i := 0; i < len(CopyStruct.AllGroup); i++ {
		if CopyStruct.AllGroup[i].ParentId == 0 {
			ParentsArr.AllGroup = append(ParentsArr.AllGroup, CopyStruct.AllGroup[i])
		} else {
			ChildrenArr.AllGroup = append(ChildrenArr.AllGroup, CopyStruct.AllGroup[i])
		}

	}
	ParentsArr = SortName(ParentsArr)
	ChildrenArr = SortName(ChildrenArr)
	ChildrenArr.AllGroup = append(ParentsArr.AllGroup, ChildrenArr.AllGroup...)
	return ChildrenArr
}

//TODO:Функция ссылки Группы.
func GetGroups(w http.ResponseWriter, r *http.Request) {
	//Data parsing from URL and request
	filter :=
		r.FormValue("filter")
	limit := r.FormValue("limit")
	//Make copy of data
	var CopyStruct ArrStruct
	CopyStruct.read()
	//Checking of params of sorting
	switch filter {
	case "sname":
		{
			CopyStruct = SortName(CopyStruct)
		}
	case "parent_with_childs":
		{
			CopyStruct = TopFamily(CopyStruct)
		}
	case

		"parents_first":
		{
			CopyStruct = TopParents(CopyStruct)
		}
	}

	//checking limit
	if limit != "" {
		n, err := strconv.ParseInt(limit, 10, 0)
		if err != nil {
			fmt.Printf("%d of type%T", n, n)
		} else {
			CopyStruct = Limit(int(n), CopyStruct)
		}
	}
	//response
	h, err := json.MarshalIndent(CopyStruct, "", "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(w, string(h))
}

//TODO:Сначало выводим отцов, потом детей
func UpperParents(CopyStruct ArrStruct) ArrStruct {
	var ParentsArr ArrStruct
	for i := 0; i < len(CopyStruct.AllGroup); i++ {
		if CopyStruct.AllGroup[i].ParentId == 0 {
			ParentsArr.AllGroup = append(ParentsArr.AllGroup, CopyStruct.AllGroup[i])
		}
	}

	return ParentsArr
}

//TODO:Функция ссылки Отцы.
func GetGroupParents(w http.ResponseWriter, r *http.Request) {
	var CopyStruct ArrStruct
	CopyStruct.read()
	//filter := r.FormValue("filter")
	limit := r.FormValue("limit")
	CopyStruct = UpperParents(CopyStruct)
	if limit != "" {
		n, err := strconv.ParseInt(limit, 10, 0)
		if err != nil {
			fmt.Printf("%d of type%T", n, n)
		} else {
			CopyStruct = Limit(int(n), CopyStruct)
		}
	}
	h, err := json.MarshalIndent(CopyStruct, "", "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(w, string(h))
}

//TODO:Функция ID группы.
func GetGroupId(w http.ResponseWriter, r *http.Request) {
	var b Group
	var CopyStruct ArrStruct
	CopyStruct.read()
	id := mux.Vars(r)["id"]
	bFound := false
	if id != "" {
		n, err := strconv.Atoi(id)
		if err != nil {
			fmt.Printf("%d of type%T", n, n)
		}
		for i := 0; i < len(CopyStruct.AllGroup); i++ {
			if CopyStruct.AllGroup[i].Id == n {
				b = CopyStruct.AllGroup[i]
				bFound = true
			}
		}

		if bFound {
			h, err := json.MarshalIndent(b, "", "")
			if err != nil {
				log.Fatal(err)
			}

			fmt.Fprint(w, string(h))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

//TODO:Функция ID группы и сыновьёв.
func GetChildId(w http.ResponseWriter, r *http.Request) {
	var b ArrStruct
	var CopyStruct ArrStruct
	CopyStruct.read()
	id := mux.Vars(r)["id"]
	if id != "" {
		n, err := strconv.Atoi(id)
		if err != nil {
			fmt.Printf("%d of type%T", n, n)
		}
		bFound := false

		for i := 0; i < len(CopyStruct.AllGroup); i++ {
			if CopyStruct.AllGroup[i].Id == n {
				bFound = true
				b.AllGroup = append(b.AllGroup, CopyStruct.AllGroup[i])
				for j := 0; j < len(CopyStruct.AllGroup); j++ {
					if CopyStruct.AllGroup[i].ParentId == i {
						b.AllGroup = append(b.AllGroup, CopyStruct.AllGroup[j])
					}
				}
			}
		}
		if !bFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h, err := json.MarshalIndent(b, "", "")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(w, string(h))
	}
}

//TODO:Создание нового элемента.
func PostNewGroup(w http.ResponseWriter, r *http.Request) {
	var CopyStruct ArrStruct
	CopyStruct.read()
	var CopyBody Group
	err := json.NewDecoder(r.Body).Decode(&CopyBody)
	if err != nil {
		fmt.Println("Cannot decode from request's json to required type.")
		w.WriteHeader(http.StatusBadRequest)
	}
	if CopyBody.GroupName == "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	sort.SliceStable(CopyStruct.AllGroup, func(i, j int) bool {
		return CopyStruct.AllGroup[i].Id < CopyStruct.AllGroup[j].Id
	})
	CopyBody.Id = CopyStruct.AllGroup[len(CopyStruct.AllGroup)-1].Id + 1
	CopyStruct.AllGroup = append(CopyStruct.AllGroup, CopyBody)
	write(CopyStruct)
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(CopyStruct.AllGroup)
	w.WriteHeader(http.StatusCreated)

}

//TODO:Обновление по ID.
func PutId(w http.ResponseWriter, r *http.Request) {
	var CopyStruct ArrStruct
	CopyStruct.read()
	id := mux.Vars(r)["id"]
	n, err := strconv.Atoi(id)
	if err != nil {
		fmt.Printf("%d of type%T", n, n)
	}
	bFound := false
	for i := 0; i < len(CopyStruct.AllGroup); i++ {
		if CopyStruct.AllGroup[i].Id == n {
			bFound = true
		}
	}
	if !bFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var CopyBody Group
	err = json.NewDecoder(r.Body).Decode(&CopyBody)
	if err != nil {
		fmt.Println("Cannot decode from request's json to required type.")
		w.WriteHeader(http.StatusBadRequest)
	}
	if CopyBody.GroupName == "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	CopyBody.Id = n
	CopyStruct.AllGroup[n] = CopyBody
	write(CopyStruct)
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(CopyBody)
	w.WriteHeader(http.StatusCreated)
}

//TODO:Уаление элемента
func Delete(w http.ResponseWriter, r *http.Request) {
	var CopyStruct ArrStruct
	CopyStruct.read()
	id := mux.Vars(r)["id"]
	n, err := strconv.Atoi(id)
	if err != nil {
		fmt.Printf("%d of type%T", n, n)
	}
	bFound := false
	for i := 0; i < len(CopyStruct.AllGroup); i++ {
		if CopyStruct.AllGroup[i].ParentId == n {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if n == CopyStruct.AllGroup[i].Id {
			n = i
			bFound = true
			break
		}
	}

	if bFound {
		CopyStruct.AllGroup = append(CopyStruct.AllGroup[:n], CopyStruct.AllGroup[n+1:]...)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	write(CopyStruct)
	w.WriteHeader(http.StatusAccepted)
}

//TODO:Сортировка тасков по группе.
func SortTasksByGroup(ArrCopy ArrStruct) ArrStruct {

	sort.SliceStable(ArrCopy.AllTask, func(i, j int) bool {
		return ArrCopy.AllTask[i].GroupId < ArrCopy.AllTask[j].GroupId
	},
	)
	return ArrCopy
}

//TODO:Сортирова тасков по имени.
func SortNameTasks(ArrCopy ArrStruct) ArrStruct {

	sort.SliceStable(ArrCopy.AllTask, func(i, j int) bool {
		return ArrCopy.AllTask[i].TaskName < ArrCopy.AllTask[j].TaskName
	},
	)
	return ArrCopy
}

//TODO:Обрезка Массива тасков.
func LimitTasks(limit int, ArrCopy ArrStruct) ArrStruct {
	var ArrLimit ArrStruct
	if limit > len(ArrCopy.AllTask) {
		limit = len(ArrCopy.AllTask)
		ArrLimit.AllTask = ArrCopy.AllTask[0:limit]
	} else {
		ArrLimit.AllTask = ArrCopy.AllTask[0:limit]
	}
	return ArrLimit
}

//TODO:Сортировка по флагу выполнения.
func SortCompleted(flag bool, CopyStruct ArrStruct) {
	var sort ArrStruct
	for i := 0; i < len(CopyStruct.AllTask); i++ {
		if CopyStruct.AllTask[i].Completed == flag {

			sort.AllTask = append(sort.AllTask, CopyStruct.AllTask[i])
		}
	}
}

//TODO:Функция вызова Тасков.
func Tasks(w http.ResponseWriter, r *http.Request) {
	var CopyStruct ArrStruct
	CopyStruct.read()
	sort := r.FormValue("sort")
	lim := r.FormValue("limit")
	state := r.FormValue("state")
	switch sort {
	case "name":
		{
			SortNameTasks(CopyStruct)
		}
	case "group":
		{
			SortTasksByGroup(CopyStruct)
		}
	}
	if lim != "" {
		n, err := strconv.ParseInt(lim, 10, 0)
		if err != nil {
			fmt.Printf("%d of type%T", n, n)
		} else {
			CopyStruct = LimitTasks(int(n), CopyStruct)
		}
	}
	if state != "" {
		switch state {
		case "completed":
			{
				SortCompleted(false, CopyStruct)
			}
		case "working":
			{
				SortCompleted(true, CopyStruct)
			}
		}
	}
	h, err := json.MarshalIndent(CopyStruct, "", "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(w, string(h))
}

//TODO:Определение Id.
func TaskIdIndefer(task string, id int) (str string) {
	task += strconv.Itoa(id)
	hsh := md5.Sum([]byte(task))
	str = fmt.Sprintf("%x", hsh)
	return str[:6]
}

//TODO:Определение времени.
func Time(CopyStruct string) string {
	CopyStruct = time.Now().Format("2006-01-02T15:04:05")
	return CopyStruct
}

//TODO:Создание нового элемента.
func PostNewTasks(w http.ResponseWriter, r *http.Request) {
	var CopyStruct ArrStruct
	CopyStruct.read()
	var CopyBody Task
	err := json.NewDecoder(r.Body).Decode(&CopyBody)
	if err != nil {
		fmt.Println("Cannot decode from request's json to required type.")
		w.WriteHeader(http.StatusBadRequest)
	}
	if CopyBody.TaskName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if CopyBody.GroupId == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	bFound := false
	for i := 0; i < len(CopyStruct.AllGroup); i++ {
		if CopyBody.GroupId == CopyStruct.AllGroup[i].Id {
			bFound = true
		}
	}
	if !bFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var indefer int
	for i := 0; i < len(CopyStruct.AllGroup); i++ {
		if CopyStruct.AllGroup[i].Id == CopyBody.GroupId {
			indefer = i
			break
		}
	}
	CopyBody.TaskId = TaskIdIndefer(CopyStruct.AllGroup[indefer].GroupDescription, CopyBody.GroupId)
	CopyStruct.AllTask = append(CopyStruct.AllTask, CopyBody)
	CopyBody.CreatedAt = Time(CopyBody.CreatedAt)

	write(CopyStruct)
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(CopyStruct.AllTask)
	w.WriteHeader(http.StatusCreated)
}

//TODO:Обновление Таска
func PutIdTask(w http.ResponseWriter, r *http.Request) {
	var CopyStruct ArrStruct
	var n int
	CopyStruct.read()
	id := mux.Vars(r)["id"]
	bFound := false
	for i := 0; i < len(CopyStruct.AllTask); i++ {
		if CopyStruct.AllTask[i].TaskId == id {
			bFound = true
			n = i
		}
	}
	if !bFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var CopyBody Task
	err := json.NewDecoder(r.Body).Decode(&CopyBody)
	if err != nil {
		fmt.Println("Cannot decode from request's json to required type.")
		w.WriteHeader(http.StatusBadRequest)
	}
	if CopyBody.TaskName == "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	CopyBody.CreatedAt = Time(CopyBody.CreatedAt)

	var indeaer int
	for i := 0; i < len(CopyStruct.AllGroup); i++ {
		if CopyStruct.AllGroup[i].Id == CopyBody.GroupId {
			indeaer = i
			break
		}
	}
	CopyBody.TaskId = TaskIdIndefer(CopyStruct.AllGroup[indeaer].GroupDescription, CopyBody.GroupId)
	CopyStruct.AllTask[n] = CopyBody
	write(CopyStruct)
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(CopyBody)
	w.WriteHeader(http.StatusCreated)
}

//TODO:Создание массива выполненных заказов.
func Grouping(CopyStruct ArrStruct, id int, b bool) ArrStruct {
	var Temporary ArrStruct
	for i := 0; i < len(CopyStruct.AllTask); i++ {
		if CopyStruct.AllTask[i].GroupId == id || CopyStruct.AllTask[i].Completed == b {
		}
		Temporary.AllTask = append(Temporary.AllTask, CopyStruct.AllTask[i])
	}
	return Temporary
}

//TODO:Вывод выполненных заданий.
func GetCompleted(w http.ResponseWriter, r *http.Request) {
	var CopyStruct ArrStruct
	CopyStruct.read()
	var n int
	id := mux.Vars(r)["id"]
	for i := 0; i < len(CopyStruct.AllTask); i++ {
		if id == CopyStruct.AllTask[i].TaskId {
			n = i
			break
		}
	}
	state := r.FormValue("state")
	switch state {
	case "completed":
		{
			CopyStruct = Grouping(CopyStruct, CopyStruct.AllTask[n].GroupId, true)
		}
	case "working":
		{
			CopyStruct = Grouping(CopyStruct, CopyStruct.AllTask[n].GroupId, false)
		}
		h, err := json.MarshalIndent(CopyStruct, "", "")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(w, string(h))
	}
}

//TODO:Изменение статуса выполнения.
func GetCompletedEdit(w http.ResponseWriter, r *http.Request) {
	var CopyStruct ArrStruct
	CopyStruct.read()
	var n int
	bFound := false
	state := r.FormValue("state")
	id := mux.Vars(r)["id"]

	for i := 0; i < len(CopyStruct.AllTask); i++ {
		if id == CopyStruct.AllTask[i].TaskId {
			n = i
			bFound = true
			break
		}
	}
	if !bFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	state2, err := strconv.ParseBool(state)
	if state2 == true {
		CopyStruct.AllTask[n].CompletedAt = Time(CopyStruct.AllTask[n].CompletedAt)
	}

	CopyStruct.AllTask[n].Completed = state2
	if err == nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	write(CopyStruct)
}

//TODO:Создание массива тасков по времени.
func TimeTaskArr(check int, CopyStruct []Task) []Task {
	var TimeArr []Task
	TimeProbs := time.Now().Format("2006-01-02T15:04:05")
	TimeDay, err := time.Parse("2006-01-02T15:04:05", TimeProbs)
	if err != nil {
		log.Fatal(err)
	}

	switch check {
	case 0:
		{
			for i := 0; i < len(CopyStruct); i++ {
				time, err := time.Parse("2006-01-02T15:04:05", CopyStruct[i].CreatedAt)
				if err != nil {
					log.Fatal(err)
				}
				YearSecond, MonthSecond, DaySecond := TimeDay.Date()
				year, month, day := time.Date()
				if YearSecond == year && MonthSecond == month && DaySecond == day {
					TimeArr = append(TimeArr, CopyStruct[i])
				}
			}
		}
	case 1:
		{
			TimeDayTwo := TimeDay
			TimeDay.AddDate(0, 0, -1)
			for i := 0; i < len(CopyStruct); i++ {
				time, err := time.Parse("2006-01-02T15:04:05", CopyStruct[i].CreatedAt)
				if err != nil {
					log.Fatal(err)
				}
				TimeDuration := TimeDay.Sub(TimeDayTwo)
				TimeDurationTwo := TimeDay.Sub(time)
				if TimeDuration > TimeDurationTwo {
					TimeArr = append(TimeArr, CopyStruct[i])
				}
			}
		}
	case 2:
		{
			TimeDayTwo := TimeDay
			TimeDay.AddDate(0, 0, -7)
			for i := 0; i < len(CopyStruct); i++ {
				time, err := time.Parse("2006-01-02T15:04:05", CopyStruct[i].CreatedAt)
				if err != nil {
					log.Fatal(err)
				}
				TimeDuration := TimeDay.Sub(TimeDayTwo)
				TimeDurationTwo := TimeDay.Sub(time)
				if TimeDuration > TimeDurationTwo {
					TimeArr = append(TimeArr, CopyStruct[i])
				}
			}
		}
	case 3:
		{
			TimeDayTwo := TimeDay
			TimeDay.AddDate(0, -1, 0)
			for i := 0; i < len(CopyStruct); i++ {
				time, err := time.Parse("2006-01-02T15:04:05", CopyStruct[i].CreatedAt)
				if err != nil {
					log.Fatal(err)
				}
				TimeDuration := TimeDay.Sub(TimeDayTwo)
				TimeDurationTwo := TimeDay.Sub(time)
				if TimeDuration > TimeDurationTwo {
					TimeArr = append(TimeArr, CopyStruct[i])
				}
			}
		}

	}

	return TimeArr
}

//TODO:Вывод тасков по времени.
func GetTime(w http.ResponseWriter, r *http.Request) {
	var CopyStruct ArrStruct
	CopyStruct.read()
	TimeArr := CopyStruct.AllTask
	filter := r.FormValue("filter")
	switch filter {
	case "yesteday":
		{
			TimeArr = TimeTaskArr(1, TimeArr)
		}
	case "day":
		{
			TimeArr = TimeTaskArr(0, TimeArr)
		}
	case "week":
		{
			TimeArr = TimeTaskArr(2, TimeArr)
		}
	case "Month":
		{
			TimeArr = TimeTaskArr(3, TimeArr)
		}
	}
	h, err := json.MarshalIndent(CopyStruct, "", "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(w, string(h))
}
