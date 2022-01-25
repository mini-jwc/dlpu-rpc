package model

// User LastSaveTime 上次登录时间，每隔3个月就要更新一次
type User struct{}

type Args struct {
	Ip      string
	Session string
	Data    string
	StuId   string
	Pwd     string
}

type Reply struct {
	Res interface{}
	Err error
}

//CTTDetails 是课程细节数组
type CTTDetails []CTTDetail

//CTT 是某个时间段的所有课程，比如早上一二节，你可能有两种课
type CTT struct {
	CTTDetails
	Id int
}

//CTTDetail 是课程细节
type CTTDetail struct {
	Name    string
	Room    string
	Teacher string
	Week    string
	Time    string
}

// ExamScore 考试分数
type ExamScore struct {
	GPA        string //绩点
	Credit     string //学分
	Property   string //属性
	Period     string //学时
	Properties string //性质
	Name       string
	Grade      string
	Detail     string
	BC         string //补考重修？
}

type ExamTime struct {
	ID   string
	Name string
	Time string
	Room string
}

type ExamScoreDetail struct {
	Ordinary      string
	OrdPercent    string
	Middle        string
	MiddlePercent string
	Final         string
	FinalPercent  string
	Total         string
}

// CultivateScheme 培养方案结构体
type CultivateScheme struct {
	Semester   string //开设学期
	Name       string //课程名称
	Credit     string //学分
	Period     string //学时
	WeekPeriod string //周学时
	College    string //开课学院
	ExamMode   string //考核方式
}
