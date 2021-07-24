package orm

func GenerateWorkerNode(hostName string, port int, _type int) int64 {
	worksNode := New()
	var newWorkerNode = WorkerNodeEntity{
		HostName: hostName,
		Port:     port,
		Type:     _type,
	}
	worksNode.db.Create(&newWorkerNode)
	return int64(newWorkerNode.Id)
}
