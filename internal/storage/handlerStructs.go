package storage

type DLoad struct {
	NodeID string `json:"id"`
}

type ListNodes struct {
	ParentID string `json:"parent_id"`
}

type Mkdir struct {
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
}
