package orm

// go ：struct内部的反引号
// https://blog.csdn.net/jason_cuijiahui/article/details/82987091
type WorkerNodeEntity struct {
	Id         uint   `grom:"colum:ID"`
	HostName   string `grom:"colum:host_name"`
	Port       int    `grom:"colum:prot"`
	Type       int    `grom:"colum:type"`
	LaunchDate int64  `grom:"colum:launch_date"`
	Updated    int64  `gorm:"autoUpdateTime:nano"` // 使用时间戳填纳秒数充更新时间
	Created    int64  `gorm:"autoCreateTime"`      // 使用时间戳秒数填充创建时间
}
