//main  获取数据的，具体的动作函数都在这里
package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	//"github.com/gooops/micadvisor_open/docker"
)

//v0.25 默认暴露8080端口
var CadvisorPort = "18080"

func getTag(_dockerdata DockerInspectStruct) string {
	// //FIXMI:some other message for container
	// return ""
	var tags []string
	for i, u := range _dockerdata.Config.Labels {
		//LogRun("gettag: " + i + u)
		tags = append(tags, i+":"+u)
	}
	return strings.Join(tags, ",")

}

func getCadvisorData() (string, error) {
	//LogRun("come in function  getCadvisorData  ....")
	var (
		resp *http.Response
		err  error
		body []byte
	)
	url := "http://localhost:" + CadvisorPort + "/api/v1.2/docker"
	if resp, err = http.Get(url); err != nil {
		LogErr(err, "Get err in getCadvisorData")
		return "", err
	}
	//LogRun("function  getCadvisorData   localhost:port/api/v1.2/docker")
	defer resp.Body.Close()
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		LogErr(err, "ReadAll err in getCadvisorData")
		return "", err
	}

	return string(body), nil
}

//getOneContainerCadvisorData  获取某个容器的cadvisor JSON信息。
func getOneContainerCadvisorData(_oneContainerID string) (string, error) {
	var (
		resp *http.Response
		err  error
		body []byte
	)
	url := "http://localhost:" + CadvisorPort + "/api/v1.2/docker/" + _oneContainerID
	if resp, err = http.Get(url); err != nil {
		//LogErr(err, "from getOneContainerCadvisorData function  ")
		return "", err
	}
	defer resp.Body.Close()
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		//LogErr(err, "from getOneContainerCadvisorData function ReadAll  ")
		return "", err
	}
	return string(body), nil
}

func getEndPoint(_containerName string) string {

	// hostname, err := os.Hostname()
	// if err != nil {
	// 	LogErr(err, "getDatas.go   get hostname err")
	// }
	// return hostname + "-" + strings.Trim(_containerName, "/")

	var EndPointPrefix string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		LogErr(err, "getDatas.go   get ipaddress  err")
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				//fmt.Println(ipnet.IP.String())
				EndPointPrefix = ipnet.IP.String() // 第一个非lookbackip被拿到之后就退出。
				break
			}

		}
	}

	return EndPointPrefix + "-" + strings.Trim(_containerName, "/")

}

// 获取容器的inspect信息，手动方法  (echo -e "GET /containers/XXXXXXXXXX/json HTTP/1.1\r\nHost: www.test.com\r\n")|nc -U //var/run/docker.sock
func getDockerData(containerId string) (string, error) {
	str, err := RequestUnixSocket("/containers/"+containerId+"/json", "GET")
	if err != nil {
		//LogErr(err, "getDatas.go  getDockerData  err")
		return "", err

	}
	return str, nil
}

//getAllDockerDataBrief  获取所有容器的基本信息，手动方法  (echo -e "GET /containers/json HTTP/1.1\r\nHost: www.test.com\r\n")|nc -U //var/run/docker.sock
func getAllDockerBrief() (string, error) {

	str, err := RequestUnixSocket("/containers/json", "GET")
	if err != nil {
		//LogErr(err, "getDatas.go  getAllDockerBrief err")
		return "", err

	}
	return str, nil
}

//RequestUnixSocket 使用docker自身的api获取数据
func RequestUnixSocket(address, method string) (string, error) {
	DOCKER_UNIX_SOCKET := "unix:///var/run/docker.sock"
	// Example: unix:///var/run/docker.sock:/images/json?since=1374067924
	unix_socket_url := DOCKER_UNIX_SOCKET + ":" + address
	u, err := url.Parse(unix_socket_url)
	if err != nil || u.Scheme != "unix" {
		LogErr(err, "getDatas.go  Error to parse unix socket url "+unix_socket_url)
		return "", err
	}

	hostPath := strings.Split(u.Path, ":")
	u.Host = hostPath[0]
	u.Path = hostPath[1]

	conn, err := net.Dial("unix", u.Host)
	if err != nil {
		LogErr(err, "getDatas.go  Error to connect to"+u.Host)
		// fmt.Println("Error to connect to", u.Host, err)
		return "", err
	}

	reader := strings.NewReader("")
	query := ""
	if len(u.RawQuery) > 0 {
		query = "?" + u.RawQuery
	}

	request, err := http.NewRequest(method, u.Path+query, reader)
	if err != nil {
		LogErr(err, "getDatas.go Error to create http request")
		// fmt.Println("Error to create http request", err)
		return "", err
	}

	client := httputil.NewClientConn(conn, nil)
	response, err := client.Do(request)
	if err != nil {
		LogErr(err, "getDatas.go  Error to achieve http request over unix socket")
		// fmt.Println("Error to achieve http request over unix socket", err)
		return "", err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		LogErr(err, "getDatas.go  Error, get invalid body in answer")
		// fmt.Println("Error, get invalid body in answer")
		return "", err
	}

	defer response.Body.Close()

	return string(body), err
}
