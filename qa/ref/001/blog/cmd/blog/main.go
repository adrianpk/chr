package main

import (
	"log"
	"net/http"

	"blog/internal/blog"
)

func main() {
	server := blog.NewServer(blog.NewPostStore(blog.SeedPosts()))

	log.Fatal(http.ListenAndServe(":8080", server.Routes()))
}
