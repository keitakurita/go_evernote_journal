package evernote_journal

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
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

	ioutil.WriteFile("template.xml", []byte(template), 0644)

	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	yesterdayNote, err := GetNoteFromNotebookByName(EvernoteAuthorToken, ns, notebook, yesterday.Format(DateFormat))

	if err != nil || yesterdayNote == "" {
		err = CreateNewNote(EvernoteAuthorToken, ns, notebook, today.Format(DateFormat), template)
		return err
	}

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

	err = CreateNewNote(EvernoteAuthorToken, ns, notebook, today.Format(DateFormat), contents)

	if err != nil {
		return err
	}

	fmt.Printf("Journal for %s successfully created\n", today.Format(DateFormat))

	return nil
}

// A class representing one section (ex. Reflections, Daily Goals, etc.) in the Journal
type Section struct {
	Title           string
	TitleDOMElement string
	Contents        string
}

// A function to parse the journal and convert it to a list of sections.
// Each section begins with a header represented by the delimiter as a regular expression.
// ex. <div># (.+?)</div>
// Where the substring encapsulated by parens represents the title.
// The section is followed by the contents, that span up to the beginning of the next section header.
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

// From a list of sections, finds and returns the section with the given title.
func findSectionByTitle(sections []Section, title string) (Section, error) {
	for _, section := range sections {
		if section.Title == title {
			return section, nil
		}
	}
	fmt.Printf("Warning: Section with title %s not found.\n", title)
	return Section{}, errors.New(fmt.Sprintf("Section with title %s not found.", title))
}

// Same as python method index of list
func indexInSlice(a string, list []string) int {
	for i, b := range list {
		if b == a {
			return i
		}
	}
	return -1
}

// Divides a notebook file into its header and contents, and returns each as a string
func divideHeaderAndBody(ennote string) (string, string) {
	re := regexp.MustCompile("<en-note.*?>(.*)</en-note>")
	indexes := re.FindAllStringIndex(ennote, -1)
	header := ennote[0:indexes[0][0]]
	body := re.FindStringSubmatch(ennote)[1]
	return header, body
}

// Takes a template journal, and populates its contents with the contents of another notebook when required.
// Replacement happens when the title of a section in the template journal is included in sectionsToReplace.
// Then, the contents of the section in the notebook with the title in sectionsToReplaceWith with the corresponding index is used to populate the contents.
func fillTemplateBody(templateBody string, noteBody string, sectionsToReplace []string, sectionsToReplaceWith []string, delimiter string) string {
	contents := bytes.Buffer{}

	templateSections := convertToSectionList(templateBody, delimiter)
	noteSections := convertToSectionList(noteBody, delimiter)

	// Filling the contetns
	for _, section := range templateSections {
		i := indexInSlice(section.Title, sectionsToReplace)
		if utf8.RuneCountInString(section.Title) > 0 {
			// replace the contents
			contents.WriteString(section.TitleDOMElement)
			if i > -1 {
				// find the section to replace to in yesteday's notes
				sec, err := findSectionByTitle(noteSections, sectionsToReplaceWith[i])
				if err == nil {
					contents.WriteString(sec.Contents)
				} else {
					contents.WriteString(section.Contents)
				}
			} else {
				contents.WriteString(section.Contents)
			}
		}
	}

	return contents.String()
}

func ConstructDailyJournalContents(template string, yesterdayNote string) (string, error) {
	contents := bytes.Buffer{}
	var err error

	header, templateBody := divideHeaderAndBody(template)
	_, noteBody := divideHeaderAndBody(yesterdayNote)

	// write the header
	contents.WriteString(header)
	contents.WriteString("<en-note>")

	sectionsToReplace := []string{"Daily Goal Checklist"}
	sectionsToReplaceWith := []string{"Goals For Tomorrow"}

	delimiter := "<div># (.+?)</div>"

	body := fillTemplateBody(templateBody, noteBody, sectionsToReplace, sectionsToReplaceWith, delimiter)
	// close the note
	contents.WriteString(body)
	contents.WriteString("</en-note>")

	return contents.String(), err
}

func ConstructWeeklyJournalContents(template string, weeklyTemplate string, yesterdayNote string) (string, error) {
	contents := bytes.Buffer{}
	var err error
	daily, err := ConstructDailyJournalContents(template, yesterdayNote)

	if err != nil {
		return template, err
	}

	header, dailyBody := divideHeaderAndBody(daily)
	_, weeklyBody := divideHeaderAndBody(weeklyTemplate)

	contents.WriteString(header)
	contents.WriteString("<en-note>")
	contents.WriteString(dailyBody)

	// Construct Weekly Review
	sectionsToReplace := []string{"Weekly Goal Checklist"}
	sectionsToReplaceWith := []string{"Weekly Goal Checklist"}
	delimiter := "<div># (.+?)</div>"

	body := fillTemplateBody(weeklyBody, dailyBody, sectionsToReplace, sectionsToReplaceWith, delimiter)

	contents.WriteString(body)

	contents.WriteString("</en-note>")
	return contents.String(), err
}

func ManageReflections(statContents string, filename string) {
	f, err := os.OpenFile(filename, os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dateString := time.Now().AddDate(0, 0, -1).Format("2016/01/02")

	// Log the data to a csv file
	internalSections := convertToSectionList(statContents, "<div>#([#$] .+?)</div>")
	for _, section := range internalSections {
		if strings.HasPrefix("$", section.Title) {
			fmt.Println(dateString + "," + section.Contents)
		}
	}
}
