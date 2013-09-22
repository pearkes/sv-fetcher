package main

import (
	"fmt"
	"github.com/pearkes/sv-frontend/data"
)

// saves a page to redis with the key page:userid:dbxuid
func savePage(page string, u data.User) error {
	conn := r.Redis.Get()
	defer conn.Close()

	keyName := fmt.Sprintf("page:%v:%s", u.Id, u.DropboxUid)
	_, err := conn.Do("SET", keyName, page)
	if err != nil {
		return err
	}
	return nil
}
