package chain_node

// type Data struct {
// 	Id      int64
// 	Version int64
// }

// // mark all new data as dirty
// func NewData(id, version int64) *Data {
// 	return &Data{Id: id, Version: version}
// }

// // check whether the data is clean
// func (data *Data) checkClean(lastCleanData *Data) bool {
// 	return data.Version == lastCleanData.Version
// }

// // find the data from the node
// // retrieve the latest version the node
// // if id is not present return 1 (one based indexing)
// func (node *ChainNode) GetDataAndVersion(id int64) int64 {
// 	data, ok := node.NodeData.GetValByKey(id)
// 	if !ok {
// 		return 1
// 	}
// 	return data.Version + 1
// }

//rpc calls to sync the old tail and the new tail
func p (){}