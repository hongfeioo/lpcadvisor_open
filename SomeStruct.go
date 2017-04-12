//main  这里只存放解析json 的结构体
//这些结构体是根据dockerAPI返回的JSON手工打造的， 目前我公司感兴趣的指标进行了细化，其它都以interface{}代替， 以后如有更多需求，只需细化这些机构体即可
//20170412 yihongfei

package main

//DockerInspectStruct  这个结构体对应Inspect json串结构
type DockerInspectStruct struct {
	ID              string `json:"Id"`
	Created         string
	Path            string
	Args            []string
	State           map[string]interface{}
	HostConfig      map[string]interface{}
	Image           string
	ResolvConfPath  string
	HostnamePath    string
	HostsPath       string
	LogPath         string
	Name            string
	RestartCount    int
	Driver          string
	MountLabel      string
	ProcessLabel    string
	ExecIDs         interface{}
	GraphDriver     map[string]interface{}
	Mounts          []interface{}
	Config          DockerInspectStructConfig
	NetworkSettings DockerInspectStructNetworkSettings
}

//DockerInspectStructNetworkSettings 本结构体对应 docker inspect 197db54aab6c0510686b3  其中的NetworkSettings 部分
type DockerInspectStructNetworkSettings struct {
	Bridge                 string
	SandboxID              string
	HairpinMode            bool
	LinkLocalIPv6Address   string
	LinkLocalIPv6PrefixLen int
	Ports                  interface{}
	SandboxKey             string
	SecondaryIPAddresses   interface{}
	SecondaryIPv6Addresses interface{}
	EndpointID             string
	Gateway                string
	GlobalIPv6Address      string
	GlobalIPv6PrefixLen    int
	IPAddress              string
	IPPrefixLen            int
	IPv6Gateway            string
	MacAddress             string
	Networks               interface{}
}

//DockerInspectStructConfig  就是 DockerInspectStruct 结构体中config的详细结构
type DockerInspectStructConfig struct {
	Hostname     string
	Domainname   string
	User         string
	AttachStdin  bool
	AttachStdout bool
	AttachStderr bool
	ExposedPorts interface{}
	Tty          bool
	OpenStdin    bool
	StdinOnce    bool
	Env          []interface{}
	Cmd          []interface{}
	Image        string
	Volumes      interface{}
	WorkingDir   string
	Entrypoint   []interface{}
	OnBuild      bool
	Labels       map[string]string
}

//AllDockerBriefStruct   存放类似docker ps 出来的json信息，  (echo -e "GET /containers/json HTTP/1.1\r\nHost: www.test.com\r\n")|nc -U //var/run/docker.sock
type AllDockerBriefStruct struct {
	ID              string `json:"Id"`
	Names           []string
	Image           string
	ImageID         string
	Command         string
	Created         int64
	Ports           []interface{}
	Labels          interface{}
	State           string
	Status          string
	HostConfig      map[string]string
	NetworkSettings map[string]interface{}
	Mounts          []interface{}
}

//
// cadvisor 可以获取宿主的信息的秘诀！   curl http://localhost:18080/api/v1.2/machine    这个结构体可以深挖一下，是否可以获取虚拟机的信息呢？
// cadvisor api 2.0 更能更强大  https://github.com/google/cadvisor/blob/master/docs/api_v2.md    完全可以用上。
