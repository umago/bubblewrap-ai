package main

func roBind(src string, dst ...string) []string {
	target := src
	if len(dst) > 0 {
		target = dst[0]
	}
	return []string{"--ro-bind", src, target}
}

func rwBind(src string, dst ...string) []string {
	target := src
	if len(dst) > 0 {
		target = dst[0]
	}
	return []string{"--bind", src, target}
}

func devBind(src string, dst ...string) []string {
	target := src
	if len(dst) > 0 {
		target = dst[0]
	}
	return []string{"--dev-bind", src, target}
}

func tmpfs(path string) []string {
	return []string{"--tmpfs", path}
}
