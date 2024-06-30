//go:build mage
// +build mage

package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func Run() error {
	mg.Deps(Build)
	return sh.RunV("bin/api/main")
}

// Builds the API project to the bin/api/main binary.
func Build() error {
	mg.Deps(GenProto)
	return sh.RunV("go", "build", "-o", "bin/api/main", "cmd/api/main.go")
}

// Generates the Golang and C# protofiles for the project.
func GenProto() error {
	return sh.RunV("protoc", "--proto_path=proto", "--csharp_out=cli/build/gen", "--csharp_opt=file_extension=.g.cs", "--go_out=internal/proto", "--go_opt=paths=source_relative", "chat/chat.proto", "errors/error.proto")
}

// Cleans up build artifacts and generated code.
func Clean() error {
	if err := cleanBuild(); err != nil {
		return err
	}
	if err := cleanProto(); err != nil {
		return err
	}
	return nil
}

func cleanBuild() error {
	return os.RemoveAll("bin")
}

func cleanProto() error {
	var files []string
	filepath.Walk("internal", func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(path, ".pb.go") {
			files = append(files, path)
		}
		return nil
	})
	filepath.Walk("cli/build/gen", func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(path, ".g.cs") {
			files = append(files, path)
		}
		return nil
	})
	log.Printf("Cleaning %d files: %v\n", len(files), files)
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	return nil
}

var Default = Build
