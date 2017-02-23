package evernote_journal

import (
	// "bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// The main function. Creates new journal entry.
// The format should correspond to how the following date would be formatted:
// Mon Jan 2 15:04:05 -0700 MST 2006
// ex. 01/02(Mon)
func CreateNewJournalEntry(EvernoteAuthorToken string, ClientKey string, ClientSecret string, NotebookName string, DateFormat string) error {
	ns, err := GetNotestoreFromCredentials(EvernoteAuthorToken, ClientKey, ClientSecret)
	HandleError(err)

	notebook, err := GetNotebookFromNotestoreByName(EvernoteAuthorToken, ns, NotebookName)
	HandleError(err)

	var contents string

	// get the necessary notebooks
	template, err := GetNoteFromNotebookByName(EvernoteAuthorToken, ns, notebook, "Daily Template")

	/* TODO: erase following code */
	err = ioutil.WriteFile("template.xml", []byte(template), 0777)
	/* End of TODO */

	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	yesterdayNote, err := GetNoteFromNotebookByName(EvernoteAuthorToken, ns, notebook, yesterday.Format(DateFormat))

	/* TODO: erase following code */
	err = ioutil.WriteFile("yesterday.xml", []byte(yesterdayNote), 0777)
	/* End of TODO */

	if yesterday.Weekday().String() == "Sunday" {
		weeklyTemplate, err := GetNoteFromNotebookByName(EvernoteAuthorToken, ns, notebook, "Weekly Template")
		HandleError(err)
		contents = ConstructWeeklyJournalContents(template, weeklyTemplate, yesterdayNote)
	} else {
		contents = ConstructDailyJournalContents(template, yesterdayNote)
	}

	CreateNewNote(EvernoteAuthorToken, ns, notebook, today.Format(DateFormat), contents)

	return nil
}

func ConstructDailyJournalContents(template string, yesterdayNote string) string {
	//var contents bytes.Buffer
	root := strings.NewReader(template)
	doc, err := goquery.NewDocumentFromReader(root)
	HandleError(err)
	fmt.Println(doc)
	return template
}

func ConstructWeeklyJournalContents(template string, weeklyTemplate string, yesterdayNote string) string {
	return template
}

func HandleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal Error: %s\n", err.Error())
		os.Exit(1)
	}
}
