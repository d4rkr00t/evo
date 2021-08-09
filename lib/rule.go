package lib

type Rule struct {
	Cmd         string
	Deps        []string
	CacheOutput bool
}
