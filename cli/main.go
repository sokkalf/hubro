package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
)

func randomString(n int) string {
    letters := []rune("abcdefghijklmnopqrstuvwxyz")
    sb := strings.Builder{}
    for i := 0; i < n; i++ {
        sb.WriteRune(letters[rand.Intn(len(letters))])
    }
    return sb.String()
}

func generateMarkdownFile(fileName string) error {
    title := fmt.Sprintf("Random Title %s", randomString(6))
    description := fmt.Sprintf("Random Description %s", randomString(8))
    author := fmt.Sprintf("Author %s", randomString(4))
	day := rand.Intn(28) + 1
	month := rand.Intn(12) + 1
	year := rand.Intn(10) + 2015
    date := fmt.Sprintf("%d-%02d-%02d", year, month, day)

    frontmatter := fmt.Sprintf(`---
title: "%s"
description: "%s"
date: %s
author: "%s"
---

`, title, description, date, author)

    randomText := fmt.Sprintf("This is some random content: %s", randomString(20))
    content := frontmatter + randomText

    f, err := os.Create(fileName)
    if err != nil {
        return err
    }
    defer f.Close()

    _, err = f.WriteString(content)
    return err
}

func main() {
    n := 5
    for i := 1; i <= n; i++ {
        fileName := fmt.Sprintf("random-file-%s.md", randomString(5))

        err := generateMarkdownFile("blog/" + fileName)
        if err != nil {
            log.Fatalf("Error generating file %s: %v", fileName, err)
        }
        fmt.Printf("Generated %s\n", fileName)
    }
}
