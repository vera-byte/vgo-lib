package common

// PageResult 分页结果
type PageResult struct {
	Page  int64       `json:"page"`  // 页码
	Size  int64       `json:"size"`  // 页大小
	Total int64       `json:"total"` // 数据总量
	Data  interface{} `json:"data"`  // 数据
}

// NewPageResult NewPageResult
func NewPageResult(page int64, size int64, total int64, data interface{}) *PageResult {
	return &PageResult{
		Page:  page,
		Size:  size,
		Total: total,
		Data:  data,
	}
}
