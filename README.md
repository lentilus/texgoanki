# TexGoAnki

> import flashcards written in latex to anki


# Outline
1. parse flashcards from source file (use texflash for now)
2. check if collection exists, if not create new
3. use anki-connect to search for flashcard with same id
4. If necessary compile using latexmk
5. finally add or update the note in anki

# tech stack
- anki connect to interface with the underlying anki database
- use go for request and file stuff
- for now use texflash to "parse" the flashcards
