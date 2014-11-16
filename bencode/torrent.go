package bencode

// 代表一个torrent文件
type Torrent struct {
	Announce     string      `bencode:"announce"`
	AnnounceList [][]string  `bencode:"announce-list,omitempty"`
	CreateBy     string      `bencode:"created by,omitempty"`
	CreateDate   int         `bencode:"creation date,omitempty"`
	Comment      string      `bencode:"comment,omitempty"`
	Encoding     string      `bencode:"encoding,omitempty"`
	Info         FileInfo    `bencode:"info"`
	Nodes        interface{} `bencode:"nodes,omitempty"`
}

// bt种子的文件信息
type FileInfo struct {
	Files        []File   `bencode:"files,omitempty"`
	Name         string   `bencode:"name"`
	Length       int      `bencode:"length,omitempty"`
	Ed2k         string   `bencode:"ed2k,omitempty"`
	Md5Sum       string   `bencode:"md5sum,omitempty"`
	FileHash     string   `bencode:"filehash,omitempty"`
	PieceLength  int      `bencode:"piece length"`
	Pieces       string   `bencode:"pieces"`
	FileDuration []int    `bencode:"file-duration,omitempty"`
	FileMedia    []int    `bencode:"file-media,omitempty"`
	Profiles     MetaData `bencode:"profiles,omitempty"`
}

// 媒体文件元数据
type MetaData struct {
	Acodec string `bencode:"acodec"`
	Vcodec string `bencode:"vcodec"`
	Height int    `bencode:"height"`
	Width  int    `bencode:"width"`
}

// 具体单个文件的路径和大小
type File struct {
	Length int      `bencode:"length"`
	Md5Sum string   `bencode:"md5sum,omitempty"`
	Path   []string `bencode:"path"`
}
