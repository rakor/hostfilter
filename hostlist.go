package main

import "sort"

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
