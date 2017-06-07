package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/russross/blackfriday"
)

var (
	title string
	style string
)

func main() {
	// Flags
	flag.StringVar(&title, "title", "Presentation", "title for the presentation")
	flag.StringVar(&style, "style", "", "additional style for presentation")
	flag.Parse()

	// Read Input
	if len(os.Args) < 2 {
		panic("No input markdown file fiven")
	}
	input := os.Args[len(os.Args)-1]
	markdown, err := ioutil.ReadFile(input)
	if err != nil {
		panic(err)
	}

	// Compile HTML
	output := blackfriday.MarkdownCommon(markdown)

	// Processing Pipeline
	output = compileSlides(output)
	output = sizeImages(output)
	output = surroundWithHTML(output)
	output = numberSlides(output)

	// Write HTML
	err = ioutil.WriteFile(input+".html", output, 0644)
	if err != nil {
		panic(err)
	}
}

// Converts horizontal rulers to slides
func compileSlides(html []byte) []byte {
	out := string(html)
	out = strings.Replace(out, "<hr />", "</div>\n</section>\n<section class='slide'>\n<div class='padding'>", -1)
	return []byte(out)
}

var openSlideRegex = regexp.MustCompile("<section class='slide'>")

// Puts numbered ids to each slide and appends the pager div
func numberSlides(html []byte) []byte {
	out := string(html)
	// Find all slides tags
	slides := openSlideRegex.FindAll(html, -1)
	for i, slide := range slides {
		s := string(slide)
		pager := fmt.Sprintf("<div class='pager'><span class='current'>%v</span><span class='separator'>/</span><span class='total'>%v</span></div>", i+1, len(slides))
		out = strings.Replace(out, s, fmt.Sprintf("%v id=\"%v\">\n%v", s[0:len(s)-1], i+1, pager), 1)
	}
	return []byte(out)
}

var imgRegex = regexp.MustCompile("<img.+/>")
var imgAltRegex = regexp.MustCompile("alt=\"(.+)\"")
var isDigitRegex = regexp.MustCompile("\\d")

// Sizes images based on their alt attributes
func sizeImages(html []byte) []byte {
	out := string(html)
	// Find all image tags
	images := imgRegex.FindAll(html, -1)
	for _, image := range images {
		img := string(image)

		// Find all alt tags
		alts := imgAltRegex.FindSubmatch(image)
		if len(alts) < 2 {
			continue
		}
		alt := alts[1]

		// Tokens are space separated in alt tags
		tokens := strings.Split(string(alt), " ")
		style := ""
		class := ""
		for _, token := range tokens {
			// If token contains a number it is assumed to be a width
			if isDigitRegex.MatchString(token) {
				style += "width: " + token + ";"
				continue
			}
			// Else it is appended as class to the image, e.g. center
			class += token + " "
		}
		// Replace old image tag with patched one
		newImg := fmt.Sprintf("%v style=\"%v\" class=\"%v\">", img[0:len(img)-3], style, class)
		out = strings.Replace(out, img, newImg, 1)
	}
	return []byte(out)
}

// Wraps all slides into the HTML base structure
func surroundWithHTML(html []byte) []byte {
	out := fmt.Sprintf(`<!DOCTYPE html>
<html>
  <head>
    <title>%v</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
		<style>%v</style>
		<script>%v</script>
  </head>
  <body>
		<section class='slide'>
			<div class='padding'>
	`, title, css, js)
	out += string(html)
	out += `</div>
		</section>
	</body>
</html>`
	return []byte(out)
}
