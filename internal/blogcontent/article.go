package blogcontent

import "time"

// Article includes both the blog post content as well as its metadata.
type Article struct {
	ID       string   // md5(content)
	URL      string   // relative to blog root, based on ID
	Metadata Metadata // format: yaml
	Content  []byte   // format: pandoc markdown
}

// Metadata includes other information like title, author, tags, summary, etc.
type Metadata struct {
	Title     string    `yaml:"title"`
	WrittenAt time.Time `yaml:"written_at"`
	Author    User      `yaml:"author"`
	Tags      []Tag     `yaml:"tags"`
	Summary   string    `yaml:"summary"`
}

// User is the name of the user/writer.
type User string

// Tag is opaque strings associated to an article.
// Usually it represents categories.
type Tag string

const (
	TagProgramming Tag = "programming"
	TagGolang      Tag = "golang"
	TagKubernetes  Tag = "kubernetes"
	TagThoughts    Tag = "thoughts"
)
