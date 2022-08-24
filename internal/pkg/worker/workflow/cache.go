package workflowruntime

// 针对 workflow 的结点的输出结果的缓存
//
// 利用已经计算好的 hash 值寻找缓存
//
// 缓存的值包括：Output, Result
type RtNodeCache interface {
	// add node to cache (by hash)
	//
	// 具体是否加入缓存取决于 node 本身
	Add(node *RtNode) error
	// check if cache exist
	Exist(node *RtNode) bool
	// assign cache to node
	Assign(node *RtNode) error
}
