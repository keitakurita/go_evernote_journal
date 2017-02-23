package evernote_journal

import (
	"bytes"
	"fmt"
	"github.com/dreampuf/evernote-sdk-golang/client"
	"github.com/dreampuf/evernote-sdk-golang/notestore"
	"github.com/dreampuf/evernote-sdk-golang/types"
	"os"
)

// Get the notestore for the user based on the developer token(EvernoteAuthorToken), evernote api key, evernote api secret
// All 3 must be first acquired from evernote to be able to use.
func GetNotestoreFromCredentials(EvernoteAuthorToken string, ClientKey string, ClientSecret string) (*notestore.NoteStoreClient, error) {
	c := client.NewClient(ClientKey, ClientSecret, client.SANDBOX)
	us, err := c.GetUserStore()
	if err != nil {
		return nil, err
	}
	url, err := us.GetNoteStoreUrl(EvernoteAuthorToken)
	if err != nil {
		return nil, err
	}

	NoteStore, err := c.GetNoteStoreWithURL(url)
	if err != nil {
		return nil, err
	}
	return NoteStore, nil
}

// Gets the id of a notebook based on the name from the notestore
// Note that this function also apparently requires the developer token
func GetNotebookFromNotestoreByName(EvernoteAuthorToken string, NoteStore *notestore.NoteStoreClient, name string) (*types.GUID, error) {
	notebooks, err := NoteStore.ListNotebooks(EvernoteAuthorToken)
	if err != nil {
		return nil, err
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

func GetNoteFromNotebookByName(EvernoteAuthorToken string, NoteStore *notestore.NoteStoreClient, notebook *types.GUID, NoteName string) (string, error) {
	// construct query to get the template note
	ascending := false
	var buffer bytes.Buffer
	buffer.WriteString("intitle:")
	// the name for the template is, by default, Template. This might be subject to change or customization in the future.
	buffer.WriteString(NoteName)
	query := buffer.String()
	filter := &notestore.NoteFilter{Words: &query, NotebookGuid: notebook, Ascending: &ascending}

	// search the notestore
	var resultSpec notestore.NotesMetadataResultSpec
	includeUpdateSequenceNum := true
	resultSpec.IncludeUpdateSequenceNum = &includeUpdateSequenceNum
	metadatas, err := NoteStore.FindNotesMetadata(EvernoteAuthorToken, filter, 0, 1, &resultSpec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	// the result should be just the single notebook
	if len(metadatas.GetNotes()) == 0 {
		fmt.Fprintf(os.Stderr, "Fatal error: No %s found in notebook.\n", NoteName)
		os.Exit(1)
	}
	metanote := metadatas.GetNotes()[0]
	template, err := NoteStore.GetNote(EvernoteAuthorToken, metanote.GUID, true, false, false, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	return *template.Content, nil
}

// Create a new entry
func CreateNewNote(EvernoteAuthorToken string, NoteStore *notestore.NoteStoreClient, notebook *types.GUID, title string, contents string) {
	note := types.Note{}
	notebookGuid := string(*notebook) // for some reason, note.NotebookGuid only accepts *string. *types.GUID is effectively *string, except with another name.

	note.Content = &contents
	note.Title = &title
	note.NotebookGuid = &notebookGuid

	// Create the notebook
	NoteStore.CreateNote(EvernoteAuthorToken, &note)
}
