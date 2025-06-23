// Package blog handles all business logic related to blog posts.
package blog

import (
	"time"
)

type Post struct {
	ID        string
	Title     string
	Content   string
	Author    string
	CreatedAt time.Time
}

var posts []Post

