package main

import (
	"reflect"
	"testing"
)

func TestRoBind(t *testing.T) {
	got := roBind("/src")
	want := []string{"--ro-bind", "/src", "/src"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("roBind(%q) = %v, want %v", "/src", got, want)
	}
}

func TestRoBindWithDst(t *testing.T) {
	got := roBind("/src", "/dst")
	want := []string{"--ro-bind", "/src", "/dst"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("roBind(%q, %q) = %v, want %v", "/src", "/dst", got, want)
	}
}

func TestRwBind(t *testing.T) {
	got := rwBind("/src")
	want := []string{"--bind", "/src", "/src"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("rwBind(%q) = %v, want %v", "/src", got, want)
	}
}

func TestRwBindWithDst(t *testing.T) {
	got := rwBind("/src", "/dst")
	want := []string{"--bind", "/src", "/dst"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("rwBind(%q, %q) = %v, want %v", "/src", "/dst", got, want)
	}
}

func TestDevBind(t *testing.T) {
	got := devBind("/dev/shm")
	want := []string{"--dev-bind", "/dev/shm", "/dev/shm"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("devBind(%q) = %v, want %v", "/dev/shm", got, want)
	}
}

func TestDevBindWithDst(t *testing.T) {
	got := devBind("/dev/shm", "/mnt/shm")
	want := []string{"--dev-bind", "/dev/shm", "/mnt/shm"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("devBind(%q, %q) = %v, want %v", "/dev/shm", "/mnt/shm", got, want)
	}
}

func TestTmpfs(t *testing.T) {
	got := tmpfs("/tmp")
	want := []string{"--tmpfs", "/tmp"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tmpfs(%q) = %v, want %v", "/tmp", got, want)
	}
}
