package evernote_journal

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	// "os"
	"regexp"
	// "strings"
	"time"
	"unicode/utf8"
)

// The main function. Creates new journal entry.
// The format should correspond to how the following date would be formatted:
// Mon Jan 2 15:04:05 -0700 MST 2006
// ex. 01/02(Mon)
func CreateNewJournalEntry(EvernoteAuthorToken string, ClientKey string, ClientSecret string, NotebookName string, DateFormat string) error {
	ns, err := GetNotestoreFromCredentials(EvernoteAuthorToken, ClientKey, ClientSecret)
	if err != nil {
		return err
	}

	notebook, err := GetNotebookFromNotestoreByName(EvernoteAuthorToken, ns, NotebookName)
	if err != nil {
		return err
	}

	var contents string

	// check to see if today's journal is not already created
	todayNote, err := GetNoteFromNotebookByName(EvernoteAuthorToken, ns, notebook, time.Now().Format(DateFormat))
	if todayNote != "" {
		return errors.New("The journal for today already exists.")
	}

	// get the necessary notebooks
	template, err := GetNoteFromNotebookByName(EvernoteAuthorToken, ns, notebook, "Daily Template")

	if template == "" {
		return errors.New("Error: No template found\n")
	}

	/* TODO: erase following code */
	err = ioutil.WriteFile("template.xml", []byte(template), 0777)
	/* End of TODO */

	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	yesterdayNote, err := GetNoteFromNotebookByName(EvernoteAuthorToken, ns, notebook, yesterday.Format(DateFormat))

	/* TODO: erase following code */
	err = ioutil.WriteFile("yesterday.xml", []byte(yesterdayNote), 0777)
	/* End of TODO */

	if today.Weekday().String() == "Sunday" {
		weeklyTemplate, err := GetNoteFromNotebookByName(EvernoteAuthorToken, ns, notebook, "Weekly Template")
		if err != nil {
			return err
		}
		contents, err = ConstructWeeklyJournalContents(template, weeklyTemplate, yesterdayNote)
		if err != nil {
			return err
		}
	} else {
		contents, err = ConstructDailyJournalContents(template, yesterdayNote)
		if err != nil {
			return err
		}
	}

	fmt.Println(contents)

	// err = CreateNewNote(EvernoteAuthorToken, ns, notebook, today.Format(DateFormat), template)

	// if err != nil {
	// 	return err
	// }

	// fmt.Printf("Journal for %s successfully created\n", today.Format(DateFormat))

	return nil
}

type Section struct {
	Title           string
	TitleDOMElement string
	Contents        string
}

func convertToSectionList(text string, delimiter string) []Section {
	re := regexp.MustCompile(delimiter)
	indexes := re.FindAllStringIndex(text, -1)
	/*
		   1 1   2
		aaa123bbb4cccc
	*/

	results := make([]Section, len(indexes))

	for i, element := range indexes {
		title := re.FindStringSubmatch(text[element[0]:element[1]])[1]
		var contents string
		if i < len(indexes)-1 {
			contents = text[element[1]:indexes[i+1][0]]
		} else {
			contents = text[element[1]:len(text)]
		}
		results[i] = Section{title, text[element[0]:element[1]], contents}
	}
	return results
}

func findSectionByTitle(sections []Section, title string) Section {
	for _, section := range sections {
		if section.Title == title {
			return section
		}
	}
	fmt.Printf("No section with name %s found\n", title)
	return Section{}
}

func indexInSlice(a string, list []string) int {
	for i, b := range list {
		if b == a {
			return i
		}
	}
	return -1
}

func divideHeaderAndBody(ennote string) (string, string) {
	re := regexp.MustCompile("<en-note>(.*)</en-note>")
	indexes := re.FindAllStringIndex(ennote, -1)
	header := ennote[0:indexes[0][0]]
	body := re.FindStringSubmatch(ennote)[1]
	return header, body
}

func ConstructDailyJournalContents(template string, yesterdayNote string) (string, error) {
	contents := bytes.Buffer{}

	header, templateBody := divideHeaderAndBody(template)
	_, noteBody := divideHeaderAndBody(yesterdayNote)

	// write the header
	contents.WriteString(header)
	contents.WriteString("<en-note>")

	sectionsToReplace := []string{"Daily Goal Checklist"}
	sectionsToReplaceWith := []string{"Goals For Tomorrow"}

	delimiter := "<div># (.+?)</div>"

	templateSections := convertToSectionList(templateBody, delimiter)
	noteSections := convertToSectionList(noteBody, delimiter)

	for j, section := range templateSections {
		//fmt.Printf("Section %d\n__________________\n", j)
		i := indexInSlice(section.Title, sectionsToReplace)
		if utf8.RuneCountInString(section.Title) > 0 {
			//fmt.Printf("Title Element: \n%s\n\n", section.TitleDOMElement)
			contents.WriteString(section.TitleDOMElement)
			if i > -1 {
				// find the section to replace to in yesteday's notes
				sec := findSectionByTitle(noteSections, sectionsToReplaceWith[i])
				//fmt.Printf("Contents: %s\n\n", sec.TitleDOMElement)
				contents.WriteString(sec.Contents)
			} else {
				//fmt.Printf("Contents: %s\n\n", section.Contents)
				contents.WriteString(section.Contents)
			}
		}
	}

	// close the note
	contents.WriteString("</en-note>")

	return contents.String(), nil
}

func ConstructWeeklyJournalContents(template string, weeklyTemplate string, yesterdayNote string) (string, error) {
	contents := bytes.Buffer{}
	daily, err := ConstructDailyJournalContents(template, yesterdayNote)

	if err != nil {
		return template, err
	}

}
