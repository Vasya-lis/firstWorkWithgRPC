package api

import md "github.com/Vasya-lis/firstWorkWithgRPC/services/models"

type TasksResponse struct {
	Tasks []*md.Task `json:"tasks"`
}
