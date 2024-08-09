package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type ServerReply struct {
	MainHost   string
	BackupHost string
}

const (
	creds       = "Basic YWRtaW46YWRtaW4="
	haproxyhost = "haproxy:5555"
)

func main() {
	main := "main"
	backup := "backup"
	for {
		srvReply, err := updateServersHaproxy()
		if err != nil {
			fmt.Println(err)
			time.Sleep(15 * time.Second)
			continue
		}
		fmt.Println(srvReply)
		err = delServer(main)
		if err != nil {
			fmt.Println(err)
			time.Sleep(15 * time.Second)
			continue
		}

		err = delServer(backup)
		if err != nil {
			fmt.Println(err)
			time.Sleep(15 * time.Second)
			continue
		}

		err = addServer(main, srvReply.MainHost, 80)
		if err != nil {
			fmt.Println(err)
			time.Sleep(15 * time.Second)
			continue
		}

		err = addServer(backup, srvReply.BackupHost, 80)
		if err != nil {
			fmt.Println(err)
			time.Sleep(15 * time.Second)
			continue
		}

		time.Sleep(15 * time.Second)
	}
}

func getVer() (int, error) {
	resp, err := http.DefaultClient.Do(&http.Request{
		Method: "GET",
		Header: http.Header{"Content-Type": []string{"application/json"}, "Authorization": {creds}},
		URL:    &url.URL{Scheme: "http", Host: haproxyhost, Path: "/v2/services/haproxy/configuration/version"},
	})
	if err != nil {
		return 0, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	bi, err := strconv.Atoi(strings.ReplaceAll(string(b), "\n", ""))
	if err != nil {
		return 0, err
	}
	return bi, nil
}

func addServer(server, addr string, port int) error {
	version, err := getVer()
	if err != nil {
		return err
	}

	fmt.Printf("addServer: %s %s %d %d\n", server, addr, port, version)
	backup := "disabled"
	if server == "backup" {
		backup = "enabled"
	}

	versionStr := fmt.Sprintf("%d", version)
	resp, err := http.DefaultClient.Do(&http.Request{
		Method: "POST",
		URL:    &url.URL{Scheme: "http", Host: haproxyhost, Path: "/v2/services/haproxy/configuration/servers", RawQuery: "backend=serverA&&version=" + versionStr},
		Header: http.Header{"Content-Type": []string{"application/json"}, "Authorization": {creds}},
		Body:   ioutil.NopCloser(strings.NewReader(fmt.Sprintf(`{"name": "%s", "address": "%s", "port" : %d, "backup": "%s"}`, server, addr, port, backup))),
	})
	br, _ := io.ReadAll(resp.Body)
	fmt.Printf("addServer: %v. Err: %v\n", string(br), err)
	return nil
}

func delServer(server string) error {
	version, err := getVer()
	if err != nil {
		return err
	}

	versionStr := fmt.Sprintf("%d", version)
	resp, err := http.DefaultClient.Do(&http.Request{
		Method: "DELETE",
		Header: http.Header{"Content-Type": []string{"application/json"}, "Authorization": {creds}},
		URL:    &url.URL{Scheme: "http", Host: haproxyhost, Path: "/v2/services/haproxy/configuration/servers/" + server, RawQuery: "backend=serverA&&version=" + versionStr},
	})
	br, _ := io.ReadAll(resp.Body)
	fmt.Printf("delServer: %v. Err: %v\n", string(br), err)
	return nil
}

func updateServersHaproxy() (ServerReply, error) {

	m := new(dns.Msg)
	m.SetQuestion("_bestserviceA._tcp.internal.", dns.TypeSRV)

	dnsClient := new(dns.Client)
	m, _, err := dnsClient.Exchange(m, "dns:1053")
	if err != nil {
		return ServerReply{}, err
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

	return srvReply, nil
}
