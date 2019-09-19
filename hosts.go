package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"sort"
)

const (
	hostsFilename    = "hosts"                // Your hostfile
	backupFilename   = "hosts.bak"            // The name of the backup to create the first time
	hostURLSFilename = "adhosts.cfg"          // this file holds the urls of hour adfilter-hostfiles
	divisionTag      = "<-hosts-separation->" // a line holding this tag divides your personal part from the adhosts
	address          = "0.0.0.0"              // This IP will be set for the ad-Hosts
	// This line will be added to hosts to separate the original hosts from the ad-hosts
	divisionLine = "# " + divisionTag + " <-- DO NOT CHANGE THIS LINE. Any changes after this line will be lost!!\n"
)

// Regexp
const (
	rexFindDivision = `^\S*#.*` + divisionTag
	rexHostname     = `^\s*((127.0.0.1|0.0.0.0)\s+)?([^#\s]+)`
)

type hostlist struct {
	sort []string
	list map[string]bool
}

func (h *hostlist) append(host string) {
	if h.list[host] == false {
		h.list[host] = true
		h.sort = append(h.sort, host)
	}
}

func (h *hostlist) initHostlist() {
	h.list = make(map[string]bool)
}

func (h *hostlist) getList() []string {
	sort.Strings(h.sort)
	return h.sort
}

func main() {
	etcPath := setPathsByOS()
	if etcPath == "" {
		log.Fatal("Could not determin your systemtype")
	}
	hosts := path.Join(etcPath, hostsFilename)
	backup := path.Join(etcPath, backupFilename)
	hostURLs := path.Join(etcPath, hostURLSFilename)

	adList := new(hostlist)
	adList.initHostlist()
	ownHosts := new([]string)

	// Backup hostsfile if no Backup is present.
	if _, err := os.Stat(backup); os.IsNotExist(err) {
		log.Printf("Creating a copy of your hosts-file %s as %s.\n", hosts, backup)
		makeBackup(hosts, backup)
	}

	fmt.Printf("Reading hosts from your hostfile in %s\n", hosts)
	ownHosts, adList = hostsUntilDivide(hosts, ownHosts, adList)

	var URLs []string
	if _, err := os.Stat(hostURLSFilename); !os.IsNotExist(err) {
		fmt.Printf("Reading ad-hosts from %s\n", hostURLSFilename)
		URLs = append(URLs, readHostURLS(hostURLSFilename)...)
	}
	if _, err := os.Stat(hostURLs); !os.IsNotExist(err) {
		fmt.Printf("Reading ad-hosts from %s\n", hostURLs)
		URLs = append(URLs, readHostURLS(hostURLs)...)
	}
	if len(URLs) == 0 {
		log.Fatal("No URLs found to download blacklists from")
	}

	for i, v := range URLs {
		fmt.Printf("[%d/%d] Downloading %s...\n", i+1, len(URLs), v)
		hostLines, err := fetchURL(v)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, k := range hostLines {
			hostname, ok := hostnameFromLine(k)
			if ok {
				adList.append(hostname)
			}
		}
	}
	fmt.Printf("Writing new hostfile in %s\n", hosts)
	if err := writeNewHosts(hosts, ownHosts, adList); err != nil {
		log.Fatal(err)
	}
}

// Returns the content of the hosts-file up to the division (So your own
// content is returned)
func hostsUntilDivide(hosts string, ownHosts *[]string, adList *hostlist) (*[]string, *hostlist) {
	file, err := os.Open(hosts)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	division := regexp.MustCompile(rexFindDivision)
	scanner := bufio.NewScanner(file)
	var zeile string
	for scanner.Scan() {
		zeile = scanner.Text()
		*ownHosts = append(*ownHosts, zeile+"\n")
		if division.MatchString(zeile) {
			break
		}
	}
	for scanner.Scan() {
		zeile = scanner.Text()
		hostname, ok := hostnameFromLine(zeile)
		if ok {
			adList.append(hostname)
		}
	}
	return ownHosts, adList
}

// Create a backup of the original hosts-file
func makeBackup(hosts string, backup string) {
	input, err := os.Open(hosts)
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()

	output, err := os.Create(backup)
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		log.Fatal(err)
	}
	if err := output.Sync(); err != nil {
		log.Fatal(err)
	}
}

// returns the hostnamepart from a hosts-file-line
func hostnameFromLine(line string) (hostname string, ok bool) {
	hasHostname := regexp.MustCompile(rexHostname)
	hostLine := hasHostname.FindStringSubmatch(line)
	if len(hostLine) != 4 {
		return "", false
	}
	if hostLine[3] == "localhost" {
		return "", false
	}
	return hostLine[3], true
}

// Write a new hosts-file
func writeNewHosts(hosts string, ownHosts *[]string, adList *hostlist) error {
	file, err := os.Create(hosts)
	if err != nil {
		return err
	}
	defer file.Close()
	// write my own hosts
	var v string
	for _, v = range *ownHosts {
		file.WriteString(v)
	}
	division := regexp.MustCompile(rexFindDivision)
	if !division.MatchString(v) {
		file.WriteString(divisionLine)
	}

	// write the advertising hosts
	for _, v := range (*adList).getList() {
		file.WriteString(address + " " + v + "\n")
	}
	return nil
}

func readHostURLS(hostURLS string) []string {
	var result []string
	file, err := os.Open(hostURLS)
	if err != nil {
		log.Fatal("Could not open URLs-File", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		zeile := scanner.Text()
		comment := regexp.MustCompile(`\s*#.*$`)
		zeile = comment.ReplaceAllString(zeile, "")
		if len(zeile) == 0 {
			continue
		}
		result = append(result, zeile)
	}
	return result
}

// Fetch the content of a given URL and return its lines as []string
func fetchURL(url string) ([]string, error) {
	var content []string
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return content, nil
}

func setPathsByOS() string {
	var etcPath string
	switch osname := runtime.GOOS; osname {
	case "linux":
		etcPath = "/etc/"
	case "freebsd":
		etcPath = "/usr/local/etc/"
	case "windows":
		windir := os.Getenv("windir")
		etcPath = path.Join(windir, "system32/drivers/etc")
	}
	return etcPath
}
