package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type ServerReply struct {
	MainHost   string
	BackupHost string
}

func main() {
	main := "main"
	backup := "backup"
	version := 36
	ptrVersion := &version
	for {
		srvReply := updateServersHaproxy()
		fmt.Println(srvReply)
		delServer(main, &version)
		delServer(backup, &version)
		addServer(main, srvReply.MainHost, 80, ptrVersion)
		addServer(backup, srvReply.BackupHost, 80, ptrVersion)
		time.Sleep(15 * time.Second)
	}
}

func addServer(server, addr string, port int, version *int) {
	fmt.Printf("addServer: %s %s %d %d\n", server, addr, port, *version)
	backup := "disabled"
	if server == "backup" {
		backup = "enabled"
	}
	*version++
	versionStr := fmt.Sprintf("%d", *version)
	resp, err := http.DefaultClient.Do(&http.Request{
		Method: "POST",
		URL:    &url.URL{Scheme: "http", Host: "localhost:5555", Path: "/v2/services/haproxy/configuration/servers", RawQuery: "backend=serverA&&version=" + versionStr},
		Header: http.Header{"Content-Type": []string{"application/json"}, "Authorization": {"Basic YWRtaW46YWRtaW4="}},
		Body:   ioutil.NopCloser(strings.NewReader(fmt.Sprintf(`{"name": "%s", "address": "%s", "port" : %d, "backup": "%s"}`, server, addr, port, backup))),
	})
	br, _ := io.ReadAll(resp.Body)
	fmt.Printf("addServer: %v. Err: %v\n", string(br), err)
}

func delServer(server string, version *int) {
	*version++
	versionStr := fmt.Sprintf("%d", *version)
	resp, err := http.DefaultClient.Do(&http.Request{
		Method: "DELETE",
		Header: http.Header{"Content-Type": []string{"application/json"}, "Authorization": {"Basic YWRtaW46YWRtaW4="}},
		URL:    &url.URL{Scheme: "http", Host: "localhost:5555", Path: "/v2/services/haproxy/configuration/servers/" + server, RawQuery: "backend=serverA&&version=" + versionStr},
	})
	br, _ := io.ReadAll(resp.Body)
	fmt.Printf("delServer: %v. Err: %v\n", string(br), err)
}

func updateServersHaproxy() ServerReply {

	m := new(dns.Msg)
	m.SetQuestion("_bestserviceA._tcp.internal.", dns.TypeSRV)

	dnsClient := new(dns.Client)
	m, _, err := dnsClient.Exchange(m, "localhost:1053")
	if err != nil {
		panic(err)
	}

	srvReply := ServerReply{}

	for _, ans := range m.Answer {
		srv := ans.(*dns.SRV)
		host := srv.Target[:len(srv.Target)-1]

		if srv.Priority == 0 {
			srvReply.MainHost = host
		} else {
			srvReply.BackupHost = host
		}
		fmt.Println(srv.Target)
	}

	return srvReply
}
