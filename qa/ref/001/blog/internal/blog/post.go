package blog

import (
	"errors"
	"slices"
	"sync"
)

var ErrPostNotFound = errors.New("post not found")

type Post struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Published bool   `json:"published"`
}

type PostStore struct {
	mu    sync.RWMutex
	posts []Post
}

func NewPostStore(posts []Post) *PostStore {
	return &PostStore{
		posts: slices.Clone(posts),
	}
}

func SeedPosts() []Post {
	return []Post{
		{
			Slug:      "hello-world",
			Title:     "Hello World",
			Body:      "The first public post.",
			Published: true,
		},
		{
			Slug:      "release-notes",
			Title:     "Release Notes",
			Body:      "A short update for readers.",
			Published: true,
		},
	}
}

func (s *PostStore) PublishedPosts() []Post {
	s.mu.RLock()
	defer s.mu.RUnlock()

	posts := make([]Post, 0, len(s.posts))
	for _, post := range s.posts {
		if post.Published {
			posts = append(posts, post)
		}
	}

	return posts
}

func (s *PostStore) PublishedPost(slug string) (Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, post := range s.posts {
		if post.Slug == slug && post.Published {
			return post, nil
		}
	}

	return Post{}, ErrPostNotFound
}
