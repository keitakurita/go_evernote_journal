package evernote_journal

import (
	"bytes"
	"fmt"
	"github.com/dreampuf/evernote-sdk-golang/client"
	"github.com/dreampuf/evernote-sdk-golang/notestore"
	"github.com/dreampuf/evernote-sdk-golang/types"
	"os"
)

func GetTemplate(name string, EvernoteAuthorToken string, ClientKey string, ClientSecret string) string {

	ns, err := GetNotestoreFromCredentials(EvernoteAuthorToken, ClientKey, ClientSecret)

	notebookGUID, err := GetNotebookFromNotestoreByName(ns, name, EvernoteAuthorToken)
	if notebookGUID == nil {
		fmt.Fprintf(os.Stderr, "Notebook with name: %s not found.", name)
	}

	// construct query to get the template note
	ascending := false
	var buffer bytes.Buffer
	buffer.WriteString("intitle:")
	// the name for the template is, by default Template. This might be subject to change or customization in the future.
	buffer.WriteString("Template")
	query := buffer.String()
	filter := &notestore.NoteFilter{Words: &query, NotebookGuid: notebookGUID, Ascending: &ascending}
	fmt.Println(filter)

	// search the notestore
	var resultSpec notestore.NotesMetadataResultSpec
	includeUpdateSequenceNum := true
	resultSpec.IncludeUpdateSequenceNum = &includeUpdateSequenceNum
	metadatas, err := ns.FindNotesMetadata(EvernoteAuthorToken, filter, 0, 1, &resultSpec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	// the result should be just the single notebook
	if len(metadatas.GetNotes()) == 0 {
		fmt.Fprintf(os.Stderr, "Fatal error: No template found in notestore.\n")
		os.Exit(1)
	}
	metanote := metadatas.GetNotes()[0]
	template, err := ns.GetNote(EvernoteAuthorToken, metanote.GUID, true, false, false, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	return *template.Content
}

// Get the notestore for the user based on the developer token(EvernoteAuthorToken), evernote api key, evernote api secret
// All 3 must be first acquired from evernote to be able to use.
func GetNotestoreFromCredentials(EvernoteAuthorToken string, ClientKey string, ClientSecret string) (*notestore.NoteStoreClient, error) {
	c := client.NewClient(ClientKey, ClientSecret, client.SANDBOX)
	us, err := c.GetUserStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
	url, err := us.GetNoteStoreUrl(EvernoteAuthorToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
	fmt.Println(url)
	ns, err := c.GetNoteStoreWithURL(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
	return ns, nil
}

// Gets the id of a notebook based on the name from the notestore
// Note that this function also apparently requires the developer token
func GetNotebookFromNotestoreByName(ns *notestore.NoteStoreClient, name string, EvernoteAuthorToken string) (*types.GUID, error) {
	notebooks, err := ns.ListNotebooks(EvernoteAuthorToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	// search for the notebook
	var notebookGUID *types.GUID
	for _, book := range notebooks {
		if *book.Name == name {
			notebookGUID = book.GUID
			break
		}
	}
	return notebookGUID, nil
}
