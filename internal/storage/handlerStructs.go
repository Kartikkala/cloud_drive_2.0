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

type Move struct {
	SrcParentID  string `json:"src_parent_id"`
	DestParentID string `json:"dest_parent_id"`
	TargetNodeID string `json:"target_id"`
}

type Delete struct {
	NodeID   string `json:"id"`
	ParentID string `json:"parent_id"`
}
