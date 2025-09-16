package api

type Task struct {
	ID      int    `db:"id" json:"id"`
	Date    string `db:"date" json:"date"`
	Title   string `db:"title" json:"title"`
	Comment string `db:"comment" json:"comment"`
	Repeat  string `db:"repeat" json:"repeat"`
}

// Структура для ответа с задачами
type TasksResp struct {
	Tasks []*Task `json:"tasks"`
}
