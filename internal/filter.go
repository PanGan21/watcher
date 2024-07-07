package watcher

import (
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
)

var filterMap = map[string]fsnotify.Op{
	"CREATE": fsnotify.Create,
	"WRITE":  fsnotify.Write,
	"REMOVE": fsnotify.Remove,
	"RENAME": fsnotify.Rename,
	"CHMOD":  fsnotify.Chmod,
}

func parseFilters(filters []string) (fsnotify.Op, error) {
	var op fsnotify.Op
	for _, f := range filters {
		event, exists := filterMap[strings.ToUpper(f)]
		if !exists {
			return 0, fmt.Errorf("invalid filter event: %s", f)
		}
		op |= event
	}

	return op, nil
}
