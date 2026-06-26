package tools

import (
    "context"
    "encoding/json"
    "os"
    "fmt"
    "strings"
)

// Note struct for storage
type note struct {
    Title   string `json:"title"`
    Content string `json:"content"`
}

const noteFile = "notes.json"

func loadNotes() ([]note, error) {
    data, readErr := os.ReadFile(noteFile)
    if readErr != nil {
        if os.IsNotExist(readErr) {
            return []note{}, nil
        }
        return nil, readErr
    }
    var notes []note
    if parseErr := json.Unmarshal(data, &notes); parseErr != nil {
        return nil, parseErr
    }
    return notes, nil
}

func saveNotes(notes []note) error {
    data, marshalErr := json.MarshalIndent(notes, "", "  ")
    if marshalErr != nil {
        return marshalErr
    }
    return os.WriteFile(noteFile, data, 0644)
}

// Handlers

func HandleListNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    notes, loadErr := loadNotes()
    if loadErr != nil {
        return err(loadErr.Error())
}

    titles := make([]string, 0, len(notes))
    for _, n := range notes {
        titles = append(titles, n.Title)

    jsonBytes, marshalErr := json.Marshal(titles)
    if marshalErr != nil {
        return err(marshalErr.Error())
}

    return ok(string(jsonBytes))
}

}

func HandleGetNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    title, _ :=getString(args, "title")
    if title == "" {
        return err("title is required")
}

    notes, loadErr := loadNotes()
    if loadErr != nil {
        return err(loadErr.Error())
}

    for _, n := range notes {
        if strings.EqualFold(n.Title, title) {
            noteJSON, marshalErr := json.Marshal(n)
            if marshalErr != nil {
                return err(marshalErr.Error())
}

            return ok(string(noteJSON))

    }
    return err(fmt.Sprintf("note with title '%s' not found", title))
}

}

func HandleCreateNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    title, _ :=getString(args, "title")
    content, _ :=getString(args, "content")
    if title == "" || content == "" {
        return err("title and content are required")
}

    notes, loadErr := loadNotes()
    if loadErr != nil {
        return err(loadErr.Error())
}

    // Check for duplicate title
    for _, n := range notes {
        if strings.EqualFold(n.Title, title) {
            return err(fmt.Sprintf("note with title '%s' already exists", title))

    }
    notes = append(notes, note{Title: title, Content: content})
    if saveErr := saveNotes(notes); saveErr != nil {
        return err(saveErr.Error())
}

    return ok(fmt.Sprintf("note '%s' created", title))
}

}

func HandleUpdateNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    title, _ :=getString(args, "title")
    content, _ :=getString(args, "content")
    if title == "" || content == "" {
        return err("title and content are required")
}

    notes, loadErr := loadNotes()
    if loadErr != nil {
        return err(loadErr.Error())
}

    found := false
    for i, n := range notes {
        if strings.EqualFold(n.Title, title) {
            notes[i].Content = content
            found = true
            break
        }
    }
    if !found {
        return err(fmt.Sprintf("note with title '%s' not found", title))
}

    if saveErr := saveNotes(notes); saveErr != nil {
        return err(saveErr.Error())
}

    return ok(fmt.Sprintf("note '%s' updated", title))
}

func HandleDeleteNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    title, _ :=getString(args, "title")
    if title == "" {
        return err("title is required")
}

    notes, loadErr := loadNotes()
    if loadErr != nil {
        return err(loadErr.Error())
}

    index := -1
    for i, n := range notes {
        if strings.EqualFold(n.Title, title) {
            index = i
            break
        }
    }
    if index == -1 {
        return err(fmt.Sprintf("note with title '%s' not found", title))
}

    notes = append(notes[:index], notes[index+1:]...)
    if saveErr := saveNotes(notes); saveErr != nil {
        return err(saveErr.Error())
}

    return ok(fmt.Sprintf("note '%s' deleted", title))
}